// go:build codegen
package gens

import (
	"log"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Model struct {
	Comment    string   `yaml:"comment,omitempty"`
	Name       string   `yaml:"name"`
	Label      string   `yaml:"label"`
	Identy     string   `yaml:"identy"`
	CollName   string   `yaml:"collName,omitempty"` // for mongodb only
	TableTag   string   `yaml:"tableTag,omitempty"` // uptrace/bun & go-pg
	Fields     Fields   `yaml:"fields"`
	Plural     string   `yaml:"plural,omitempty"`
	OIDCat     string   `yaml:"oidcat,omitempty"`
	OIDKey     string   `yaml:"oidKey,omitempty"`
	StoHooks   Tags     `yaml:"hooks,omitempty"`
	SpecExtras Fields   `yaml:"specExtras,omitempty"`
	Sifters    []string `yaml:"sifters,omitempty"`
	HookNs     string   `yaml:"hookNs,omitempty"` // prefix of hook store
	SpecNs     string   `yaml:"specNs,omitempty"` // prefix of model in stores, default empty
	SpecUp     string   `yaml:"specUp,omitempty"` // spec.{specUp}(ctx,obj) // deprecated
	Descr      string   `yaml:"descr,omitempty"`

	DiscardUnknown bool `yaml:"discardUnknown,omitempty"` // 忽略未知的列
	WithCompare    bool `yaml:"withCompare,omitempty"`    // 允许实现比较
	WithPlural     bool `yaml:"withPlural,omitempty"`     // 允许复数定义
	WithForeignKey bool `yaml:"withFK,omitempty"`         // 允许在创建时关联外键
	WithColumnGet  bool `yaml:"withColumnGet,omitempty"`  // Get时允许定制列
	WithColumnList bool `yaml:"withColumnList,omitempty"` // List时允许定制列
	DbTriggerSave  bool `yaml:"dbTriggerSave,omitempty"`  // 已存在保存时生效的数据表触发器
	WithCreatedSet bool `yaml:"withCreatedSet,omitempty"` // 开放created的设置
	ForceCreate    bool `yaml:"forceCreate,omitempty"`    // 强行创建不报错
	PostNew        bool `yaml:"postNew,omitempty"`
	PreSet         bool `yaml:"preSet,omitempty"`
	PostSet        bool `yaml:"postSet,omitempty"`
	DisableLog     bool `yaml:"disableLog,omitempty"` // 不记录model的日志
	Bsonable       bool `yaml:"bson,omitempty"`       // for mongodb only
	RegLoader      bool `yaml:"regLoader,omitempty"`  // 允许注册加载器
	WithSet        bool `yaml:"withSet,omitempty"`

	ExportOne  bool `yaml:"export1,omitempty"` // for alias in store
	ExportMore bool `yaml:"export2,omitempty"` // for alias in store

	doc    *Document
	pkg    string
	prefix string
}

func (m *Model) String() string {
	return m.Name
}

func (m *Model) GetPlural() string {
	if m.Plural == "" {
		return Plural(m.Name)
	}
	return m.Plural
}

func (m *Model) tableName() string {
	if m.TableTag == "" {
		return Underscore(m.GetPlural())
	}
	tt := m.TableTag
	if pos := strings.Index(tt, "table:"); pos > -1 {
		tt = tt[pos+6:]
	}
	if pos := strings.Index(tt, ","); pos > 0 {
		tt = tt[0:pos]
	}
	return tt
}

func (m *Model) tableAlias() string {
	if m.TableTag == "" {
		return Underscore(m.GetPlural())
	}
	tt := m.TableTag
	if pos := strings.Index(tt, "alias:"); pos > -1 {
		return tt[pos+6:]
	}
	if pos := strings.Index(tt, "table:"); pos > -1 {
		tt = tt[pos+6:]
	}
	return tt
}

func (m *Model) getIdenty() string {
	if len(m.Identy) > 0 {
		return LcFirst(m.Identy)
	}
	return LcFirst(m.prefix + m.Name)
}

func (m *Model) getLabel() string {
	if len(m.Label) > 0 {
		return LcFirst(m.Label)
	}
	return LcFirst(m.Name)
}

func (m *Model) TableField() jen.Code {
	tt := m.TableTag
	if tt == "" {
		tt = Underscore(m.GetPlural())
	}
	if m.doc.IsPG10() {
		if m.DiscardUnknown && !strings.Contains(tt, "discard_unknown_columns") {
			tt += ",discard_unknown_columns"
		}
		return jen.Id("tableName").Add(jen.Struct()).Tag(Tags{"pg": tt}).Line()
	}
	return jen.Id("comm.BaseModel").Tag(Tags{"json": "-", "bun": "table:" + tt}).Line()
}

type UniField struct {
	Field
	Column string
}

func (uf *UniField) Op() string {
	if uf.IgnoreCase {
		return "ILIKE"
	}
	return "="
}

func (m *Model) UniqueOne() (u UniField, onlyOne bool) {
	var count int
	for _, field := range m.Fields {
		if cn, _, ok := field.ColName(); ok {
			count++
			u.Field = field
			u.Column = cn
		}
	}
	onlyOne = count == 1
	return
}

func (m *Model) ChangablCodes() (members []jen.Code, imples []jen.Code, rets []jen.Code) {
	if m.PreSet {
		imples = append(imples, jen.Id("z").Dot("PreSet").Call(jen.Op("&").Id("o")))
	}
	bsonable := m.IsBsonable()
	if bsonable {
		rets = append(rets, jen.Id("base.BM"))
		imples = append(imples, jen.Id("m").Op(":=").Id("base.BM").Op("{}"))
	}
	var hasMeta bool
	var hasOwner bool
	var idx int
	for _, field := range m.Fields {
		if !field.IsSet || field.isEmbed() {
			if field.isMeta() {
				hasMeta = true
			} else if field.isOwner() {
				hasOwner = true
			}
			continue
		}
		idx++
		code := jen.Empty()
		if jc := field.commentCode(); jc != nil {
			code.Add(jc)
		}
		code.Id(field.Name)
		cn, isInDb, _ := field.ColName()
		qn, tn, isptr := field.cutType()

		jcond := jen.Id("o").Dot(field.Name).Op("!=").Nil()
		if field.isScalar() {
			jcond.Op("&&").Id("z").Dot(field.Name).Op("!=").Op("*").Id("o").Dot(field.Name)
		} else if field.Compare == CompareEqualTo {
			jarg := jen.Empty()
			if isptr {
				jarg.Op("*")
			}
			jarg.Id("z").Dot(field.Name)
			jcond.Op("&&!").Id("o").Dot(field.Name).Dot("EqualTo").Call(jarg)
		} else if field.Compare == CompareSliceCmp {
			jcond.Op("&&").Qual("slices", "Compare").Call(
				jen.Id("z").Dot(field.Name),
				jen.Id("*o").Dot(field.Name),
			).Op("!=0")
		}

		if qn == "oid" && tn == "OID" {
			field.Type = "string"
			field.isOid = true
			field.Qual = ""
			tn = field.Type
		}
		if len(field.Qual) > 0 {
			code.Op("*").Qual(field.Qual, tn)
		} else {
			code.Op("*").Id(tn)
		}
		tags := field.Tags.Clone()
		tags.CleanKeys("bson", "bun", "pg", "binding", "validate")
		if tags.Has("json") {
			if field.isScalar() {
				tags.FillKey("form", "json")
			}
			// tags := Tags{"json": s}
			tags.extOrder(idx, len(m.Fields))
			code.Tag(tags)
		}
		members = append(members, code)

		bsonName, bsonOK := field.BsonName()

		jcsa := jen.Id("z").Dot("SetChange").Call(jen.Lit(cn))

		imples = append(imples, jen.If(jcond).BlockFunc(func(g *jen.Group) {
			if isInDb && !m.DisableLog && !field.isOid {
				g.Id("z").Dot("LogChangeValue").Call(jen.Lit(cn), jen.Id("z").Dot(field.Name), jen.Id("o").Dot(field.Name))
			}
			if field.isOid {
				g.If(jen.Id("id").Op(":=").Id("oid").Dot("Cast").Call(jen.Op("*").Id("o").Dot(field.Name)).Op(";").
					Id("z").Dot(field.Name).Op("!=").Id("id")).BlockFunc(func(g1 *jen.Group) {
					if isInDb && !m.DisableLog {
						g1.Id("z").Dot("LogChangeValue").Call(jen.Lit(cn), jen.Id("z").Dot(field.Name), jen.Id("id"))
					}
					g1.Id("z").Dot(field.Name).Op("=").Id("id")
					if isInDb && m.DisableLog {
						g1.Add(jcsa)
					}
				})
			} else if field.IsChangeWith {
				g.Id("z").Dot(field.Name).Dot("ChangeWith").Call(jen.Id("o").Dot(field.Name))
			} else if isptr {
				g.Id("z").Dot(field.Name).Op("=").Id("o").Dot(field.Name)
			} else {
				g.Id("z").Dot(field.Name).Op("=").Op("*").Id("o").Dot(field.Name)
			}
			if isInDb && !field.isOid && m.DisableLog {
				g.Add(jcsa)
			}
			if bsonable && bsonOK {
				g.Id("m").Index(jen.Lit(bsonName)).Op("=").Id("z").Dot(field.Name)
			}
		}))
	}

	if m.WithCreatedSet {
		idx++
		// CreatedAt time.Time `bson:"createdAt" json:"createdAt" form:"createdAt" pg:"created,notnull,default:now()" extensions:"x-order=["` // 创建时间
		members = append(members, createdUpCode(idx))
		imples = append(imples, jen.If(jen.Id("o").Dot(createdField).Op("!=").Nil().BlockFunc(func(g *jen.Group) {
			g.Id("z").Dot(createdField).Op("=").Op("*").Id("o").Dot(createdField)
			g.Id("z").Dot("SetChange").Call(jen.Lit(createdColumn))
		})))
	}

	if hasMeta {
		name := "MetaDiff"
		jmetaup := []jen.Code{jen.Id("z").Dot("SetChange").Call(jen.Lit("meta"))}
		if bsonable {
			jmetaup = append(jmetaup, jen.Id("m").Index(jen.Lit("meta")).Op("=").Id("z").Dot("Meta"))
		}
		members = append(members, metaUpCode())
		imples = append(imples, jen.If(jen.Id("o").Dot(name).Op("!=").Nil().Op("&&").Id("z").Dot("MetaUp").Call(jen.Id("o").Dot(name))).Block(
			jmetaup...,
		))
	}
	if hasOwner {
		idx++
		name := "OwnerID"
		members = append(members, ownerUpCode(idx))
		imples = append(imples, jen.If(jen.Id("o").Dot(name).Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			g.If(jen.Id("id").Op(":=").Id("oid").Dot("Cast").Call(jen.Op("*").Id("o").Dot(name)).Op(";").
				Id("z").Dot(name).Op("!=").Id("id")).BlockFunc(func(g1 *jen.Group) {
				if !m.DisableLog {
					g1.Id("z").Dot("LogChangeValue").Call(jen.Lit("owner_id"), jen.Id("z").Dot(name), jen.Id("id"))
				}
				g1.Id("z").Dot("SetOwnerID").Call(jen.Id("id"))
				if m.DisableLog {
					g1.Id("z").Dot("SetChange").Call(jen.Lit("owner_id"))
				}
			})
		}))
	}

	if m.PostSet {
		imples = append(imples, jen.Id("z").Dot("PostSet").Call(jen.Op("&").Id("o")))
	}
	if bsonable {
		imples = append(imples, jen.Return(jen.Id("m")))
	}
	return
}

