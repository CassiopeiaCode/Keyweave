# 31 — 性能与容量设计

## 1. 目标环境

首要目标是单用户、单机、几十到数百 Credential，而不是多租户 hyperscale Gateway。

参考设计容量而非硬承诺：

```text
Credential Templates:     1 - 200
Credential Instances:     1 - 10,000
Credential Groups:        1 - 1,000
Concurrent requests:      1 - 1,000
Typical personal deploy:  < 100 concurrent
```

## 2. Candidate 索引

维护：

```text
templateId -> []instanceId
instanceId -> immutable config snapshot
```

Group 展开不应每次扫描全部实例。

## 3. Template 编译缓存

Key：

```text
templateId + version
```

缓存 compiled artifact。一次请求 pin 版本。

## 4. Transport Pool

HTTP client/transport 需要按代理 identity 复用：

```text
transportKey = proxy scheme + host + user identity + TLS policy
```

不能每个请求新建 TCP Transport，否则连接复用丢失。

同时不能把含不同 proxy credential 的 transport 错误复用。

## 5. Lock 范围

锁内仅做：

- 读取 active counts；
- hard limit validation；
- scheduler 的短纯选择；
- increment lease；
- RR cursor/state 的小型更新。

禁止锁内：

- SQLite I/O；
- Template JS；
- HTTP；
- JSON 大对象序列化；
- OAuth refresh。

## 6. Template 状态写

不要每请求都持久化 RR 位置或 activeRequests。

只有业务上真正需要跨重启保留的 Template state 才写 DB。

## 7. `/v1/models` 缓存

可按：

```text
groupId + protocol + configRevision
```

缓存 model list。任何相关 Template/Instance/Group 更新提高 `configRevision`。

## 8. Streaming 内存

使用固定大小 copy buffer；不把 SSE 聚合成完整字符串。

Template stream transform 若存在，必须采用 iterator/transform stream 模式。

## 9. 日志采样

Debug body logging 默认关闭。即使开启，也要：

- 截断大小；
- redact secrets；
- multimodal/binary 不直接记录。
