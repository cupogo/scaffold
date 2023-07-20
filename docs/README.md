`yaml`文档结构说明
===

- 在此docs目录下，`yaml` 格式的文档，用于表示一套相关联的业务中所涉及到的若干对象以及接口的相关定义；

- 使用脚手架项目中的代码生成脚本工具(`codegen`)，会通过读取此文档，并按文档中所描述的结构和定义，在当前项目约定位置生成相对应的业务实现代码。

### 文档结构主要包括三部分内容

1. 模型定义
2. 存储接口定义
3. Web接口定义

除此之外，还有若干通用定义

- `modelpkg`: 约定模型所在的包名，推荐使用简约可识别的词

- `depends`: 字典类型，指定依赖的包

- `enums`: 集合类型，定义若干枚举类型

## 模型定义 `models`

字段名 描述 是否必需

- `name` 对象名称， 必需

- `comment` 对象注释和描述，建议写

- `tableTag` 表名标签，

- `fields` 字段列表
  
  - `comment` 字段注释说明，必需，名称和描述用空格分开

  - `name` 字段名，如果提供是内嵌类型，可以只定义`type`， 否则必需
  
  - `type` 类型名， 如果省略，则为内嵌类型（会继承类型的方法）
  
  - `tags` map类型，用于Go标签，例如json等
  
  - `basic` 布尔类型，是否是创建时必需的字段
  
  - `isset` 布尔类型，字段内容是否允许变更，一般用于更新操作

  - `query` 字串类型，字段可作为查询参数，有多种值，见后

  - `changeWith` 布尔类型，此字段有自己的更新方法，签名为 `ChangeWith(other) bool`

  - `icse` 布尔类型，此字段在查询匹配时忽略大小写 Ignore case sensitivity equality

  - `compare` 字串类型，此字段有自己的比较方法，可选值为`scalar`和`equalTo`，其中后者的签名为 `EqualTo(other) bool`

- `plural`: 复数形式名称，如不指定，会自动生成

- `oidcat`:  指定使用在oid包中定义的类型名称

- `discardUnknown`: 布尔类型，忽略未知的列，只`pg`需要，`bun`设定在连接时

- `withColumnGet`: 布尔类型，Get时允许定制列

- `dbTriggerSave`: 布尔类型，已存在保存时生效的数据表触发器

- `forceCreate`: 布尔类型，强行创建不报错

- `postNew`: 布尔类型，在函数`NewModelWithBasic`结束前调用

- `preSet`: 布尔类型，在方法`SetWith`开始前调用

- `postSet`: 布尔类型，在方法`SetWith`结束前调用

- `hooks`：字典类型，钩子方法集

- `disableLog`: 布尔类型，不记录模型的日志

**注意**：所有必需的定义都需要可导出，也即首字母为大写

- 大多数模型都会以字段的形式嵌入 `comm.DefaultModel` 这个默认模型结构体，由此会自动添加 `id`,`created`,`updated`和 `creator_id` 等字段，如果继续嵌入 `comm.MetaField` 则会添加 `meta` 支持添加更多元信息

### query 字段查询参数定义

- 查询定义分了两个部分：方法和扩展

- 方法:
   - `equal` 相等，此项有扩展，之间用逗号分隔
   - `ice` ignore case equal 相等并忽略大小写
   - `match` 进行模式匹配, `abc` = `abc*` = `abc%`
   - `date` 日期范围匹配, see also `sqlutil.DateRange`
   - `great` 大于
   - `less` 小于

- 扩展:
   - `decode` 此类型有自己的解码方法 `(t *T) Decode(string) error`
   - `hasVals` 整数类型可多选，只适用于位枚举 `(t T) Vals() []T`
   - `ints` 可多选的整数类型
   - `strs` 可多选的字串类型
   - `oids` 可多选的 `OID` 类型


### 模型存储的钩子说明

 - 所有自定义函数建议使用统一风格的名称
 - 位于事务中的，是纯函数，统一参数：
 	  1. `ctx context.Context` 上下文
	  2. `db ormDB` 数据库富指针
	  3. `obj Model` 当前操作对象指针

 - `beforeCreating` = "事务，在创建前"
 - `beforeUpdating` = "事务，在更新前"
 - `beforeSaving`   = "事务，在保存前"
 - `afterSaving`    = "事务，在保存后"
 - `beforeDeleting` = "事务，在删除前"
 - `afterDeleting`  = "事务，在删除后"
 - `afterCreated`   = "非事务，创建后"，参数和上面几位一致
 - `upsertES` = 非事务 更新后调用外部接口
 - `deleteES` = 非事务 删除后调用外部接口

非事务中的，是存储对象方法：
 - `afterLoad`      = "主键查询后"，参数：`ctx Context`，`obj Model`
 - `afterList`      = "列表查询后"，参数：`ctx Context`，`spec ModelSpec`，`data Slice`

## 存储接口定义 `stores`

字段名 描述 是否必需

- `name`: 存储对象名称 必需

- `iname`: 接口对象名称 如不提供会由`name`推导出

- `embed`: 嵌入接口名称

- `hodBread`: 集合类型 俱备浏览、读取、编辑、添加、删除全部功能的对象清单

- `hodPrdb`: 集合类型 俱备浏览、读取、存储、删除全部功能的对象清单

