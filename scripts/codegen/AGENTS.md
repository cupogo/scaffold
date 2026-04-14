# AGENTS.md

详细配置格式说明见 [docs/README.md](../../docs/README.md)

`codegen` 是一个从 YAML 配置生成 Go 项目代码的工具，用于快速生成 Models、Stores 和 Web API。

## 使用方法

```bash
go run ./scripts/codegen -spec=<flags> <yaml-file>
```

### 参数

- `-drop`: 是否先删除已存在的生成文件
- `-spec`: 指定生成内容（按位掩码）
  - `1` = Models (TgModel)
  - `2` = Stores (TgStore)
  - `4` = Web API (TgWeb)
  - 默认: `7` (全部生成)

## 配置文件示例

参考 `templates/docs/demo.yaml`

```yaml
depends:
  comm: 'github.com/cupogo/andvari/models/comm'
  oid: 'github.com/cupogo/andvari/models/oid'

dbcode: bun
modelpkg: demo

enums:
  - comment: 状态
    name: Status
    type: int8
    values:
      - label: 运行中
        suffix: Running
      - label: 已停止
        suffix: Stopped

models:
  - name: Task
    tableTag: 'demo_task,alias:t'
    oidcat: event
    fields:
      - type: comm.DefaultModel
      - name: Name
        type: string
        tags: {json: 'name', pg: ',notnull'}
        isset: true
        query: 'match'
      - name: Status
        type: Status
        isset: true
        query: 'equal'
      - type: comm.MetaField

stores:
  - name: demoStore
    hods:
      - { name: Task, type: LGCUD }

webapi:
  pkg: api_v1
  uris:
    - model: Task
      prefix: '/api/v1/demo'
```

## 生成输出

- **Models**: `pkg/models/<modelpkg>/*_gen.go`
- **Stores**: `pkg/services/stores/*_gen.go`
- **Web API**: `pkg/web/<pkg>/handle_*.go`

## 支持的数据库

- `bun` - Uptrace/Bun
- `pgx` - go-pg
- `mgm` - MongoDB (MGM)

## 支持的 Web 框架

- `gin` - gin-gonic/gin
- `chi` - go-chi/chi/v5

## 关键文件

- `main.go` - 入口
- `gens/generator.go` - 主生成器
- `gens/type_document.go` - 文档解析
- `gens/type_model.go` - 模型代码生成
- `gens/type_store.go` - Store 代码生成
- `gens/type_webapi.go` - Web API 代码生成
- `gens/type_enum.go` - 枚举代码生成