func (m *Model) Codes() jen.Code {
	st := jen.Empty()
	basicName := m.Name + "Basic"
	var cs []jen.Code
	isTable := m.IsTable()
	if isTable {
		cs = append(cs, m.TableField())
	}
	bsonable := m.IsBsonable()

	st.Comment("consts of " + m.Name + " " + m.shortComment()).Line()
	st.Const().DefsFunc(func(g *jen.Group) {
		if isTable {
			g.Id(m.Name + "Table").Op("=").Lit(m.tableName())
			g.Id(m.Name + "Alias").Op("=").Lit(m.tableAlias())
		} else if bsonable {
			g.Id(m.Name + "Collection").Op("=").Lit(m.CollectionName())
		}

		g.Id(m.Name + "Label").Op("=").Lit(m.getLabel())
		g.Id(m.Name + "TypID").Op("=").Lit(m.getIdenty())
	}).Line()

	mcs, bcs := m.Fields.Codes(basicName, isTable, bsonable)
	cs = append(cs, mcs...)
	st.Comment(m.Name + " " + m.Comment).Line()
	jcodeDesc(st, m.Descr, "@Description ")

	st.Type().Id(m.Name).Struct(cs...).Add(jen.Comment("@name " + LcFirst(m.prefix+m.Name))).Line().Line()

	if len(bcs) > 0 && (isTable || bsonable) {
		st.Type().Id(basicName).Struct(bcs...).Add(jen.Comment("@name " + LcFirst(m.prefix+basicName))).Line().Line()
	}

	pluralName := m.GetPlural()
	withPlual := pluralName != m.Name && (isTable || bsonable || len(m.Plural) > 0 || m.WithPlural)
	if withPlual {
		pname := m.GetPlural()
		st.Type().Id(pname).Index().Id(m.Name).Line().Line()
	}

	if jhk := m.hookModelCodes(); jhk != nil {
		st.Add(jhk)
	}
	if m.DisableLog {
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("DisableLog").Params().Bool().Block(jen.Return(jen.Lit(true)))
		st.Line()
	}
	if isTable || bsonable {
		if jc := m.basicCodes(); jc != nil {
			st.Add(jc)
		}
	}

	if ic := m.identityCode(); ic != nil {
		st.Add(ic)
	}

	if m.WithForeignKey {
		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("WithFK").Params().Bool().Block(jen.Return(jen.Lit(true)))
		st.Line()
	}

	if fields, stmts, rets := m.ChangablCodes(); len(fields) > 0 && (isTable || bsonable || m.WithSet) {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(fields...).Add(jen.Comment("@name " + LcFirst(m.prefix+changeSetName))).Line().Line()
		// scs = append(scs, jen.Return(jen.Id("z").Dot("CountChange").Call().Op(">0")))
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("SetWith").Params(jen.Id("o").Id(changeSetName)).Params(rets...).Block(
			stmts...,
		).Line()
	}
	if isTable || bsonable {
		if jc := m.metaAddCodes(); jc != nil {
			st.Add(jc)
		}
	}
	if jc := m.codeEqualTo(); jc != nil {
		st.Add(jc).Line()
		if withPlual {
			st.Func().Params(jen.Id("z").Id(pluralName)).Id("EqualTo").
				Params(jen.Id("o").Id(pluralName)).Bool().BlockFunc(func(g *jen.Group) {
				g.If(jen.Len(jen.Id("z")).Op("!=").Len(jen.Id("o"))).Block(
					jen.Return(jen.False()))
				g.For(jen.Id("i").Op(":=0;").Id("i").Op("<").Len(jen.Id("z")).Op(";").Id("i").Op("++")).Block(
					jen.If(jen.Op("!").Id("z").Index(jen.Id("i")).Dot("EqualTo").Call(jen.Id("o").Index(jen.Id("i")))).Block(
						jen.Return(jen.False())),
				)
				g.Return(jen.True())
			})
		}
	}
	return st
}

func (m *Model) hasBasic() bool {
	for i := range m.Fields {
		if m.Fields[i].IsBasic || m.Fields[i].IsSet {
			return true
		}
	}
	return false
}

func (m *Model) hasMeta() bool {
	for i := range m.Fields {
		if m.Fields[i].isMeta() {
			return true
		}
	}
	return false
}

func (m *Model) hasAudit() bool {
	for i := range m.Fields {
		if m.Fields[i].isAudit() {
			return true
		}
	}
	return false
}

func (m *Model) hasModHook() (ok bool, idf string, dtf string) {
	for _, field := range m.Fields {
		typ := field.getType()
		if _, b, ok := strings.Cut(typ, "."); ok && len(b) > 0 {
			typ = b
		}
		if typ == modelDefault {
			return true, modelDefault, modelDefault
		}
		if typ == modelDunce {
			return true, modelDunce, modelDunce
		}
		if typ == modelSerial {
			return false, modelSerial, modelSerial
		}

		if strings.HasPrefix(typ, "IDField") { // .IDField, .IDFieldStr
			idf = typ
		} else if typ == "SerialField" {
			idf = "SerialField"
		} else if typ == "DateFields" {
			dtf = "DateFields"
		}
	}

	if len(idf) > 0 && len(dtf) > 0 {
		ok = (idf == "IDField" || idf == "IDFieldStr")
	}
	return
}

func (mod *Model) IsTable() bool {
	if _, idf, _ := mod.hasModHook(); len(idf) > 0 && len(mod.TableTag) > 0 {
		return true
	}
	return false
}

