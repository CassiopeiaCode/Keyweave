# 01 — 术语与核心不变量

## 1. 术语

### Credential Template

一种“可执行 Schema”，包含 metadata、protocol capabilities、default fields、custom fields、persistent state、JS source 和 hooks。

### Credential Instance

某个具体账号/Token/API Key/OAuth credential。Instance 必须属于且只属于一个 Template。

### Credential Group

Client API Key 的路由池。Group 通过 `templateIds[]` 定义候选范围，并拥有 Scheduler。

### Access Key

客户端调用 Gateway 时使用的 Key。其作用同时是认证与路由入口。

### Availability

Credential Instance 的持久化调度分值。数值越高，默认越优先。

### Decrease

当前实例每持有一个 active lease 时，对其 `effectiveAvailability` 施加的临时惩罚。

### Lease

一次 attempt 对某个 Credential Instance 的占用。

### Protocol Adapter

入站 API 的薄适配层，只提取路由必须的信息，并保留 Raw Request。

### Host HTTP

Template 唯一允许使用的网络出口。

## 2. 强制不变量

- **INV-001** 一个 Credential Instance 只能属于一个 Credential Template。
- **INV-002** 一个 Credential Template 可以拥有 0..N 个 Credential Instance。
- **INV-003** 一个 Credential Instance 可以暴露 0..N 个模型。
- **INV-004** 一个 Credential Template 可以声明 1..N 个协议能力。
- **INV-005** 一个有效 Access Key 唯一映射一个 Credential Group。
- **INV-006** Credential Group 的静态候选边界由 `templateIds[]` 决定。
- **INV-007** Custom Scheduler 只能返回 candidates 中已有的 `instanceId`。
- **INV-008** Quota 不是 Core 官方字段。
- **INV-009** 一次请求期间的并发惩罚不得通过 Core 自动写回 Instance `availability`。
- **INV-010** 模板的所有网络访问默认必须经过 `ctx.http`。
- **INV-011** `ctx.http` 默认强制绑定当前 Credential Instance 的 proxy policy。
- **INV-012** 下游 Client Key 默认必须被移除并替换为上游 credential。
- **INV-013** Streaming / WebSocket attempt 的 Lease 必须持续到连接结束。
- **INV-014** Secret 永远通过 secret accessor 暴露给模板，而不是普通 JSON field。
- **INV-015** 请求开始后绑定 Template Version；热更新不得改变正在执行请求的代码。
- **INV-016** Retry 的每次上游调用都视为新的 Attempt，有自己的 Lease 与 trace。
- **INV-017** Scheduler 锁内不能执行网络 I/O、Template JS、数据库慢查询。
- **INV-018** Group Scheduler 不理解 Provider 业务语义。

## 3. Availability 常量

持久化不要依赖 IEEE `-Infinity`。建议：

```text
AvailabilityDefault = 100
AvailabilityDisabledFloor = -9_000_000_000_000_000
```

Template `canSchedule=false` 优先于极低 availability；availability 主要用于排序。
