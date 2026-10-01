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