func (m *Model) IsBsonable() bool {
	return m.Bsonable || len(m.CollName) > 0
}

func (m *Model) CollectionName() string {
	if len(m.CollName) > 0 {
		return m.CollName
	}
	return strings.ToLower(m.GetPlural())
}

func (m *Model) hookModelCodes() (st *jen.Statement) {
	if hasHooks, idF, dtF := m.hasModHook(); hasHooks {
		st = new(jen.Statement)
		st.Comment("Creating function call to it's inner fields defined hooks").Line()
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("Creating").Params().Error().Block(
			jen.If(jen.Id("z").Dot("IsZeroID").Call()).BlockFunc(func(g *jen.Group) {
				oidcat := CamelCased(m.OIDCat)
				if (len(m.OIDKey) >= 2 || len(oidcat) > 0) && (idF == modelDefault || idF == "IDField") {
					if len(oidcat) == 0 {
						oidcat = "Default"
					}
					oidQual, _ := m.doc.getQual("oid")
					if len(m.OIDKey) >= 2 {
						g.Id("id,ok").Op(":=").Qual(oidQual, "NewWithCode").Call(
							jen.Id(m.Name + "Label"))
						g.If(jen.Op("!").Id("ok")).Block(
							jen.Id("id").Op("=").Qual(oidQual, "NewID").Call(
								jen.Qual(oidQual, "Ot"+oidcat),
							))
						g.Id("z").Dot("SetID").Call(jen.Id("id"))
					} else {
						g.Id("z").Dot("SetID").Call(
							jen.Qual(oidQual, "NewID").Call(
								jen.Qual(oidQual, "Ot"+oidcat)),
						)
					}
				} else {
					g.Return(jen.Id("comm").Dot("ErrEmptyID"))
				}
			}).Line(),
			jen.Return(jen.Id("z").Dot(dtF).Dot("Creating").Call()),
		).Line()

	}
	return st
}

func (m *Model) basicCodes() (st *jen.Statement) {
	if !m.hasBasic() {
		return
	}
	st = new(jen.Statement)
	basicName := m.Name + "Basic"
	st.Func().Id("New" + m.Name + "WithBasic").Params(jen.Id("in").Id(basicName)).Op("*").Id(m.Name).BlockFunc(func(g *jen.Group) {
		g.Id("obj").Op(":=&").Id(m.Name).Block(jen.Id(basicName).Op(":").Id("in").Op(","))
		if m.hasMeta() {
			g.Op("_=").Id("obj").Dot("MetaUp").Call(jen.Id("in").Dot("MetaDiff"))
		}
		if m.PostNew {
			g.Id("obj").Dot("PostNew").Call(jen.Op("&").Id("in"))
		}
		g.Return(jen.Id("obj"))
	})
	st.Line()
	st.Func().Id("New" + m.Name + "WithID").Params(jen.Id("id").Any()).Op("*").Id(m.Name).BlockFunc(func(g *jen.Group) {
		g.Id("obj").Op(":=").New(jen.Id(m.Name))
		if m.IsBsonable() {
			g.Id("obj").Dot("SetID").Call(jen.Id("id"))
		} else {
			g.Op("_=").Id("obj").Dot("SetID").Call(jen.Id("id"))
		}
		g.Return(jen.Id("obj"))
	})
	st.Line()
	return st
}

func (m *Model) specFields() (out Fields) {
	for _, f := range m.Fields {
		if q, ok := f.parseQuery(); ok {
			// log.Printf("name: %s, sfn: %q, ext: %q", f.Name, sfn, ext)
			f.siftExt = q.ext
			if q.add {
				ftyp := "string"
				if q.ext == "oids" {
					ftyp = "oid.OIDsStr"
				}
				argTag := Plural(f.getArgTag())
				f0 := Field{
					Comment:  f.Comment + " (多值逗号分隔)",
					Type:     ftyp,
					Name:     Plural(f.Name),
					Tags:     Tags{"form": argTag, "json": argTag + ",omitempty"},
					siftExt:  q.ext,
					multable: true,
				}
				// log.Printf("f0: %+v", f0)
				out = append(out, f0)
			} else if q.ext == "decode" {
				f.qtype = "string"
				f.Comment += " (支持混合解码)"
			} else if q.ext == "hasVals" {
				f.Comment += " (多值数字相加)"
			}
			if f.Type == "oid.OID" {
				f.Type = "string"
				f.isOid = true
				if q.sift == "siftOIDs" {
					f.siftFn = q.sift
				} else {
					f.siftFn = "siftOID"
				}
			} else if f.Type == "oid.OIDs" {
				f.Type = "string"
				f.isOid = true
				f.siftFn = "sift"
				f.siftOp = "any"
				f.Comment += " (多值逗号分隔)"
			} else if strings.HasSuffix(f.Type, "DateTime") {
				f.Type = "string"
				f.isDate = true
				f.isIntDt = true
				f.siftFn = "siftDate"
			} else if strings.HasSuffix(f.Type, "Time") {
				f.Type = "string"
				f.isDate = true
				f.siftFn = "siftDate"
			} else if f.Type == "bool" || q.ext == "str" {
				f.Type = "string"
				f.siftFn = q.sift
			} else if q.custom {
				if f.isOwner() {
					f.Type = "string"
					f.Name = "OwnerID"
					if len(f.Comment) == 0 {
						f.Comment = "所有者编号"
						f.qtype = "string"
						f.Tags = Tags{"form": "owner", "json": "owner,omitempty"}
					}
				}
			} else {
				f.siftFn = q.sift
			}

			out = append(out, f)
		} else if f.isOwner() && f.Query != "ignore" {
			sfn := "siftOIDs"
			cmt := "所有者编号 (多值使用逗号分隔)"
			if f.Query == "x" {
				sfn = ""
				if len(f.Comment) > 0 {
					cmt = f.Comment
				} else {
					cmt = "所有者编号"
				}
			}
			f0 := Field{
				Comment: cmt,
				Type:    "string",
				Name:    "OwnerID",
				Tags:    Tags{"form": "owner", "json": "owner,omitempty"},
				siftFn:  sfn,
				colname: "owner_id",
				bson:    m.IsBsonable(),
			}
			out = append(out, f0)
		}
	}
	return
}

func (m *Model) sortableColumns() (cs []string) {
	for _, f := range m.Fields {
		if f.isEmbed() {
			if f.isOwner() && f.Sortable {
				cs = append(cs, "owner_id")
			}
			continue
		}

		if cn, ok, _ := f.ColName(); ok && len(cn) > 0 && f.Sortable {
			cs = append(cs, cn)
		}
	}
	return
}

func (m *Model) getExportName(ss ...string) string {
	return getExportName(m.Name, m.SpecNs, ss...)
}

func getExportName(name, specNs string, ss ...string) string {
	ns := ToExported(specNs)
	if ns == name {
		ns = ""
	}
	if len(ss) > 0 && len(ss[0]) > 0 {
		return ns + name + ss[0]
	}
	return ns + name
}

func (m *Model) getSpecName() string {
	return m.getExportName("Spec")
}

func (m *Model) jSpecBasic() (name, parent string, args []jen.Code, rets []jen.Code,
	jfsc func(on string) jen.Code) {
	name = m.getSpecName()
	if m.IsBsonable() || m.doc.IsMongo() {
		parent = "base.ModelSpec"
		args = append(args, jen.Id("q").Id("BD"))
		rets = append(rets, jen.Id("BD"))
		jfsc = func(on string) jen.Code {
			return jen.Id("q").Op("=").Id("spec").Dot(on).Dot("Sift").Call(jen.Id("q"))
		}
	} else {
		parent = "ModelSpec"
		args = append(args, jen.Id("q").Op("*").Id("ormQuery"))
		rets = append(rets, jen.Op("*").Id("ormQuery"))
		if m.doc.IsPG10() {
			rets = append(rets, jen.Error())
			jfsc = func(on string) jen.Code {
				return jen.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot(name).Dot("Sift").Call(jen.Id("q"))
			}
		} else { // bun
			jfsc = func(on string) jen.Code {
				return jen.Id("q").Op("=").Id("spec").Dot(on).Dot("Sift").Call(jen.Id("q"))
			}
		}

	}

	return
}

