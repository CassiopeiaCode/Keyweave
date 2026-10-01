# 32 — 管理 UI 与 Template Debugger

## 1. 页面

建议最小管理 UI：

```text
Dashboard
Templates
Credentials
Groups
Client Keys
Sources
Request Logs
Template Debugger
Settings
```

## 2. Credentials 列表

显示：

- ID / name；
- template；
- models；
- stored availability；
- current active requests（runtime）；
- calculated effective availability（仅观察值）；
- proxy mode（secret masked）；
- last success/error；
- custom summary。

不显示完整 Key。

## 3. Template 编辑器

至少支持：

- JS source；
- manifest；
- syntax/compile validation；
- version diff；
- save as new version；
- rollback by creating another version；
- Test/Dry Run。

## 4. Debugger 输入

```json
{
  "templateId": "generic-openai",
  "instanceId": "cred-1",
  "protocol": "openai.chat",
  "model": "gpt-x",
  "request": {
    "method": "POST",
    "path": "/v1/chat/completions",
    "headers": {},
    "body": {}
  },
  "networkMode": "disabled"
}
```

## 5. Debugger 输出

```text
Resolved template version
Resolved fields
Resolved model mapping
supports result
availability
decrease
schedule hints
resolved proxy (masked)
request before rewrite
request after rewrite (secrets masked)
logs
host API calls
optional upstream status
```

## 6. Dry Run 网络语义

`networkMode=disabled`：`ctx.http` 返回显式 `dry_run_network_disabled`，并记录拟发出的 URL/method/masked headers。

可选 `networkMode=real` 只能管理员显式开启，并显示警告，因为可能消耗额度或执行有副作用的 API 请求。

## 7. Key 创建

Client API Key 只在创建时显示一次完整 secret：

```text
ptpg_xxxxxxxxx
```

数据库只保存 hash + prefix。

## 8. Group 调试

提供 Route Preview：

```text
Key/Group + Protocol + Model + Session
 -> 展示所有候选
 -> 展示被过滤原因
 -> availability/decrease/active/effective
 -> scheduler 结果
```

Preview 默认不 acquire real lease，也不发网络请求。
