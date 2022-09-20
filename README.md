# scaffold

Go 项目脚手架

## Features

* 依据文档定义生成相应的模型
* 依据参数生成相应的数据访问层
* 依据参数生成相应的API层

## 模型生成，使用 `yaml`

详见 [源文档结构说明](docs/)

```bash
go run -tags=codegen ./scripts/pgcodegen docs/cms.yaml
```

或者
```bash
make codegen MDs=docs/cms.yaml SPEC=7
```

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
