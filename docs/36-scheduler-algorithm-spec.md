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
