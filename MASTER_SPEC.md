# Programmable Token Pool / Credential Pool Gateway — Master Specification v0.1

> 冻结日期：2026-10-01  
> 本文件由模块化规格合并生成，方便单文件交给 Coding Agent。冲突优先级仍以 `AGENT_IMPLEMENTATION_GUIDE.md` 为准：Invariants > Contracts > Module Docs > Master Spec > Examples。

## 文档索引

- `docs/00-overview.md` — 00 — 系统概览
- `docs/01-terminology-and-invariants.md` — 01 — 术语与核心不变量
- `docs/02-architecture.md` — 02 — 总体架构
- `docs/03-domain-model.md` — 03 — Domain Model
- `docs/04-auth-routing.md` — 04 — 认证与路由
- `docs/05-scheduler-concurrency.md` — 05 — Scheduler、Availability 与并发
- `docs/06-template-runtime.md` — 06 — Credential Template Runtime
- `docs/07-protocol-gateway.md` — 07 — Protocol Gateway
- `docs/08-transport-proxy.md` — 08 — Transport 与 Proxy
- `docs/09-streaming-websocket.md` — 09 — Streaming 与 WebSocket
- `docs/10-persistence.md` — 10 — SQLite 持久化
- `docs/11-admin-api.md` — 11 — 管理 API
- `docs/12-sources-classification.md` — 12 — Credential Source 与自动分类
- `docs/13-retry-cooldown.md` — 13 — Retry、Cooldown 与失败处理
- `docs/14-security-sandbox.md` — 14 — 安全与 JavaScript Sandbox
- `docs/15-observability.md` — 15 — 可观测性
- `docs/16-cliproxyapi-compatibility.md` — 16 — CLIProxyAPI 非插件核心兼容目标
- `docs/17-testing.md` — 17 — 测试规格
- `docs/18-versioning-migrations.md` — 18 — Versioning、热更新与 Migration
- `docs/19-deployment.md` — 19 — 部署
- `docs/20-implementation-plan.md` — 20 — 详细实现计划
- `docs/21-decisions-nongoals.md` — 21 — Design Decisions 与 Non-goals
- `docs/22-threat-model.md` — 22 — Threat Model
- `docs/23-error-model.md` — 23 — 统一错误模型
- `docs/24-template-library.md` — 24 — 官方 Template 与 Scheduler Library
- `docs/25-end-to-end-sequences.md` — 25 — 端到端时序与热路径
- `docs/26-state-machines.md` — 26 — 状态机
- `docs/27-configuration-runtime-flags.md` — 27 — 进程配置与运行参数
- `docs/28-reference-package-layout.md` — 28 — 推荐代码包结构与依赖方向
- `docs/29-acceptance-matrix.md` — 29 — 验收矩阵
- `docs/30-operational-failure-recovery.md` — 30 — 故障、恢复与一致性
- `docs/31-performance-capacity.md` — 31 — 性能与容量设计
- `docs/32-admin-ui-debugger.md` — 32 — 管理 UI 与 Template Debugger
- `docs/33-secret-crypto-design.md` — 33 — Secret 与加密设计
- `docs/34-js-runtime-abi-semantics.md` — 34 — JavaScript Runtime ABI 精确定义
- `docs/35-protocol-details.md` — 35 — 入站协议实现细节
- `docs/36-scheduler-algorithm-spec.md` — 36 — Scheduler 算法规范

---

<!-- BEGIN docs/00-overview.md -->

# 00 — 系统概览

## 1. 产品定义

系统名称暂称 **Programmable Token Pool Gateway（PTPG）**。

它同时承担四类职责：

1. **Credential Registry**：管理不同渠道获得的凭据。
2. **Programmable Provider Adapter**：通过 Template JS 适配不同上游。
3. **Routing & Scheduling Gateway**：根据 Client Key、协议、模型和状态选择凭据。
4. **Protocol Gateway**：提供 OpenAI/Anthropic 等标准入站 API，并透传/转换到上游。

它不是：

- 单纯 API Key Vault；
- 单纯 Round Robin Proxy；
- 单纯 OpenAI-compatible Reverse Proxy；
- Workflow Engine；
- 通用 Plugin Host。

## 2. 核心抽象

```text
AccessKey
   │ 1:1
   ▼
CredentialGroup
   │ scope = templateIds[]
   ▼
CredentialTemplate 1:N CredentialInstance
```

运行请求时：

```text
Request
  -> Protocol Adapter
  -> Authenticate AccessKey
  -> Load Group
  -> Expand templateIds to instances
  -> Filter by protocol/model/supports()
  -> Scheduler.select()
  -> Lease
  -> Template Runtime
  -> Host HTTP
  -> Upstream
```

## 3. “协议 / 模型 / 额度”重新定义

不设计成：

```text
Protocol -> Model -> Quota -> Credential
```

而是：

- **Protocol**：请求和 Template 能力之间的匹配条件。
- **Model**：请求和 Credential 能力之间的匹配条件。
- **Quota**：Provider-specific 状态；默认属于 custom fields。
- **Credential**：真正的核心资源。

## 4. 可编程性的边界

### Credential Template

高扩展性。负责 authentication、request rewrite、model mapping、upstream URL、OAuth refresh、quota check、availability、proxy selection、response handling 和 retry advice。

### Credential Group Scheduler

低扩展性。只负责从系统提供的候选集合中选择一个 Credential Instance。不能访问任意网络，也不能越过候选集合。

## 5. 单机优先

v0.x 明确以个人部署为目标：

```text
1 process
1 SQLite DB
1 in-memory Lease manager
N template workers
```

多节点不是 MVP 的约束条件。

<!-- END docs/00-overview.md -->

---

<!-- BEGIN docs/01-terminology-and-invariants.md -->

# 01 — 术语与核心不变量

## 1. 术语

### Credential Template

一种“可执行 Schema”，包含 metadata、protocol capabilities、default fields、custom fields、persistent state、JS source 和 hooks。

### Credential Instance

某个具体账号/Token/API Key/OAuth credential。Instance 必须属于且只属于一个 Template。

### Credential Group

Client API Key 的路由池。Group 通过 `templateIds[]` 定义候选范围，并拥有 Scheduler。

### Access Key

客户端调用 Gateway 时使用的 Key。其作用同时是认证与路由入口。

### Availability

Credential Instance 的持久化调度分值。数值越高，默认越优先。

### Decrease

当前实例每持有一个 active lease 时，对其 `effectiveAvailability` 施加的临时惩罚。

### Lease

一次 attempt 对某个 Credential Instance 的占用。

### Protocol Adapter

入站 API 的薄适配层，只提取路由必须的信息，并保留 Raw Request。

### Host HTTP

Template 唯一允许使用的网络出口。

## 2. 强制不变量

- **INV-001** 一个 Credential Instance 只能属于一个 Credential Template。
- **INV-002** 一个 Credential Template 可以拥有 0..N 个 Credential Instance。
- **INV-003** 一个 Credential Instance 可以暴露 0..N 个模型。
- **INV-004** 一个 Credential Template 可以声明 1..N 个协议能力。
- **INV-005** 一个有效 Access Key 唯一映射一个 Credential Group。
- **INV-006** Credential Group 的静态候选边界由 `templateIds[]` 决定。
- **INV-007** Custom Scheduler 只能返回 candidates 中已有的 `instanceId`。
- **INV-008** Quota 不是 Core 官方字段。
- **INV-009** 一次请求期间的并发惩罚不得通过 Core 自动写回 Instance `availability`。
- **INV-010** 模板的所有网络访问默认必须经过 `ctx.http`。
- **INV-011** `ctx.http` 默认强制绑定当前 Credential Instance 的 proxy policy。
- **INV-012** 下游 Client Key 默认必须被移除并替换为上游 credential。
- **INV-013** Streaming / WebSocket attempt 的 Lease 必须持续到连接结束。
- **INV-014** Secret 永远通过 secret accessor 暴露给模板，而不是普通 JSON field。
- **INV-015** 请求开始后绑定 Template Version；热更新不得改变正在执行请求的代码。
- **INV-016** Retry 的每次上游调用都视为新的 Attempt，有自己的 Lease 与 trace。
- **INV-017** Scheduler 锁内不能执行网络 I/O、Template JS、数据库慢查询。
- **INV-018** Group Scheduler 不理解 Provider 业务语义。

## 3. Availability 常量

持久化不要依赖 IEEE `-Infinity`。建议：

```text
AvailabilityDefault = 100
AvailabilityDisabledFloor = -9_000_000_000_000_000
```

Template `canSchedule=false` 优先于极低 availability；availability 主要用于排序。

<!-- END docs/01-terminology-and-invariants.md -->

---

<!-- BEGIN docs/02-architecture.md -->

# 02 — 总体架构

## 1. 组件图

```mermaid
flowchart LR
  C[Client] --> P[Protocol Adapter]
  P --> A[Access Key Auth]
  A --> G[Credential Group Resolver]
  G --> B[Candidate Builder]
  B --> S[Scheduler]
  S --> L[Lease Manager]
  L --> R[Template Runtime]
  R --> H[Host HTTP]
  H --> X[Proxy Transport]
  X --> U[Upstream]
  U --> H
  H --> R
  R --> P
  P --> C

  DB[(SQLite)] --> A
  DB --> G
  DB --> B
  DB --> R
  SEC[Secret Store] --> R
```

## 2. 模块职责

### protocol

负责 route、model、stream、session metadata、Client Key 的提取，并保留 raw request。  
不负责 Provider-specific rewrite、上游 endpoint 决策或 quota。

### auth

输入 Client API Key，输出：

```text
AccessKeyRecord + CredentialGroupID
```

### routing/candidate

- 展开 Group.templateIds
- 查询 instances
- 做静态 model/protocol filter
- 调用 Template `supports`
- 构造 CandidateSnapshot

### scheduler

输入 RequestRoutingContext 和 CandidateSnapshot，输出一个候选 instanceId。

### lease

保证 select + activeRequests++ 原子。

### runtime

- 加载并编译指定 Template Version
- 合并 fields
- 暴露 Host API
- 执行 hooks
- enforce timeout/capability

### transport

- HTTP direct
- HTTP proxy
- CONNECT
- cancellation
- stream copy
- optional SOCKS5

### store

SQLite persistence、optimistic concurrency、transaction、JSON state。

### secrets

encrypt/decrypt/redaction。

## 3. 请求阶段

```text
Ingress
  |
  +-- Parse
  +-- Auth
  +-- Build candidates
  +-- Acquire scheduling lock
  |     +-- snapshot active leases
  |     +-- score
  |     +-- select
  |     +-- create lease
  +-- Release scheduling lock
  |
  +-- Runtime execute
  +-- Upstream
  +-- stream/response
  +-- hooks
  +-- release lease
```

## 4. 快路径原则

调度锁内禁止 JS execution、HTTP/DNS、DB writes、大 JSON parsing、model discovery。锁内仅允许内存 snapshot、排序/选择和 lease mutation。

<!-- END docs/02-architecture.md -->

---

<!-- BEGIN docs/03-domain-model.md -->

# 03 — Domain Model

## 1. CredentialTemplate

推荐领域结构：

```ts
interface CredentialTemplate {
  id: string;
  name: string;
  description?: string;
  protocols: ProtocolCapability[];

  officialDefaults: {
    address?: string;
    baseUrl?: string;
    models?: ModelDefinition[];
    availability?: number;
    proxy?: string;
  };

  customFields: Record<string, unknown>;
  state: Record<string, unknown>;

  jsSource: string;
  version: number;

  createdAt: string;
  updatedAt: string;
}
```

Template 级 `key` 不建议放进普通 defaults；如确实要默认 Secret，应走 Template Secret Store，而不是 `officialDefaults.key` 明文 JSON。

## 2. CredentialInstance

