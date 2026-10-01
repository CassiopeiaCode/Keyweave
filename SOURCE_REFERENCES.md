# 外部参考与兼容性基线

冻结日期：2026-10-01。

CLIProxyAPI 仅用于兼容性目标参考，不意味着复制其内部架构或代码。

- Project: https://github.com/router-for-me/CLIProxyAPI
- Example config: https://github.com/router-for-me/CLIProxyAPI/blob/main/config.example.yaml
- Management API docs: https://github.com/router-for-me/CLIProxyAPIDocs/blob/main/docs/en/management/api.md

截至冻结日期公开文档可观察到的、与本项目相关的能力包括：

- round-robin / weighted-round-robin / fill-first；
- universal session affinity + TTL + unavailable failover；
- API-key / OAuth/file credentials；
- per-entry Base URL、Proxy、Headers、model alias/exclusion；
- OpenAI-compatible custom providers；
- credential model definitions/status/metadata management；
- retry/cooldown/quota-related runtime behavior；
- streaming/non-streaming，以及部分 WebSocket 场景。

本项目的兼容策略见 `docs/16-cliproxyapi-compatibility.md`。
