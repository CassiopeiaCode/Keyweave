# 16 — CLIProxyAPI 非插件核心兼容目标

> 本文依据截至 2026-10-01 可见的 CLIProxyAPI README/config 与 CLIProxyAPIDocs 管理 API 文档制定。目标是“行为可表达”，不是内部结构逐字段复制。

## 1. 兼容分类

### A — Core/Library 必须可表达

- 多 API Account
- 多 Credential 调度
- round-robin
- weighted-round-robin
- fill-first
- universal session affinity + failover
- per-credential priority/weight
- model alias
- model exclusions
- client-visible alias -> upstream model
- response model force mapping
- per-credential proxy
- HTTP/HTTPS/SOCKS5（SOCKS5 可 Phase 2）
- custom headers
- dynamic downstream-header copy
- retries
- cooldown
- streaming
- non-streaming
- supported WebSocket paths
- OpenAI-compatible custom provider
- OAuth credential refresh
- auth credential enable/disable 等价能力
- credential metadata fields
- model listing per credential/group

### B — 通过 Template 实现，不进入 Core Provider 特例

- Gemini key auth
- Claude key auth
- Codex-specific auth
- xAI request special behavior
- provider OAuth refresh
- provider-specific quota behavior
- payload rules
- request-scoped error classification
- provider model discovery

### C — 不属于初始兼容目标

- CLIProxyAPI 自身插件生态 ABI
- 其 UI/管理端 endpoint 路径完全相同
- 其内部 auth file 格式一比一兼容
- 其特定商业/Home 服务逻辑

## 2. 能力映射

| CLIProxyAPI 概念 | 本系统 |
|---|---|
| Provider/API-key group | Credential Template |
| Individual key/auth file | Credential Instance |
| API client key | Access Key |
| Routing scope | Credential Group |
| routing.strategy | Scheduler Library |
| weight | custom field -> weighted scheduler attribute |
| priority | availability 或 scheduler attribute |
| proxy-url | instance.proxy |
| models | ModelDefinition[] |
| alias | ModelDefinition.id/upstream |
| excluded-models | Template model filtering |
| OAuth auth file | Credential Instance + secrets/custom fields |
| auth auto refresh | Template startup/background helper |
| cooldown | custom state + availability/supports |
| request-retry | Retry Engine + Template policy |
| session-affinity | Session Affinity Scheduler |
| custom provider headers | Template beforeRequest/execute |
| payload rules | Template request rewrite |
| WebSocket | Host transport |
| auth-file metadata PATCH | Instance custom fields patch |

## 3. Round Robin

CLIProxyAPI 当前公开配置支持 `round-robin`，并支持 `weighted-round-robin`、`fill-first`。本项目分别作为 Scheduler Library 实现。

## 4. Session Affinity

公开配置描述 universal session-sticky routing，优先显式 session metadata，并在绑定 credential 不可用时 failover。

本项目使用：

```text
SessionIdentityExtractor
+ SessionAffinityScheduler
```

表达。

## 5. Retry / Cooldown

CLIProxyAPI 当前公开配置提供 request retry rounds、max retry credentials、cooldown 等。

本项目使用：

```text
Retry Engine
+ Template ErrorAction
+ custom cooldown fields
```

表达。

## 6. Proxy

CLIProxyAPI 当前公开配置支持 HTTP/HTTPS/SOCKS5、per-entry proxy override 与 direct bypass。

本项目：

- MVP：HTTP + Basic auth + HTTPS CONNECT + direct；
- Phase 2：SOCKS5 / HTTPS proxy；
- Instance proxy 可被 Template `getProxy` 覆盖。

## 7. Model Alias / Exclusion

本系统 ModelDefinition：

```json
{
  "id": "gpt-fast",
  "upstream": "provider/gpt-actual",
  "forceResponseMapping": true
}
```

Wildcard exclusion 由 Template Library 提供。

## 8. Header Copy

Template 可将 downstream header 动态复制：

```js
const v = ctx.request.headers.get("ABC");
if (v) headers["X-Upstream-Session"] = v;
```

## 9. OpenAI-compatible Provider

Generic OpenAI Template 直接覆盖：

```text
baseUrl
key
models
headers
proxy
```

用户通常不需要自己写 JS。

## 10. 管理 API 等价能力

CLIProxyAPIDocs 当前公开管理能力包括：

- credential/auth-file models
- credential enable/disable
- metadata/headers update
- OAuth model aliases
- OpenAI-compatible provider entries

本项目不要求路径兼容，但 Admin API 应有等价表达能力。

## 11. 当前公开资料

- https://github.com/router-for-me/CLIProxyAPI
- https://github.com/router-for-me/CLIProxyAPI/blob/main/config.example.yaml
- https://github.com/router-for-me/CLIProxyAPIDocs/blob/main/docs/en/management/api.md
- https://github.com/router-for-me/CLIProxyAPIDocs/blob/main/docs/en/configuration/basic.md

## 12. 验收原则

如果 CLIProxyAPI 的某个**非插件 Provider/credential routing** 行为无法通过：

```text
Template code
Template fields
Group scheduler
Host HTTP/stream/ws
```

表达，优先检查 Host Runtime 是否缺一个通用 primitive。

只有确实是通用 primitive 缺失时才扩 Core；不能直接为 Provider 加特例。
