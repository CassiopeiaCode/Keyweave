# 27 — 进程配置与运行参数

本章只定义 Core 运行配置；Provider 行为不要放入全局配置。

## 1. 推荐配置文件

```yaml
server:
  listen: "127.0.0.1:8317"
  publicBaseUrl: "http://127.0.0.1:8317"
  shutdownGrace: "30s"

database:
  path: "./data/token-pool.db"
  wal: true
  busyTimeout: "5s"

runtime:
  templateEngine: "goja"
  templateWallTimeout: "5s"
  templateCpuBudget: "250ms"
  maxTemplateMemoryBytes: 33554432
  maxLogBytesPerRequest: 65536

routing:
  schedulingShards: 64
  defaultScheduler: "availability-round-robin"
  maxAttempts: 3

transport:
  dialTimeout: "10s"
  responseHeaderTimeout: "60s"
  idleConnTimeout: "90s"
  maxIdleConns: 256
  maxIdleConnsPerHost: 32
  allowDirectNetworkFromTemplates: false

secrets:
  masterKeyProvider: "file-or-keychain"
  masterKeyId: "local-v1"

admin:
  enabled: true
  listen: "127.0.0.1:8318"
  remoteAccess: false

observability:
  level: "info"
  metrics: true
  requestLog: true
```

## 2. 环境变量

环境变量只用于部署级 Secret/路径：

```text
PTPG_CONFIG
PTPG_DB_PATH
PTPG_MASTER_KEY
PTPG_ADMIN_TOKEN
PTPG_LOG_LEVEL
```

Provider credential 不建议通过环境变量批量注入，因为会绕过统一 Secret Store 与管理面。

## 3. 启动顺序

```text
parse config
 -> validate
 -> initialize master key
 -> open SQLite
 -> migration
 -> repositories
 -> load/compile templates
 -> run template onSystemStart/onInstanceStart
 -> initialize lease manager/schedulers
 -> start admin listener
 -> start gateway listener
```

如果任一 migration 或 master-key 初始化失败，Gateway 不得进入 ready 状态。

单个非关键 Template compile 失败时可进入 degraded 状态，但对应实例不得进入候选。

## 4. Readiness

`/health/live`：进程存活。

`/health/ready` 至少检查：

- DB writable/readable；
- master key loaded；
- gateway router initialized；
- migration current。

不要把所有上游 Provider 可达性作为 readiness 条件，否则一个 Provider 故障会让整个网关被重启。

## 5. 优雅退出

收到 SIGTERM：

1. readiness=false；
2. 停止接收新请求；
3. 等待 active request 至 grace deadline；
4. cancel remaining request；
5. close transports；
6. flush logs/metrics；
7. close SQLite。

## 6. 热加载

MVP 可对以下对象支持管理 API 热加载：

- Template source/version；
- Credential Instance fields；
- Group scheduler/config；
- Access Key revoke/create。

进程级配置例如 listen address、master key provider 可以要求 restart。
