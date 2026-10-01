# 13 — Retry、Cooldown 与失败处理

## 1. 核心原则

Core 提供 retry engine，但不把 Provider-specific 状态码语义写死。

Template：

```js
onError(ctx, error)
```

返回标准 action。

## 2. ErrorAction

```ts
type ErrorAction =
  | { action: "stop" }
  | {
      action: "retry";
      target: "same" | "another";
      cooldownMs?: number;
    };
```

持久化 availability/cooldown 状态仍由 Template 控制。

## 3. 默认 Library Policy

官方 Generic HTTP Template 可以预置：

- 408 -> retry another
- 429 -> retry another + optional cooldown
- 500/502/503/504 -> retry another
- 401/403 -> 默认 stop + mark unhealthy（可配置）

这只是 Template Library 默认，不是 Core 硬规则。

## 4. Retry State

```ts
interface RetryState {
  requestId: string;
  round: number;
  attempt: number;
  attemptedInstanceIds: string[];
  downstreamStarted: boolean;
}
```

## 5. Cooldown

Template custom field：

```json
{"cooldownUntil": 1790000000000}
```

Template：

```js
getAvailability(ctx) {
  if (ctx.fields.cooldownUntil > ctx.clock.now()) return -1e12;
  return ctx.instance.availability;
}
```

或：

```js
supports(ctx) {
  return ctx.fields.cooldownUntil <= ctx.clock.now();
}
```

Core 不需要理解 cooldown 字段。

## 6. Request Scoped Errors

Template config 可描述：

```json
{
  "requestScopedErrors": [
    {
      "status": 400,
      "contains": ["context_length_exceeded"],
      "action": "stop"
    }
  ]
}
```

避免错误地认为所有 400 都说明 credential 坏了。

## 7. Retry 安全

- 已经向下游发送 stream 字节后，不自动跨 credential retry；
- 非幂等 endpoint 必须由 Template 显式标记 retry-safe；
- cancellation 不 retry；
- deadline exceeded 通常不 retry；
- max attempt/round 必须有硬上限；
- retry 不能造成无限 credential cycling。

## 8. Retry Candidate Filter

`target=another` 时默认排除：

```text
attemptedInstanceIds in current retry round
```

下一 round 是否允许重新尝试，由 retry policy 决定。

## 9. Upstream Error 标准化

Transport 统一区分：

- DNS
- connect timeout
- TLS
- proxy auth
- response header timeout
- HTTP status
- stream reset
- downstream cancellation

Template `onError` 获得结构化错误，而非只能解析字符串。
