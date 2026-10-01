# 35 — 入站协议实现细节

## 1. 原则

协议 Adapter 的职责是提取路由元数据并保持 Raw Passthrough，不要把所有 Provider 强制转成巨大统一 AST。

统一提取：

```text
protocol id
requested model
stream flag
session metadata
client auth key
request method/path/query/headers/body
```

## 2. OpenAI Chat Completions

Endpoint：

```text
POST /v1/chat/completions
```

提取：

```text
body.model
body.stream
```

默认透传所有未知字段，包括未来新增参数、tools、多模态内容。

## 3. OpenAI Responses

Endpoint：

```text
POST /v1/responses
```

提取：

```text
body.model
body.stream
body.conversation / prompt_cache_key（如存在，用于 session metadata）
```

不要为了路由而删掉 Responses 特有字段。

## 4. Anthropic Messages

Endpoint：

```text
POST /v1/messages
```

提取 `model`、`stream`。Authentication adapter 接受项目定义的 Client API Key 方式；转上游时由 Template 重建 Provider auth header。

## 5. Auth header

Inbound 支持可配置：

```text
Authorization: Bearer <client-key>
x-api-key: <client-key>
```

如果两个同时出现且不一致，应拒绝请求，避免模糊认证。

## 6. Model mapping

模型匹配产生：

```text
clientModel = request body model
upstreamModel = ModelDefinition.upstream ?? id
```

Template beforeRequest/default passthrough 必须把上游 body 中的 model 改为 `upstreamModel`。

## 7. `/v1/models`

推荐返回 OpenAI-compatible list：

```json
{
  "object": "list",
  "data": [
    {"id":"client-model","object":"model","owned_by":"token-pool"}
  ]
}
```

若同一 client model 来自多个实例，只显示一次。

## 8. Error Mapping

内部错误不要直接把 Go/JS stack 返回客户端。

OpenAI style 示例：

```json
{
  "error": {
    "message": "No eligible credential is available for model x",
    "type": "routing_error",
    "code": "no_eligible_credential"
  }
}
```

## 9. Unknown Endpoint

MVP 未注册 endpoint 返回 404，而不是把任意客户端路径盲目代理到 Provider。

后续可由显式 Protocol Plugin/Template Route capability 扩展。
