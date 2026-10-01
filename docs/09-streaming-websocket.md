# 09 — Streaming 与 WebSocket

## 1. SSE

SSE 不能先完整读取再返回：

```text
upstream Body
 -> optional event transform
 -> downstream writer
```

必须具备：

- flush each event/chunk
- backpressure
- cancel propagation
- error accounting
- lease held until EOF

## 2. Bootstrap Retry

如果上游在**任何下游字节发送前**失败，可以安全 retry。

一旦下游已经发送响应内容：

```text
默认禁止透明切换 credential
```

因此 runtime 必须维护：

```text
downstreamStarted bool
```

Retry policy 必须接收该状态。

## 3. Streaming Keepalive

可选支持 SSE comment ping，但只能由 Protocol Adapter 明确支持时启用，不能向任意字节流注入数据。

## 4. Non-stream Keepalive

某些长耗时非流式接口如果需要 whitespace keepalive，应作为特定协议行为，不进入通用 Core。

## 5. Stream Transform

Template 可以请求 SSE 事件级 transform：

```js
return ctx.http.passthrough({
  transformSSE(event) { ... }
});
```

Host 应逐 event 调用；禁止让 JS 持有无限长度 buffer。

## 6. WebSocket

长期兼容目标：

```text
downstream WS
 <-> Gateway
 <-> upstream WS
```

Credential Lease 从 upstream dial 前建立，到任意一端 close 后释放。

## 7. WebSocket Proxy

HTTP proxy/CONNECT 情况下，WS dialer 同样必须绑定 instance proxy。

## 8. WebSocket Template API

未来推荐：

```ts
ctx.http.websocket(url, options): Promise<HostWebSocket>
```

Template 不直接拿 OS socket。

## 9. Retry 规则

WebSocket handshake 失败且下游尚未 upgrade：

- 可以 retry other credential。

已经 upgrade：

- 不透明 retry；
- 正常 close/error 传播。

## 10. Backpressure

实现必须避免：

```text
upstream read fast
-> unbounded in-memory queue
-> slow downstream
```

默认应让读写链路自然 backpressure，或设有严格 bounded buffer。
