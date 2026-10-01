# 28 — 推荐代码包结构与依赖方向

## 1. Go 包结构

```text
cmd/gateway/main.go
cmd/admin/main.go                # 可与 gateway 合并

internal/domain/                 # 纯结构和值对象
internal/store/                  # SQLite repository
internal/secrets/                # encryption + redaction
internal/auth/                   # Client API Key authenticate
internal/protocol/               # inbound adapters
  openai/
  anthropic/
internal/routing/                # candidate resolve
internal/scheduler/              # builtin scheduler + ABI
internal/lease/                  # in-memory concurrency reservation
internal/template/               # compile/cache/lifecycle
internal/runtime/                # sandbox host ABI
internal/transport/              # HTTP/SSE/WS + proxy
internal/retry/                  # request attempt controller
internal/source/                 # credential source/classifier
internal/admin/                  # CRUD/debugger APIs
internal/observability/          # logs/metrics/tracing

pkg/templateapi/                 # stable public ABI structs if needed
pkg/schedulerapi/

templates/builtin/
schedulers/builtin/
migrations/
web/
```

## 2. 依赖方向

必须保持单向：

```text
protocol/auth
     |
     v
routing -> scheduler -> lease
     |          |
     v          v
 template/runtime -> transport
     |
     v
 store/secrets
```

`domain` 不依赖其他内部包。

Transport 不依赖具体 Provider Template。

Scheduler 不依赖 HTTP Transport。

## 3. Repository 接口

建议最小接口：

```go
type TemplateRepository interface {
    Get(ctx context.Context, id string) (Template, error)
    List(ctx context.Context) ([]Template, error)
    Create(ctx context.Context, in CreateTemplate) (Template, error)
    UpdateCAS(ctx context.Context, id string, version int64, patch TemplatePatch) (Template, error)
}

type InstanceRepository interface {
    Get(ctx context.Context, id string) (Instance, error)
    ListByTemplateIDs(ctx context.Context, ids []string) ([]Instance, error)
    UpdateCAS(ctx context.Context, id string, version int64, patch InstancePatch) (Instance, error)
}
```

管理面所有写操作推荐使用 optimistic concurrency：`If-Match` 或请求字段 `version`。

## 4. Routing Service

```go
type Resolver interface {
    BuildCandidates(ctx context.Context, req RouteRequest, group Group) ([]CandidateSnapshot, error)
}
```

Resolver 不负责 select，也不 acquire lease。

## 5. Scheduling Service

```go
type SchedulingService interface {
    SelectAndAcquire(ctx context.Context, in SelectInput) (Selection, error)
    Release(leaseID string)
}
```

这样可以确保调用方无法忘记原子 acquire。

## 6. Runtime Service

```go
type TemplateExecutor interface {
    Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResponse, error)
}
```

ExecutionRequest 必须包含已固定：

- template ID/version；
- instance ID/version snapshot；
- model mapping；
- lease/attempt metadata。

## 7. 错误依赖

底层返回 typed error；只有协议边界负责映射为 HTTP/Provider-compatible error body。