```ts
interface CredentialInstance {
  id: string;
  templateId: string;
  name?: string;

  address?: string;
  baseUrl?: string;
  keySecretRef?: string;
  models?: ModelDefinition[];
  availability: number;
  proxy?: string;

  customFields: Record<string, unknown>;
  state: Record<string, unknown>;

  version: number;
  createdAt: string;
  updatedAt: string;
}
```

## 3. Official Fields

官方用户级字段只保留：

| 字段 | 含义 |
|---|---|
| address | Provider/account 语义地址，可选 |
| baseUrl | 默认上游基址 |
| key | Secret，物理存储为 secret ref/ciphertext |
| models | 可调用模型定义 |
| availability | 持久化调度分值 |
| proxy | 当前实例默认代理 |

内部 metadata 如 `id/templateId/version` 不属于“可编程业务字段”。

## 4. ModelDefinition

```ts
interface ModelDefinition {
  id: string;              // client-visible
  upstream?: string;       // upstream model id
  displayName?: string;

  protocols?: string[];

  forceResponseMapping?: boolean;

  inputModalities?: string[];
  outputModalities?: string[];

  metadata?: Record<string, unknown>;
}
```

如果 `upstream` 为空，默认等于 `id`。

允许多个不同 Credential Instance 暴露同一个 `id`；这是合法的模型池。

## 5. Field Merge

Effective Fields：

```text
template official defaults
<- instance official override
template custom fields
<- instance custom override
```

默认深合并规则：

- scalar：instance override
- object：recursive merge
- array：replace
- `null`：explicit clear
- missing：inherit

Template 可在 `initialize` hook 中派生运行时字段，但不应偷偷写数据库。

## 6. CredentialGroup

```ts
interface CredentialGroup {
  id: string;
  name: string;
  templateIds: string[];

  schedulerType: string;
  schedulerCode?: string;

  config: Record<string, unknown>;
  state: Record<string, unknown>;

  version: number;
}
```

## 7. AccessKey

```ts
interface AccessKey {
  id: string;
  name: string;
  secretHash: string;
  secretPrefix?: string;
  groupId: string;
  createdAt: string;
  revokedAt?: string;
}
```

## 8. CredentialSource

```ts
interface CredentialSource {
  id: string;
  name: string;
  type: string;
  templateId?: string;
  config: Record<string, unknown>;
  code?: string;
  state: Record<string, unknown>;
}
```

Source 只进入写路径，不进入 inference 热路径。

## 9. 数据一致性

- 删除 Template 时，如果仍有 Instance，默认拒绝；
- 删除 Group 时，如果仍有 AccessKey，默认拒绝；
- 删除 Credential Instance 不应自动修改 Group，因为 Group 引用的是 Template；
- Template ID 应稳定，不以显示名称做外键；
- Models 的 client-visible ID 在同一 Group 内可以重名；
- Instance 的 `availability` 必须是有限数值，NaN/Infinity 禁止持久化。

<!-- END docs/03-domain-model.md -->

---

<!-- BEGIN docs/04-auth-routing.md -->

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

<!-- END docs/04-auth-routing.md -->

---

<!-- BEGIN docs/05-scheduler-concurrency.md -->

# 05 — Scheduler、Availability 与并发

## 1. 三个状态必须分开

### Stored Availability

持久化：

```text
instance.availability
```

由 Template 或 Admin 修改。

### Active Requests

进程内：

```text
leaseManager.activeCount(instanceId)
```

### Effective Availability

调度快照：

```text
effective = availability - activeRequests * decrease
```

不写数据库。

## 2. Decrease

默认 Template Manifest 可以声明：

```json
{"defaultDecrease": 1}
```

Template 可以动态：

```js
export function getDecrease(ctx) {
  if (ctx.fields.maxConcurrency === 1) return 1_000_000;
  return 10;
}
```

`decrease` 是一次 lease 的临时影响，不是永久扣分。

## 3. 默认调度器

默认：

```text
Availability First
+ Round Robin tie-break
+ active Lease decrease
```

伪代码：

```text
eligible = candidates where canSchedule

for c in eligible:
    c.effective =
        c.availability -
        c.activeRequests * c.decrease

max = max(effective)
top = all where effective == max

selected = rrCursor.next(top)
```

RR cursor 应以 Group + routing bucket 为键，而不是全局一个 cursor。

推荐 bucket：

```text
groupId + protocol + requestedModel
```

## 4. 原子 Select + Lease

错误实现：

```text
selected = scheduler.select()
unlock()
lease(selected)
```

正确：

```text
lock group scheduling shard
snapshot lease counts
selected = scheduler.select(snapshot)
leaseManager.acquire(selected)
unlock
```

## 5. Sharded Lock

单机 MVP 可以按 `groupId` 做 mutex。

后续高并发优化：

```text
hash(groupId) % 64
```

避免一个全局锁。

## 6. Template 动态计算

`supports/getAvailability/getDecrease/getScheduleHints` 可能涉及 JS，应在锁外计算为不可变 Candidate Snapshot。

锁内只读取：

- Snapshot
- 最新 active count
- group scheduler in-memory cursor/state 的最小部分

Custom Scheduler 如果本身是 JS，不能在全局锁内执行任意 JS。推荐两阶段：

1. 锁外构建候选与计算 scheduler preference；
2. 锁内用一个受限、纯选择的 scheduler implementation，或者为 custom scheduler 使用 group-serialized actor；
3. 最终 membership/hard concurrency 再校验；
4. 原子 lease。

Custom Scheduler ABI 因此应被限制为纯函数风格，禁止 I/O。

## 7. CandidateSnapshot

```ts
interface CandidateSnapshot {
  instanceId: string;
  templateId: string;

  availability: number;
  decrease: number;

  canSchedule: boolean;
  clientModel: string;
  upstreamModel: string;

  hardMaxConcurrency?: number;
  attributes: Record<string, unknown>;
}
```

Scheduler 调用时 Core 附加：

```ts
activeRequests: number;
effectiveAvailability: number;
```

## 8. 最大并发

Core 不提供官方 `maxConcurrency` 持久字段。

Template 可以将 custom field：

```json
{"maxConcurrency": 2}
```

投射成：

```ts
scheduleHints.hardMaxConcurrency = 2
```

Core 锁内检查：

```text
activeRequests < hardMaxConcurrency
```

从而保证硬限制。

## 9. Lease

```ts
interface Lease {
  id: string;
  requestId: string;
  attempt: number;
  groupId: string;
  templateId: string;
  instanceId: string;
  acquiredAt: number;
}
```

进程崩溃后 in-memory lease 自动清零。

## 10. Retry 与 Lease

一次请求：

```text
Attempt 1 -> Credential A -> release
Attempt 2 -> Credential B -> release
Attempt 3 -> Credential C -> release
```

失败实例默认加入本次 request 的 `attemptedSet`，除非 Template 明确要求 same-credential retry。

## 11. Session Affinity

作为 Scheduler Library，不属于 Core Credential 语义。

```text
sessionKey -> instanceId, expiresAt
```

选择：

1. 有绑定且绑定实例仍在候选且可调度 -> 选绑定；
2. 否则走 fallback scheduler；
3. 成功建立新绑定。

绑定实例暂时不可用时允许 failover 并重绑。

## 12. Weighted RR

不是用 `availability` 伪装权重。

Weighted RR 是独立 scheduler，可以读取：

```text
candidate.attributes.weight
```

`weight` 可来自 Template custom field。

## 13. Fill First

排序：

1. availability
2. template-defined priority attribute
3. stable credential order

然后尽可能连续使用首选 credential，直到其 Template 将其 availability/canSchedule 改变。

## 14. 并发正确性验收

必须验证：

- 100 并发请求不会把 hard max=1 的实例同时 lease 两次；
- cancellation 100% 释放 lease；
- upstream stream 中断释放 lease；
- panic/runtime error 使用 defer/finally 释放 lease；
- scheduler 返回 scope 外 instanceId 时拒绝执行。

<!-- END docs/05-scheduler-concurrency.md -->

---

<!-- BEGIN docs/06-template-runtime.md -->

# 06 — Credential Template Runtime

## 1. Template 是可执行 Provider Adapter

Template 是受限 JavaScript Module。

推荐代码形式：

```js
export const manifest = { ... };

export async function getAvailability(ctx) { ... }
export async function getDecrease(ctx) { ... }
export async function getScheduleHints(ctx) { ... }
export async function supports(ctx) { ... }
export async function beforeRequest(ctx) { ... }
export async function execute(ctx) { ... }
export async function afterResponse(ctx, result) { ... }
export async function onError(ctx, err) { ... }
export async function onFinish(ctx) { ... }
```

全部 hook 均 optional。

## 2. 运行模式

Template Version 保存 JS source。

运行时：

```text
(templateId, version)
  -> compile
  -> cached immutable program
```

Request 开始时固定 version。

## 3. Context 分层

```ts
ctx.request
ctx.route
ctx.fields
ctx.secrets
ctx.instance
ctx.template
ctx.runtime
ctx.http
ctx.log
ctx.clock
```

职责：

- `request`：请求元数据 + 受控 body accessor
- `route`：group/model/protocol/session
- `fields`：effective fields，read-only
- `secrets`：secret accessor
- `instance`：受控 state/field patch
- `template`：受控 template state patch
- `runtime`：request/lease runtime facts
- `http`：唯一网络出口
- `log`：redacted logger
- `clock`：时间 API

## 4. 请求对象

不要简单把整个原始 body 永久复制成 JS object，防止大上传导致双倍内存。

建议：

```ts
ctx.request.json()
ctx.request.text()
ctx.request.bodyStream()
ctx.request.headers
ctx.request.method
ctx.request.path
ctx.request.query
```

Host 可缓存一次解析结果。

## 5. Secret

```js
const key = await ctx.secrets.get("key");
```

返回 SecretString wrapper。

Logger 对 SecretString 自动 redaction。

真正 secret 不出现在 `ctx.fields`。

## 6. 默认 execute

如果 Template 没实现 `execute`，使用 default passthrough：

```text
resolve URL
-> clone safe downstream headers
-> strip client auth
-> inject upstream auth
-> rewrite model if configured
-> ctx.http.request()
-> return response
```

普通 OpenAI-compatible 模板可完全不自己写 execute。

## 7. HTTP Host API

推荐：

```ts
interface HostHttp {
  request(input: {
    url: string;
    method?: string;
    headers?: Record<string,string>;
    body?: Uint8Array | string | HostReadableStream;
    timeoutMs?: number;
    proxyMode?: "instance" | "direct";
  }): Promise<HostResponse>;
}
```

默认 `proxyMode="instance"`。

是否允许 Template 请求 `direct` 必须由 manifest capability 明确授权。

## 8. Instance Mutation

禁止给模板直接 DB handle。

提供 CAS 风格：

```js
await ctx.instance.patch({
  availability: 50,
  customFields: {
    cooldownUntil: Date.now() + 60000
  }
});
```

Host 内部使用 optimistic version。

## 9. Atomic Helpers

建议：

```js
await ctx.instance.atomic.increment("customFields.failCount", 1);
await ctx.instance.atomic.max("customFields.lastSeenAt", ctx.clock.now());
```

避免并发 hook 写丢状态。

## 10. Runtime Limits

每次 hook 至少需要：

- wall clock timeout
- CPU/instruction budget（运行时支持时）
- memory ceiling（worker 级更可靠）
- Host call count limit
- log bytes limit
- response rewrite size limit

## 11. Go + JS Runtime 建议

如果使用 Go：

- Goja 易嵌入，但异步 Host API、硬内存隔离需要额外工程；
- QuickJS 更接近完整 JS runtime，但 Go binding 和 worker 生命周期需谨慎；
- 最安全的长期方案是 `Gateway Core <RPC> Template Worker Process`；
- MVP 可以先用可信模板 + Goja/QuickJS，但 ABI 必须保留未来 worker 隔离空间。

## 12. 生命周期顺序

