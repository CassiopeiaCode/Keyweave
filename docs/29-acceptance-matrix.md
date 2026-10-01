# 29 — 验收矩阵

此文件是 Coding Agent 的最终验收表。MVP 标记 MUST 的项目全部通过才算完成。

| ID | 场景 | 级别 | 验收条件 |
|---|---|---|---|
| A001 | Client Key -> Group | MUST | 一个有效 Key 精确解析到一个 Group |
| A002 | 无效 Key | MUST | 不泄露 Group/credential 信息；401 |
| A003 | revoked Key | MUST | 401，不进入 candidate resolution |
| A004 | Group scope | MUST | Scheduler 永远不能选 `templateIds[]` 外实例 |
| A005 | 多模型实例 | MUST | 一个实例可匹配多个 client model |
| A006 | model mapping | MUST | client model 可映射到不同 upstream model |
| A007 | 多协议模板 | MUST | 一个 Template 可声明多个 protocol |
| A008 | availability | MUST | 高 effective availability 优先 |
| A009 | RR tie break | MUST | 同分候选公平轮询，不固定首项 |
| A010 | decrease | MUST | active lease 导致临时 score 下降，不写 DB |
| A011 | hard concurrency | MUST | 100 并发下 max=1 不发生双 lease |
| A012 | lease cancel | MUST | 客户端取消后 active count 回到 0 |
| A013 | panic cleanup | MUST | runtime panic 后 lease 释放 |
| A014 | generic passthrough | MUST | 客户端 key 被替换为 instance secret |
| A015 | header sanitize | MUST | client Authorization 不会误透传 |
| A016 | proxy HTTP | MUST | upstream 请求实际经过 HTTP proxy |
| A017 | proxy auth | MUST | Basic authenticated HTTP proxy 可用 |
| A018 | HTTPS CONNECT | MUST | HTTPS upstream 经 HTTP CONNECT proxy |
| A019 | template network isolation | MUST | template 无原生 socket/fetch/file access |
| A020 | SSE streaming | MUST | chunk 低延迟透传，不缓冲整个响应 |
| A021 | stream lease lifetime | MUST | EOF/cancel/error 前 lease 不释放 |
| A022 | retry before downstream | MUST | 未开始下游时可换 credential 重试 |
| A023 | no retry after byte | MUST | 下游已发送字节后默认不换凭据重试 |
| A024 | 429 template policy | MUST | Core 不写 Provider-specific quota 逻辑 |
| A025 | session affinity | SHOULD | 同 session 在 credential 可用时稳定绑定 |
| A026 | affinity failover | SHOULD | 绑定实例不可用时可重新选择/绑定 |
| A027 | template update pinning | MUST | 在途请求继续使用开始时版本 |
| A028 | template compile failure | MUST | 新版本失败不破坏旧 active 版本 |
| A029 | secret at rest | MUST | SQLite 中无明文 upstream key |
| A030 | secret logs | MUST | 日志/trace/debugger 无完整 secret |
| A031 | `/v1/models` scoped | MUST | 仅返回当前 Client Key 可达模型 |
| A032 | source transactionality | SHOULD | 分类失败不产生半实例 |
| A033 | optimistic admin update | SHOULD | stale version 更新返回 conflict |
| A034 | websocket lease | PHASE2 | WS close 前 lease 保持 |
| A035 | SOCKS5 | PHASE2 | instance proxy 支持 socks5 |
| A036 | OAuth refresh singleflight | PHASE2 | 并发请求不会同时刷新同 token |

## 1. 强制测试命令

参考 Go 实现最终应支持：

```bash
go test -race ./...
```

并建立集成测试：

```bash
./scripts/integration-test.sh
```

## 2. 性能底线

个人部署 MVP 不是高性能竞赛，但调度热路径应满足：

- 不做每次请求全表扫描；
- 不在 group lock 内执行网络 I/O；
- 不在 group lock 内执行任意长 JS；
- streaming 不复制完整 body；
- active lease 查询为 O(1)；
- candidate expand 可通过 template_id index 完成。

## 3. 安全验收

至少主动测试：

- malicious template infinite loop；
- template 试图读取 env/file；
- SSRF 到 loopback/link-local（按部署 policy）；
- proxy credential redaction；
- management endpoint 未授权访问；
- secret interpolation 后错误日志泄露；
- custom scheduler 返回任意 scope 外 ID。
