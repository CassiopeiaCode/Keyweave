# 19 — 部署

## 1. 单机默认

```text
ptpg
  --listen 127.0.0.1:8317
  --db ./data/ptpg.db
  --templates ./templates
```

Admin 默认：

```text
127.0.0.1 only
```

Inference 可按用户配置监听 LAN。

## 2. 数据目录

```text
data/
  gateway.db
  gateway.db-wal
  gateway.db-shm
  backups/
  runtime/
```

Master encryption key 不放在此目录明文。

## 3. Shutdown

收到 SIGTERM：

1. stop accepting new requests；
2. wait grace period；
3. cancel remaining；
4. release/clear in-memory leases；
5. flush observability；
6. close DB。

## 4. Upgrade

- backup DB；
- run migration；
- validate templates compile；
- start readiness；
- only then expose traffic。

## 5. Docker

- data volume；
- secret/master key mount；
- admin port not published by default；
- template worker 无额外 Linux capabilities。

## 6. 配置层级

推荐：

```text
compiled defaults
< config file
< environment
< CLI flags
```

但凭据 Secret 不建议通过大而杂的 config 文件长期维护；应进入 Secret Store。

## 7. 单实例限制

v0.x 不保证两个 Gateway 进程同时打开同一个 DB 并共享 Lease 语义。

如果用户启动多进程，应显式报 warning 或阻止，直到实现分布式/DB-backed lease。