- `hodGL`: 集合类型 俱备只读即浏览和读取功能的对象清单

- `hods`: 集合类型 详细指定每个数据模型的增删改查列等方法，元素定义如下：

    - `name` string： 模型名
    - `type` string： 具体的方法代号，G=取，L=浏览，C=创建，U=更新，D=删除
    - `export` string： 是否导出上面的方法

- `methods`: 方法列表，如果提供了`hodBread`或`hodPrdb`，此项可省略，not in `hod*` only
  
  - `name`: 方法名称 必需
  
  - `simple`: 是否使用简单实现

- `postNew`: 布尔类型，若为true，且不存在`new{Name}Store`函数会自动创建，若存在strap方法会在返回前调用

- 注意: 当检测到ES相关的Hook时，若不存在`new{Name}Store`函数会自动创建并在返回前逐行运行指定模型


### 存储方法名称规则 （仅非hod集）

**not in `hod*` only**

- 名称由动词和一个名词对象组成，名词对象必须是前面的models里已经定义过的同名对象，即结构体，故需导出，也是说名称首字母要为大写，动词也是；

- 以 `Create`、`Update`、`Delete` 开头，表示创建、更新和删除记录；

- 以 `Put` 开头，表示会根据主键是否为空(`IzZeroID()`)来判断是创建还是更新；

- 以 `List` 开头，则表示按分页参数列出记录集合；

- 以 `Get` 开头，则表示按主键查询一条记录；

- 以上方法的参数和返回值形式和数量都是事先约定的；规则如下
  
  - `Get(ctx , id string) (*Object, error)`
  
  - `Delete(ctx, id string) error`
  
  - `Create(ctx, *Object) error`
  
  - `Update/Put (ctx, id string, obj *ObjectSet) error`
  
  - `List(ctx , spec *ObjectSpec) (Objects, int, error)`

### 存储接口的补充工作

在完成存储接口定义生成之后，还有少量的修补工作。

1. 钩子方法，目前代码生成工具暂时还搞不定，稍后完善

## Web接口定义 `webapi`

- `pkg`: 定义目录名（去除下划线等字母后即是包名）；

- `uris`: 集合类型， 来用定义路径，条目如下：
  - `model`: 模型名称
  - `uri`: 表示此模型数据的接口路径，会优先使用，如省略会使用前缀
  - `prefix`: 前缀，如果定义了 `uri`，此项会忽略
  - `batch`: 使用一个字母表示是否启动批量操作，目前支持 `C`(`Create`), `U`(`Update`)

- `handles`: 实现接口的方法定义以及`swagger`相关信息，如果在`uris`已经存在，此处要省略；
  
  - `name`: 方法名称 必需，一般不需要导出，可以小写字母开头；
  
  - `id`: 用于确定权限的编号，如不需要可以不加；
  
  - `store`:  即前面定义的在`Storage`接口中的方法名；，必需
  
  - `method`: 对应的存储接口方法名（不含参数），必需
  
  - `summary`: 接口功能摘要，必需
  
  - `route`: 接口路由地址，格式: 完整路径 [请求方法]，必需
  
  - `needAuth`: 是否需要登录身份

  - `needPerm`: 是否需要权限授权，此项仅控制确保有api.id

具体请参阅：[示例文档](cms.yaml)

## 生成的代码可以运行的约定

- 依赖 `andvari` 项目的 `pkg/models/{comm,oid}` 若干包

- 依赖 `andvari` 项目的 `pkg/stores/pgx` 包

- `pkg/servies/stores/wrap.go` 可提前准备或由模版生成

- `pkg/web/apixx/api.go` 可提前准备或由模版生成




## 设置全文检索

1. 模型添加嵌入字段 `type: comm.TextSearchField`；
2. 指定文本类型字段的查询方法 `query: 'fts'` 或 `query: 'match,fts'`，添加 `match`是为了保留单独查询此字段的能力；
3. 设置更新方式：
   1. 方法一：使用触发器（推荐），生效需要设置模型选项 `dbTriggerSave: true`
   2. 方法二：自动，暂只支持更新操作

### 触发器示例

1. `database/procedure/pg_20_article_trigger.sql`

```sql
-- 触发器: article 更新时保存 ts_vec 字段
CREATE OR REPLACE FUNCTION article_save_trigger()
RETURNS TRIGGER AS $$

BEGIN
	IF TG_OP = 'UPDATE' OR TG_OP = 'INSERT' THEN
		IF NEW.ts_cfg <> '' AND EXISTS(SELECT oid FROM pg_ts_config WHERE cfgname = NEW.ts_cfg) THEN
		NEW.ts_vec = to_tsvector(NEW.ts_cfg::regconfig, jsonb_build_array(
			NEW.title, NEW.subtitle, NEW.summary, NEW.content, NEW.source, NEW.author)
			);
		END IF;
	END IF;
	RETURN NEW;
END;

$$
LANGUAGE plpgsql;
```

2. `database/triggers/pg_20_article.sql`

```sql
DROP TRIGGER IF EXISTS article_insert_or_update_trigger ON cms_article;
CREATE TRIGGER article_insert_or_update_trigger BEFORE INSERT OR UPDATE ON cms_article
FOR EACH ROW EXECUTE PROCEDURE article_save_trigger();
```
