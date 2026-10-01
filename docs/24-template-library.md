# 24 — 官方 Template 与 Scheduler Library

## 1. 目标

用户日常使用时不应该为了 OpenRouter/OpenAI-compatible 这种普通场景自己写 JS。

可编程能力是上限，不是使用门槛。

## 2. Template Library 第一批

```text
generic-bearer
generic-header-auth
generic-openai
openrouter
anthropic-api-key
gemini-api-key
```

每个 Library Template 应包含：

- manifest；
- JSON field schema；
- defaults；
- JS implementation；
- compatibility tests；
- example instances；
- README。

## 3. Generic Bearer

字段：

```text
baseUrl
key
models
proxy
authHeader(default Authorization)
authScheme(default Bearer)
```

## 4. Generic OpenAI

协议：

```text
openai.chat
openai.responses
```

行为：

- downstream auth strip；
- upstream bearer；
- baseUrl + incoming path；
- model alias；
- raw passthrough；
- SSE passthrough。

## 5. OpenRouter

基于 generic-openai，增加可选：

```text
HTTP-Referer
X-Title
quota/rate-limit response parsing
```

这些仍然是 Template 行为。

## 6. Scheduler Library

第一批：

```text
availability-round-robin
weighted-round-robin
fill-first
least-active
session-affinity
```

Scheduler 配置必须有 Schema，避免任意 JSON 靠猜。

## 7. Library 版本

Library Template 更新不能直接覆盖用户已编辑代码。

建议区分：

```text
library template definition
installed template
```

安装时复制或引用一个版本；用户本地修改后升级要显示 diff。
