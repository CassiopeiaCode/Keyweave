# 08 — Transport 与 Proxy

## 1. Proxy 是 Credential 官方字段

每个 Instance 可以有：

```text
proxy
```

例如：

```text
http://127.0.0.1:8080
http://user:pass@proxy.example.com:3128
direct
```

后续可增加：

```text
https://
socks5://
```

## 2. Proxy Resolution

默认：

```text
Template.getProxy()
  -> if returned: use it
  -> else instance.proxy
  -> else template default proxy
  -> else optional global proxy
  -> else direct
```

如果 Instance 显式 `direct`，禁止继承全局代理。

## 3. 强制绑定

Template：

```js
await ctx.http.request(...)
```

Host 在 request 创建时已经绑定 credential execution context。模板不能通过自己构造 socket 绕过 proxy。

## 4. HTTP Proxy

普通 HTTP：

```text
Client -> proxy
Request-URI = absolute URI
```

HTTPS：

```text
Gateway -> Proxy CONNECT upstream:443
Proxy -> 200
Gateway TLS -> upstream through tunnel
```

## 5. Proxy Authentication

支持 URL userinfo：

```text
http://user:password@host:port
```

Host 生成 Proxy-Authorization。Proxy password 视为 Secret。

## 6. Connection Pool Isolation

连接池 key 至少包含：

```text
scheme
upstream authority
proxy identity
TLS policy
```

不能把通过 Proxy A 创建的 keep-alive connection 错误复用到 Proxy B。

## 7. Template 自定义 Proxy

Template 可以：

```js
export function getProxy(ctx) {
  return ctx.fields.region === "jp"
    ? ctx.fields.jpProxy
    : ctx.fields.proxy;
}
```

返回值经过 Host validation。

## 8. SSRF 防护

默认建议：

- scheme allowlist；
- 禁止 cloud metadata 地址；
- 可配置 deny CIDR；
- DNS rebinding 二次校验；
- redirect target 重新校验；
- 禁止 `file://`、`gopher://` 等非 HTTP scheme。

## 9. Timeouts

分别配置：

```text
connectTimeout
tlsHandshakeTimeout
responseHeaderTimeout
idleConnTimeout
requestOverallTimeout
```

Streaming 的 overall timeout 不能使用普通短请求默认值。

## 10. Cancellation

下游 cancel 必须传播：

```text
downstream cancel
 -> upstream request cancel
 -> body close
 -> lease release
```
