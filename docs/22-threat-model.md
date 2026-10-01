# 22 — Threat Model

## 1. 资产

- upstream API keys
- OAuth access/refresh tokens
- proxy credentials
- Client API keys
- Template source
- user traffic
- usage metadata

## 2. Template code 风险

风险：

- secret exfiltration；
- SSRF；
- infinite loop；
- memory exhaustion；
- log exfiltration。

控制：

- sandbox；
- host-only network；
- resource quotas；
- redaction；
- capability declaration。

## 3. Admin API 风险

风险：

- remote credential theft；
- template replacement；
- routing takeover。

控制：

- separate management auth；
- localhost default；
- TLS when remote；
- audit；
- CSRF when browser auth。

## 4. Inference API 风险

风险：

- brute force key；
- request body DoS；
- slowloris；
- expensive upstream fan-out。

控制：

- high entropy keys；
- body limits；
- server timeouts；
- retry caps；
- concurrency caps。

## 5. Proxy 风险

风险：

- credential in proxy logs；
- malicious proxy MITM；
- wrong connection pool reuse。

控制：

- TLS verification；
- connection pool isolation；
- no proxy secret logs。

## 6. Template Trust Levels

建议：

```text
system
trusted
untrusted
```

- `system`：官方签名，可获更多 capability；
- `trusted`：用户手写，可用 credential-bound HTTP；
- `untrusted`：未来 worker 强隔离，默认最少 capability。

## 7. 安全默认值

- Admin localhost only；
- templates cannot direct network；
- secrets masked；
- no raw body logging；
- deny cloud metadata address；
- max request body；
- max retry attempts；
- template timeout。

## 8. 失败原则

安全相关校验失败时：

```text
fail closed
```

例如：

- proxy URL 无法解析；
- template capability 不允许 direct；
- secret decrypt 失败；
- scheduler 返回 scope 外 credential；
- template version 不兼容。
