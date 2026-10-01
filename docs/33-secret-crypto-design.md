# 33 — Secret 与加密设计

## 1. Secret 类型

至少包括：

- upstream API key；
- OAuth access/refresh token；
- client secret；
- proxy password；
- provider-specific secret custom field。

普通 custom fields 不应该偷偷保存 Secret。

## 2. Envelope

推荐格式：

```text
ciphertext
nonce
key_id
algorithm = AES-256-GCM
```

每条 Secret 独立随机 nonce。

AAD 推荐：

```text
owner_type + "\0" + owner_id + "\0" + secret_name
```

防止 ciphertext 被复制到其他对象后仍可透明解密。

## 3. Master Key

优先：

1. OS keychain / secret service；
2. 外部 secret manager；
3. 仅个人开发使用环境变量/权限严格的文件。

禁止把 master key 写进 SQLite。

## 4. Client API Key

Client key 不需要可逆：

```text
random 256-bit secret
 -> display once
 -> store keyed hash/HMAC or slow hash
```

如果需要高性能查找，可存：

```text
prefix -> candidate rows -> constant-time verify
```

也可使用 HMAC-SHA-256(masterLookupKey, token) 做唯一索引。

## 5. Template Secret API

模板只通过：

```ts
ctx.secrets.get("key")
```

获得 Secret handle。

Host 应尽可能避免 Secret 变成可枚举普通字段。

## 6. Redaction

结构化日志 redactor 必须处理：

```text
Authorization
Proxy-Authorization
api-key
x-api-key
cookie (configurable)
known secret values
proxy userinfo
```

错误对象在序列化前也经过 redactor。

## 7. Backup

SQLite 备份包含 ciphertext。恢复时必须同时具备对应 master key，否则 Secret 无法使用。

提供可选 encrypted export bundle，用新的 export passphrase 重新封装 Secret。
