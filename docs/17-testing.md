# 17 — 测试规格

## 1. 测试层次

```text
Unit
Contract
Integration
Concurrency
Fault Injection
Security
Compatibility
```

## 2. Domain Unit

必须测试：

- field merge；
- model alias；
- version conflict；
- access key hash lookup；
- secret redaction；
- invalid availability NaN/Infinity；
- Template 删除依赖约束。

## 3. Scheduler Unit

至少：

1. availability 高者优先；
2. 同分 Round Robin；
3. active lease decrease 生效；
4. hard max concurrency 生效；
5. candidate scope enforcement；
6. weighted RR 统计近似；
7. session binding；
8. session failover；
9. retry attempted-set；
10. 无候选的稳定错误。

## 4. 并发 Race

Go 必须：

```bash
go test -race ./...
```

场景：

- 1000 goroutine；
- 3 credentials；
- max concurrency 1/2/unlimited；
- random cancel；
- random upstream delay。

最终：

```text
all active lease count == 0
```

## 5. Streaming

Mock upstream：

- 立即失败；
- response header 后失败；
- 第一个 SSE event 前失败；
- 第一个 event 后失败；
- 客户端中途断开。

验证 retry boundary 与 lease cleanup。

## 6. Proxy

启动本地 test proxy：

- 无认证；
- Basic auth；
- auth failure；
- CONNECT；
- direct bypass。

检查 upstream 观察到的请求确实经过对应 proxy。

## 7. Runtime

错误/恶意 Template：

```js
while(true){}
```

```js
throw new Error("x")
```

```js
ctx.log.info(secret)
```

```js
await ctx.http.request("http://169.254.169.254/")
```

验证 timeout/redaction/SSRF policy。

## 8. Secret Snapshot

对日志、Admin API response、debug trace 做 fixture secret 全文搜索，必须 0 命中。

## 9. Test Vectors

`test-vectors/` 中 JSON 是跨语言 contract test 输入。

实现方可以增加字段，但不得改变已有向量的期望语义。

## 10. E2E

启动：

```text
Gateway
Mock Provider A
Mock Provider B
Mock HTTP Proxy
```

真实 curl/OpenAI SDK 调用，验证：

- model route；
- key replacement；
- stream；
- retries；
- proxy；
- model list。

## 11. 性能基线

不要求极致性能，但必须测：

- 无 Template 网络调用的路由开销；
- 100/500/1000 concurrent streams；
- SQLite admin write 对 inference read 的影响；
- Template compile cache 命中率。

性能优化不能破坏 lease correctness。