```text
load version
initialize context
supports
getAvailability
getDecrease
getScheduleHints
[select + lease]
beforeRequest
execute
afterResponse OR onError
onFinish
release lease
```

`onFinish` 无论成功失败都执行，但其失败不能覆盖主请求结果，只记录 secondary error。

## 13. Startup Hooks

```js
export async function onSystemStart(ctx) {}
export async function onInstanceStart(ctx) {}
```

用途：

- 清理模板自己持久化的旧并发标记
- token 预热
- metadata migration
- state sanity check

Startup hook 必须有超时和错误隔离。

<!-- END docs/06-template-runtime.md -->

---

<!-- BEGIN docs/07-protocol-gateway.md -->

# 07 — Protocol Gateway

## 1. 原则：薄 Adapter + Raw Passthrough

不要把所有 Provider 请求统一转换成一个“超级内部消息格式”。

这样会持续丢失：

- 新字段
- tool payload
- multimodal
- reasoning metadata
- provider extension
- future protocol fields

正确方式：

```text
Raw Request
+ normalized routing metadata
```

## 2. Normalized RequestContext

```ts
interface RequestContext {
  requestId: string;
  protocol: string;
  model: string;
  stream: boolean;

  clientKeyId: string;
  groupId: string;

  method: string;
  path: string;
  query: string;

  headers: Record<string,string[]>;
  session?: SessionMetadata;
}
```

## 3. MVP 协议

### OpenAI Chat

```text
POST /v1/chat/completions
protocol = openai.chat
model = body.model
stream = body.stream === true
```

### OpenAI Responses

```text
POST /v1/responses
protocol = openai.responses
model = body.model
stream = body.stream === true
```

### Anthropic Messages

```text
POST /v1/messages
protocol = anthropic.messages
model = body.model
stream = body.stream === true
```

### Model Listing

```text
GET /v1/models
```

由当前 Group 动态聚合。

## 4. 入站认证 Header

OpenAI 默认：

```text
Authorization: Bearer <CLIENT_KEY>
```

Anthropic：

```text
x-api-key: <CLIENT_KEY>
```

允许配置兼容方式，但路由层最终只收到归一化 AccessKey ID。

## 5. Session Metadata

Adapter 尽量提取显式 session：

```ts
interface SessionMetadata {
  id?: string;
  conversationId?: string;
  promptCacheKey?: string;
  parentSessionId?: string;
  clientRequestId?: string;
  derivedHash?: string;
}
```

Session Scheduler 决定绑定逻辑。

## 6. Header Policy

默认 passthrough：

1. 删除 hop-by-hop headers；
2. 删除 client authentication；
3. 删除 gateway-only headers；
4. 保留安全业务 header；
5. Template 再注入上游认证及自定义 header。

必须正确处理：

```text
Connection
Proxy-Connection
Keep-Alive
Transfer-Encoding
Upgrade
TE
Trailer
```

WebSocket 例外由专用 transport 处理。

## 7. Body Rewrite

普通 passthrough 如果需要 model alias：

- JSON body 只改 `model`；
- 其他字段原样；
- 不做 schema normalize。

如果 body 非法 JSON，返回 protocol-specific 4xx，不进入 Scheduler。

## 8. Response Rewrite

默认不改。

仅在 `forceResponseMapping=true` 或 Template hook 要求时，对已知 response model 字段做轻量 rewrite。

Streaming 中需要逐 SSE event 处理，不能 buffer 整个流。

## 9. 请求大小

Adapter 必须设置：

- JSON body size limit
- header count/size limit
- decompression size limit

避免在进入 Template Runtime 前就被大 body 打爆。

<!-- END docs/07-protocol-gateway.md -->

---

<!-- BEGIN docs/08-transport-proxy.md -->

# 08 — Transport 与 Proxy

## 1. Proxy 是 Credential 官方字段

每个 Instance 可以有：

```text
proxy
```

例如：

```text
http://127.0.0.1:8080
http://user:pass@proxy.example.com:3128
direct
```

后续可增加：

```text
https://
socks5://
```

## 2. Proxy Resolution

默认：

```text
Template.getProxy()
  -> if returned: use it
  -> else instance.proxy
  -> else template default proxy
  -> else optional global proxy
  -> else direct
```

如果 Instance 显式 `direct`，禁止继承全局代理。

## 3. 强制绑定

Template：

```js
await ctx.http.request(...)
```

Host 在 request 创建时已经绑定 credential execution context。模板不能通过自己构造 socket 绕过 proxy。

## 4. HTTP Proxy

普通 HTTP：

```text
Client -> proxy
Request-URI = absolute URI
```

HTTPS：

```text
Gateway -> Proxy CONNECT upstream:443
Proxy -> 200
Gateway TLS -> upstream through tunnel
```

## 5. Proxy Authentication

支持 URL userinfo：

```text
http://user:password@host:port
```

Host 生成 Proxy-Authorization。Proxy password 视为 Secret。

## 6. Connection Pool Isolation

连接池 key 至少包含：

```text
scheme
upstream authority
proxy identity
TLS policy
```

不能把通过 Proxy A 创建的 keep-alive connection 错误复用到 Proxy B。

## 7. Template 自定义 Proxy

Template 可以：

```js
export function getProxy(ctx) {
  return ctx.fields.region === "jp"
    ? ctx.fields.jpProxy
    : ctx.fields.proxy;
}
```

返回值经过 Host validation。

## 8. SSRF 防护

默认建议：

- scheme allowlist；
- 禁止 cloud metadata 地址；
- 可配置 deny CIDR；
- DNS rebinding 二次校验；
- redirect target 重新校验；
- 禁止 `file://`、`gopher://` 等非 HTTP scheme。

## 9. Timeouts

分别配置：

```text
connectTimeout
tlsHandshakeTimeout
responseHeaderTimeout
idleConnTimeout
requestOverallTimeout
```

Streaming 的 overall timeout 不能使用普通短请求默认值。

## 10. Cancellation

下游 cancel 必须传播：

```text
downstream cancel
 -> upstream request cancel
 -> body close
 -> lease release
```

<!-- END docs/08-transport-proxy.md -->

---

<!-- BEGIN docs/09-streaming-websocket.md -->

# 09 — Streaming 与 WebSocket

## 1. SSE

SSE 不能先完整读取再返回：

```text
upstream Body
 -> optional event transform
 -> downstream writer
```

必须具备：

- flush each event/chunk
- backpressure
- cancel propagation
- error accounting
- lease held until EOF

## 2. Bootstrap Retry

如果上游在**任何下游字节发送前**失败，可以安全 retry。

一旦下游已经发送响应内容：

```text
默认禁止透明切换 credential
```

因此 runtime 必须维护：

```text
downstreamStarted bool
```

Retry policy 必须接收该状态。

## 3. Streaming Keepalive

可选支持 SSE comment ping，但只能由 Protocol Adapter 明确支持时启用，不能向任意字节流注入数据。

## 4. Non-stream Keepalive

某些长耗时非流式接口如果需要 whitespace keepalive，应作为特定协议行为，不进入通用 Core。

## 5. Stream Transform

Template 可以请求 SSE 事件级 transform：

```js
return ctx.http.passthrough({
  transformSSE(event) { ... }
});
```

Host 应逐 event 调用；禁止让 JS 持有无限长度 buffer。

## 6. WebSocket

长期兼容目标：

```text
downstream WS
 <-> Gateway
 <-> upstream WS
```

Credential Lease 从 upstream dial 前建立，到任意一端 close 后释放。

## 7. WebSocket Proxy

HTTP proxy/CONNECT 情况下，WS dialer 同样必须绑定 instance proxy。

## 8. WebSocket Template API

未来推荐：

```ts
ctx.http.websocket(url, options): Promise<HostWebSocket>
```

Template 不直接拿 OS socket。

## 9. Retry 规则

WebSocket handshake 失败且下游尚未 upgrade：

- 可以 retry other credential。

已经 upgrade：

- 不透明 retry；
- 正常 close/error 传播。

## 10. Backpressure

实现必须避免：

```text
upstream read fast
-> unbounded in-memory queue
-> slow downstream
```

默认应让读写链路自然 backpressure，或设有严格 bounded buffer。

<!-- END docs/09-streaming-websocket.md -->

---

<!-- BEGIN docs/10-persistence.md -->

# 10 — SQLite 持久化

## 1. 选择 SQLite WAL

个人部署优先：

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

优点：

- 单文件；
- transaction 完整；
- 运维简单；
- 支持并发读；
- 足够承载个人 API gateway metadata。

## 2. JSON 字段

Template/Instance Custom Fields、state、models 等先使用 JSON TEXT。

不要为了 Provider-specific 类型提前拆几十张表。

## 3. Secret

Secret 不放入：

```text
custom_fields_json
```

官方 key 使用：

```text
key_ciphertext
key_nonce
key_kid
```

Custom secret 长期建议独立：

```text
secret_items(owner_type, owner_id, name, ciphertext...)
```

## 4. Optimistic Version

Template / Instance / Group：

```text
version INTEGER NOT NULL
```

Patch：

```sql
UPDATE credential_instances
SET ..., version = version + 1
WHERE id = ? AND version = ?;
```

0 row = conflict。

## 5. Runtime State 与 Persistent State

默认内存：

- active leases
- RR cursor
- compiled template cache
- short-lived session affinity
- in-flight spans

持久化：

- availability
- template custom state
- explicit cooldown（若 Template 选择）
- OAuth tokens
- imported credentials

## 6. Transaction 边界

以下操作应使用 transaction：

- 创建 credential + secret；
- rotate access key；
- rebind template；
- delete template dependency check；
- admin patch + audit log。

不要让 inference 请求长期持有 write transaction。

## 7. WAL 维护

提供：

- WAL size metric；
- background checkpoint；
- graceful shutdown checkpoint 可选。

## 8. Backup

提供 SQLite backup API 或 `VACUUM INTO`/在线备份机制，不建议用户直接复制活跃 DB 文件。

## 9. Integrity

启动时：

- migrations version check；
- foreign key check；
- optionally `PRAGMA quick_check`；
- key encryption master key 可用性检查。

<!-- END docs/10-persistence.md -->

---

<!-- BEGIN docs/11-admin-api.md -->

# 11 — 管理 API

## 1. 管理面和推理面分离

建议：

```text
/v1/...                 inference
/api/admin/v1/...       management
```

Management Auth 使用独立密钥。默认只监听 localhost，除非显式允许远程管理。

## 2. Templates

```text
GET    /api/admin/v1/templates
POST   /api/admin/v1/templates
GET    /api/admin/v1/templates/{id}
PATCH  /api/admin/v1/templates/{id}
DELETE /api/admin/v1/templates/{id}
POST   /api/admin/v1/templates/{id}/validate
POST   /api/admin/v1/templates/{id}/test
```

Template PATCH 必须带 version。

## 3. Instances

```text
GET    /api/admin/v1/credentials
POST   /api/admin/v1/credentials
GET    /api/admin/v1/credentials/{id}
PATCH  /api/admin/v1/credentials/{id}
DELETE /api/admin/v1/credentials/{id}

POST /api/admin/v1/credentials/{id}/test
POST /api/admin/v1/credentials/{id}/rebind-template
```

GET 默认：

```text
key: masked
proxy password: masked
```

## 4. Groups

```text
GET    /api/admin/v1/groups
POST   /api/admin/v1/groups
GET    /api/admin/v1/groups/{id}
PATCH  /api/admin/v1/groups/{id}
DELETE /api/admin/v1/groups/{id}
POST   /api/admin/v1/groups/{id}/simulate-route
```

`simulate-route` 不发真实上游请求，可以返回：

```json
{
  "candidates": [
    {
      "instanceId":"a",
      "availability":100,
      "activeRequests":1,
      "decrease":10,
      "effectiveAvailability":90
    }
  ],
  "selected":"a"
}
```

## 5. Access Keys

