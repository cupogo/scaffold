---
title: 创建 API 记录集生成脚本
type: feat
status: completed
date: 2026-04-14
---

# 创建 API 记录集生成脚本

## 概述

扩展现有的 `scripts/sqlgen/` 工具，添加 `api` 子命令，用于从 `docs/swagger.yaml` 解析并生成所有 API 端点的记录集初始化 SQL。新功能需要覆盖**全部接口**（包括公开接口），并提取更完整的元数据（parameters、responses）存入 `jsonb` 字段。

## 问题陈述 / 动机

当前项目仅有 `scripts/sqlgen` 生成 `auth_permission` 权限数据，其覆盖面有限：
- 仅包含有 `operationId` 的接口（跳过公开接口如 `/api/ping`）
- 只提取 `operationId` 和 `summary`，缺少 `description`、`parameters`、`responses` 等详细元数据
- 没有统一的 API 记录集表可供审计、文档同步、网关注册等场景使用

通过新增 API 记录集生成脚本，可以建立完整的 API 目录，为后续的 API 治理、自动化文档同步和网关集成打下基础。

## 数据模型

目标表结构：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `bigint` | OID 主键，使用 `oid.NewID(oid.OtDefault)` 生成 |
| `operation_id` | `varchar` | `operationId`，公开接口可为空 |
| `endpoint` | `varchar` | API 路径，如 `/api/accounts/{id}` |
| `method` | `varchar` | GET/POST/PUT/DELETE 等，统一大写 |
| `summary` | `text` | `summary` 字段内容 |
| `description` | `text` | `description` 字段内容，为空时回退为 `summary` |
| `parameters` | `jsonb` | 参数结构（YAML 数组的 JSON 表示） |
| `responses` | `jsonb` | 响应结构（YAML map 的 JSON 表示） |
| `updated` | `timestamptz` | 最后更新时间 |

DDL 建议：

```sql
CREATE TABLE IF NOT EXISTS api_record (
    id            BIGINT PRIMARY KEY,
    operation_id  VARCHAR NOT NULL DEFAULT '',
    endpoint      VARCHAR NOT NULL,
    method        VARCHAR NOT NULL,
    summary       TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    parameters    JSONB NOT NULL DEFAULT '[]',
    responses     JSONB NOT NULL DEFAULT '{}',
    creator_id BIGINT NOT NULL DEFAULT 0,
    created       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_record_endpoint_method
    ON api_record (endpoint, method);
```

## 技术方案

### 脚本结构

沿用现有 `scripts/sqlgen/` 目录，添加子命令参数区分生成类型：

```
go run ./scripts/sqlgen perm    # 生成 auth_permission（现有行为）
go run ./scripts/sqlgen api    # 生成 api_record
```

### 核心实现逻辑

#### 1. Swagger 解析扩展

扩展 `swag.go` 中的 `apiEntry` 结构体，添加 `description`、`parameters` 和 `responses` 字段：

```go
type apiEntry struct {
	OperationID string                 `json:"operationId" yaml:"operationId"`
	Summary     string                 `json:"summary" yaml:"summary"`
	Description string                 `json:"description" yaml:"description"`
	Parameters  []any                  `json:"parameters" yaml:"parameters"`
	Responses   map[string]any         `json:"responses" yaml:"responses"`
}
```

#### 2. 数据结构

参考 `sqlgen/main.go` 中 `Permission` 嵌入 `comm.DunceModel` 的模式：

```go
type ApiRecord struct {
	comm.DefaultModel

	OperationID string
	Endpoint    string
	Method      string
	Summary     string
	Description string
	Parameters  []any
	Responses   map[string]any
}
```

#### 3. 记录生成

遍历 `doc.Paths`，对每个 `path` → `method` → `entry`：
- 嵌入 `comm.DefaultModel` 提供 `id`、`created`、`updated` 等字段（ID 通过 `Creating()` hook 自动生成）
- `operation_id` = `entry.OperationID`（可为空）
- `endpoint` = `path`
- `method` = `strings.ToUpper(method)`
- `summary` = `entry.Summary`
- `description` = `coalesce(entry.Description, entry.Summary)`
- `parameters` = JSON 序列化 `entry.Parameters`（空数组存 `[]`）
- `responses` = JSON 序列化 `entry.Responses`（空对象存 `{}`）

