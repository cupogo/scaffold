# scaffold

Go 项目脚手架, 专门用于快速实现 RESTful 接口, 包含代码生成器和示例

Go project scaffolding for RESTful API, including a code generator and examples

## Features

* 依据文档定义生成相应的模型 `pkg/models/xxx/doc_gen.go`
* 依据参数生成相应的数据访问层 `pkg/services/stores/doc_gen.go`
* 依据参数生成相应的Web API层 `pkg/web/api_vx/handle_doc_gen.go`
* 以上代码如无误可一次生成，并允许无限更新
* 生成的API支持swagger描述文档

## 模型生成，使用 `yaml`

详见 [源文档结构说明](docs/) 和[示例](docs/cms.yaml)

### 生成初始文档

```bash
go run ./scripts/beafup -name mytask
```

### 生成指令

```bash
go run -tags=codegen ./scripts/codegen docs/cms.yaml
```

或者
```bash
make codegen MDs=docs/cms.yaml SPEC=7
```

### 新项目操作示例 Example for a new project

```bash
cd ~/myworkspace
test -d scaffold || git clone https://github.com/cupogo/scaffold
test -d myproject || mkdir myproject
cd myproject
test -f go.mod || go mod init mycom/mywork/myproject
test -d docs || mkdir docs

test -d pkg/web || mkdir -p pkg/web
test -d pkg/web/resp || cp -r ../scaffold/pkg/web/resp pkg/web/
test -d pkg/web/routes || cp -r ../scaffold/pkg/web/routes pkg/web/

code docs/cms.yaml
# or
go run ../scaffold/scripts/beafup -name cms

go run -tags=codegen ../scaffold/scripts/codegen docs/cms.yaml

```