```text
GET    /api/admin/v1/access-keys
POST   /api/admin/v1/access-keys
DELETE /api/admin/v1/access-keys/{id}
POST   /api/admin/v1/access-keys/{id}/rotate
```

新 key 只在创建/rotate 响应中返回一次完整值。

## 6. Sources

```text
GET/POST/PATCH/DELETE /api/admin/v1/sources
POST /api/admin/v1/sources/{id}/pull
POST /api/admin/v1/sources/{id}/preview
```

## 7. Debugger

请求：

```json
{
  "templateId": "openrouter",
  "instanceId": "or-1",
  "protocol": "openai.responses",
  "model": "fast",
  "request": {
    "method": "POST",
    "path": "/v1/responses",
    "headers": {},
    "body": {}
  },
  "dryRun": true
}
```

响应允许显示：

- effective fields（secret masked）
- protocol match
- model resolution
- availability
- decrease
- schedule hints
- proxy（password masked）
- final URL
- final headers（auth masked）
- logs
- hook durations

## 8. Rebind Template

必须先 dry-run validation：

```json
{
  "compatible": true,
  "preservedFields": [],
  "missingRequiredFields": [],
  "droppedFields": []
}
```

用户确认后再真正切 templateId。

## 9. Audit

管理写操作记录：

```text
actor
action
objectType
objectId
timestamp
redactedDiff
```

## 10. 错误格式

Admin API 统一：

```json
{
  "error": {
    "code": "version_conflict",
    "message": "credential changed since version 4",
    "details": {}
  }
}
```

不要把内部 stack trace 默认返回给客户端。

<!-- END docs/11-admin-api.md -->

---

<!-- BEGIN docs/12-sources-classification.md -->

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

<!-- END docs/12-sources-classification.md -->

---

<!-- BEGIN docs/13-retry-cooldown.md -->

# 13 — Retry、Cooldown 与失败处理

## 1. 核心原则

Core 提供 retry engine，但不把 Provider-specific 状态码语义写死。

Template：

```js
onError(ctx, error)
```

返回标准 action。

## 2. ErrorAction

```ts
type ErrorAction =
  | { action: "stop" }
  | {
      action: "retry";
      target: "same" | "another";
      cooldownMs?: number;
    };
```

持久化 availability/cooldown 状态仍由 Template 控制。

## 3. 默认 Library Policy

官方 Generic HTTP Template 可以预置：

- 408 -> retry another
- 429 -> retry another + optional cooldown
- 500/502/503/504 -> retry another
- 401/403 -> 默认 stop + mark unhealthy（可配置）

这只是 Template Library 默认，不是 Core 硬规则。

## 4. Retry State

```ts
interface RetryState {
  requestId: string;
  round: number;
  attempt: number;
  attemptedInstanceIds: string[];
  downstreamStarted: boolean;
}
```

## 5. Cooldown

Template custom field：

```json
{"cooldownUntil": 1790000000000}
```

Template：

```js
getAvailability(ctx) {
  if (ctx.fields.cooldownUntil > ctx.clock.now()) return -1e12;
  return ctx.instance.availability;
}
```

或：

```js
supports(ctx) {
  return ctx.fields.cooldownUntil <= ctx.clock.now();
}
```

Core 不需要理解 cooldown 字段。

## 6. Request Scoped Errors

Template config 可描述：

```json
{
  "requestScopedErrors": [
    {
      "status": 400,
      "contains": ["context_length_exceeded"],
      "action": "stop"
    }
  ]
}
```

避免错误地认为所有 400 都说明 credential 坏了。

## 7. Retry 安全

- 已经向下游发送 stream 字节后，不自动跨 credential retry；
- 非幂等 endpoint 必须由 Template 显式标记 retry-safe；
- cancellation 不 retry；
- deadline exceeded 通常不 retry；
- max attempt/round 必须有硬上限；
- retry 不能造成无限 credential cycling。

## 8. Retry Candidate Filter

`target=another` 时默认排除：

```text
attemptedInstanceIds in current retry round
```

下一 round 是否允许重新尝试，由 retry policy 决定。

## 9. Upstream Error 标准化

Transport 统一区分：

- DNS
- connect timeout
- TLS
- proxy auth
- response header timeout
- HTTP status
- stream reset
- downstream cancellation

Template `onError` 获得结构化错误，而非只能解析字符串。

<!-- END docs/13-retry-cooldown.md -->

---

<!-- BEGIN docs/14-security-sandbox.md -->

# 14 — 安全与 JavaScript Sandbox

## 1. 信任模型

即使系统定位个人使用，Template 代码仍可能：

- 来自模板库；
- 被 Agent 生成；
- 从网络复制；
- 写错形成死循环；
- 意外泄露 Token。

因此不应直接使用宿主原生 `eval()`。

## 2. 禁止能力

默认 Sandbox 不提供：

```text
process
require
fs
os
net
tls
child_process
Deno
Bun
native fetch
native WebSocket
environment variables
```

## 3. 允许能力

仅通过 Host Capability：

```text
ctx.http
ctx.secrets
ctx.instance
ctx.template
ctx.log
ctx.clock
```

Capability 按 template manifest 授权。

## 4. 网络能力

普通 Template：

```text
network = credential-bound
```

即只能通过绑定 instance proxy 的 Host HTTP。

特殊系统 Template 可以声明 direct network，但管理员安装/启用时必须显式授权。

## 5. Secret Redaction

SecretString 应带 taint。

任何：

```text
ctx.log.info(secret)
```

必须输出 `[REDACTED]`。

仅依赖 wrapper 不够：字符串拼接可能丢 taint。日志层还应维护当前 request 的 known-secret set 做精确/前后缀 redaction。

## 6. Template Supply Chain

Template Library 条目包含：

```text
id
version
source hash
signature(optional)
author
capabilities
```

升级前显示 capability diff。

## 7. 资源隔离

长期最稳妥：

```text
Gateway Core
  <RPC>
Template Worker Process
```

Worker：

- memory limit
- CPU time
- no filesystem
- restricted syscalls
- no direct network

Host HTTP 请求经 RPC 回 Core。

## 8. 管理 API

- 默认 localhost；
- browser UI 做 CSRF 防护；
- management key 与 inference key 分离；
- response 不返回 secrets；
- import bundle 不能自动执行未知 Template code。

## 9. 数据加密

Key、OAuth token、proxy password：

- AES-256-GCM；
- random nonce；
- key id；
- master key outside DB。

## 10. SSRF

Host HTTP：

- scheme allowlist；
- optional private CIDR deny；
- metadata IP deny；
- DNS resolution validation；
- redirect target 重新校验。

## 11. Template 安装审批

新 Template 如果 capability 比旧版增加：

```text
directNetwork
mutateTemplate
backgroundTask
```

必须阻止静默自动升级。

<!-- END docs/14-security-sandbox.md -->

---

<!-- BEGIN docs/15-observability.md -->

# 15 — 可观测性

## 1. Request Log

每个 request：

```text
requestId
clientKeyId
groupId
protocol
requestedModel
stream
startAt
endAt
status
latencyMs
attemptCount
```

每个 attempt：

```text
attempt
templateId
templateVersion
instanceId
clientModel
upstreamModel
availability
decrease
activeRequestsBefore
effectiveAvailability
scheduler
proxyMode
upstreamStatus
timeToFirstByte
duration
errorClass
retryAction
```

## 2. 不能记录

- complete API Key
- Authorization
- OAuth refresh token
- proxy password
- raw request body by default
- raw response body by default

## 3. Metrics

建议：

```text
gateway_requests_total
gateway_request_duration_seconds
gateway_attempts_total
gateway_upstream_errors_total
gateway_active_leases
gateway_template_hook_duration_seconds
gateway_template_errors_total
gateway_scheduler_select_seconds
gateway_stream_active
gateway_db_busy_total
gateway_proxy_errors_total
```

Labels 控制 cardinality，默认不要把 instanceId/任意 model 字符串作为 Prometheus 高基数 label。

## 4. Debug Trace

管理员可按 requestId 查看：

```text
candidate set
filter reasons
scores
selected
hook timings
retry chain
```

Secrets masked。

## 5. Health

```text
/health/live
/health/ready
```

ready 至少检查：

- DB；
- migrations；
- runtime pool；
- encryption key loaded。

## 6. Event 类型

建议内部 structured events：

```text
request.started
routing.candidates_built
scheduler.selected
lease.acquired
template.hook
upstream.headers
stream.started
retry.scheduled
lease.released
request.completed
```

方便将来接入 tracing。

<!-- END docs/15-observability.md -->

---

<!-- BEGIN docs/16-cliproxyapi-compatibility.md -->

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

<!-- END docs/16-cliproxyapi-compatibility.md -->

---

<!-- BEGIN docs/17-testing.md -->

# 17 — 测试规格

## 1. 测试层次

```text
Unit
Contract
Integration
Concurrency
Fault Injection
Security
Compatibility
```

## 2. Domain Unit

必须测试：

- field merge；
- model alias；
- version conflict；
- access key hash lookup；
- secret redaction；
- invalid availability NaN/Infinity；
- Template 删除依赖约束。

## 3. Scheduler Unit

至少：

1. availability 高者优先；
2. 同分 Round Robin；
3. active lease decrease 生效；
4. hard max concurrency 生效；
5. candidate scope enforcement；
6. weighted RR 统计近似；
7. session binding；
8. session failover；
9. retry attempted-set；
10. 无候选的稳定错误。

## 4. 并发 Race

Go 必须：

```bash
go test -race ./...
```

场景：

- 1000 goroutine；
- 3 credentials；
- max concurrency 1/2/unlimited；
- random cancel；
- random upstream delay。

最终：

```text
all active lease count == 0
```

## 5. Streaming

Mock upstream：

- 立即失败；
- response header 后失败；
- 第一个 SSE event 前失败；
- 第一个 event 后失败；
- 客户端中途断开。

验证 retry boundary 与 lease cleanup。

## 6. Proxy

启动本地 test proxy：

- 无认证；
- Basic auth；
- auth failure；
- CONNECT；
- direct bypass。

检查 upstream 观察到的请求确实经过对应 proxy。

## 7. Runtime

错误/恶意 Template：

```js
while(true){}
```

```js
throw new Error("x")
```

```js
ctx.log.info(secret)
```

```js
await ctx.http.request("http://169.254.169.254/")
```

验证 timeout/redaction/SSRF policy。

## 8. Secret Snapshot

对日志、Admin API response、debug trace 做 fixture secret 全文搜索，必须 0 命中。

## 9. Test Vectors

`test-vectors/` 中 JSON 是跨语言 contract test 输入。

实现方可以增加字段，但不得改变已有向量的期望语义。

## 10. E2E

启动：

```text
Gateway
Mock Provider A
Mock Provider B
Mock HTTP Proxy
```

真实 curl/OpenAI SDK 调用，验证：

- model route；
- key replacement；
- stream；
- retries；
- proxy；
- model list。

## 11. 性能基线

不要求极致性能，但必须测：

- 无 Template 网络调用的路由开销；
- 100/500/1000 concurrent streams；
- SQLite admin write 对 inference read 的影响；
- Template compile cache 命中率。

性能优化不能破坏 lease correctness。

<!-- END docs/17-testing.md -->

---

<!-- BEGIN docs/18-versioning-migrations.md -->

# 18 — Versioning、热更新与 Migration

## 1. Template Version

每次 `jsSource`、manifest 或影响执行语义的默认配置变更：

```text
version++
```

Request 绑定：

```text
templateId + version
```

Compiled cache key 同样如此。

## 2. 热更新

旧请求继续旧 program；新请求使用新 program。

旧 version 没有 active runtime 后可从 cache 驱逐。

## 3. Group Version

Scheduler code/config 更新也需要 version。

Session binding 默认建议在 group version 改变后失效；也可以在下一次使用时重新验证 membership。

## 4. Instance Version

用于 Admin/template patch 的 optimistic concurrency。