func (m *Model) genSiftCode(field, fieldM Field, isPG10, withRel bool) jen.Code {
	if field.siftExt == "skip" {
		return nil
	}
	utilsQual, _ := m.doc.getQual("utils")
	cn, indb, _ := field.ColName()
	if !indb && len(field.siftFn) == 0 {
		return nil
	}
	acn := cn
	if !isPG10 && withRel {
		// cn = m.tableAlias() + "." + cn
		acn = "?TableAlias." + cn
	}
	jSV := jen.Id("spec").Dot(field.Name)
	jq := field.siftCode(m.IsBsonable() || m.doc.IsMongo())
	jSiftVals := jen.Id("q").Op(",").Id("_").Op("=").Id("sift").Call(jen.Id("q"), jen.Lit(cn), jen.Lit("IN"), jen.Id("vals"), jen.Lit(false))
	if field.siftExt == "decode" {
		if m.IsBsonable() || m.doc.IsMongo() {
			jv := jen.Id("v")
			if field.isTagJsonString() {
				jv = jen.Id("fmt.Sprintf").Call(jen.Lit("%d"), jv)
			}
			jq = jen.Id("q").Op("=").Id("mg"+ToExported(field.siftFn)).Call(
				jen.Id("q"), jen.Lit(cn), jv)
		} else {
			jq = jen.Id("q").Op("=").Id("q").Dot("Where").Call(jen.Lit(acn+" = ?"), jen.Id("v"))
		}
		return jen.If(jen.Len(jSV).Op(">0")).Block(
			jen.Var().Id("v").Add(field.typeCode(m.doc.getModQual(field.getType()))),
			jen.If(jen.Err().Op(":=").Id("v").Dot("Decode").Call(jSV).Op(";").Err().Op("==").Nil()).Block(jq),
		)
	}
	if field.siftOp == "any" {
		return jen.If(jen.Id("vals").Op(":=").Qual("strings", "Split").Call(jSV, jen.Lit(",")).Op(";").Len(jSV).Op(">0").Op("&&").Len(jen.Id("vals")).Op(">0")).Block(
			// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
			jen.Id("q").Op(",").Id("_").Op("=").Id("sift").Call(jen.Id("q"), jen.Lit(cn), jen.Lit(field.siftOp), jen.Id("vals"), jen.Lit(false)),
		)
	}
	if field.siftExt == "hasVals" {
		return jen.If(jen.Id("vals").Op(":=").Id("spec").Dot(field.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
			// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
			jSiftVals,
		).Else().Block(jq)
	}
	if field.siftExt == "ints" {
		return jen.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseInts").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
			// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
			jSiftVals,
		).Else().Block(jq)
	}
	if field.siftExt == "strs" {
		return jen.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseStrs").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
			// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
			jSiftVals,
		).Else().Block(jq)
	}
	if field.siftExt == "oids" {
		return jen.If(jen.Id("vals").Op(":=").Id("spec").Dot(fieldM.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
			// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
			jSiftVals,
		).Else().Block(jq)
	}

	return jen.Add(jq)
}

func (m *Model) getSpecCodes() jen.Code {
	var fcs []jen.Code
	tname, parent, args, rets, jfSiftCall := m.jSpecBasic()
	fcs = append(fcs, jen.Id("PageSpec"), jen.Id(parent))
	if m.hasAudit() {
		fcs = append(fcs, jen.Id("AuditSpec"))
	}
	for _, sifter := range m.Sifters {
		fcs = append(fcs, jen.Id(sifter))
	}
	colTS, okTS := m.HasTextSearch()
	if okTS || len(colTS) > 0 {
		fcs = append(fcs, jen.Id("TextSearchSpec"))
	}

	var idx int
	specFields := m.specFields()
	if len(specFields) > 0 {
		// log.Printf("specFields: %+v", specFields)
		fcs = append(fcs, jen.Empty())
		for _, field := range specFields {
			fcs = append(fcs, field.queryCode(idx, m.doc.getModQual(field.getType())))
			idx++
		}

	}

	var withRel string
	var wrTyp string
	_, okAL := m.hasStoreHook(afterList)
	relFields := m.Fields.relHasOne()
	relations := m.Fields.Relations()
	if len(relFields) > 0 || len(relations) > 0 {
		// wrTyp = "bool"
		// if okAL || len(relations) > 1 {
		wrTyp = "string"
		// }
		withRel = "WithRel"
		jtag := "rel"
		field := &Field{
			Name: withRel,
			Type: wrTyp, Tags: Tags{"json": jtag},
			Comment: "include relation name"}
		if len(relations) > 0 {
			field.Comment += "s: " + strings.Join(highlights(relations), ",") + ",..."
		}
		fcs = append(fcs, jen.Empty(), field.queryCode(idx))
		idx++
	}
	for _, field := range m.SpecExtras {
		fcs = append(fcs, field.queryCode(idx, m.doc.getModQual(field.getType())))
		idx++
	}

	st := jen.Type().Id(tname).Struct(fcs...).Line()
	if len(fcs) > 2 {
		isPG10 := m.doc.IsPG10()
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("Sift").Params(args...).Params(rets...)
		st.BlockFunc(func(g *jen.Group) {
			// if m.IsBsonable() || m.doc.IsMongo() {
			// 	g.Var().Id("qd").Id("BD")
			// }
			if len(relFields) > 0 && !okAL {

				if len(relFields) == 1 {
					g.If(jen.Id("spec").Dot(withRel).Op("==").Lit("1").Op("||").Id("spec").Dot(withRel).Op("==").Lit(relFields[0].Name)).Block(
						jen.Id("q").Dot("Relation").Call(jen.Lit(relFields[0].Name)),
					)
				} else {
					jcond := jen.Len(jen.Id("spec").Dot(withRel)).Op(">0")
					jrels := make([]jen.Code, len(relFields))
					for i, relField := range relFields {
						jrels[i] = jen.Lit(relField.Name)
					}
					g.If(jcond).BlockFunc(func(g2 *jen.Group) {
						g2.For(jen.Id("_,rel").Op(":=").Range().Qual("strings", "Split").Call(jen.Id("spec").Dot(withRel), jen.Lit(","))).BlockFunc(func(g3 *jen.Group) {
							g3.Switch(jen.Id("rel")).BlockFunc(func(gs *jen.Group) {
								gs.Case(jrels...).Block(
									jen.Id("q").Dot("Relation").Call(jen.Id("rel")),
								)
							})
						})
					})

					// g.Switch(jen.Id("spec").Dot(withRel)).BlockFunc(func(gs *jen.Group) {
					// 	for _, relField := range relFields {
					// 		gs.Case(jen.Lit(relField.Name)).Block(
					// 			jen.Id("q").Dot("Relation").Call(jen.Lit(relField.Name)),
					// 		)
					// 	}
					// })

				}
				g.Line()
			}
			g.Add(jfSiftCall("ModelSpec"))

			if m.hasAudit() {
				g.Add(jfSiftCall("AuditSpec"))
			}
			for _, sifter := range m.Sifters {
				g.Add(jfSiftCall(sifter))
			}

			for i := 0; i < len(specFields); i++ {
				field := specFields[i]
				fieldM := field
				if field.multable { // ints, strs, oids
					field = specFields[i+1]
					i++
				}
				jscs := m.genSiftCode(field, fieldM, isPG10, len(withRel) > 0)
				if jscs != nil {
					g.Add(jscs)
				}

			}
			if okTS || len(colTS) > 0 {
				// g.Add(jfSiftCall("TextSearchSpec"))
				g.Id("q").Op("=").Id("spec").Dot("TextSearchSpec").Dot("SiftTS").Call(
					jen.Id("q"), jen.Op("!spec").Dot("HasColumn").Call())
			}
			g.Line()

			if isPG10 {
				g.Return(jen.Id("q"), jen.Nil())
			} else {
				g.Return(jen.Id("q"))
			}

		}).Line()
	}

	if cols := m.sortableColumns(); len(cols) > 0 {
		log.Printf("sortable: %+v", cols)
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("CanSort").Params(jen.Id("k").Id("string")).Bool()
		st.BlockFunc(func(g *jen.Group) {
			g.Switch(jen.Id("k")).BlockFunc(func(g1 *jen.Group) {
				g1.Case(jen.ListFunc(func(g2 *jen.Group) {
					for _, s := range cols {
						g2.Lit(s)
					}
				})).Return(jen.True())
				g1.Default().Return(jen.Id("spec").Dot("ModelSpec").Dot("CanSort").Call(jen.Id("k")))
			})
		}).Line()
	}

	return st
}

func (m *Model) hasStoreHook(k string) (sh storeHook, ok bool) {
	var v string
	if v, ok = m.StoHooks[k]; ok {
		sh, ok = ParseHook(m.Name, m.HookNs, k, v)
	}
	return
}

func (m *Model) hasAnyStoreHook(a ...string) bool {
	for _, k := range a {
		if _, ok := m.hasStoreHook(k); ok {
			return true
		}
	}
	return false
}

