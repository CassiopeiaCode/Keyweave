# 30 — 故障、恢复与一致性

## 1. SQLite Busy

启用 WAL + busy timeout。请求热路径应优先使用已加载的只读配置快照；不要因为一次普通请求而同步写 DB。

Template 如果在 `afterResponse` 修改状态失败：

- 响应已经成功时，默认不要把客户端成功响应改成失败；
- 记录 structured error；
- 对安全关键写入（例如 OAuth refresh token 更新）允许模板声明 `mustPersist=true` 并在发送响应前完成。

## 2. DB corruption / migration failure

进程启动失败并明确报错，不允许在未知 schema 上继续运行。

建议运维：

- 定期 SQLite online backup；
- migration 前自动备份；
- migration 不可逆时提供显式说明。

## 3. JS Runtime 崩溃/超时

单次 runtime 异常不得崩溃宿主进程。

映射为：

```text
template_compile_error
template_timeout
template_runtime_error
template_memory_limit
```

并确保 Lease release。

## 4. Proxy 故障

区分：

- proxy URL invalid；
- DNS failure；
- proxy auth failed；
- CONNECT rejected；
- upstream through proxy timeout。

这些错误可交给 Template `onError`，但 Core 不自动永久降低 availability。

## 5. Upstream 故障

Retry Controller 受以下共同限制：

```text
maxAttempts
request deadline
downstreamStarted
attemptedSet
template ErrorAction
```

不得无限重试。

## 6. 进程 crash

因为 Lease 是 in-memory：

- restart 后 activeRequests=0；
- 不需要回滚 lease 数据；
- Template 如果自行持久化 transient state，必须在 startup hook 清理或带 TTL。

## 7. Secret master key 丢失

如果无法解密已有 credentials：

- Gateway 不应 silently 将 Secret 当空字符串使用；
- 实例进入不可执行状态；
- Admin API 显示 `secret_unavailable`（不回显 ciphertext）；
- 支持用户重新导入/重写 Secret。

## 8. 版本冲突

Admin 修改 Instance 时使用 CAS：

```text
expected version=7
DB current version=8
=> 409 conflict
```

不要 last-write-wins 覆盖别人/其他自动化刚写入的状态。

Template 高频 runtime state 建议使用原子字段 API，而不是读-改-写整个 JSON。

## 9. Clock

过期时间、cooldown、TTL 使用 Host `ctx.clock`，便于测试。

持久化时间使用 UTC RFC3339 或 Unix milliseconds，并统一规定一种格式。

## 10. Backpressure

Streaming copy 必须尊重 downstream backpressure。客户端消费慢时，不应无限把 upstream 数据堆在内存。
