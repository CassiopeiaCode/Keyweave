# 14 — 安全与 JavaScript Sandbox

## 1. 信任模型

即使系统定位个人使用，Template 代码仍可能：

- 来自模板库；
- 被 Agent 生成；
- 从网络复制；
- 写错形成死循环；
- 意外泄露 Token。

因此不应直接使用宿主原生 `eval()`。

## 2. 禁止能力

默认 Sandbox 不提供：

```text
process
require
fs
os
net
tls
child_process
Deno
Bun
native fetch
native WebSocket
environment variables
```

## 3. 允许能力

仅通过 Host Capability：

```text
ctx.http
ctx.secrets
ctx.instance
ctx.template
ctx.log
ctx.clock
```

Capability 按 template manifest 授权。

## 4. 网络能力

普通 Template：

```text
network = credential-bound
```

即只能通过绑定 instance proxy 的 Host HTTP。

特殊系统 Template 可以声明 direct network，但管理员安装/启用时必须显式授权。

## 5. Secret Redaction

SecretString 应带 taint。

任何：

```text
ctx.log.info(secret)
```

必须输出 `[REDACTED]`。

仅依赖 wrapper 不够：字符串拼接可能丢 taint。日志层还应维护当前 request 的 known-secret set 做精确/前后缀 redaction。

## 6. Template Supply Chain

Template Library 条目包含：

```text
id
version
source hash
signature(optional)
author
capabilities
```

升级前显示 capability diff。

## 7. 资源隔离

长期最稳妥：

```text
Gateway Core
  <RPC>
Template Worker Process
```

Worker：

- memory limit
- CPU time
- no filesystem
- restricted syscalls
- no direct network

Host HTTP 请求经 RPC 回 Core。

## 8. 管理 API

- 默认 localhost；
- browser UI 做 CSRF 防护；
- management key 与 inference key 分离；
- response 不返回 secrets；
- import bundle 不能自动执行未知 Template code。

## 9. 数据加密

Key、OAuth token、proxy password：

- AES-256-GCM；
- random nonce；
- key id；
- master key outside DB。

## 10. SSRF

Host HTTP：

- scheme allowlist；
- optional private CIDR deny；
- metadata IP deny；
- DNS resolution validation；
- redirect target 重新校验。

## 11. Template 安装审批

新 Template 如果 capability 比旧版增加：

```text
directNetwork
mutateTemplate
backgroundTask
```

必须阻止静默自动升级。
