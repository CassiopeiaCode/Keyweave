# ADR-0002 — Availability、Decrease 与 Lease 分离

Status: Accepted

## Decision

`availability` 是持久化、模板控制的排序信号；`decrease` 是活动请求造成的临时调度惩罚；并发占用由 in-memory Lease Manager 记录。

```text
effective = availability - activeRequests * decrease
```

Core 不为轮询效果每请求修改数据库 availability。

## Rationale

避免高频 DB 写、崩溃残留、并发竞争与“健康度/轮询计数”语义混杂。
