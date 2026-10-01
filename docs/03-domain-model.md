# 03 — Domain Model

## 1. CredentialTemplate

推荐领域结构：

```ts
interface CredentialTemplate {
  id: string;
  name: string;
  description?: string;
  protocols: ProtocolCapability[];

  officialDefaults: {
    address?: string;
    baseUrl?: string;
    models?: ModelDefinition[];
    availability?: number;
    proxy?: string;
  };

  customFields: Record<string, unknown>;
  state: Record<string, unknown>;

  jsSource: string;
  version: number;

  createdAt: string;
  updatedAt: string;
}
```

Template 级 `key` 不建议放进普通 defaults；如确实要默认 Secret，应走 Template Secret Store，而不是 `officialDefaults.key` 明文 JSON。

## 2. CredentialInstance

```ts
interface CredentialInstance {
  id: string;
  templateId: string;
  name?: string;

  address?: string;
  baseUrl?: string;
  keySecretRef?: string;
  models?: ModelDefinition[];
  availability: number;
  proxy?: string;

  customFields: Record<string, unknown>;
  state: Record<string, unknown>;

  version: number;
  createdAt: string;
  updatedAt: string;
}
```

## 3. Official Fields

官方用户级字段只保留：

| 字段 | 含义 |
|---|---|
| address | Provider/account 语义地址，可选 |
| baseUrl | 默认上游基址 |
| key | Secret，物理存储为 secret ref/ciphertext |
| models | 可调用模型定义 |
| availability | 持久化调度分值 |
| proxy | 当前实例默认代理 |

内部 metadata 如 `id/templateId/version` 不属于“可编程业务字段”。

## 4. ModelDefinition

```ts
interface ModelDefinition {
  id: string;              // client-visible
  upstream?: string;       // upstream model id
  displayName?: string;

  protocols?: string[];

  forceResponseMapping?: boolean;

  inputModalities?: string[];
  outputModalities?: string[];

  metadata?: Record<string, unknown>;
}
```

如果 `upstream` 为空，默认等于 `id`。

允许多个不同 Credential Instance 暴露同一个 `id`；这是合法的模型池。

## 5. Field Merge

Effective Fields：

```text
template official defaults
<- instance official override
template custom fields
<- instance custom override
```

默认深合并规则：

- scalar：instance override
- object：recursive merge
- array：replace
- `null`：explicit clear
- missing：inherit

Template 可在 `initialize` hook 中派生运行时字段，但不应偷偷写数据库。

## 6. CredentialGroup

```ts
interface CredentialGroup {
  id: string;
  name: string;
  templateIds: string[];

  schedulerType: string;
  schedulerCode?: string;

  config: Record<string, unknown>;
  state: Record<string, unknown>;

  version: number;
}
```

## 7. AccessKey

```ts
interface AccessKey {
  id: string;
  name: string;
  secretHash: string;
  secretPrefix?: string;
  groupId: string;
  createdAt: string;
  revokedAt?: string;
}
```

## 8. CredentialSource

```ts
interface CredentialSource {
  id: string;
  name: string;
  type: string;
  templateId?: string;
  config: Record<string, unknown>;
  code?: string;
  state: Record<string, unknown>;
}
```

Source 只进入写路径，不进入 inference 热路径。

## 9. 数据一致性

- 删除 Template 时，如果仍有 Instance，默认拒绝；
- 删除 Group 时，如果仍有 AccessKey，默认拒绝；
- 删除 Credential Instance 不应自动修改 Group，因为 Group 引用的是 Template；
- Template ID 应稳定，不以显示名称做外键；
- Models 的 client-visible ID 在同一 Group 内可以重名；
- Instance 的 `availability` 必须是有限数值，NaN/Infinity 禁止持久化。
