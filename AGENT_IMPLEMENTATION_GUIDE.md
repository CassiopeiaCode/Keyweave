# Coding Agent 实现指南

## 1. Agent 的任务

根据本包中的规范实现一个单机优先、可长期演进的 Programmable Token Pool / Credential Pool Gateway。

除非规格明确允许，否则不要“为了快速跑起来”改变核心对象关系。

## 2. 设计优先级

按以下顺序处理冲突：

1. `docs/01-terminology-and-invariants.md` 中的 Core Invariants
2. `contracts/` 中的接口契约
3. `docs/` 中具体模块设计
4. `MASTER_SPEC.md`
5. `examples/` 与 `reference/`（仅为参考，不高于契约）

## 3. 强制实现原则

### 3.1 不要把 Provider 写死在 Core

禁止：

```go
if provider == "openrouter" { ... }
if provider == "codex" { ... }
```

允许：

```go
template.Execute(ctx)
scheduler.Select(ctx, candidates)
```

Provider-specific 行为必须尽可能进入 Credential Template Library。

### 3.2 Core 不理解 Quota 语义

Core 可以观察：

- availability
- decrease
- activeRequests
- template output
- retry action

Core 不应出现：

```go
if quotaRemaining <= 0 { ... }
```

Quota 只是 custom field，由 Template 转换成 availability/canSchedule/retry 行为。

### 3.3 Lease 是并发正确性的核心

选择实例和建立 Lease 必须属于同一个原子调度临界区：

```text
candidate snapshot
-> score
-> select
-> lease++
```

不能先 select，之后再异步 lease。

### 3.4 Streaming 生命周期

Lease 释放点必须是：

- 非流：最终 response body 已处理完成；
- SSE：下游 stream EOF / cancel / error；
- WebSocket：connection closed；
- 上游连接建立失败：失败清理后立即释放。

### 3.5 Template 网络隔离

Template JS：

- 不得拥有原生 `fetch`
- 不得拥有 filesystem
- 不得拥有 process/env
- 不得拥有 raw socket
- 不得拥有 DB handle

网络只能使用 `ctx.http`。

## 4. 推荐工程结构

```text
cmd/gateway/
internal/
  domain/
  store/
  auth/
  protocol/
    openai/
    anthropic/
  routing/
  scheduler/
  lease/
  runtime/
  template/
  transport/
  proxy/
  retry/
  secrets/
  admin/
  observability/
pkg/
  templateapi/
templates/
schedulers/
migrations/
web/
```

## 5. 实现顺序

### Milestone 1 — Domain + Store

实现：

- migrations
- CRUD repositories
- encrypted secret storage
- optimistic version
- Access Key hash lookup

验收：

- migration 可重复检查；
- CRUD round-trip；
- Secret 不明文入库；
- Client API Key 不明文入库。

### Milestone 2 — Candidate Resolution

实现：

```text
API Key -> Group -> templateIds -> instances
```

以及：

- protocol filter
- model filter
- template metadata filter

暂时可以使用静态 template capabilities，不需要 JS。

### Milestone 3 — Scheduler + Lease

实现：

- `effectiveAvailability`
- atomic select + lease
- RR tie-break
- active request accounting
- cancellation-safe release

必须先通过并发 test vectors，再做网络代理。

### Milestone 4 — Transport

实现：

- direct HTTP
- HTTP proxy
- proxy basic auth
- HTTPS CONNECT
- response body streaming
- request cancellation

### Milestone 5 — JS Runtime

实现：

- Template compile/cache
- ABI
- timeout
- restricted host API
- `ctx.http`
- field merge
- secrets accessor
- instance patch API
- hook lifecycle

### Milestone 6 — Protocol Gateway

依次：

1. OpenAI Chat Completions
2. OpenAI Responses
3. Anthropic Messages
4. `/v1/models`

优先 Raw Passthrough，协议 Adapter 只负责提取必要 metadata。

### Milestone 7 — Retry / Cooldown / Session

实现为 library，不要把 Provider 行为写死。

### Milestone 8 — Admin + Debugger

最后实现管理 UI/API 和 Template Debugger。

## 6. 每次提交必须满足

- `go test -race ./...`
- migration test
- no plaintext secret snapshot
- no goroutine leak on cancelled stream
- lease count returns to zero
- invalid custom scheduler cannot escape candidate scope

## 7. Definition of Done

系统至少可完成以下真实调用：

```text
OpenAI client
  -> local gateway
  -> client key
  -> group(openrouter template)
  -> instance-2
  -> authenticated HTTP proxy
  -> OpenRouter
  -> SSE response
```

同时另一个并发请求优先调度到其他实例（取决于 effective availability）。

并且：

- 429 可触发 Template retry action；
- retry 可以换 credential；
- Session Affinity scheduler 可以稳定绑定；
- `/v1/models` 只显示当前 Client Key 的可用模型；
- 日志不可出现 upstream key。

## 8. 架构变更要求

以下修改必须新增 ADR，不允许作为普通重构直接做：

- 增加/修改官方字段；
- 允许一个 Access Key 对多个 Group；
- 允许 custom scheduler 选择 scope 外凭据；
- 允许 template 使用 raw network；
- 把 quota 变成 core 概念。