func (m *Model) storeHookName(k, v string) (string, bool) {
	sh, ok := ParseHook(m.Name, m.HookNs, k, v)
	return sh.FunName, ok
}

func (m *Model) StoreHooks() (out []storeHook) {
	for k, v := range m.StoHooks {
		if len(v) == 0 {
			continue
		}
		sh, ok := ParseHook(m.Name, m.HookNs, k, v)
		if !ok {
			continue
		}
		sh.m = m

		out = append(out, sh)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].FunName > out[j].FunName
	})

	return out
}

func (m *Model) shortComment() string {
	s := m.Comment
	if a, _, ok := strings.Cut(s, " "); ok {
		return a
	}
	return s
}

func metaUpCode(a ...bool) jen.Code {
	tags := Tags{"json": "metaUp,omitempty", TagSwaggerIgnore: "true"}
	if len(a) > 0 && a[0] {
		tags["bson"] = "-"
		tags["bun"] = "-"
		tags["pg"] = "-"
	}
	code := jen.Comment("for meta update").Line()
	code.Id("MetaDiff").Op("*").Add(jen.Id("comm.MetaDiff"))
	code.Tag(tags)
	return code
}

func ownerUpCode(idx int) jen.Code {
	tags := Tags{"json": "ownerID,omitempty"}
	tags.extOrder(idx, 0)
	code := jen.Comment("仅用于更新所有者(负责人)").Line()
	code.Id("OwnerID").Op("*").Id("string")
	code.Tag(tags)
	return code
}

func createdUpCode(idx int) jen.Code {
	code := jen.Comment("创建时间").Line()

	tags := Tags{"json": "created,omitempty"}
	tags.extOrder(idx, 0)
	code.Id(createdField).Op("*").Qual("time", "Time")
	code.Tag(tags)
	return code
}

func (m *Model) HasTextSearch() (cols []string, ok bool) {
	if m.IsBsonable() {
		return
	}
	var hasTs bool
	for _, field := range m.Fields {
		if strings.HasSuffix(field.Query, "fts") {
			cn, _, _ := field.ColName()
			cols = append(cols, cn)
		}
		if matchs(textSearchField, field.Name, field.Type) {
			hasTs = true
		}
	}
	ok = hasTs

	return
}

func (mod *Model) textSearchCodes(id string, isup bool) (jen.Code, bool) {
	st := jen.Empty()
	if cols, ok := mod.HasTextSearch(); ok {
		st.If(jen.Id("tscfg").Op(",").Id("ok").Op(":=").Id("DbTsCheck").Call().Op(";").Id("ok")).BlockFunc(func(g *jen.Group) {
			g.Id(id).Dot("TsCfgName").Op("=").Id("tscfg")
			if !mod.DbTriggerSave && len(cols) > 0 {
				g.Id(id).Dot("SetTsColumns").Call(jen.ListFunc(func(g1 *jen.Group) {
					for _, s := range cols {
						g1.Lit(s)
					}
				}))
			}
			if isup {
				g.Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
			}
		})
		// if id == "exist" {
		// 	// st.Else().Block(jen.Id(id).Dot("TsCfgName").Op("=").Lit(""))
		// 	st.Line().Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
		// }

		return st, true
	}

	return st, false
}

func (m *Model) getIPath() string {
	if m.doc != nil {
		return m.doc.modipath
	}
	return m.pkg
}

func (m *Model) codeNilInstance() jen.Code {
	return jen.Call(jen.Op("*").Qual(m.getIPath(), m.Name)).Call(jen.Nil())
}

type jPair struct {
	p1 jen.Code
	p2 jen.Code
}

func (m *Model) codeRegSto() (ctable jen.Code, cload jPair) {
	if m.IsTable() {
		ctable = m.codeNilInstance()
		if m.RegLoader {
			cload = jPair{
				jen.Qual(m.getIPath(), m.Name+"TypID"),
				jen.Op("*").Qual(m.getIPath(), m.Name),
			}
		}
	}
	return
}

func (m *Model) dbTxFn() string {
	if m.doc.IsPG10() {
		return "RunInTransaction"
	}
	return "RunInTx"
}

func (m *Model) jvdbcall(c rune) (db jen.Code, cn string, isBson bool) {
	if m.IsBsonable() { // mongodb
		isBson = true
		db = jen.Id("s.w.mdb")
		if s, ok := methodsMongo[c]; ok {
			cn = s
		}
	} else { // pgx
		db = jen.Id("s").Dot("w").Dot("db")
		if s, ok := methodsPGx[c]; ok {
			cn = s
		}
	}
	return
}

