# scaffold

Go 项目脚手架

## Features

* 依据文档定义生成相应的模型
* 依据参数生成相应的数据访问层
* TODO: 依据参数生成相应的API层

## 模型生成(已废弃)

1. 创建 docs/{model}.md 文档

2. `make modcodegen` 或 `make modcodegen MDs={model}`

## 模型生成，使用 `yaml`

[源文档结构说明](docs/)

```bash
go run -tags=codegen ./scripts/pgcodegen docs/cms.yaml
```
