# 07 — Protocol Gateway

## 1. 原则：薄 Adapter + Raw Passthrough

不要把所有 Provider 请求统一转换成一个“超级内部消息格式”。

这样会持续丢失：

- 新字段
- tool payload
- multimodal
- reasoning metadata
- provider extension
- future protocol fields

正确方式：

```text
Raw Request
+ normalized routing metadata
```

## 2. Normalized RequestContext

```ts
interface RequestContext {
  requestId: string;
  protocol: string;
  model: string;
  stream: boolean;

  clientKeyId: string;
  groupId: string;

  method: string;
  path: string;
  query: string;

  headers: Record<string,string[]>;
  session?: SessionMetadata;
}
```

## 3. MVP 协议

### OpenAI Chat

```text
POST /v1/chat/completions
protocol = openai.chat
model = body.model
stream = body.stream === true
```

### OpenAI Responses

```text
POST /v1/responses
protocol = openai.responses
model = body.model
stream = body.stream === true
```

### Anthropic Messages

```text
POST /v1/messages
protocol = anthropic.messages
model = body.model
stream = body.stream === true
```

### Model Listing

```text
GET /v1/models
```

由当前 Group 动态聚合。

## 4. 入站认证 Header

OpenAI 默认：

```text
Authorization: Bearer <CLIENT_KEY>
```

Anthropic：

```text
x-api-key: <CLIENT_KEY>
```

允许配置兼容方式，但路由层最终只收到归一化 AccessKey ID。

## 5. Session Metadata

Adapter 尽量提取显式 session：

```ts
interface SessionMetadata {
  id?: string;
  conversationId?: string;
  promptCacheKey?: string;
  parentSessionId?: string;
  clientRequestId?: string;
  derivedHash?: string;
}
```

Session Scheduler 决定绑定逻辑。

## 6. Header Policy

默认 passthrough：

1. 删除 hop-by-hop headers；
2. 删除 client authentication；
3. 删除 gateway-only headers；
4. 保留安全业务 header；
5. Template 再注入上游认证及自定义 header。

必须正确处理：

```text
Connection
Proxy-Connection
Keep-Alive
Transfer-Encoding
Upgrade
TE
Trailer
```

WebSocket 例外由专用 transport 处理。

## 7. Body Rewrite

普通 passthrough 如果需要 model alias：

- JSON body 只改 `model`；
- 其他字段原样；
- 不做 schema normalize。

如果 body 非法 JSON，返回 protocol-specific 4xx，不进入 Scheduler。

## 8. Response Rewrite

默认不改。

仅在 `forceResponseMapping=true` 或 Template hook 要求时，对已知 response model 字段做轻量 rewrite。

Streaming 中需要逐 SSE event 处理，不能 buffer 整个流。

## 9. 请求大小

Adapter 必须设置：

- JSON body size limit
- header count/size limit
- decompression size limit

避免在进入 Template Runtime 前就被大 body 打爆。
