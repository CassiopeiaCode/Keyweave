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
