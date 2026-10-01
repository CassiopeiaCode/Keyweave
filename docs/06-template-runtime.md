# 06 — Credential Template Runtime

## 1. Template 是可执行 Provider Adapter

Template 是受限 JavaScript Module。

推荐代码形式：

```js
export const manifest = { ... };

export async function getAvailability(ctx) { ... }
export async function getDecrease(ctx) { ... }
export async function getScheduleHints(ctx) { ... }
export async function supports(ctx) { ... }
export async function beforeRequest(ctx) { ... }
export async function execute(ctx) { ... }
export async function afterResponse(ctx, result) { ... }
export async function onError(ctx, err) { ... }
export async function onFinish(ctx) { ... }
```

全部 hook 均 optional。

## 2. 运行模式

Template Version 保存 JS source。

运行时：

```text
(templateId, version)
  -> compile
  -> cached immutable program
```

Request 开始时固定 version。

## 3. Context 分层

```ts
ctx.request
ctx.route
ctx.fields
ctx.secrets
ctx.instance
ctx.template
ctx.runtime
ctx.http
ctx.log
ctx.clock
```

职责：

- `request`：请求元数据 + 受控 body accessor
- `route`：group/model/protocol/session
- `fields`：effective fields，read-only
- `secrets`：secret accessor
- `instance`：受控 state/field patch
- `template`：受控 template state patch
- `runtime`：request/lease runtime facts
- `http`：唯一网络出口
- `log`：redacted logger
- `clock`：时间 API

## 4. 请求对象

不要简单把整个原始 body 永久复制成 JS object，防止大上传导致双倍内存。

建议：

```ts
ctx.request.json()
ctx.request.text()
ctx.request.bodyStream()
ctx.request.headers
ctx.request.method
ctx.request.path
ctx.request.query
```

Host 可缓存一次解析结果。

## 5. Secret

```js
const key = await ctx.secrets.get("key");
```

返回 SecretString wrapper。

Logger 对 SecretString 自动 redaction。

真正 secret 不出现在 `ctx.fields`。

## 6. 默认 execute

如果 Template 没实现 `execute`，使用 default passthrough：

```text
resolve URL
-> clone safe downstream headers
-> strip client auth
-> inject upstream auth
-> rewrite model if configured
-> ctx.http.request()
-> return response
```

普通 OpenAI-compatible 模板可完全不自己写 execute。

## 7. HTTP Host API

推荐：

```ts
interface HostHttp {
  request(input: {
    url: string;
    method?: string;
    headers?: Record<string,string>;
    body?: Uint8Array | string | HostReadableStream;
    timeoutMs?: number;
    proxyMode?: "instance" | "direct";
  }): Promise<HostResponse>;
}
```

默认 `proxyMode="instance"`。

是否允许 Template 请求 `direct` 必须由 manifest capability 明确授权。

## 8. Instance Mutation

禁止给模板直接 DB handle。

提供 CAS 风格：

```js
await ctx.instance.patch({
  availability: 50,
  customFields: {
    cooldownUntil: Date.now() + 60000
  }
});
```

Host 内部使用 optimistic version。

## 9. Atomic Helpers

建议：

```js
await ctx.instance.atomic.increment("customFields.failCount", 1);
await ctx.instance.atomic.max("customFields.lastSeenAt", ctx.clock.now());
```

避免并发 hook 写丢状态。

## 10. Runtime Limits

每次 hook 至少需要：

- wall clock timeout
- CPU/instruction budget（运行时支持时）
- memory ceiling（worker 级更可靠）
- Host call count limit
- log bytes limit
- response rewrite size limit

## 11. Go + JS Runtime 建议

如果使用 Go：

- Goja 易嵌入，但异步 Host API、硬内存隔离需要额外工程；
- QuickJS 更接近完整 JS runtime，但 Go binding 和 worker 生命周期需谨慎；
- 最安全的长期方案是 `Gateway Core <RPC> Template Worker Process`；
- MVP 可以先用可信模板 + Goja/QuickJS，但 ABI 必须保留未来 worker 隔离空间。

## 12. 生命周期顺序

```text
load version
initialize context
supports
getAvailability
getDecrease
getScheduleHints
[select + lease]
beforeRequest
execute
afterResponse OR onError
onFinish
release lease
```

`onFinish` 无论成功失败都执行，但其失败不能覆盖主请求结果，只记录 secondary error。

## 13. Startup Hooks

```js
export async function onSystemStart(ctx) {}
export async function onInstanceStart(ctx) {}
```

用途：

- 清理模板自己持久化的旧并发标记
- token 预热
- metadata migration
- state sanity check

Startup hook 必须有超时和错误隔离。
