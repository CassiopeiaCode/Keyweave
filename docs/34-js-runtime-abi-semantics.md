# 34 — JavaScript Runtime ABI 精确定义

本章补充 `contracts/template-runtime.d.ts` 的运行语义。

## 1. Module 形式

Template source 使用 ES module 风格导出：

```js
export const manifest = {...};
export async function execute(ctx) {...}
```

如果所选 JS 引擎不原生支持 ESM，可在编译层转换，但对模板作者保持此接口。

## 2. Hook 缺省

Hook 不存在时使用 Core/Library 默认值：

```text
supports           => true (前提是静态 protocol/model 已匹配)
getAvailability    => instance.availability
getDecrease        => manifest.defaultDecrease ?? 1
getScheduleHints   => {}
getProxy           => instance.proxy ?? template default proxy ?? global policy
beforeRequest      => no-op
execute            => default passthrough（仅模板声明可用时）
afterResponse      => no-op
onError            => {action:"stop"}
onFinish           => no-op
```

## 3. Hook 调用上下文

每次 Attempt 创建独立 `TemplateContext`。Context 不能跨请求保存。

`ctx.fields` 是：

```text
template official defaults
 + template custom fields
 + instance official overrides
 + instance custom fields
```

Secret 不参与此 merge。

## 4. Body 读取

`ctx.request.json()` / `text()` 如果会消耗 request stream，Host 必须实现可重放缓存或明确限制只能调用一次。

MVP 对 AI JSON 请求建议在协议层设置最大 body 后读取为 replayable bytes，再交 Runtime。

## 5. HostResponse ownership

`execute()` 返回 HostResponse 后：

- 非流 body 可由 Gateway 消费；
- 流 body ownership 转移给 Gateway；
- Template 不得在 return 后再次并发读取 body。

## 6. afterResponse 与流

`afterResponse` 默认只保证看到 response metadata，不应默认要求读取完整 stream。

若 Template 需要 stream inspection，应使用专门 `transformResponseStream` Phase 2 ABI；不要在 `afterResponse` 内偷偷全量读取流。

## 7. Patch

`ctx.instance.patch()` 必须：

- 校验官方字段；
- JSON-safe；
- 使用 CAS/原子 merge 语义；
- 不允许修改 `id/templateId`；
- 不允许写 Secret 明文到 customFields。

## 8. Atomic State

`atomic.increment(path, delta)` 用于数字计数器。

path 采用受限 dotted path，例如：

```text
runtime.failureCount
quota.window.429Count
```

必须阻止 prototype pollution 字段：`__proto__`, `constructor`, `prototype`。

## 9. 网络

`ctx.http.request`：

- 默认 proxyMode=`instance`；
- URL 必须绝对地址；
- 自动继承 request cancellation；
- response body streaming；
- Host 限制 header/body 大小；
- 默认禁止访问文件协议等非 HTTP(S) scheme。

## 10. Runtime timeout

区分：

- JS CPU/同步执行 budget；
- hook wall timeout；
- HTTP request timeout。

等待 `ctx.http` 不应该全部计入很短的 CPU budget，但仍受 request deadline。

## 11. Determinism

Scheduler JS 比 Template JS 更严格：

- 无网络；
- 无 Secret；
- 无 instance patch；
- 候选只读；
- 应在极短 deadline 内完成。
