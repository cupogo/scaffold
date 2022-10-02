# scaffold

Go 项目脚手架, 包含代码生成器和示例

Go project scaffolding, including code generators and examples

## Features

* 依据文档定义生成相应的模型 `pkg/models/xxx/doc_gen.go`
* 依据参数生成相应的数据访问层 `pkg/services/stores/doc_gen.go`
* 依据参数生成相应的Web API层 `pkg/web/api_vx/handle_doc_gen.go`
* 以上代码如无误可一次生成，并允许无限更新
* 生成的API支持swagger描述文档

## 模型生成，使用 `yaml`

详见 [源文档结构说明](docs/)

### 生成指令

```bash
go run -tags=codegen ./scripts/codegen docs/cms.yaml
```

或者
```bash
make codegen MDs=docs/cms.yaml SPEC=7
```
