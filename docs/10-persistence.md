# 10 — SQLite 持久化

## 1. 选择 SQLite WAL

个人部署优先：

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

优点：

- 单文件；
- transaction 完整；
- 运维简单；
- 支持并发读；
- 足够承载个人 API gateway metadata。

## 2. JSON 字段

Template/Instance Custom Fields、state、models 等先使用 JSON TEXT。

不要为了 Provider-specific 类型提前拆几十张表。

## 3. Secret

Secret 不放入：

```text
custom_fields_json
```

官方 key 使用：

```text
key_ciphertext
key_nonce
key_kid
```

Custom secret 长期建议独立：

```text
secret_items(owner_type, owner_id, name, ciphertext...)
```

## 4. Optimistic Version

Template / Instance / Group：

```text
version INTEGER NOT NULL
```

Patch：

```sql
UPDATE credential_instances
SET ..., version = version + 1
WHERE id = ? AND version = ?;
```

0 row = conflict。

## 5. Runtime State 与 Persistent State

默认内存：

- active leases
- RR cursor
- compiled template cache
- short-lived session affinity
- in-flight spans

持久化：

- availability
- template custom state
- explicit cooldown（若 Template 选择）
- OAuth tokens
- imported credentials

## 6. Transaction 边界

以下操作应使用 transaction：

- 创建 credential + secret；
- rotate access key；
- rebind template；
- delete template dependency check；
- admin patch + audit log。

不要让 inference 请求长期持有 write transaction。

## 7. WAL 维护

提供：

- WAL size metric；
- background checkpoint；
- graceful shutdown checkpoint 可选。

## 8. Backup

提供 SQLite backup API 或 `VACUUM INTO`/在线备份机制，不建议用户直接复制活跃 DB 文件。

## 9. Integrity

启动时：

- migrations version check；
- foreign key check；
- optionally `PRAGMA quick_check`；
- key encryption master key 可用性检查。
