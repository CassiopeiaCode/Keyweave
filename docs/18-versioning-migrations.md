# 18 — Versioning、热更新与 Migration

## 1. Template Version

每次 `jsSource`、manifest 或影响执行语义的默认配置变更：

```text
version++
```

Request 绑定：

```text
templateId + version
```

Compiled cache key 同样如此。

## 2. 热更新

旧请求继续旧 program；新请求使用新 program。

旧 version 没有 active runtime 后可从 cache 驱逐。

## 3. Group Version

Scheduler code/config 更新也需要 version。

Session binding 默认建议在 group version 改变后失效；也可以在下一次使用时重新验证 membership。

## 4. Instance Version

用于 Admin/template patch 的 optimistic concurrency。

## 5. DB Migration

每个 migration：

- 单调递增 ID；
- transaction；
- `schema_migrations` 记录；
- 不修改已发布 migration；
- 大数据 rewrite 使用 shadow column/table。

## 6. Template State Migration

Template code 可以定义逻辑层迁移：

```js
migrateState(fromVersion, toVersion, state)
```

这与 DB schema migration 分开。

状态迁移失败应：

- 保留旧 state；
- 记录明确错误；
- 默认阻止新 version 接管该实例，而不是破坏数据。

## 7. Bundle Version

整个规格和模板库应有独立版本：

```text
coreApiVersion
templateApiVersion
schedulerApiVersion
```

例如：

```text
ptpg.template/v1
ptpg.scheduler/v1
```

Core 升级时必须校验兼容范围。
