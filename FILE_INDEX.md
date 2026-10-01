# File Index

本文件说明实现包中每类文件的用途。`MASTER_SPEC.md` 是单文件阅读入口，但真正实现时应以 `contracts/`、invariants 与模块文档为准。

## Top-level entry files

- `README.md`
- `MASTER_SPEC.md`
- `AGENT_PROMPT.md`
- `AGENT_IMPLEMENTATION_GUIDE.md`
- `SOURCE_REFERENCES.md`
- `VERSION`

## Contracts

- `contracts/config.schema.json`
- `contracts/domain.ts`
- `contracts/errors.schema.json`
- `contracts/openapi.yaml`
- `contracts/scheduler-runtime.d.ts`
- `contracts/template-manifest.schema.json`
- `contracts/template-runtime.d.ts`

## Module docs

- `docs/00-overview.md`
- `docs/01-terminology-and-invariants.md`
- `docs/02-architecture.md`
- `docs/03-domain-model.md`
- `docs/04-auth-routing.md`
- `docs/05-scheduler-concurrency.md`
- `docs/06-template-runtime.md`
- `docs/07-protocol-gateway.md`
- `docs/08-transport-proxy.md`
- `docs/09-streaming-websocket.md`
- `docs/10-persistence.md`
- `docs/11-admin-api.md`
- `docs/12-sources-classification.md`
- `docs/13-retry-cooldown.md`
- `docs/14-security-sandbox.md`
- `docs/15-observability.md`
- `docs/16-cliproxyapi-compatibility.md`
- `docs/17-testing.md`
- `docs/18-versioning-migrations.md`
- `docs/19-deployment.md`
- `docs/20-implementation-plan.md`
- `docs/21-decisions-nongoals.md`
- `docs/22-threat-model.md`
- `docs/23-error-model.md`
- `docs/24-template-library.md`
- `docs/25-end-to-end-sequences.md`
- `docs/26-state-machines.md`
- `docs/27-configuration-runtime-flags.md`
- `docs/28-reference-package-layout.md`
- `docs/29-acceptance-matrix.md`
- `docs/30-operational-failure-recovery.md`
- `docs/31-performance-capacity.md`
- `docs/32-admin-ui-debugger.md`
- `docs/33-secret-crypto-design.md`
- `docs/34-js-runtime-abi-semantics.md`
- `docs/35-protocol-details.md`
- `docs/36-scheduler-algorithm-spec.md`

## Architecture decisions

- `adr/0001-core-provider-boundary.md`
- `adr/0002-availability-decrease-lease.md`
- `adr/0003-sandbox-network.md`
- `adr/0004-access-key-group.md`
- `adr/0005-single-node-first.md`

## Database migrations

- `migrations/001_init.sql`

## Examples

- `examples/config/example.json`
- `examples/schedulers/availability-round-robin.js`
- `examples/schedulers/session-affinity.js`
- `examples/templates/generic-openai.js`
- `examples/templates/oauth-refresh.js`
- `examples/templates/openrouter.js`
- `examples/templates/quota-aware.js`
- `examples/templates/request-rewrite.js`

## Reference implementation

- `reference/go/cmd/specdemo/main.go`
- `reference/go/go.mod`
- `reference/go/internal/domain/types.go`
- `reference/go/internal/lease/manager.go`
- `reference/go/internal/retry/controller.go`
- `reference/go/internal/retry/controller_test.go`
- `reference/go/internal/routing/candidates.go`
- `reference/go/internal/routing/candidates_test.go`
- `reference/go/internal/runtime/interfaces.go`
- `reference/go/internal/scheduler/scheduler.go`
- `reference/go/internal/scheduler/scheduler_test.go`
- `reference/go/internal/transport/httpclient.go`

## Test vectors

- `test-vectors/basic-routing.json`
- `test-vectors/candidate-scope.json`
- `test-vectors/hard-concurrency.json`
- `test-vectors/retry-429.json`
- `test-vectors/session-affinity.json`
- `test-vectors/stream-no-retry-after-byte.json`

## Verification scripts

- `scripts/verify.py`
- `scripts/verify.sh`