func (m *Model) codeStoreList(_ Method) ([]jen.Code, []jen.Code, *jen.Statement) {
	// TODO: export
	jdataptr := jen.Op("&").Id("data")
	jspec := jen.Id("spec")
	swdb, mList, isBson := m.jvdbcall('L')
	jargs := []jen.Code{jen.Id("ctx")}
	if isBson {
		jargs = append(jargs, swdb, jen.Qual(m.getIPath(), m.Name+"Collection"), jspec, jdataptr)
	} else {
		jargs = append(jargs, jspec, jdataptr)
	}
	return []jen.Code{jen.Id("spec").Op("*").Id(m.getSpecName())},
		[]jen.Code{jen.Id("data").Qual(m.getIPath(), m.GetPlural()),
			jen.Id("total").Int(), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			if cols, ok := m.HasTextSearch(); ok || len(cols) > 0 {
				if ok {
					g.Id("spec").Dot("SetTsConfig").Call(jen.Id("s.w.db.GetTsCfg").Call())
				}
				if len(cols) > 0 {
					g.Id("spec").Dot("SetTsFallback").Call(jen.ListFunc(func(g1 *jen.Group) {
						for _, s := range cols {
							g1.Lit(s)
						}
					}))
				}
			}

			isPG10 := m.doc.IsPG10()
			if hkBL, okBL := m.hasStoreHook(beforeList); okBL {
				var jcall jen.Code
				if isPG10 {
					jcall = jen.Dot("ModelContext").Call(jen.Id("ctx"), jdataptr)
				} else {
					jcall = jen.Dot("NewSelect").Call().Dot("Model").Call(jdataptr)
				}
				g.Id("q").Op(":=").Add(swdb, jcall) //.Dot("Apply").Call(jen.Id("spec").Dot("Sift"))
				g.If(jen.Err().Op("=").Id("s").Dot(hkBL.FunName).Call(jen.Id("ctx"), jspec, jen.Id("q")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("total").Op(",").Err().Op("=").Id("queryPager").Call(jen.Id("ctx"), jspec, jen.Id("q"))
			} else {
				g.Id("total").Op(",").Id("err").Op("=").Id(mList).Call(
					jargs...,
				)
			}

			if hkAL, okAL := m.hasStoreHook(afterList); okAL {
				jb := new(jen.Statement)
				args := []jen.Code{jen.Id("ctx"), jen.Id("spec")}
				if hkAL.isPtr {
					args = append(args, jen.Id("&data"))
				} else {
					args = append(args, jen.Id("data"))
				}
				isT := hkAL.isTot || strings.HasSuffix(hkAL.FunName, "T")
				if isT {
					jb.Id("total").Op(",")
					args = append(args, jen.Id("total"))
				}
				jb.Err().Op("=").Id("s").Dot(hkAL.FunName).Call(args...)
				g.If(jen.Err().Op("==").Nil()).Block(jb)
			}
			g.Return()
		})
}

func (mod *Model) codeStoreGet(mth Method) (arg []jen.Code, ret []jen.Code, addition jen.Code, blkcode *jen.Statement) {

	utilsQual, _ := mod.doc.getQual("utils")
	arg = []jen.Code{jen.Id("id").String()}
	ret = []jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()}

	jload := jen.Id("err").Op("=")
	swdb, fnGet, isBson := mod.jvdbcall('G')

	jaf := func(g *jen.Group, jdb jen.Code) {
		g.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))

		args := []jen.Code{jen.Id("ctx"), jdb, jen.Id("obj"), jen.Id("id")}
		if mth.Export || mth.ColGet {
			args = append(args, jen.Id("cols").Op("..."))
		}
		jload.Id(fnGet).Call(args...)
		if uf, isuniq := mod.UniqueOne(); isuniq {
			args := []jen.Code{jen.Id("ctx"), jdb, jen.Id("obj")}
			var ukey string
			var fnGet2 string
			// ukey := fmt.Sprintf("%s %s ?", uf.Column, uf.Op())
			if isBson {
				ukey, _ = uf.BsonName()
				fnGet2 = "mgGetWithKey"
				args = append(args, jen.Lit(ukey), jen.Id("id"))
			} else {
				ukey = uf.Column
				fnGet2 = "dbGetWith"
				args = append(args, jen.Lit(ukey), jen.Lit(uf.Op()), jen.Id("id"))
			}
			if mth.Export || mth.ColGet {
				args = append(args, jen.Id("cols").Op("..."))
			}
			g.If(jen.Err().Op("=").Id(fnGet2).Call(args...).
				Op(";").Err().Op("!=").Nil()).Block(jload)
		} else {
			g.Add(jload)
		}
	}
	if mth.Export {
		args := []jen.Code{jactx, jadbO}
		args = append(args, arg...)
		args = append(args, jen.Id("cols").Op("...").String())
		addition = jen.Func().Id(mth.Name).Params(args...).Params(ret...).BlockFunc(func(g *jen.Group) {
			jaf(g, jen.Id("db"))
			g.Return()
		}).Line()
	}

	blkcode = jen.BlockFunc(func(g *jen.Group) {
		if mth.Export && !isBson {
			args := []jen.Code{jen.Id("ctx"), swdb, jen.Id("id")}
			if mth.ColGet {
				args = append(args, jen.Id("ColumnsFromContext").Call(jen.Id("ctx")).Op("..."))
			}
			g.Id("obj").Op(",").Err().Op("=").Id(mth.Name).Call(args...)
		} else {
			if mth.ColGet {
				g.Id("cols").Op(":=").Id("ColumnsFromContext").Call(jen.Id("ctx"))
			}
			jaf(g, swdb)
		}
		jer := jen.Empty()
		if mod.doc.hasQualErrors() {
			jer.If(jen.Id("errorIs").Call(jen.Err(), jen.Id("ErrNotFound"))).Block(
				jen.Err().Op("=").Add(mod.doc.qual("errors.NewErrNotFound")).
					Call(jen.Lit(mod.getLabel()), jen.Id("id")),
			)
		}

		if hkEL, okEL := mod.hasStoreHook(errorLoad); okEL {
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Err().Op("=").Id("s").Dot(hkEL.FunName).Call(jen.Id("ctx"), jen.Id("id"), jen.Err(), jen.Id("obj")),
			)
		}

		if hkAL, okAL := mod.hasStoreHook(afterLoad); okAL {
			g.If(jen.Err().Op("==").Nil()).Block(
				jen.Err().Op("=").Id("s").Dot(hkAL.FunName).Call(jen.Id("ctx"), jen.Id("obj")),
			)
			if mod.doc.hasQualErrors() {
				g.Add(jer)
			}
		} else if rels := mod.Fields.relHasOne(); len(rels) > 0 {
			g.If(jen.Err().Op("!=").Nil()).BlockFunc(func(g1 *jen.Group) {
				if mod.doc.hasQualErrors() {
					g1.Add(jer)
				}
				g1.Return()
			})
			g.For().Op("_,").Id("rn").Op(":=").Range().Id("RelationFromContext").Call(jen.Id("ctx")).BlockFunc(func(g2 *jen.Group) {
				for _, rf := range rels {
					lastName := rf.Name + "ID"
					var jck jen.Code
					if fieldI, ok := mod.Fields.withName(lastName); ok && fieldI.isOID() {
						jck = jen.Id("obj." + lastName).Dot("Valid").Call()
					} else {
						jck = jen.Op("!").Qual(utilsQual, "IsZero").Call(jen.Id("obj." + lastName))
					}
					g2.If(jen.Id("rn").Op("==").Lit(rf.Name).Op("&&").Add(jck)).Block(
						jen.Id("ro").Op(":=").New(rf.typeCode(mod.getIPath())),
						jen.If(jen.Err().Op("=").Id("dbGetWithPKID").Call(
							jen.Id("ctx"), swdb, jen.Id("ro"), jen.Id("obj").Dot(lastName)).Op(";").Err().Op("==").Nil()).Block(
							jen.Id("obj").Dot(rf.Name).Op("=").Id("ro"),
							jen.Continue(),
						),
					)
				}

			})
		} else {
			g.Add(jer)
		}

		g.Return()
	})
	return
}

func (mod *Model) codeStoreCreate(mth Method) (arg []jen.Code, ret []jen.Code, addition jen.Code, blkcode *jen.Statement) {
	tname := mod.Name + "Basic"

	hkBC, okBC := mod.hasStoreHook(beforeCreating)
	hkAC, okAC := mod.hasStoreHook(afterCreating)
	hkBS, okBS := mod.hasStoreHook(beforeSaving)
	hkAS, okAS := mod.hasStoreHook(afterSaving)
	hookTxing := okBC || okAC || okBS || okAS

	isPG10 := mod.doc.IsPG10()

	swdb, fnCreate, _ := mod.jvdbcall('C')

	arg = []jen.Code{jen.Id("in").Qual(mod.getIPath(), tname)}
	ret = []jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()}
	jaf := func(g *jen.Group, jdb jen.Code) {
		nname := "New" + mod.Name + "WithBasic"
		g.Id("obj").Op("=").Qual(mod.getIPath(), nname).Call(jen.Id("in"))

		targs := []jen.Code{jen.Id("ctx"), jdb, jen.Id("obj")}
		jfCheck := func() {
			unfd, isuniq := mod.UniqueOne()
			if isuniq && !mod.IsBsonable() {
				var jcond jen.Code
				if unfd.isOID() {
					jcond = jen.Id("obj").Dot(unfd.Name).Dot("IsZero").Call()
				} else {
					jcond = jen.Id("obj").Dot(unfd.Name).Op("==").Lit("")
				}
				g.If(jcond).Block(
					jen.Err().Op("=").Id("ErrEmptyKey"),
					jen.Return())
				targs = append(targs, jen.Lit(unfd.Column))
			} else if mod.ForceCreate {
				targs = append(targs, jen.Lit(true))
			}
		}

		if jt, ok := mod.textSearchCodes("obj", false); ok {
			g.Add(jt)
		}
		if hookTxing {
			if okBC {
				g.If(jen.Err().Op("=").Id(hkBC.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
					jen.Return(),
				)
			} else if okBS {
				g.If(jen.Err().Op("=").Id(hkBS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
					jen.Return(),
				)
			}
			jfCheck()
			mod.codeMetaUp(g, jdb, "obj")

			g.Err().Op("=").Id(fnCreate).Call(targs...)
			if okAC {
				g.If(jen.Err().Op("==")).Nil().Block(
					jen.Err().Op("=").Id(hkAC.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
				)
			} else if okAS {
				g.If(jen.Err().Op("==")).Nil().Block(
					jen.Err().Op("=").Id(hkAS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
				)
			}

		} else {
			jfCheck()
			mod.codeMetaUp(g, jdb, "obj")

			g.Err().Op("=").Id(fnCreate).Call(targs...)
		}
	}
	efname := mth.getExportAction() + mod.getExportName()
	if mth.Export {
		args := []jen.Code{jactx, jadbO}
		args = append(args, arg...)
		addition = jen.Func().Id(efname).Params(args...).Params(ret...).BlockFunc(func(g *jen.Group) {
			jaf(g, jen.Id("db"))
			g.Return()
		}).Line()
	}

	blkcode = jen.BlockFunc(func(g *jen.Group) {
		jbf := func(g2 *jen.Group, jdb jen.Code) {
			if mth.Export {
				args := []jen.Code{jen.Id("ctx"), jdb, jen.Id("in")}
				g2.Id("obj").Op(",").Err().Op("=").Id(efname).Call(args...)
			} else {
				jaf(g2, jdb)
			}
		}
		if okBC || okBS || okAS {
			g.Err().Op("=").Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
				g1.Id("ctx")
				jxf := func(g3 *jen.Group) {
					jbf(g3, jen.Id("tx"))
					g3.Return(jen.Err())
				}
				if isPG10 {
					g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jxf)
				} else {
					g1.Nil()
					g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jxf)
				}

			})

		} else {
			jbf(g, swdb)
		}

		if hk, ok := mod.hasStoreHook(afterCreated); ok {
			g.If(jen.Err().Op("==").Nil()).Block(
				jen.Err().Op("=").Id("s").Dot(hk.FunName).Call(jen.Id("ctx"), jen.Id("obj")),
			)
		}

		if fc, ok := mod.hasStoreHook(upsertES); ok {
			g.If(jen.Err().Op("==").Nil()).Block(
				jen.Err().Op("=").Id("s").Dot(fc.FunName).Call(jen.Id("ctx"), jen.Id("obj")),
			)
		}

		g.Return()
	})
	return
}