## 5. DB Migration

每个 migration：

- 单调递增 ID；
- transaction；
- `schema_migrations` 记录；
- 不修改已发布 migration；
- 大数据 rewrite 使用 shadow column/table。

## 6. Template State Migration

Template code 可以定义逻辑层迁移：

```js
migrateState(fromVersion, toVersion, state)
```

这与 DB schema migration 分开。

状态迁移失败应：

- 保留旧 state；
- 记录明确错误；
- 默认阻止新 version 接管该实例，而不是破坏数据。

## 7. Bundle Version

整个规格和模板库应有独立版本：

```text
coreApiVersion
templateApiVersion
schedulerApiVersion
```

例如：

```text
ptpg.template/v1
ptpg.scheduler/v1
```

Core 升级时必须校验兼容范围。

<!-- END docs/18-versioning-migrations.md -->

---

<!-- BEGIN docs/19-deployment.md -->

# 19 — 部署

## 1. 单机默认

```text
ptpg
  --listen 127.0.0.1:8317
  --db ./data/ptpg.db
  --templates ./templates
```

Admin 默认：

```text
127.0.0.1 only
```

Inference 可按用户配置监听 LAN。

## 2. 数据目录

```text
data/
  gateway.db
  gateway.db-wal
  gateway.db-shm
  backups/
  runtime/
```

Master encryption key 不放在此目录明文。

## 3. Shutdown

收到 SIGTERM：

1. stop accepting new requests；
2. wait grace period；
3. cancel remaining；
4. release/clear in-memory leases；
5. flush observability；
6. close DB。

## 4. Upgrade

- backup DB；
- run migration；
- validate templates compile；
- start readiness；
- only then expose traffic。

## 5. Docker

- data volume；
- secret/master key mount；
- admin port not published by default；
- template worker 无额外 Linux capabilities。

## 6. 配置层级

推荐：

```text
compiled defaults
< config file
< environment
< CLI flags
```

但凭据 Secret 不建议通过大而杂的 config 文件长期维护；应进入 Secret Store。

## 7. 单实例限制

v0.x 不保证两个 Gateway 进程同时打开同一个 DB 并共享 Lease 语义。

如果用户启动多进程，应显式报 warning 或阻止，直到实现分布式/DB-backed lease。

<!-- END docs/19-deployment.md -->

---

<!-- BEGIN docs/20-implementation-plan.md -->

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

<!-- END docs/20-implementation-plan.md -->

---

<!-- BEGIN docs/21-decisions-nongoals.md -->

# 21 — Design Decisions 与 Non-goals

## ADR-001：Credential 是核心实体

理由：一个 credential 可以跨多个模型/协议，Provider/Model 都不应成为它的数据库父级。

## ADR-002：Quota 不进 Core

理由：不同上游 quota 语义不同，统一 schema 会快速膨胀。

## ADR-003：Availability 持久、Decrease 临时

理由：把并发轮询直接写 availability 会产生高频 DB 写、崩溃残留和语义混乱。

## ADR-004：Group Scheduler 可编程

理由：必须支持 Session Affinity、自定义调度，以及 CLIProxyAPI 同类策略。

## ADR-005：Scheduler 可编程性低于 Template

Scheduler 只选 candidate；不允许网络和任意资源访问。

## ADR-006：Raw Passthrough 优先

避免统一中间协议因 Provider 新字段持续落后。

## ADR-007：Proxy 由 Host 强制

不能依赖 Template 作者“记得使用代理”。

## ADR-008：单机优先

v0.x 不为了未来多节点引入 Redis/etcd/consensus。

## ADR-009：Access Key 一对一映射 Group

避免认证后出现第二个权限/路由配置体系，保持“认证即路由”。

## ADR-010：模型重名允许存在

同一个 client-visible model 可以形成跨 Provider 候选池。消歧由 Group scope/alias/scheduler 完成。

## Non-goals

- 通用脚本自动化平台；
- arbitrary plugin ecosystem；
- 云租户计费平台；
- 完整 WAF；
- 多租户强隔离 SaaS；
- 自己实现所有 LLM protocol semantics；
- 第一版实现分布式一致性。

<!-- END docs/21-decisions-nongoals.md -->

---

<!-- BEGIN docs/22-threat-model.md -->

# 22 — Threat Model

## 1. 资产

- upstream API keys
- OAuth access/refresh tokens
- proxy credentials
- Client API keys
- Template source
- user traffic
- usage metadata

## 2. Template code 风险

风险：

- secret exfiltration；
- SSRF；
- infinite loop；
- memory exhaustion；
- log exfiltration。

控制：

- sandbox；
- host-only network；
- resource quotas；
- redaction；
- capability declaration。

## 3. Admin API 风险

风险：

- remote credential theft；
- template replacement；
- routing takeover。

控制：

- separate management auth；
- localhost default；
- TLS when remote；
- audit；
- CSRF when browser auth。

## 4. Inference API 风险

风险：

- brute force key；
- request body DoS；
- slowloris；
- expensive upstream fan-out。

控制：

- high entropy keys；
- body limits；
- server timeouts；
- retry caps；
- concurrency caps。

## 5. Proxy 风险

风险：

- credential in proxy logs；
- malicious proxy MITM；
- wrong connection pool reuse。

控制：

- TLS verification；
- connection pool isolation；
- no proxy secret logs。

## 6. Template Trust Levels

建议：

```text
system
trusted
untrusted
```

- `system`：官方签名，可获更多 capability；
- `trusted`：用户手写，可用 credential-bound HTTP；
- `untrusted`：未来 worker 强隔离，默认最少 capability。

## 7. 安全默认值

- Admin localhost only；
- templates cannot direct network；
- secrets masked；
- no raw body logging；
- deny cloud metadata address；
- max request body；
- max retry attempts；
- template timeout。

## 8. 失败原则

安全相关校验失败时：

```text
fail closed
```

例如：

- proxy URL 无法解析；
- template capability 不允许 direct；
- secret decrypt 失败；
- scheduler 返回 scope 外 credential；
- template version 不兼容。

<!-- END docs/22-threat-model.md -->

---

<!-- BEGIN docs/23-error-model.md -->

# 23 — 统一错误模型

## 1. 为什么需要统一错误类型

Transport、Protocol、Template、Scheduler 如果只用字符串错误，会导致 retry、安全和 HTTP 映射难以稳定。

内部统一：

```ts
interface GatewayError {
  code: string;
  category: string;
  message: string;
  retryable: boolean;
  upstreamStatus?: number;
  cause?: unknown;
  details?: Record<string, unknown>;
}
```

## 2. Error Category

建议：

```text
auth
routing
scheduler
template
transport
proxy
upstream_http
stream
timeout
cancelled
storage
security
```

## 3. 典型错误码

```text
auth_missing
auth_invalid
access_key_revoked
group_not_found
model_not_routable
no_credential_available
scheduler_invalid_selection
template_compile_failed
template_timeout
template_capability_denied
proxy_invalid
proxy_auth_failed
upstream_connect_failed
upstream_timeout
upstream_http_error
stream_reset
client_cancelled
version_conflict
secret_decrypt_failed
```

## 4. Protocol 映射

Protocol Adapter 负责把内部错误转换成对应协议的 HTTP/body 形式。

不能让 Template 直接决定 Gateway 身份验证失败应返回什么格式。

## 5. Retry 输入

Template `onError` 收到的是结构化 error snapshot：

```json
{
  "code":"upstream_http_error",
  "category":"upstream_http",
  "upstreamStatus":429,
  "retryable":true
}
```

而不是让 Template 解析 `"429 Too Many Requests"` 字符串。

<!-- END docs/23-error-model.md -->

---

<!-- BEGIN docs/24-template-library.md -->

# 24 — 官方 Template 与 Scheduler Library

## 1. 目标

用户日常使用时不应该为了 OpenRouter/OpenAI-compatible 这种普通场景自己写 JS。

可编程能力是上限，不是使用门槛。

## 2. Template Library 第一批

```text
generic-bearer
generic-header-auth
generic-openai
openrouter
anthropic-api-key
gemini-api-key
```

每个 Library Template 应包含：

- manifest；
- JSON field schema；
- defaults；
- JS implementation；
- compatibility tests；
- example instances；
- README。

## 3. Generic Bearer

字段：

```text
baseUrl
key
models
proxy
authHeader(default Authorization)
authScheme(default Bearer)
```

## 4. Generic OpenAI

协议：

```text
openai.chat
openai.responses
```

行为：

- downstream auth strip；
- upstream bearer；
- baseUrl + incoming path；
- model alias；
- raw passthrough；
- SSE passthrough。

## 5. OpenRouter

基于 generic-openai，增加可选：

```text
HTTP-Referer
X-Title
quota/rate-limit response parsing
```

这些仍然是 Template 行为。

## 6. Scheduler Library

第一批：

```text
availability-round-robin
weighted-round-robin
fill-first
least-active
session-affinity
```

Scheduler 配置必须有 Schema，避免任意 JSON 靠猜。

## 7. Library 版本

Library Template 更新不能直接覆盖用户已编辑代码。

建议区分：

```text
library template definition
installed template
```

安装时复制或引用一个版本；用户本地修改后升级要显示 diff。

<!-- END docs/24-template-library.md -->

---

<!-- BEGIN docs/25-end-to-end-sequences.md -->

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

<!-- END docs/25-end-to-end-sequences.md -->

---

<!-- BEGIN docs/26-state-machines.md -->

# 26 — 状态机

## 1. Credential Instance 状态

Core 不增加持久化 `status` 官方字段；状态由可调度结果推导。

逻辑状态：

```text
                   +------------------+
                   |    SCHEDULABLE    |
                   +------------------+
                    |        |       |
       supports=no  |        |       | hard concurrency full
                    v        |       v
              UNSUPPORTED    |    BUSY_NOW
                             |
              template disables / cooldown
                             v
                       UNAVAILABLE
```

`UNAVAILABLE` 可由以下任意条件导致：

- Template `supports()` 返回 false；
- Template `getScheduleHints()` 标明不可调度；
- Template 以 custom field 表达 cooldown；
- instance/template 已禁用（如实现 disabled custom convention）；
- 模型不匹配；
- 协议不匹配。

Core 不能根据 Provider 名称推断状态。

## 2. Request 状态

```text
RECEIVED
 -> AUTHENTICATED
 -> CANDIDATES_READY
 -> LEASED
 -> UPSTREAM_PENDING
 -> HEADERS_RECEIVED
 -> STREAMING? -----------+
 -> COMPLETED             |
                          |
任何阶段 ----------------> FAILED
客户端取消 --------------> CANCELLED
                          |
                          v
                       CLEANUP
                          |
                          v
                        DONE
```

`CLEANUP` 必须幂等。

## 3. Attempt 状态

```text
CREATED
 -> LEASED
 -> EXECUTING
 -> RESPONSE
 -> SUCCESS

EXECUTING/RESPONSE
 -> RETRYABLE_FAILURE
 -> RELEASED
 -> next attempt

EXECUTING/RESPONSE
 -> TERMINAL_FAILURE
 -> RELEASED
```

一个 Request 可有多个 Attempt，但同一时刻默认只有一个 active Attempt。

## 4. Lease 状态

```text
NONE -> ACQUIRED -> RELEASED
```

规则：

- 不允许 `ACQUIRED -> ACQUIRED`；
- `release()` 必须幂等；
- debug build 可对 double release 发出告警；
- lease 不持久化到 SQLite；
- process restart 后全部 lease 视为不存在。

## 5. Template Version 状态

```text
DRAFT/UNVALIDATED
 -> VALIDATED
 -> COMPILED
 -> ACTIVE
 -> SUPERSEDED
```

如果 compile 失败，旧 ACTIVE 版本继续服务。

## 6. OAuth Token 生命周期（模板级）

OAuth 只是 Template Library 行为，推荐状态：

