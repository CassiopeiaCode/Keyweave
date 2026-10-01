# Programmable Token Pool / Credential Pool Gateway — Agent Implementation Pack v0.1

> 文档冻结日期：2026-10-01  
> 目标读者：负责直接实现系统的 Coding Agent / 工程师  
> 参考实现栈：Go + SQLite(WAL) + JavaScript Sandbox Runtime  
> 架构原则：**小 Core、可编程 Credential Template、可编程 Credential Group Scheduler**

## 0. 一句话定义

这是一个以“凭据”为核心的可编程 AI API Gateway。

客户端以标准协议（OpenAI / Anthropic 等）发起请求，并携带本系统的 Client API Key。Client API Key 唯一映射到一个 Credential Group。Group 展开一组 Credential Template 下的所有 Credential Instance，按照协议、模型和模板能力过滤，再由 Scheduler 选出一个实例。该实例对应的 Template 在隔离 JS Runtime 内执行，通过系统提供的、强制遵循该实例代理配置的 HTTP Client 调用上游。

```text
Client Request
  -> Client API Key
  -> Credential Group
  -> Template Scope
  -> Credential Instances
  -> Protocol/Model Filter
  -> Group Scheduler
  -> Atomic Lease
  -> Credential Template Runtime
  -> Proxy-bound Host HTTP
  -> Upstream
  -> Streaming / Non-streaming Response
```

## 1. 本包内容

- `MASTER_SPEC.md`：完整总规格，适合单文件交给 Agent。
- `docs/`：按模块拆分的工程设计。
- `contracts/`：TypeScript 风格 ABI、JSON Schema、OpenAPI 契约。
- `migrations/`：SQLite 初始 Schema。
- `examples/templates/`：Credential Template 示例。
- `examples/schedulers/`：默认调度和 Session Affinity 示例。
- `reference/go/`：参考 Go 接口与核心算法骨架。
- `test-vectors/`：实现必须通过的关键场景。
- `AGENT_IMPLEMENTATION_GUIDE.md`：实现顺序、禁止事项与 Definition of Done。
- `AGENT_PROMPT.md`：可直接复制给 Coding Agent 的任务入口。
- `adr/`：已经冻结的关键架构决策。
- `SOURCE_REFERENCES.md`：CLIProxyAPI 兼容性参考基线。
- `docs/29-acceptance-matrix.md`：MVP 最终验收矩阵。

## 2. 最重要的设计约束

1. **Credential Template 是 Provider 行为边界**。不要在 Core 中新增 `if provider == ...`。
2. **Credential Instance 是具体账号/Token/Key/OAuth 凭据**；一个实例可支持多个模型和多个接口协议。
3. **Client API Key = 路由入口**，一个 Key 只映射一个 Credential Group。
4. **Credential Group = 候选范围 + Scheduler**。
5. **Protocol / Model 是筛选维度，不是数据库父子层级。**
6. **Quota 不是官方字段**；由 Template custom fields + hooks 自己表达。
7. `availability` 是持久化的模板可控分值；`decrease` 是请求占用期间的临时调度惩罚。
8. Core 不因一次请求擅自写回 `availability`。
9. 所有模板网络访问必须经 `ctx.http`；`ctx.http` 强制遵循当前实例的 proxy。
10. Streaming / WebSocket 必须持有 Lease 到连接结束。
11. Custom Scheduler 只能从 Core 传入的 candidates 中选择，不能越权选择其他实例。
12. Secret 不能出现在普通日志、管理 API 明文返回或 Template Debugger 明文输出中。

## 3. 推荐阅读顺序

从 `docs/00-overview.md` 开始，按编号顺序阅读。实现时先看 `contracts/`，再看 `AGENT_IMPLEMENTATION_GUIDE.md`。

## 4. MVP 边界

MVP 必须实现：

- Credential Template / Instance / Group / Access Key
- SQLite WAL
- OpenAI Chat Completions
- OpenAI Responses
- Anthropic Messages
- `/v1/models`
- Streaming SSE + Non-streaming
- Availability / Decrease / Lease
- Availability-first Round Robin
- HTTP / HTTPS-over-CONNECT authenticated proxy
- JavaScript Template Runtime
- Generic OpenAI-compatible Template
- OpenRouter Template
- Admin API
- Template Debugger 基础版

MVP 可以暂缓：

- 多节点
- 分布式 Lease
- 插件 ABI
- Kubernetes
- 所有 OAuth Provider
- 复杂前端

## 5. CLIProxyAPI 兼容目标

目标不是复制 CLIProxyAPI 的内部实现，而是保证其**非插件核心 Provider / Credential / Routing 能力**能够被本系统的：

```text
Credential Template
+ Credential Instance
+ Credential Group Scheduler
+ Host Runtime
```

表达。兼容性细目见 `docs/16-cliproxyapi-compatibility.md`。


## 6. 交给 Agent 的最短方式

把整个目录交给 Agent，并直接附上 `AGENT_PROMPT.md`。如果 Agent 上下文有限，优先提供：

```text
AGENT_PROMPT.md
MASTER_SPEC.md
contracts/
migrations/
test-vectors/
reference/go/
```

## 7. 本包自检

压缩包生成前执行：

```bash
./scripts/verify.sh
```

它会验证 JSON/YAML/SQLite migration、Go reference tests、文档关键文件以及 SHA-256 manifest。