func (mod *Model) codeStoreUpdate(mth Method) (arg []jen.Code, ret []jen.Code, addition jen.Code, blkc *jen.Statement) {
	fnGet := "dbGetWithPKID"
	swdb, fnUpdate, isBson := mod.jvdbcall('U')
	if isBson {
		fnGet = "mgGet"
	}
	hkBU, okBU := mod.hasStoreHook(beforeUpdating)
	hkAU, okAU := mod.hasStoreHook(afterUpdating)
	hkBS, okBS := mod.hasStoreHook(beforeSaving)
	hkAS, okAS := mod.hasStoreHook(afterSaving)
	hookTxing := okBU || okAU || okBS || okAS

	hkAX, okAX := mod.hasStoreHook(afterUpdated)
	hkue, okue := mod.hasStoreHook(upsertES)
	hookTxDone := okAX || okue

	tname := mod.Name + "Set"
	arg = []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)}
	ret = []jen.Code{jen.Error()}

	isPG10 := mod.doc.IsPG10()
	jretf := func(cs ...jen.Code) jen.Code {
		if mth.Export {
			if len(cs) > 0 {
				return jen.Err().Op("=").Add(cs...).Line().Return()
			}
			return jen.Return()
		}
		if len(cs) == 0 {
			cs = []jen.Code{jen.Err()}
		}
		return jen.Return(cs...)
	}
	jaf := func(g *jen.Group, jdb jen.Code, inTx bool) {
		eop := "="
		if !inTx && !mth.Export {
			eop = ":="
		}
		if mth.Export || inTx && hookTxDone {
			g.Id("exist").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
		} else {
			g.Id("exist").Op(":=").New(jen.Qual(mod.getIPath(), mod.Name))
		}
		g.If(jen.Err().Op(eop).Id(fnGet).Call(
			jen.Id("ctx"), jdb, jen.Id("exist"), jen.Id("id"),
		).Op(";").Err().Op("!=").Nil()).Block(jretf())

		if mod.IsBsonable() {
			g.Id("up").Op(eop).Id("exist").Dot("SetWith").Call(jen.Id("in"))
		} else {
			g.Id("exist").Dot("SetIsUpdate").Call(jen.Lit(true))
			g.Id("exist").Dot("SetWith").Call(jen.Id("in"))
		}

		if jt, ok := mod.textSearchCodes("exist", true); ok {
			g.Add(jt)
		}

		jupArgs := []jen.Code{jen.Id("ctx"), jdb, jen.Id("exist")}
		if mod.IsBsonable() {
			jupArgs = append(jupArgs, jen.Id("up"))
		}

		jup := jen.Id(fnUpdate).Call(jupArgs...)

		if hookTxing {
			jcondf := func(eop string, jnbc jen.Code) jen.Code {
				return jen.If(jen.Err().Op(eop).Add(jnbc).Op(";").Err().Op("!=")).Nil().Block(jretf())
			}
			if okBU {
				g.Add(jcondf(eop, jen.Id(hkBU.FunName).Call(jen.Id("ctx"), jdb, jen.Id("exist"))))
			} else if okBS {
				g.Add(jcondf(eop, jen.Id(hkBS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("exist"))))
			}

			mod.codeMetaUp(g, jdb, "exist")

			if okAU {
				g.Add(jcondf(eop, jup))
				g.Add(jretf(jen.Id(hkAU.FunName).Call(jen.Id("ctx"), jdb, jen.Id("exist"))))
			} else if okAS {
				g.Add(jcondf(eop, jup))
				g.Add(jretf(jen.Id(hkAS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("exist"))))
			} else if mth.Export {
				g.Err().Op("=").Add(jup)
				g.Return()
			} else {
				g.Return(jup)
			}
		} else {
			mod.codeMetaUp(g, jdb, "exist")

			if hookTxDone {
				jst := jen.Return()
				if !mth.Export {
					jst.Err()
				}
				g.If(jen.Err().Op(eop).Add(jup).Op(";").Err().Op("!=").Nil()).Block(
					jst,
				)
				if mth.Export {
					g.Return()
				}
			} else {
				g.Add(jretf(jup))
			}

		}
	}
	efname := mth.action + mod.getExportName()
	if mth.Export {
		args := []jen.Code{jactx, jadbO}
		args = append(args, arg...)
		rets := []jen.Code{jen.Id("exist").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()}
		addition = jen.Func().Id(efname).Params(args...).Params(rets...).BlockFunc(func(g *jen.Group) {
			jaf(g, jen.Id("db"), false)
		}).Line()
	}

	blkc = jen.BlockFunc(func(g *jen.Group) {
		jbf := func(g2 *jen.Group, jdb jen.Code, inTx bool) {
			if mth.Export {
				op := ":="
				if inTx {
					op = "="
				}
				args := []jen.Code{jen.Id("ctx"), jdb, jen.Id("id"), jen.Id("in")}
				if hookTxDone {
					g2.Id("exist").Op(",").Err().Op(op).Id(efname).Call(args...)
				} else {
					g2.Id("_").Op(",").Err().Op(op).Id(efname).Call(args...)
				}
			} else {
				jaf(g2, jdb, inTx)
			}
		}
		if hookTxing {
			jfbd := jen.Empty()
			if hookTxDone {
				g.Var().Id("exist").Op("*").Qual(mod.getIPath(), mod.Name)
			}
			jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
				g1.Id("ctx")
				jxf := func(g3 *jen.Group) {
					jbf(g3, jen.Id("tx"), true)
					if mth.Export {
						g3.Return(jen.Err())
					}
				}
				if isPG10 {
					g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jxf)
				} else {
					g1.Nil()
					g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jxf)
				}
			})
			if hookTxDone {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
			} else {
				g.Return(jfbd)
			}

		} else {
			jbf(g, swdb, false)
			if mth.Export && hookTxDone {
				g.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
			}
		}

		if okAX && okue {
			callau := jen.Id("s").Dot(hkAX.FunName).Call(jen.Id("ctx"), jen.Id("exist"))
			g.If(jen.Err().Op(":=").Add(callau).Op(";").Err().Op("!=").Nil()).Block(
				jen.Return(jen.Err()),
			)
			callke := jen.Id("s").Dot(hkue.FunName).Call(jen.Id("ctx"), jen.Id("exist"))
			g.Return(callke)
		} else if okAX {
			callau := jen.Id("s").Dot(hkAX.FunName).Call(jen.Id("ctx"), jen.Id("exist"))
			g.Return(callau)
		} else if okue {
			callke := jen.Id("s").Dot(hkue.FunName).Call(jen.Id("ctx"), jen.Id("exist"))
			g.Return(callke)
		} else if mth.Export && !hookTxing {
			g.Return(jen.Err())
		}
	})
	return
}

func (mod *Model) codeStorePut(isSimp bool) ([]jen.Code, []jen.Code, *jen.Statement) {
	jqset := jen.Qual(mod.getIPath(), mod.Name+"Set")
	jqobp := jen.Op("*").Qual(mod.getIPath(), mod.Name)
	var jret *jen.Statement
	if isSimp {
		jret = jen.Id("nid").String()
	} else {
		jret = jen.Id("obj").Add(jqobp)
	}
	// log.Printf("jret: %s, %+v", mod.Name, jret)
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Add(jqset)},
		[]jen.Code{jret, jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {

			if isSimp {
				g.Var().Id("obj").Add(jqobp)
			}
			cpms := []jen.Code{
				jen.Id("ctx"), swdb, jen.Id("in"),
			}

			uf, isuniq := mod.UniqueOne()
			if isuniq {
				g.If(jen.Id("in").Dot(uf.Name).Op("==").Nil().Op("||*").Id("in").Dot(uf.Name).Op("==").Lit("")).Block(
					jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("need "+LcFirst(uf.Name))),
					jen.Return())
				cpms = append(cpms, jen.Op("*").Id("in").Dot(uf.Name), jen.Lit(uf.Column))
			} else {
				cpms = append(cpms, jen.Id("id"))
			}
			pgxQual, _ := mod.doc.getQual("pgx")
			jpre := jen.Id("obj").Op(",").Err().Op("=").Qual(pgxQual, "StoreWithSet").Index(jen.Add(jqobp))
			jCallStore := func() jen.Code {
				arg := make([]jen.Code, 4)
				copy(arg, cpms[0:3])
				arg[3] = jen.Id("id")
				return jen.If(jen.Len(jen.Id("id")).Op(">0")).Block(
					jpre.Clone().Call(arg...)).Else().Block(
					jpre.Clone().Call(cpms...))
			}

			hkBS, okBS := mod.hasStoreHook(beforeSaving)
			hkAS, okAS := mod.hasStoreHook(afterSaving)
			if okBS || okAS {
				g.Err().Op("=").Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					cpms[1] = jdb
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBS {
							g2.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
							g2.Id("obj").Id("SetID").Call(jen.Id("id"))
							g2.Id("obj").Id("SetWith").Call(jen.Id("in"))
							g2.If(jen.Err().Op("=").Id(hkBS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(jen.Err()),
							)
						}

						if isuniq {
							g2.Add(jCallStore())
						} else {
							g2.Add(jpre.Clone().Call(cpms...))
						}

						if okAS {
							g2.If(jen.Err().Op("==")).Nil().Block(
								jen.Err().Op("=").Id(hkAS.FunName).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
							)
						}

						g2.Return(jen.Err())
					}
					g1.Nil()
					g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
				})
			} else {
				if isuniq {
					g.Add(jCallStore())
				} else {
					g.Add(jpre.Clone().Call(cpms...))
				}
			}

			if isSimp {
				g.Id("nid").Op("=").Id("obj").Dot("StringID").Call()
			}
			g.Return()
		})
}

