# 11 — 管理 API

## 1. 管理面和推理面分离

建议：

```text
/v1/...                 inference
/api/admin/v1/...       management
```

Management Auth 使用独立密钥。默认只监听 localhost，除非显式允许远程管理。

## 2. Templates

```text
GET    /api/admin/v1/templates
POST   /api/admin/v1/templates
GET    /api/admin/v1/templates/{id}
PATCH  /api/admin/v1/templates/{id}
DELETE /api/admin/v1/templates/{id}
POST   /api/admin/v1/templates/{id}/validate
POST   /api/admin/v1/templates/{id}/test
```

Template PATCH 必须带 version。

## 3. Instances

```text
GET    /api/admin/v1/credentials
POST   /api/admin/v1/credentials
GET    /api/admin/v1/credentials/{id}
PATCH  /api/admin/v1/credentials/{id}
DELETE /api/admin/v1/credentials/{id}

POST /api/admin/v1/credentials/{id}/test
POST /api/admin/v1/credentials/{id}/rebind-template
```

GET 默认：

```text
key: masked
proxy password: masked
```

## 4. Groups

```text
GET    /api/admin/v1/groups
POST   /api/admin/v1/groups
GET    /api/admin/v1/groups/{id}
PATCH  /api/admin/v1/groups/{id}
DELETE /api/admin/v1/groups/{id}
POST   /api/admin/v1/groups/{id}/simulate-route
```

`simulate-route` 不发真实上游请求，可以返回：

```json
{
  "candidates": [
    {
      "instanceId":"a",
      "availability":100,
      "activeRequests":1,
      "decrease":10,
      "effectiveAvailability":90
    }
  ],
  "selected":"a"
}
```

## 5. Access Keys

```text
GET    /api/admin/v1/access-keys
POST   /api/admin/v1/access-keys
DELETE /api/admin/v1/access-keys/{id}
POST   /api/admin/v1/access-keys/{id}/rotate
```

新 key 只在创建/rotate 响应中返回一次完整值。

## 6. Sources

```text
GET/POST/PATCH/DELETE /api/admin/v1/sources
POST /api/admin/v1/sources/{id}/pull
POST /api/admin/v1/sources/{id}/preview
```

## 7. Debugger

请求：

```json
{
  "templateId": "openrouter",
  "instanceId": "or-1",
  "protocol": "openai.responses",
  "model": "fast",
  "request": {
    "method": "POST",
    "path": "/v1/responses",
    "headers": {},
    "body": {}
  },
  "dryRun": true
}
```

响应允许显示：

- effective fields（secret masked）
- protocol match
- model resolution
- availability
- decrease
- schedule hints
- proxy（password masked）
- final URL
- final headers（auth masked）
- logs
- hook durations

## 8. Rebind Template

必须先 dry-run validation：

```json
{
  "compatible": true,
  "preservedFields": [],
  "missingRequiredFields": [],
  "droppedFields": []
}
```

用户确认后再真正切 templateId。

## 9. Audit

管理写操作记录：

```text
actor
action
objectType
objectId
timestamp
redactedDiff
```

## 10. 错误格式

Admin API 统一：

```json
{
  "error": {
    "code": "version_conflict",
    "message": "credential changed since version 4",
    "details": {}
  }
}
```

不要把内部 stack trace 默认返回给客户端。
