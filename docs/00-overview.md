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
