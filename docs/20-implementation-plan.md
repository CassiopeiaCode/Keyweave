# 20 — 详细实现计划

## Phase 0 — Contract Freeze

产物：

- domain types；
- Template ABI；
- Scheduler ABI；
- OpenAPI；
- DB migration；
- test vectors。

任何业务实现之前先保证这些契约一致。

## Phase 1 — Storage / Secrets

实现 repositories：

```go
TemplateRepository
CredentialRepository
GroupRepository
AccessKeyRepository
SourceRepository
```

Secrets：

```go
Seal(plaintext) -> ciphertext
Open(ciphertext) -> Secret
```

Acceptance：

- DB 中无 upstream key 明文；
- version conflict 有测试；
- audit diff 不含 secret。

## Phase 2 — Static Router

Template 暂不运行 JS，使用 metadata：

```text
Access Key -> Group -> Instances -> Model/Protocol -> first
```

先贯通协议到 mock upstream。

## Phase 3 — Lease Scheduler

加入：

- active count；
- effective availability；
- RR cursor；
- atomic lease；
- retry exclusion set。

通过 race tests 后才能进入下一阶段。

## Phase 4 — Transport

加入：

- direct；
- HTTP proxy auth；
- CONNECT；
- streams；
- cancellation。

## Phase 5 — Template Runtime

先支持：

- manifest；
- getAvailability；
- getDecrease；
- getScheduleHints；
- beforeRequest；
- execute via host.http；
- afterResponse/onError；
- patch state。

之后再加：

- startup hooks；
- background refresh；
- richer capabilities。

## Phase 6 — Protocols

依次：

1. chat completions；
2. responses；
3. anthropic messages；
4. model listing。

## Phase 7 — Scheduler Library

- availability RR；
- weighted RR；
- fill first；
- least active；
- session affinity。

## Phase 8 — Library Templates

- generic bearer；
- generic header auth；
- generic OpenAI；
- OpenRouter；
- Anthropic API key；
- Gemini API key。

## Phase 9 — Retry/Cooldown/OAuth

OAuth 必须仍然 Template-driven。

## Phase 10 — Admin/Debugger

优先 API，最后 UI。

## Phase 11 — CLIProxyAPI Compatibility Tests

逐项从 compatibility matrix 建 fixture。

## 实现切分建议

每个 phase 都应该可运行，不做“大爆炸式”一次性实现。

尤其禁止同时第一次引入：

```text
JS runtime + proxy + streaming + retry
```

否则很难定位 lease 泄漏和网络问题。