#### 4. SQL 生成

输出文件：`database/schemas/pg_11_api_records.sql`

SQL 模板（与用户调整后的 DDL 一致）：

```sql
INSERT INTO api_record ("id", "operation_id", "endpoint", "method", "summary", "description", "parameters", "responses", "creator_id", "created", "updated")
VALUES
(1234567890123456789, 'v1-accounts-get', '/api/v1/accounts', 'GET', '列出账号', '列出账号', '[...]'::jsonb, '{...}'::jsonb, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (endpoint, method) DO UPDATE SET
    operation_id = EXCLUDED.operation_id,
    summary      = EXCLUDED.summary,
    description  = EXCLUDED.description,
    parameters   = EXCLUDED.parameters,
    responses    = EXCLUDED.responses,
    updated      = CURRENT_TIMESTAMP;
```

可选清理语句（与 `sqlgen` 保持一致）：

```sql
DELETE FROM api_record WHERE updated < CURRENT_DATE - INTERVAL '1 day';
```

### Makefile 集成

在 `Makefile` 中新增 `gen-api-records` 目标，调用现有 `sqlgen` 脚本的 `api` 子命令：

```makefile
gen-api-records:
	go run ./scripts/sqlgen api
```

## 系统范围影响

- **数据库层面**：新增 `api_record` 表，建议在应用启动或 schema 初始化时创建。
- **CI/CD**：如果 CI 中执行 schema 初始化，需要将 `pg_11_api_records.sql` 纳入执行列表。
- **无运行时影响**：该脚本为离线工具，不影响应用运行时性能。

## 验收标准

### 功能标准

- [ ] 扩展 `scripts/sqlgen/swag.go` 添加 `description`、`parameters`、`responses` 字段
- [ ] 在 `scripts/sqlgen/main.go` 中添加 `api` 子命令生成 `api_record`
- [ ] 脚本能解析 `docs/swagger.yaml` 中的**全部**接口路径（包括无 `operationId` 的公开接口）
- [ ] 有 `operationId` 的接口正确填充 `operation_id`；无 `operationId` 的接口 `operation_id` 留空
- [ ] `id` 使用 `comm.DefaultModel` + `Creating()` hook 自动生成，类型为 `bigint`
- [ ] `method` 统一为大写字符串
- [ ] `description` 为空时，使用 `summary` 回退填充
- [ ] `parameters` 以 JSON 数组形式存入 `jsonb`，空值存 `[]`
- [ ] `responses` 以 JSON 对象形式存入 `jsonb`，空值存 `{}`
- [ ] 生成 SQL 文件 `database/schemas/pg_11_api_records.sql`
- [ ] SQL 具备幂等性（`ON CONFLICT (endpoint, method) DO UPDATE`）
- [ ] `Makefile` 中新增 `gen-api-records` 目标
- [ ] 脚本执行失败时返回非零退出码

### 测试标准

- [ ] 为 `genApiRecords()` 编写单元测试：验证有/无 `operationId` 的解析、OID 生成、JSON 序列化
- [ ] 在测试数据库中执行生成的 SQL，验证首次插入和重复执行的幂等性

## 依赖与风险

### 依赖

- 复用项目已有的 `github.com/cupogo/andvari/models/oid` 包，无需引入新依赖
- 需要确认 `api_record` 表的 DDL 是否在项目的数据库初始化流程中创建

### 风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| `parameters`/`responses` JSON 结构过大 | 中 | PostgreSQL `jsonb` 无严格大小限制，但若异常庞大可考虑压缩或截断 |
| 向后兼容风险 | 低 | 现有 `perm` 子命令行为保持不变 |

## 来源与参考

- 现有模式参考：`scripts/sqlgen/main.go:42-65`（权限 SQL 生成逻辑）
- 现有模式参考：`scripts/sqlgen/swag.go:9-24`（Swagger 解析结构）
- Swagger 文件：`docs/swagger.yaml:381-500`（paths / parameters / responses 结构示例）
- 数据库 schema 目录：`database/schemas/`
