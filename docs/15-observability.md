# 15 — 可观测性

## 1. Request Log

每个 request：

```text
requestId
clientKeyId
groupId
protocol
requestedModel
stream
startAt
endAt
status
latencyMs
attemptCount
```

每个 attempt：

```text
attempt
templateId
templateVersion
instanceId
clientModel
upstreamModel
availability
decrease
activeRequestsBefore
effectiveAvailability
scheduler
proxyMode
upstreamStatus
timeToFirstByte
duration
errorClass
retryAction
```

## 2. 不能记录

- complete API Key
- Authorization
- OAuth refresh token
- proxy password
- raw request body by default
- raw response body by default

## 3. Metrics

建议：

```text
gateway_requests_total
gateway_request_duration_seconds
gateway_attempts_total
gateway_upstream_errors_total
gateway_active_leases
gateway_template_hook_duration_seconds
gateway_template_errors_total
gateway_scheduler_select_seconds
gateway_stream_active
gateway_db_busy_total
gateway_proxy_errors_total
```

Labels 控制 cardinality，默认不要把 instanceId/任意 model 字符串作为 Prometheus 高基数 label。

## 4. Debug Trace

管理员可按 requestId 查看：

```text
candidate set
filter reasons
scores
selected
hook timings
retry chain
```

Secrets masked。

## 5. Health

```text
/health/live
/health/ready
```

ready 至少检查：

- DB；
- migrations；
- runtime pool；
- encryption key loaded。

## 6. Event 类型

建议内部 structured events：

```text
request.started
routing.candidates_built
scheduler.selected
lease.acquired
template.hook
upstream.headers
stream.started
retry.scheduled
lease.released
request.completed
```

方便将来接入 tracing。