```text
VALID
 -> NEAR_EXPIRY
 -> REFRESHING
 -> VALID

REFRESHING -> REFRESH_FAILED -> optional cooldown/unavailable
```

为避免并发 refresh 风暴，应使用 template/instance atomic state 或 singleflight：

```text
refresh lock key = instanceId + provider-defined token slot
```

## 7. Source 状态

```text
IDLE
 -> FETCHING
 -> CLASSIFYING
 -> APPLYING
 -> IDLE

any -> ERROR -> IDLE(next run)
```

Source 运行失败不能影响 Gateway 请求面可用性。

<!-- END docs/26-state-machines.md -->

---

<!-- BEGIN docs/27-configuration-runtime-flags.md -->

# 27 — 进程配置与运行参数

本章只定义 Core 运行配置；Provider 行为不要放入全局配置。

## 1. 推荐配置文件

```yaml
server:
  listen: "127.0.0.1:8317"
  publicBaseUrl: "http://127.0.0.1:8317"
  shutdownGrace: "30s"

database:
  path: "./data/token-pool.db"
  wal: true
  busyTimeout: "5s"

runtime:
  templateEngine: "goja"
  templateWallTimeout: "5s"
  templateCpuBudget: "250ms"
  maxTemplateMemoryBytes: 33554432
  maxLogBytesPerRequest: 65536

routing:
  schedulingShards: 64
  defaultScheduler: "availability-round-robin"
  maxAttempts: 3

transport:
  dialTimeout: "10s"
  responseHeaderTimeout: "60s"
  idleConnTimeout: "90s"
  maxIdleConns: 256
  maxIdleConnsPerHost: 32
  allowDirectNetworkFromTemplates: false

secrets:
  masterKeyProvider: "file-or-keychain"
  masterKeyId: "local-v1"

admin:
  enabled: true
  listen: "127.0.0.1:8318"
  remoteAccess: false

observability:
  level: "info"
  metrics: true
  requestLog: true
```

## 2. 环境变量

环境变量只用于部署级 Secret/路径：

```text
PTPG_CONFIG
PTPG_DB_PATH
PTPG_MASTER_KEY
PTPG_ADMIN_TOKEN
PTPG_LOG_LEVEL
```

Provider credential 不建议通过环境变量批量注入，因为会绕过统一 Secret Store 与管理面。

## 3. 启动顺序

```text
parse config
 -> validate
 -> initialize master key
 -> open SQLite
 -> migration
 -> repositories
 -> load/compile templates
 -> run template onSystemStart/onInstanceStart
 -> initialize lease manager/schedulers
 -> start admin listener
 -> start gateway listener
```

如果任一 migration 或 master-key 初始化失败，Gateway 不得进入 ready 状态。

单个非关键 Template compile 失败时可进入 degraded 状态，但对应实例不得进入候选。

## 4. Readiness

`/health/live`：进程存活。

`/health/ready` 至少检查：

- DB writable/readable；
- master key loaded；
- gateway router initialized；
- migration current。

不要把所有上游 Provider 可达性作为 readiness 条件，否则一个 Provider 故障会让整个网关被重启。

## 5. 优雅退出

收到 SIGTERM：

1. readiness=false；
2. 停止接收新请求；
3. 等待 active request 至 grace deadline；
4. cancel remaining request；
5. close transports；
6. flush logs/metrics；
7. close SQLite。

## 6. 热加载

MVP 可对以下对象支持管理 API 热加载：

- Template source/version；
- Credential Instance fields；
- Group scheduler/config；
- Access Key revoke/create。

进程级配置例如 listen address、master key provider 可以要求 restart。

<!-- END docs/27-configuration-runtime-flags.md -->

---

<!-- BEGIN docs/28-reference-package-layout.md -->

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

<!-- END docs/28-reference-package-layout.md -->

---

<!-- BEGIN docs/29-acceptance-matrix.md -->

# 29 — 验收矩阵

此文件是 Coding Agent 的最终验收表。MVP 标记 MUST 的项目全部通过才算完成。

| ID | 场景 | 级别 | 验收条件 |
|---|---|---|---|
| A001 | Client Key -> Group | MUST | 一个有效 Key 精确解析到一个 Group |
| A002 | 无效 Key | MUST | 不泄露 Group/credential 信息；401 |
| A003 | revoked Key | MUST | 401，不进入 candidate resolution |
| A004 | Group scope | MUST | Scheduler 永远不能选 `templateIds[]` 外实例 |
| A005 | 多模型实例 | MUST | 一个实例可匹配多个 client model |
| A006 | model mapping | MUST | client model 可映射到不同 upstream model |
| A007 | 多协议模板 | MUST | 一个 Template 可声明多个 protocol |
| A008 | availability | MUST | 高 effective availability 优先 |
| A009 | RR tie break | MUST | 同分候选公平轮询，不固定首项 |
| A010 | decrease | MUST | active lease 导致临时 score 下降，不写 DB |
| A011 | hard concurrency | MUST | 100 并发下 max=1 不发生双 lease |
| A012 | lease cancel | MUST | 客户端取消后 active count 回到 0 |
| A013 | panic cleanup | MUST | runtime panic 后 lease 释放 |
| A014 | generic passthrough | MUST | 客户端 key 被替换为 instance secret |
| A015 | header sanitize | MUST | client Authorization 不会误透传 |
| A016 | proxy HTTP | MUST | upstream 请求实际经过 HTTP proxy |
| A017 | proxy auth | MUST | Basic authenticated HTTP proxy 可用 |
| A018 | HTTPS CONNECT | MUST | HTTPS upstream 经 HTTP CONNECT proxy |
| A019 | template network isolation | MUST | template 无原生 socket/fetch/file access |
| A020 | SSE streaming | MUST | chunk 低延迟透传，不缓冲整个响应 |
| A021 | stream lease lifetime | MUST | EOF/cancel/error 前 lease 不释放 |
| A022 | retry before downstream | MUST | 未开始下游时可换 credential 重试 |
| A023 | no retry after byte | MUST | 下游已发送字节后默认不换凭据重试 |
| A024 | 429 template policy | MUST | Core 不写 Provider-specific quota 逻辑 |
| A025 | session affinity | SHOULD | 同 session 在 credential 可用时稳定绑定 |
| A026 | affinity failover | SHOULD | 绑定实例不可用时可重新选择/绑定 |
| A027 | template update pinning | MUST | 在途请求继续使用开始时版本 |
| A028 | template compile failure | MUST | 新版本失败不破坏旧 active 版本 |
| A029 | secret at rest | MUST | SQLite 中无明文 upstream key |
| A030 | secret logs | MUST | 日志/trace/debugger 无完整 secret |
| A031 | `/v1/models` scoped | MUST | 仅返回当前 Client Key 可达模型 |
| A032 | source transactionality | SHOULD | 分类失败不产生半实例 |
| A033 | optimistic admin update | SHOULD | stale version 更新返回 conflict |
| A034 | websocket lease | PHASE2 | WS close 前 lease 保持 |
| A035 | SOCKS5 | PHASE2 | instance proxy 支持 socks5 |
| A036 | OAuth refresh singleflight | PHASE2 | 并发请求不会同时刷新同 token |

## 1. 强制测试命令

参考 Go 实现最终应支持：

```bash
go test -race ./...
```

并建立集成测试：

```bash
./scripts/integration-test.sh
```

## 2. 性能底线

个人部署 MVP 不是高性能竞赛，但调度热路径应满足：

- 不做每次请求全表扫描；
- 不在 group lock 内执行网络 I/O；
- 不在 group lock 内执行任意长 JS；
- streaming 不复制完整 body；
- active lease 查询为 O(1)；
- candidate expand 可通过 template_id index 完成。

## 3. 安全验收

至少主动测试：

- malicious template infinite loop；
- template 试图读取 env/file；
- SSRF 到 loopback/link-local（按部署 policy）；
- proxy credential redaction；
- management endpoint 未授权访问；
- secret interpolation 后错误日志泄露；
- custom scheduler 返回任意 scope 外 ID。

<!-- END docs/29-acceptance-matrix.md -->

---

<!-- BEGIN docs/30-operational-failure-recovery.md -->

# 30 — 故障、恢复与一致性

## 1. SQLite Busy

启用 WAL + busy timeout。请求热路径应优先使用已加载的只读配置快照；不要因为一次普通请求而同步写 DB。

Template 如果在 `afterResponse` 修改状态失败：

- 响应已经成功时，默认不要把客户端成功响应改成失败；
- 记录 structured error；
- 对安全关键写入（例如 OAuth refresh token 更新）允许模板声明 `mustPersist=true` 并在发送响应前完成。

## 2. DB corruption / migration failure

进程启动失败并明确报错，不允许在未知 schema 上继续运行。

建议运维：

- 定期 SQLite online backup；
- migration 前自动备份；
- migration 不可逆时提供显式说明。

## 3. JS Runtime 崩溃/超时

单次 runtime 异常不得崩溃宿主进程。

映射为：

```text
template_compile_error
template_timeout
template_runtime_error
template_memory_limit
```

并确保 Lease release。

## 4. Proxy 故障

区分：

- proxy URL invalid；
- DNS failure；
- proxy auth failed；
- CONNECT rejected；
- upstream through proxy timeout。

这些错误可交给 Template `onError`，但 Core 不自动永久降低 availability。

## 5. Upstream 故障

Retry Controller 受以下共同限制：

```text
maxAttempts
request deadline
downstreamStarted
attemptedSet
template ErrorAction
```

不得无限重试。

## 6. 进程 crash

因为 Lease 是 in-memory：

- restart 后 activeRequests=0；
- 不需要回滚 lease 数据；
- Template 如果自行持久化 transient state，必须在 startup hook 清理或带 TTL。

## 7. Secret master key 丢失

如果无法解密已有 credentials：

- Gateway 不应 silently 将 Secret 当空字符串使用；
- 实例进入不可执行状态；
- Admin API 显示 `secret_unavailable`（不回显 ciphertext）；
- 支持用户重新导入/重写 Secret。

## 8. 版本冲突

Admin 修改 Instance 时使用 CAS：

```text
expected version=7
DB current version=8
=> 409 conflict
```

不要 last-write-wins 覆盖别人/其他自动化刚写入的状态。

Template 高频 runtime state 建议使用原子字段 API，而不是读-改-写整个 JSON。

## 9. Clock

过期时间、cooldown、TTL 使用 Host `ctx.clock`，便于测试。

持久化时间使用 UTC RFC3339 或 Unix milliseconds，并统一规定一种格式。

## 10. Backpressure

Streaming copy 必须尊重 downstream backpressure。客户端消费慢时，不应无限把 upstream 数据堆在内存。

<!-- END docs/30-operational-failure-recovery.md -->

---

<!-- BEGIN docs/31-performance-capacity.md -->

# 31 — 性能与容量设计

## 1. 目标环境

首要目标是单用户、单机、几十到数百 Credential，而不是多租户 hyperscale Gateway。

参考设计容量而非硬承诺：

```text
Credential Templates:     1 - 200
Credential Instances:     1 - 10,000
Credential Groups:        1 - 1,000
Concurrent requests:      1 - 1,000
Typical personal deploy:  < 100 concurrent
```

## 2. Candidate 索引

维护：

```text
templateId -> []instanceId
instanceId -> immutable config snapshot
```

Group 展开不应每次扫描全部实例。

## 3. Template 编译缓存

Key：

```text
templateId + version
```

缓存 compiled artifact。一次请求 pin 版本。

## 4. Transport Pool

HTTP client/transport 需要按代理 identity 复用：

```text
transportKey = proxy scheme + host + user identity + TLS policy
```

不能每个请求新建 TCP Transport，否则连接复用丢失。

同时不能把含不同 proxy credential 的 transport 错误复用。

## 5. Lock 范围

锁内仅做：

- 读取 active counts；
- hard limit validation；
- scheduler 的短纯选择；
- increment lease；
- RR cursor/state 的小型更新。

