# ADR-0003 — Template 网络必须走 Host HTTP

Status: Accepted

## Decision

Template Sandbox 不提供 native fetch、raw socket、filesystem、process/env。所有网络调用通过 `ctx.http`，默认绑定当前 Credential Instance 的 proxy policy。

## Consequences

系统能够保证 per-credential proxy、认证代理、取消传播、审计、超时和 Secret redaction。
