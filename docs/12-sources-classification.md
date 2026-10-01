# 12 — Credential Source 与自动分类

## 1. Source 的角色

Source 负责从外部渠道获得 raw credential。

它不是 inference runtime：

```text
Source
 -> RawCredential
 -> Classifier
 -> Template
 -> Credential Instance
```

## 2. Source 类型

官方先支持：

- manual
- JSON file import
- directory scan
- HTTP pull
- webhook push
- script
- OAuth login flow

## 3. RawCredential

```ts
interface RawCredential {
  sourceId: string;
  externalId?: string;
  observedAt: string;
  payload: unknown;
}
```

## 4. Classifier

Template Library 可以暴露 classifier：

```js
export function classify(raw) {
  return {
    confidence: 0.95,
    fields: {...},
    secrets: {...}
  }
}
```

多个 template 命中：

- 高 confidence 胜出；
- 差距过小进入人工确认；
- classifier 默认不能执行任意网络请求。

## 5. 去重

推荐优先：

```text
sourceId + externalId
```

否则使用安全 fingerprint，不能保存 secret 明文。

导入结果：

```text
new
update_existing
duplicate_candidate
conflict
ignored
```

## 6. 重分组 vs 重分类

### 重分组

不改变 Template，只改变哪些 Client Key/Group 能调度到对应 Template 范围。

### 重分类

Instance 迁移到另一个 Template。

两者在 UI/API 必须分开。

## 7. OAuth Source 更新

- refresh token 变化 -> update secret；
- metadata 变化 -> patch custom fields；
- credential removed -> 默认归档/降低可调度性，不直接物理删除；
- external identity 稳定时应更新原实例，不创建无限重复凭据。

## 8. Source 定时拉取

Source scheduler 与 inference scheduler 完全独立。

建议：

```text
source interval
jitter
timeout
lastSuccess
lastError
```

Source 失败不能让 Gateway inference 主路径阻塞。
