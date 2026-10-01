# 04 — 认证与路由

## 1. Authentication = Routing Entry

用户的 Client API Key 不仅用于“是否允许请求”，还直接确定请求进入哪个 Credential Group。

```text
Authorization: Bearer tp_live_xxx
                   |
                   v
             access_keys
                   |
                 group_id
                   |
                   v
          credential_groups
```

不存在默认的“认证成功后再从所有 Provider 中任选”的第二阶段。

## 2. Key 解析

入站适配器允许不同协议采用不同标准 Header：

- OpenAI：`Authorization: Bearer <key>`
- Anthropic：可接受 `x-api-key: <key>`；也可兼容 bearer
- 管理 API：必须使用独立 management auth，不能复用 inference client key

归一化后：

```ts
interface ClientAuth {
  rawPresentedKey: SecretString;
  keySource: "authorization" | "x-api-key";
}
```

原始值不可进入普通日志。

## 3. Hash 策略

Client Key 是高熵随机值，推荐：

```text
tp_live_<random 32 bytes base64url>
```

DB 保存：

```text
secret_hash = HMAC-SHA-256(serverLookupKey, presentedKey)
secret_prefix
```

这样可做 O(1) 查询且数据库泄露时不能直接使用 key。

Management key 可使用独立方案；不要把 management credential 和 inference credential 混用。

## 4. Candidate Build

```text
group.templateIds
  -> instance rows WHERE template_id IN (...)
  -> static protocol filter
  -> static model filter
  -> dynamic template.supports()
  -> candidates
```

## 5. Protocol Filter

Template 有声明能力：

```json
["openai.chat", "openai.responses", "anthropic.messages"]
```

如果请求协议不在列表，静态剔除。

动态 Template 可以进一步 `supports=false`。

## 6. Model Filter

每个实例的 effective model definitions 来自：

```text
template.models defaults
+ instance.models override
+ optional template getModels()
```

请求 `model` 必须匹配一个 client-visible `ModelDefinition.id`。

如果 Template 声明 dynamic model passthrough，可以允许 wildcard；必须显式声明，不能默认任何模型都匹配。

## 7. 模型别名

请求：

```json
{"model":"fast"}
```

某候选：

```json
{"id":"fast","upstream":"provider/model-x"}
```

Runtime 传给 Template：

```text
requestedModel = fast
resolvedUpstreamModel = provider/model-x
```

是否重写 response 中 model 字段，由 `forceResponseMapping` 或 Template 决定。

## 8. `/v1/models`

必须基于当前 Client Key 的 Group 生成：

```text
expand group
-> filter protocol/model-enabled instances
-> union client-visible models
-> dedupe by id
-> return catalog
```

默认不暴露用户无权路由到的模板模型。

## 9. Ambiguous Alias

同一个 group 中允许：

```text
Template A -> model "fast"
Template B -> model "fast"
```

这是合法情况，表示同一 client-visible model 可以由多个 upstream 实现。Scheduler 在候选 credentials 之间选择。

## 10. 路由错误分类

推荐统一内部错误，再由 Protocol Adapter 映射：

- `AuthMissing`
- `AuthInvalid`
- `GroupNotFound`
- `ModelMissing`
- `ModelNotRoutable`
- `NoCredentialAvailable`
- `SchedulerRejected`
- `TemplateUnavailable`

避免上层根据字符串判断。

## 11. 不做 Provider Pinning 特例

如果需要严格指定某个后端，使用：

- 独立 model alias；
- prefix convention；
- Group scope；
- custom scheduler；
- model metadata；

不要加 `provider=` 核心字段。
