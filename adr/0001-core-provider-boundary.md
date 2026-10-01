# ADR-0001 — Core 与 Provider 行为边界

Status: Accepted

## Decision

Core 不出现 Provider 名称分支。Provider-specific authentication、URL、model mapping、OAuth refresh、quota/cooldown、request/response rewrite 都进入 Credential Template 或模板库。

## Consequences

优点：新增 Provider 不需要改 Core；CLIProxyAPI 类能力可通过模板表达。

代价：必须认真设计 Sandbox ABI、Template debugger 和版本管理。
