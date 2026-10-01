# 25 — 端到端时序与热路径

本章定义请求从进入 Gateway 到响应结束的**规范时序**。实现可以内部重构，但外部可观察语义不得改变。

## 1. 普通非流请求

```text
Client
  | POST /v1/chat/completions + Client Key
  v
Protocol Adapter
  | authenticate + parse protocol/model/session metadata
  v
AccessKey Resolver
  | key -> exactly one CredentialGroup
  v
Candidate Resolver
  | expand group.templateIds
  | load instances
  | merge fields
  | resolve model mapping
  | run supports/getAvailability/getDecrease/getScheduleHints
  v
Candidate Snapshot[]
  |
  | atomically: scheduler select + lease acquire
  v
Credential Lease
  |
  v
Template Runtime
  | beforeRequest
  | execute/default passthrough
  v
Bound Host HTTP Client
  | current credential proxy policy
  v
Upstream
  | response
  v
Template Runtime
  | afterResponse
  v
Gateway
  | release lease in finally
  v
Client
```

### 1.1 关键顺序

以下顺序是 MUST：

1. 完成客户端认证后才能读取 Group。
2. Candidate scope 必须先由 Group 的 `templateIds[]` 限定。
3. Scheduler 只能看到 scope 内的候选。
4. Select 与 Lease acquire 必须原子化。
5. Template 执行前 Lease 必须已经存在。
6. `onFinish` 无论成功、错误、取消都执行一次。
7. Lease release 必须位于最外层 `finally/defer`。

## 2. SSE 流式请求

SSE 不允许把 response body 全量读入内存。

```text
select + acquire lease
      |
      v
upstream headers received
      |
      v
send downstream headers
      |
      +---- downstreamStarted = true
      |
      v
copy/transform stream chunks
      |
      +-- EOF ---------> finish
      +-- client cancel -> cancel upstream -> finish
      +-- read error ---> onError(stream_error) -> finish
      +-- write error --> cancel upstream -> finish
                           |
                           v
                      release lease
```

一旦第一个响应字节已经写给客户端，默认禁止切换 Credential 后重试，因为无法安全重放一个已经部分发送的 HTTP 响应。

## 3. Upstream 建连失败

如果还没有向下游发送任何字节：

```text
Attempt 1 / Credential A
  -> connection failure
  -> template.onError
  -> release A lease
  -> retry action = another
  -> mark A in attemptedSet
  -> reschedule
Attempt 2 / Credential B
```

每个 Attempt 都拥有独立 Lease。

## 4. 429 场景

Core 不把 429 等同于“额度耗尽”。

规范路径：

```text
upstream status=429
 -> Template afterResponse/onError
 -> Template optionally persists cooldown/quota custom state
 -> Template returns retry advice
 -> Core releases lease
 -> retry controller applies retry budget
 -> scheduler selects candidate not in attemptedSet
```

Template 可以把当前实例 availability 调低，也可以不修改；这属于模板策略。

## 5. Session Affinity

```text
request sessionKey=S
 -> group scheduler
 -> binding S -> instance B exists?
       yes -> B still candidate and schedulable?
                yes -> B
                no  -> fallback scheduler -> C -> rebind S=C
       no  -> fallback scheduler -> X -> bind S=X
```

绑定永远不能突破 Group scope、protocol/model filter 或硬并发限制。

## 6. `/v1/models`

模型列表必须与 Client Key 路由范围一致：

```text
Client Key
 -> Group
 -> Templates
 -> Instances
 -> effective model definitions
 -> protocol visibility/filtering
 -> dedupe by client-facing model id
 -> response
```

不得列出当前 Key 永远无法路由到的模型。

## 7. 管理面更新与请求面隔离

管理 API 更新 Template 时：

1. 验证 source/manifest；
2. 编译新版本；
3. DB transaction 写入 `version + 1`；
4. 原子切换 active compiled artifact；
5. 新请求使用新版本；
6. 已开始请求保持 pinned old version，直至完成。

## 8. 凭据导入

```text
Source Poll/Push
 -> Raw Credential
 -> optional classifier
 -> candidate Template
 -> validate/normalize
 -> dry-run preview
 -> create/update Credential Instance
 -> secret extraction/encryption
 -> optional startup/instance initialization
```

导入失败不得创建半初始化实例。

## 9. 取消传播

客户端取消必须沿调用链传播：

```text
client context cancelled
 -> template context cancelled
 -> ctx.http request cancelled
 -> upstream socket/stream closed
 -> onFinish
 -> lease release
```

Template 不能屏蔽 Host 的 cancellation signal。
