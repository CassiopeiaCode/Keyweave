# 23 — 统一错误模型

## 1. 为什么需要统一错误类型

Transport、Protocol、Template、Scheduler 如果只用字符串错误，会导致 retry、安全和 HTTP 映射难以稳定。

内部统一：

```ts
interface GatewayError {
  code: string;
  category: string;
  message: string;
  retryable: boolean;
  upstreamStatus?: number;
  cause?: unknown;
  details?: Record<string, unknown>;
}
```

## 2. Error Category

建议：

```text
auth
routing
scheduler
template
transport
proxy
upstream_http
stream
timeout
cancelled
storage
security
```

## 3. 典型错误码

```text
auth_missing
auth_invalid
access_key_revoked
group_not_found
model_not_routable
no_credential_available
scheduler_invalid_selection
template_compile_failed
template_timeout
template_capability_denied
proxy_invalid
proxy_auth_failed
upstream_connect_failed
upstream_timeout
upstream_http_error
stream_reset
client_cancelled
version_conflict
secret_decrypt_failed
```

## 4. Protocol 映射

Protocol Adapter 负责把内部错误转换成对应协议的 HTTP/body 形式。

不能让 Template 直接决定 Gateway 身份验证失败应返回什么格式。

## 5. Retry 输入

Template `onError` 收到的是结构化 error snapshot：

```json
{
  "code":"upstream_http_error",
  "category":"upstream_http",
  "upstreamStatus":429,
  "retryable":true
}
```

而不是让 Template 解析 `"429 Too Many Requests"` 字符串。
