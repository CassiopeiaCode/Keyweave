# 26 — 状态机

## 1. Credential Instance 状态

Core 不增加持久化 `status` 官方字段；状态由可调度结果推导。

逻辑状态：

```text
                   +------------------+
                   |    SCHEDULABLE    |
                   +------------------+
                    |        |       |
       supports=no  |        |       | hard concurrency full
                    v        |       v
              UNSUPPORTED    |    BUSY_NOW
                             |
              template disables / cooldown
                             v
                       UNAVAILABLE
```

`UNAVAILABLE` 可由以下任意条件导致：

- Template `supports()` 返回 false；
- Template `getScheduleHints()` 标明不可调度；
- Template 以 custom field 表达 cooldown；
- instance/template 已禁用（如实现 disabled custom convention）；
- 模型不匹配；
- 协议不匹配。

Core 不能根据 Provider 名称推断状态。

## 2. Request 状态

```text
RECEIVED
 -> AUTHENTICATED
 -> CANDIDATES_READY
 -> LEASED
 -> UPSTREAM_PENDING
 -> HEADERS_RECEIVED
 -> STREAMING? -----------+
 -> COMPLETED             |
                          |
任何阶段 ----------------> FAILED
客户端取消 --------------> CANCELLED
                          |
                          v
                       CLEANUP
                          |
                          v
                        DONE
```

`CLEANUP` 必须幂等。

## 3. Attempt 状态

```text
CREATED
 -> LEASED
 -> EXECUTING
 -> RESPONSE
 -> SUCCESS

EXECUTING/RESPONSE
 -> RETRYABLE_FAILURE
 -> RELEASED
 -> next attempt

EXECUTING/RESPONSE
 -> TERMINAL_FAILURE
 -> RELEASED
```

一个 Request 可有多个 Attempt，但同一时刻默认只有一个 active Attempt。

## 4. Lease 状态

```text
NONE -> ACQUIRED -> RELEASED
```

规则：

- 不允许 `ACQUIRED -> ACQUIRED`；
- `release()` 必须幂等；
- debug build 可对 double release 发出告警；
- lease 不持久化到 SQLite；
- process restart 后全部 lease 视为不存在。

## 5. Template Version 状态

```text
DRAFT/UNVALIDATED
 -> VALIDATED
 -> COMPILED
 -> ACTIVE
 -> SUPERSEDED
```

如果 compile 失败，旧 ACTIVE 版本继续服务。

## 6. OAuth Token 生命周期（模板级）

OAuth 只是 Template Library 行为，推荐状态：

```text
VALID
 -> NEAR_EXPIRY
 -> REFRESHING
 -> VALID

REFRESHING -> REFRESH_FAILED -> optional cooldown/unavailable
```

为避免并发 refresh 风暴，应使用 template/instance atomic state 或 singleflight：

```text
refresh lock key = instanceId + provider-defined token slot
```

## 7. Source 状态

```text
IDLE
 -> FETCHING
 -> CLASSIFYING
 -> APPLYING
 -> IDLE

any -> ERROR -> IDLE(next run)
```

Source 运行失败不能影响 Gateway 请求面可用性。