禁止锁内：

- SQLite I/O；
- Template JS；
- HTTP；
- JSON 大对象序列化；
- OAuth refresh。

## 6. Template 状态写

不要每请求都持久化 RR 位置或 activeRequests。

只有业务上真正需要跨重启保留的 Template state 才写 DB。

## 7. `/v1/models` 缓存

可按：

```text
groupId + protocol + configRevision
```

缓存 model list。任何相关 Template/Instance/Group 更新提高 `configRevision`。

## 8. Streaming 内存

使用固定大小 copy buffer；不把 SSE 聚合成完整字符串。

Template stream transform 若存在，必须采用 iterator/transform stream 模式。

## 9. 日志采样

Debug body logging 默认关闭。即使开启，也要：

- 截断大小；
- redact secrets；
- multimodal/binary 不直接记录。

<!-- END docs/31-performance-capacity.md -->

---

<!-- BEGIN docs/32-admin-ui-debugger.md -->

# 32 — 管理 UI 与 Template Debugger

## 1. 页面

建议最小管理 UI：

```text
Dashboard
Templates
Credentials
Groups
Client Keys
Sources
Request Logs
Template Debugger
Settings
```

## 2. Credentials 列表

显示：

- ID / name；
- template；
- models；
- stored availability；
- current active requests（runtime）；
- calculated effective availability（仅观察值）；
- proxy mode（secret masked）；
- last success/error；
- custom summary。

不显示完整 Key。

## 3. Template 编辑器

至少支持：

- JS source；
- manifest；
- syntax/compile validation；
- version diff；
- save as new version；
- rollback by creating another version；
- Test/Dry Run。

## 4. Debugger 输入

```json
{
  "templateId": "generic-openai",
  "instanceId": "cred-1",
  "protocol": "openai.chat",
  "model": "gpt-x",
  "request": {
    "method": "POST",
    "path": "/v1/chat/completions",
    "headers": {},
    "body": {}
  },
  "networkMode": "disabled"
}
```

## 5. Debugger 输出

```text
Resolved template version
Resolved fields
Resolved model mapping
supports result
availability
decrease
schedule hints
resolved proxy (masked)
request before rewrite
request after rewrite (secrets masked)
logs
host API calls
optional upstream status
```

## 6. Dry Run 网络语义

`networkMode=disabled`：`ctx.http` 返回显式 `dry_run_network_disabled`，并记录拟发出的 URL/method/masked headers。

可选 `networkMode=real` 只能管理员显式开启，并显示警告，因为可能消耗额度或执行有副作用的 API 请求。

## 7. Key 创建

Client API Key 只在创建时显示一次完整 secret：

```text
ptpg_xxxxxxxxx
```

数据库只保存 hash + prefix。

## 8. Group 调试

提供 Route Preview：

```text
Key/Group + Protocol + Model + Session
 -> 展示所有候选
 -> 展示被过滤原因
 -> availability/decrease/active/effective
 -> scheduler 结果
```

Preview 默认不 acquire real lease，也不发网络请求。

<!-- END docs/32-admin-ui-debugger.md -->

---

<!-- BEGIN docs/33-secret-crypto-design.md -->

# 33 — Secret 与加密设计

## 1. Secret 类型

至少包括：

- upstream API key；
- OAuth access/refresh token；
- client secret；
- proxy password；
- provider-specific secret custom field。

普通 custom fields 不应该偷偷保存 Secret。

## 2. Envelope

推荐格式：

```text
ciphertext
nonce
key_id
algorithm = AES-256-GCM
```

每条 Secret 独立随机 nonce。

AAD 推荐：

```text
owner_type + "\0" + owner_id + "\0" + secret_name
```

防止 ciphertext 被复制到其他对象后仍可透明解密。

## 3. Master Key

优先：

1. OS keychain / secret service；
2. 外部 secret manager；
3. 仅个人开发使用环境变量/权限严格的文件。

禁止把 master key 写进 SQLite。

## 4. Client API Key

Client key 不需要可逆：

```text
random 256-bit secret
 -> display once
 -> store keyed hash/HMAC or slow hash
```

如果需要高性能查找，可存：

```text
prefix -> candidate rows -> constant-time verify
```

也可使用 HMAC-SHA-256(masterLookupKey, token) 做唯一索引。

## 5. Template Secret API

模板只通过：

```ts
ctx.secrets.get("key")
```

获得 Secret handle。

Host 应尽可能避免 Secret 变成可枚举普通字段。

## 6. Redaction

结构化日志 redactor 必须处理：

```text
Authorization
Proxy-Authorization
api-key
x-api-key
cookie (configurable)
known secret values
proxy userinfo
```

错误对象在序列化前也经过 redactor。

## 7. Backup

SQLite 备份包含 ciphertext。恢复时必须同时具备对应 master key，否则 Secret 无法使用。

提供可选 encrypted export bundle，用新的 export passphrase 重新封装 Secret。

<!-- END docs/33-secret-crypto-design.md -->

---

<!-- BEGIN docs/34-js-runtime-abi-semantics.md -->

# 34 — JavaScript Runtime ABI 精确定义

本章补充 `contracts/template-runtime.d.ts` 的运行语义。

## 1. Module 形式

Template source 使用 ES module 风格导出：

```js
export const manifest = {...};
export async function execute(ctx) {...}
```

如果所选 JS 引擎不原生支持 ESM，可在编译层转换，但对模板作者保持此接口。

## 2. Hook 缺省

Hook 不存在时使用 Core/Library 默认值：

```text
supports           => true (前提是静态 protocol/model 已匹配)
getAvailability    => instance.availability
getDecrease        => manifest.defaultDecrease ?? 1
getScheduleHints   => {}
getProxy           => instance.proxy ?? template default proxy ?? global policy
beforeRequest      => no-op
execute            => default passthrough（仅模板声明可用时）
afterResponse      => no-op
onError            => {action:"stop"}
onFinish           => no-op
```

## 3. Hook 调用上下文

每次 Attempt 创建独立 `TemplateContext`。Context 不能跨请求保存。

`ctx.fields` 是：

```text
template official defaults
 + template custom fields
 + instance official overrides
 + instance custom fields
```

Secret 不参与此 merge。

## 4. Body 读取

`ctx.request.json()` / `text()` 如果会消耗 request stream，Host 必须实现可重放缓存或明确限制只能调用一次。

MVP 对 AI JSON 请求建议在协议层设置最大 body 后读取为 replayable bytes，再交 Runtime。

## 5. HostResponse ownership

`execute()` 返回 HostResponse 后：

- 非流 body 可由 Gateway 消费；
- 流 body ownership 转移给 Gateway；
- Template 不得在 return 后再次并发读取 body。

## 6. afterResponse 与流

`afterResponse` 默认只保证看到 response metadata，不应默认要求读取完整 stream。

若 Template 需要 stream inspection，应使用专门 `transformResponseStream` Phase 2 ABI；不要在 `afterResponse` 内偷偷全量读取流。

## 7. Patch

`ctx.instance.patch()` 必须：

- 校验官方字段；
- JSON-safe；
- 使用 CAS/原子 merge 语义；
- 不允许修改 `id/templateId`；
- 不允许写 Secret 明文到 customFields。

## 8. Atomic State

`atomic.increment(path, delta)` 用于数字计数器。

path 采用受限 dotted path，例如：

```text
runtime.failureCount
quota.window.429Count
```

必须阻止 prototype pollution 字段：`__proto__`, `constructor`, `prototype`。

## 9. 网络

`ctx.http.request`：

- 默认 proxyMode=`instance`；
- URL 必须绝对地址；
- 自动继承 request cancellation；
- response body streaming；
- Host 限制 header/body 大小；
- 默认禁止访问文件协议等非 HTTP(S) scheme。

## 10. Runtime timeout

区分：

- JS CPU/同步执行 budget；
- hook wall timeout；
- HTTP request timeout。

等待 `ctx.http` 不应该全部计入很短的 CPU budget，但仍受 request deadline。

## 11. Determinism

Scheduler JS 比 Template JS 更严格：

- 无网络；
- 无 Secret；
- 无 instance patch；
- 候选只读；
- 应在极短 deadline 内完成。

<!-- END docs/34-js-runtime-abi-semantics.md -->

---

<!-- BEGIN docs/35-protocol-details.md -->

# 35 — 入站协议实现细节

## 1. 原则

协议 Adapter 的职责是提取路由元数据并保持 Raw Passthrough，不要把所有 Provider 强制转成巨大统一 AST。

统一提取：

```text
protocol id
requested model
stream flag
session metadata
client auth key
request method/path/query/headers/body
```

## 2. OpenAI Chat Completions

Endpoint：

```text
POST /v1/chat/completions
```

提取：

```text
body.model
body.stream
```

默认透传所有未知字段，包括未来新增参数、tools、多模态内容。

## 3. OpenAI Responses

Endpoint：

```text
POST /v1/responses
```

提取：

```text
body.model
body.stream
body.conversation / prompt_cache_key（如存在，用于 session metadata）
```

不要为了路由而删掉 Responses 特有字段。

## 4. Anthropic Messages

Endpoint：

```text
POST /v1/messages
```

提取 `model`、`stream`。Authentication adapter 接受项目定义的 Client API Key 方式；转上游时由 Template 重建 Provider auth header。

## 5. Auth header

Inbound 支持可配置：

```text
Authorization: Bearer <client-key>
x-api-key: <client-key>
```

如果两个同时出现且不一致，应拒绝请求，避免模糊认证。

## 6. Model mapping

模型匹配产生：

```text
clientModel = request body model
upstreamModel = ModelDefinition.upstream ?? id
```

Template beforeRequest/default passthrough 必须把上游 body 中的 model 改为 `upstreamModel`。

## 7. `/v1/models`

推荐返回 OpenAI-compatible list：

```json
{
  "object": "list",
  "data": [
    {"id":"client-model","object":"model","owned_by":"token-pool"}
  ]
}
```

若同一 client model 来自多个实例，只显示一次。

## 8. Error Mapping

内部错误不要直接把 Go/JS stack 返回客户端。

OpenAI style 示例：

```json
{
  "error": {
    "message": "No eligible credential is available for model x",
    "type": "routing_error",
    "code": "no_eligible_credential"
  }
}
```

## 9. Unknown Endpoint

MVP 未注册 endpoint 返回 404，而不是把任意客户端路径盲目代理到 Provider。

后续可由显式 Protocol Plugin/Template Route capability 扩展。

<!-- END docs/35-protocol-details.md -->

---

<!-- BEGIN docs/36-scheduler-algorithm-spec.md -->

# 36 — Scheduler 算法规范

## 1. Candidate Eligibility

在进入 scheduler 前，Core 至少过滤：

```text
candidate.belongsToGroupScope == true
candidate.protocolMatched == true
candidate.modelMatched == true
candidate.canSchedule == true
candidate.hardMaxConcurrency == nil OR active < hardMaxConcurrency
```

Scheduler 不负责“把不合法候选变合法”。

## 2. Effective Availability

```text
E(c) = A(c) - Active(c) * D(c)
```

其中：

- `A` = Template 解析后的 availability；
- `D` = decrease；
- `Active` = 原子 Lease manager 当前计数。

`A`、`D` 必须是 finite number。

## 3. Availability Round Robin

算法：

1. 计算所有 eligible 的 E；
2. 找最大值 `M`；
3. `top = {c | E(c)=M}`；
4. 对 top 以稳定 key 排序；
5. 用 `(group, protocol, model)` bucket cursor 选择下一个；
6. acquire lease；
7. 更新 cursor。

浮点比较推荐规范化为固定精度或只允许整数 score，MVP 更推荐整数，避免 `0.1+0.2` 比较问题。

## 4. 原子竞争

如果 scheduler 选择后，在 acquire 前 candidate 状态被其他 goroutine 改变，必须在同一锁/actor 内重新验证 hard max 与 active count。