func (mod *Model) codeStoreDelete() ([]jen.Code, []jen.Code, *jen.Statement) {
	swdb, mDelete, _ := mod.jvdbcall('D')

	jqual := jen.Qual(mod.getIPath(), mod.Name)
	jtabl := jen.Qual(mod.getIPath(), mod.Name+"Table")
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			jfbd := jen.Empty()
			g.Id("obj").Op(":=").New(jqual)
			hkBD, okBD := mod.hasStoreHook(beforeDeleting)
			hkAD, okAD := mod.hasStoreHook(afterDeleting)
			if okBD || okAD {
				g.If(jen.Id("err").Op(":=").Id("dbGetWithPKID").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"),
				).Op(";").Id("err").Op("!=").Nil()).Block(jen.Return(jen.Err()))

				jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBD {
							g2.If(jen.Err().Op("=").Id(hkBD.FunName).Call(jen.Id("ctx"), jen.Id("tx"),
								jen.Id("obj")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
						}

						g2.Err().Op("=").Id("dbDeleteM").Call(jen.Id("ctx"), jen.Id("tx"),
							jen.Add(swdb).Dot("Schema").Call(),
							jen.Add(swdb).Dot("SchemaCrap").Call(),
							jen.Id("obj"))
						if okAD {
							g2.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
							g2.Return(jen.Id(hkAD.FunName).Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj")))
						} else {
							g2.Return()
						}
					}
					if mod.doc.IsPG10() {
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					} else {
						g1.Nil()
						g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					}

				})
			} else {
				if mod.doc.IsPG10() {
					g.If(jen.Op("!").Id("obj").Dot("SetID").Call(jen.Id("id"))).Block(
						jen.Return().Qual("fmt", "Errorf").Call(jen.Lit("id: '%s' is invalid"), jen.Id("id")),
					)
					jfbd.Id("s").Dot("w").Dot("db").Dot("OpDeleteAny").Call(
						jen.Id("ctx"), jtabl, jen.Id("obj").Dot("ID"),
					)
				} else {
					jfbd.Id(mDelete).Call(
						jen.Id("ctx"), jen.Id("obj"), jen.Id("id"),
					)
				}

			}

			hkad, okad := mod.hasStoreHook(afterDeleted)
			hkde, okde := mod.hasStoreHook(deleteES)

			if okad && okde {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)

				callad := jen.Id("s").Dot(hkad.FunName).Call(jen.Id("ctx"), jen.Id("obj"))
				g.If(jen.Err().Op(":=").Add(callad).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)

				callde := jen.Id("s").Dot(hkde.FunName).Call(jen.Id("ctx"), jen.Id("obj"))
				g.Return(callde)

			} else if okad {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callad := jen.Id("s").Dot(hkad.FunName).Call(jen.Id("ctx"), jen.Id("obj"))
				g.Return(callad)
			} else if okde {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callde := jen.Id("s").Dot(hkde.FunName).Call(jen.Id("ctx"), jen.Id("obj"))
				g.Return(callde)
			} else {
				g.Return(jfbd)
			}
		})
}

func (m *Model) identityCode() (st *jen.Statement) {
	if m.IsTable() {
		st = new(jen.Statement)
		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityLabel").Params().String().Op("{").Return(jen.Id(m.Name + "Label")).Op("}").Line()
		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityModel").Params().String().Op("{").Return(jen.Id(m.Name + "TypID")).Op("}").Line()

		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityTable").Params().String().Op("{").Return(jen.Id(m.Name + "Table")).Op("}").Line()

		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityAlias").Params().String().Op("{").Return(jen.Id(m.Name + "Alias")).Op("}").Line()
	} else if m.IsBsonable() {
		st = new(jen.Statement)
		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("CollectionName").Params().String().Block(
			jen.Return(jen.Id(m.Name + "Collection")),
		).Line()
	}
	return st
}

func (m *Model) metaAddCodes() (st *jen.Statement) {
	if m.hasMeta() {
		st = new(jen.Statement)
		func(st *jen.Statement, args ...string) {
			for _, suf := range args {
				st.Func().Params(jen.Id("in").Op("*").Id(m.Name+suf)).Id("MetaAddKVs").Params(
					jen.Id("args").Op("...").Any()).Op("*").Id(m.Name+suf).Block(
					jen.Id("in").Dot("MetaDiff").Op("=").Id("comm.MetaDiffAddKVs").Call(
						jen.Id("in").Dot("MetaDiff"), jen.Id("args").Op("...")),
					jen.Return().Id("in"),
				).Line()
			}
		}(st, "Basic", "Set")
	}
	return
}

func (m *Model) codeMetaUp(g *jen.Group, jdb jen.Code, id string) {
	if m.hasMeta() && !m.IsBsonable() {
		g.Id("dbMetaUp").Call(jen.Id("ctx"), jdb, jen.Id(id))
	}
}

func (m *Model) canCompare() bool {
	if !m.WithCompare {
		return false
	}
	for _, field := range m.Fields {
		typ := field.getType()
		if strings.HasSuffix(typ, modelDefault) || strings.HasSuffix(typ, modelDunce) || strings.HasSuffix(typ, modelSerial) {
			return false
		}
		if len(typ) > 0 && typ[0] == '*' { // ptr
			return false
		}
		if !field.isScalar() && !field.isOID() && len(field.Compare) == 0 {
			return false
		}
	}

	return len(m.Fields) > 0
}

func (m *Model) codeEqualTo() (st *jen.Statement) {
	if !m.canCompare() {
		// log.Printf("model %s cannot compare", m.Name)
		return
	}
	var jelems []jen.Code
	for _, field := range m.Fields {
		if field.isScalar() || field.isOID() || field.Compare == CompareScalar {
			jelems = append(jelems, jen.Id("z").Dot(field.Name).Op("==").Id("o").Dot(field.Name))
		} else if field.Compare == CompareEqualTo {
			jelems = append(jelems, jen.Id("z").Dot(field.Name).Dot("EqualTo").Call(jen.Id("o").Dot(field.Name)))
		} else if field.Compare == CompareSliceCmp {
			jelems = append(jelems, jen.Qual("slices", "Compare").Call(
				jen.Id("z").Dot(field.Name),
				jen.Id("o").Dot(field.Name),
			).Op("==0"))
		}
	}
	count := len(jelems)
	if count > 0 {
		rets := jen.Empty()
		for i, jel := range jelems {
			if i > 0 && i < count {
				rets.Op("&&")
			}
			rets.Add(jel)
		}
		st = new(jen.Statement)
		st.Func().Params(
			jen.Id("z").Id(m.Name)).Id("EqualTo").Params(
			jen.Id("o").Id(m.Name)).Bool().Block(
			jen.Return(rets),
		)
	}
	return
}

func (m *Model) init(doc *Document) {
	m.doc = doc
	m.pkg = doc.ModelPkg
	if m.pkg != m.getLabel() {
		m.prefix = m.pkg
	}
	for j := range m.Fields {
		m.Fields[j].mod = m
		f := m.Fields[j]
		if k, _, _ := f.cutType(); len(k) > 0 && len(f.Qual) == 0 {
			if p, ok := doc.Qualified[k]; ok {
				m.Fields[j].Qual = p
			}
		}
	}
	for j := range m.SpecExtras {
		m.SpecExtras[j].mod = m
	}
}
