# 可直接交给 Coding Agent 的任务提示

你要实现本目录定义的 Programmable Token Pool / Credential Pool Gateway。

先阅读顺序：

1. `README.md`
2. `docs/01-terminology-and-invariants.md`
3. `MASTER_SPEC.md`
4. `contracts/`
5. `AGENT_IMPLEMENTATION_GUIDE.md`
6. `test-vectors/`
7. `reference/go/`

实现时必须遵守：

- 不在 Core 写 Provider-specific 分支；
- Client API Key 精确映射一个 Credential Group；
- Group 的 `templateIds[]` 是 Scheduler 不可突破的候选边界；
- Credential Template 可以支持多个协议、多个模型；
- Quota 不是 Core 官方字段；
- persisted `availability` 与 runtime `decrease/lease` 分开；
- select + lease 必须原子；
- template 网络只能经 Host `ctx.http`，并遵循 credential proxy policy；
- streaming lease 持有到 EOF/cancel/error；
- 已开始向客户端发送响应后，不得默认换 credential 重试；
- Secret 不明文入库、日志或管理 API；
- custom scheduler 返回 scope 外 instance 必须拒绝。

参考技术栈是 Go + SQLite WAL + JS sandbox。可以替换具体 JS 引擎，但 ABI 不得随意改变。

请按 `docs/20-implementation-plan.md` 和 `AGENT_IMPLEMENTATION_GUIDE.md` 分 Milestone 实现。每个 Milestone 先补测试，再实现；最终至少通过 `docs/29-acceptance-matrix.md` 的所有 MUST 项。

任何需要改变 Core invariant 的决定，先新增 ADR，再修改契约与文档；不要只在代码中偷偷改变语义。