## 5. Scheduler 返回验证

Custom Scheduler 返回：

```text
selectedInstanceId
```

Core 必须检查：

- ID 在传入 candidates 中；
- 当前仍满足 hard concurrency；
- instance 未被删除/revoked；
- request context 未取消。

否则当作 scheduler invalid result，不执行该 credential。

## 6. Weighted Round Robin

推荐 Smooth Weighted Round Robin。权重来自 `candidate.attributes.weight`，非正权重视为当前策略下不参与。

不要把 weight 写进 persistent availability，因为两者语义不同：

- availability = 当前偏好/健康/额度投影；
- weight = 长期流量比例。

可以先按最大 availability tier 筛选，再在 tier 内 weighted RR；也可以作为独立库策略明确文档化。

## 7. Fill First

用于希望先耗尽首选 credential 的场景：

```text
sort by availability DESC
then priority DESC
then stable ID ASC
pick first schedulable
```

Active decrease 仍可由 Template 调整是否把流量溢出到其他实例。

## 8. Session Affinity

Affinity 是 scheduler wrapper：

```text
affinity lookup
 -> valid bound candidate ? bound
 -> fallback scheduler
 -> bind result
```

TTL 从 Group config 读取。

## 9. Scheduler State

RR cursor、affinity cache 默认是 runtime state；是否持久化由 scheduler library 决定。

个人部署默认无需把 RR cursor 每请求写 SQLite。

## 10. 公平性

在以下条件下：

```text
same availability
same decrease
same request duration approximation
no affinity
```

N 个 credential 的长期选择次数应近似均匀。

提供统计测试，例如 10,000 次顺序调度，最大/最小选择次数差不超过 1（无并发、候选恒定时）。

<!-- END docs/36-scheduler-algorithm-spec.md -->

---

# Appendix A — Contract Files

## contracts/domain.ts

```ts
export type ProtocolId =
  | "openai.chat"
  | "openai.responses"
  | "anthropic.messages"
  | string;

export interface ModelDefinition {
  id: string;
  upstream?: string;
  displayName?: string;
  protocols?: ProtocolId[];
  forceResponseMapping?: boolean;
  inputModalities?: string[];
  outputModalities?: string[];
  metadata?: Record<string, unknown>;
}

export interface CredentialOfficialFields {
  address?: string | null;
  baseUrl?: string | null;
  models?: ModelDefinition[] | null;
  availability: number;
  proxy?: string | null;
}

export interface CredentialTemplate {
  id: string;
  name: string;
  description?: string;
  protocols: ProtocolId[];
  officialDefaults: Partial<CredentialOfficialFields>;
  customFields: Record<string, unknown>;
  state: Record<string, unknown>;
  jsSource: string;
  version: number;
}

export interface CredentialInstance extends CredentialOfficialFields {
  id: string;
  templateId: string;
  name?: string;
  keySecretRef?: string;
  customFields: Record<string, unknown>;
  state: Record<string, unknown>;
  version: number;
}

export interface CredentialGroup {
  id: string;
  name: string;
  templateIds: string[];
  schedulerType: string;
  schedulerCode?: string;
  config: Record<string, unknown>;
  state: Record<string, unknown>;
  version: number;
}

export interface AccessKey {
  id: string;
  name: string;
  secretHash: string;
  secretPrefix?: string;
  groupId: string;
  revokedAt?: string | null;
}

export interface CredentialSource {
  id: string;
  name: string;
  type: string;
  templateId?: string | null;
  config: Record<string, unknown>;
  code?: string;
  state: Record<string, unknown>;
  version: number;
}
```

## contracts/template-runtime.d.ts

```ts
import type { ModelDefinition, ProtocolId } from "./domain";

export interface TemplateManifest {
  id: string;
  apiVersion: "ptpg.template/v1";
  protocols: ProtocolId[];
  defaultDecrease?: number;
  capabilities?: {
    credentialBoundNetwork?: boolean;
    directNetwork?: boolean;
    mutateInstance?: boolean;
    mutateTemplate?: boolean;
    setSecrets?: boolean;
  };
}

export interface SessionMetadata {
  id?: string;
  conversationId?: string;
  promptCacheKey?: string;
  parentSessionId?: string;
  clientRequestId?: string;
  derivedHash?: string;
}

export interface HeaderMap {
  get(name: string): string | undefined;
  set(name: string, value: string): void;
  delete(name: string): void;
  entries(): Array<[string, string]>;
}

export interface TemplateRequest {
  requestId: string;
  protocol: ProtocolId;
  model: string;
  upstreamModel: string;
  stream: boolean;
  method: string;
  path: string;
  query: string;
  headers: HeaderMap;
  session?: SessionMetadata;

  json<T = unknown>(): Promise<T>;
  text(): Promise<string>;
}

export interface SecretString {
  readonly __secretBrand: unique symbol;
  reveal(): string;
}

export interface SecretAccessor {
  get(name: string): Promise<SecretString | undefined>;
}

export interface HostReadableStream {
  [Symbol.asyncIterator](): AsyncIterator<Uint8Array>;
}

export interface HostResponse {
  status: number;
  headers: HeaderMap;
  body: HostReadableStream;
  json?<T = unknown>(): Promise<T>;
  text?(): Promise<string>;
}

export interface HostHttp {
  request(input: {
    url: string;
    method?: string;
    headers?: Record<string, string>;
    body?: string | Uint8Array | HostReadableStream;
    timeoutMs?: number;
    proxyMode?: "instance" | "direct";
  }): Promise<HostResponse>;

  websocket?(
    url: string,
    options?: Record<string, unknown>
  ): Promise<unknown>;
}

export interface AtomicStateApi {
  increment(path: string, delta: number): Promise<number>;
  max(path: string, value: number): Promise<number>;
}

export interface MutableInstanceApi {
  readonly id: string;
  readonly templateId: string;
  readonly availability: number;
  readonly activeRequests: number;

  patch(input: {
    availability?: number;
    customFields?: Record<string, unknown>;
    state?: Record<string, unknown>;
  }): Promise<void>;

  setSecret?(name: string, value: string): Promise<void>;
  atomic: AtomicStateApi;
}

export interface MutableTemplateApi {
  readonly id: string;
  readonly version: number;
  patchState(statePatch: Record<string, unknown>): Promise<void>;
  atomic: AtomicStateApi;
}

export interface TemplateContext {
  request: TemplateRequest;
  fields: Readonly<Record<string, unknown>>;
  secrets: SecretAccessor;

  instance: MutableInstanceApi;
  template: MutableTemplateApi;

  route: {
    groupId: string;
    clientKeyId: string;
    requestedModel: string;
    upstreamModel: string;
  };

  runtime: {
    attempt: number;
    activeRequests: number;
    downstreamStarted: boolean;
  };

  http: HostHttp;

  log: {
    debug(...args: unknown[]): void;
    info(...args: unknown[]): void;
    warn(...args: unknown[]): void;
    error(...args: unknown[]): void;
  };

  clock: {
    now(): number;
  };
}

export interface ScheduleHints {
  hardMaxConcurrency?: number;
  attributes?: Record<string, unknown>;
}

export type ErrorAction =
  | { action: "stop" }
  | {
      action: "retry";
      target: "same" | "another";
      cooldownMs?: number;
    };

export declare const manifest: TemplateManifest;

export declare function onSystemStart?(
  ctx: TemplateContext
): Promise<void>;

export declare function onInstanceStart?(
  ctx: TemplateContext
): Promise<void>;

export declare function supports?(
  ctx: TemplateContext
): boolean | Promise<boolean>;

export declare function getModels?(
  ctx: TemplateContext
): ModelDefinition[] | Promise<ModelDefinition[]>;

export declare function getAvailability?(
  ctx: TemplateContext
): number | Promise<number>;

export declare function getDecrease?(
  ctx: TemplateContext
): number | Promise<number>;

export declare function getScheduleHints?(
  ctx: TemplateContext
): ScheduleHints | Promise<ScheduleHints>;

export declare function getProxy?(
  ctx: TemplateContext
): string | null | Promise<string | null>;

export declare function beforeRequest?(
  ctx: TemplateContext
): Promise<void>;

export declare function execute?(
  ctx: TemplateContext
): Promise<HostResponse>;

export declare function afterResponse?(
  ctx: TemplateContext,
  response: HostResponse
): Promise<void>;

export declare function onError?(
  ctx: TemplateContext,
  error: unknown
): Promise<ErrorAction>;

export declare function onFinish?(
  ctx: TemplateContext
): Promise<void>;
```

## contracts/scheduler-runtime.d.ts

```ts
export interface SchedulerRequest {
  requestId: string;
  groupId: string;
  protocol: string;
  requestedModel: string;
  session?: {
    id?: string;
    conversationId?: string;
    promptCacheKey?: string;
    parentSessionId?: string;
    clientRequestId?: string;
    derivedHash?: string;
  };
}

export interface Candidate {
  instanceId: string;
  templateId: string;

  availability: number;
  decrease: number;

  activeRequests: number;
  effectiveAvailability: number;

  clientModel: string;
  upstreamModel: string;

  hardMaxConcurrency?: number;

  attributes: Record<string, unknown>;
}

export interface SchedulerContext {
  request: SchedulerRequest;
  groupConfig: Readonly<Record<string, unknown>>;
  groupState: Readonly<Record<string, unknown>>;

  patchGroupState(
    patch: Record<string, unknown>
  ): Promise<void>;

  clock: { now(): number };
}

export type SchedulerResult =
  | { selectedInstanceId: string }
  | { selectedInstanceId: null; reason: string };

export declare function select(
  ctx: SchedulerContext,
  candidates: ReadonlyArray<Candidate>
): Promise<SchedulerResult>;
```

## migrations/001_init.sql

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE credential_templates (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  protocols_json TEXT NOT NULL,
  official_defaults_json TEXT NOT NULL DEFAULT '{}',
  custom_fields_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',
  js_source TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE credential_instances (
  id TEXT PRIMARY KEY,
  template_id TEXT NOT NULL REFERENCES credential_templates(id)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  name TEXT,

  address TEXT,
  base_url TEXT,

  key_ciphertext BLOB,
  key_nonce BLOB,
  key_kid TEXT,

  models_json TEXT,
  availability REAL NOT NULL DEFAULT 100,
  proxy TEXT,

  custom_fields_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',

  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,

  CHECK(availability = availability)
);

CREATE INDEX idx_credential_instances_template
  ON credential_instances(template_id);

CREATE INDEX idx_credential_instances_availability
  ON credential_instances(availability DESC);

CREATE TABLE credential_groups (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  template_ids_json TEXT NOT NULL,
  scheduler_type TEXT NOT NULL,
  scheduler_code TEXT,
  config_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE access_keys (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  secret_hash BLOB NOT NULL UNIQUE,
  secret_prefix TEXT,
  group_id TEXT NOT NULL REFERENCES credential_groups(id)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  created_at TEXT NOT NULL,
  revoked_at TEXT
);

CREATE INDEX idx_access_keys_group ON access_keys(group_id);

CREATE TABLE credential_sources (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  template_id TEXT REFERENCES credential_templates(id)
    ON UPDATE CASCADE ON DELETE SET NULL,
  config_json TEXT NOT NULL DEFAULT '{}',
  code TEXT,
  state_json TEXT NOT NULL DEFAULT '{}',
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE secret_items (
  id TEXT PRIMARY KEY,
  owner_type TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  name TEXT NOT NULL,
  ciphertext BLOB NOT NULL,
  nonce BLOB NOT NULL,
  key_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(owner_type, owner_id, name)
);

CREATE TABLE audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor TEXT NOT NULL,
  action TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT,
  redacted_diff_json TEXT,
  created_at TEXT NOT NULL
);

INSERT OR IGNORE INTO schema_migrations(version, applied_at)
VALUES (1, datetime('now'));
```

