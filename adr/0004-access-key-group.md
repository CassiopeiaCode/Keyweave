# ADR-0004 — 一个 Client API Key 映射一个 Credential Group

Status: Accepted

## Decision

MVP 中 Access Key 必须精确映射到一个 Group。Group 再定义多个 Template 的 candidate scope 和 Scheduler。

## Consequences

认证与路由入口统一；复杂组合通过 Group 实现，而不是让 Key 自己变成脚本对象。
