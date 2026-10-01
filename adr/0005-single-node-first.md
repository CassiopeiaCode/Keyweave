# ADR-0005 — 单机优先

Status: Accepted

## Decision

v0.x 以单进程 + SQLite WAL + in-memory lease 为正式架构，不提前引入分布式锁或一致性协议。

## Evolution

未来多节点时需要重新设计：distributed lease、scheduler state、affinity state、template version distribution；不得假装当前 in-memory lease 能透明扩展到多节点。
