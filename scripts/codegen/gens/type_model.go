// go:build codegen
package gens

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Model struct {
	Comment    string   `yaml:"comment,omitempty"`
	Name       string   `yaml:"name"`
	Label      string   `yaml:"label"`
	CollName   string   `yaml:"collName,omitempty"` // for mongodb only
	TableTag   string   `yaml:"tableTag,omitempty"` // uptrace/bun & go-pg
	Fields     Fields   `yaml:"fields"`
	Plural     string   `yaml:"plural,omitempty"`
	OIDCat     string   `yaml:"oidcat,omitempty"`
	OIDKey     string   `yaml:"oidKey,omitempty"`
	StoHooks   Tags     `yaml:"hooks,omitempty"`
	SpecExtras Fields   `yaml:"specExtras,omitempty"`
	Sifters    []string `yaml:"sifters,omitempty"`
	SpecUp     string   `yaml:"specUp,omitempty"`

	DiscardUnknown bool `yaml:"discardUnknown,omitempty"` // 忽略未知的列
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

	doc *Document
	pkg string
}

func (m *Model) String() string {
	return m.Name
}

func (m *Model) GetPlural() string {
	if m.Plural == "" {
		m.Plural = Plural(m.Name)
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
	Name       string
	Column     string
	IgnoreCase bool
	Tags       Tags
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
			u.Name = field.Name
			u.Column = cn
			u.IgnoreCase = field.IgnoreCase
			u.Tags = field.Tags
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
		var code *jen.Statement
		if len(field.Comment) > 0 {
			code = jen.Comment(field.Comment).Line()
		} else {
			code = jen.Empty()
		}
		code.Id(field.Name)
		cn, isInDb, _ := field.ColName()
		qn, tn, isptr := field.cutType()

		jcond := jen.Id("o").Dot(field.Name).Op("!=").Nil()
		if field.isScalar() {
			jcond.Op("&&").Id("z").Dot(field.Name).Op("!=").Op("*").Id("o").Dot(field.Name)
		} else if field.Compare == CompareEqualTo {
			jcond.Op("&&!").Id("z").Dot(field.Name).Dot("EqualTo").Call(jen.Id("o").Dot(field.Name))
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
		if s, ok := field.Tags["json"]; ok {
			tags := Tags{"json": s}
			tags.extOrder(idx)
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

	st.Comment("consts of " + m.Name + " " + m.shortComment()).Line()
	st.Const().DefsFunc(func(g *jen.Group) {
		if isTable {
			g.Id(m.Name + "Table").Op("=").Lit(m.tableName())
			g.Id(m.Name + "Alias").Op("=").Lit(m.tableAlias())
		} else if m.IsBsonable() {
			g.Id(m.Name + "Collection").Op("=").Lit(m.CollectionName())
		}

		g.Id(m.Name + "Label").Op("=").Lit(m.getLabel())
	}).Line()

	mcs, bcs := m.Fields.Codes(basicName, m.IsBsonable())
	cs = append(cs, mcs...)
	st.Comment(m.Name + " " + m.Comment).Line()
	var prefix string
	if m.doc != nil {
		prefix = m.doc.ModelPkg
	}

	st.Type().Id(m.Name).Struct(cs...).Add(jen.Comment("@name " + prefix + m.Name)).Line().Line()

	if len(bcs) > 0 {
		st.Type().Id(basicName).Struct(bcs...).Add(jen.Comment("@name " + prefix + basicName)).Line().Line()
	}

	st.Type().Id(m.GetPlural()).Index().Id(m.Name).Line().Line()

	if jhk := m.hookModelCodes(); jhk != nil {
		st.Add(jhk)
	}
	if m.DisableLog {
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("DisableLog").Params().Bool().Block(jen.Return(jen.Lit(true)))
		st.Line()
	}
	if jc := m.basicCodes(); jc != nil {
		st.Add(jc)
	}
	if ic := m.identityCode(); ic != nil {
		st.Add(ic)
	}

	if fields, stmts, rets := m.ChangablCodes(); len(fields) > 0 {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(fields...).Add(jen.Comment("@name " + prefix + changeSetName)).Line().Line()
		// scs = append(scs, jen.Return(jen.Id("z").Dot("CountChange").Call().Op(">0")))
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("SetWith").Params(jen.Id("o").Id(changeSetName)).Params(rets...).Block(
			stmts...,
		).Line()
	}
	if jc := m.metaAddCodes(); jc != nil {
		st.Add(jc)
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
		if strings.HasSuffix(typ, modelDefault) {
			return true, modelDefault, modelDefault
		}
		if strings.HasSuffix(typ, modelDunce) {
			return true, modelDunce, modelDunce
		}
		if strings.HasSuffix(typ, modelSerial) {
			return false, modelSerial, modelSerial
		}

		if strings.Contains(typ, "IDField") { // .IDField, .IDFieldStr
			idf = "IDField"
		} else if strings.HasSuffix(typ, "SerialField") {
			idf = "SerialField"
		} else if strings.HasSuffix(typ, "DateFields") {
			dtf = "DateFields"
		}
	}

	if len(idf) > 0 && len(dtf) > 0 {
		ok = idf == "IDField"
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
	return Underscore(m.GetPlural())
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
		if sfn, ext, ok := f.parseQuery(); ok {
			// log.Printf("name: %s, sfn: %q, ext: %q", f.Name, sfn, ext)
			f.siftExt = ext
			if ext == "ints" || ext == "strs" || ext == "oids" {
				ftyp := "string"
				if ext == "oids" {
					ftyp = "oid.OIDsStr"
				}
				argTag := Plural(f.getArgTag())
				f0 := Field{
					Comment:  f.Comment + " (多值逗号分隔)",
					Type:     ftyp,
					Name:     Plural(f.Name),
					Tags:     Tags{"form": argTag, "json": argTag + ",omitempty"},
					siftExt:  ext,
					multable: true,
				}
				// log.Printf("f0: %+v", f0)
				out = append(out, f0)
			} else if ext == "decode" {
				f.qtype = "string"
				f.Comment += " (支持混合解码)"
			} else if ext == "hasVals" {
				f.Comment += " (多值数字相加)"
			}
			if f.Type == "oid.OID" {
				f.Type = "string"
				f.isOid = true
				if sfn == "siftOIDs" {
					f.siftFn = sfn
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
			} else if f.Type == "bool" {
				f.Type = "string"
				f.siftFn = sfn
			} else {
				f.siftFn = sfn
			}

			out = append(out, f)
		} else if f.isOwner() {
			f0 := Field{
				Comment: "所有者编号 (多值使用逗号分隔)",
				Type:    "string",
				Name:    "OwnerID",
				Tags:    Tags{"form": "owner", "json": "owner,omitempty"},
				siftFn:  "siftOIDs",
				colname: "owner_id",
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

func (m *Model) jSpecBasic() (name, parent string, args []jen.Code, rets []jen.Code,
	jfsc func(on string) jen.Code) {
	name = m.Name + "Spec"
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
		return jen.If(jen.Len(jSV).Op(">0")).Block(
			jen.Var().Id("v").Add(field.typeCode(m.doc.getModQual(field.getType()))),
			jen.If(jen.Err().Op(":=").Id("v").Dot("Decode").Call(jSV).Op(";").Err().Op("==").Nil()).Block(
				jen.Id("q").Op("=").Id("q").Dot("Where").Call(jen.Lit(acn+" = ?"), jen.Id("v")),
			),
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
	if len(relFields) > 0 || len(relations) > 0 || okAL {
		wrTyp = "bool"
		if okAL || len(relations) > 1 {
			wrTyp = "string"
		}
		withRel = "WithRel"
		jtag := "rel"
		field := &Field{
			Name: withRel,
			Type: wrTyp, Tags: Tags{"json": jtag},
			Comment: "include relation column"}
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
				log.Printf("%s belongsTo Names %+v", m.Name, relFields)
				// g.Var().Id("pre").String()
				var jcond jen.Code
				if wrTyp == "bool" {
					jcond = jen.Id("spec").Dot(withRel)
				} else {
					jcond = jen.Len(jen.Id("spec").Dot(withRel)).Op(">0")
				}
				g.If(jcond).BlockFunc(func(g *jen.Group) {
					for _, relField := range relFields {
						g.Id("q").Dot("Relation").Call(jen.Lit(relField.Name))
					}
					// g.Id("pre").Op("=").Lit("?TableAlias.")
				}).Line()
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
				g.Add(jfSiftCall("TextSearchSpec"))
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

func (m *Model) hasStoreHook(k string) (v string, ok bool) {
	if v, ok = m.StoHooks[k]; ok {
		v, ok = m.storeHookName(k, v)
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
	return HookMethod(m.Name, k, v)
}

func (m *Model) StoreHooks() (out []storeHook) {
	for k, v := range m.StoHooks {
		if len(v) == 0 {
			continue
		}
		fn, ok := m.storeHookName(k, v)
		if !ok {
			continue
		}

		out = append(out, storeHook{
			FunName: fn,
			ObjName: m.Name,
			k:       k,
			m:       m,
		})
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
	tags.extOrder(idx)
	code := jen.Comment("仅用于更新所有者(负责人)").Line()
	code.Id("OwnerID").Op("*").Id("string")
	code.Tag(tags)
	return code
}

func createdUpCode(idx int) jen.Code {
	code := jen.Comment("创建时间").Line()

	tags := Tags{"json": "created,omitempty"}
	tags.extOrder(idx)
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

func (mod *Model) textSearchCodes(id string) (jen.Code, bool) {
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
			g.Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
		})
		// if id == "exist" {
		// 	// st.Else().Block(jen.Id(id).Dot("TsCfgName").Op("=").Lit(""))
		// 	st.Line().Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
		// }

		return st, true
	}

	return st, false
}

var (
	swdb  jen.Code = jen.Id("s").Dot("w").Dot("db")
	jactx jen.Code = jen.Id("ctx").Id("context.Context")
	jadb  jen.Code = jen.Id("db").Id("ormDB")
)

func (m *Model) getIPath() string {
	if m.doc != nil {
		return m.doc.modipath
	}
	return m.pkg
}

func (m *Model) codeNilInstance() jen.Code {
	return jen.Call(jen.Op("*").Qual(m.getIPath(), m.Name)).Call(jen.Nil())
}

func (m *Model) dbTxFn() string {
	if m.doc.IsPG10() {
		return "RunInTransaction"
	}
	return "RunInTx"
}

func (m *Model) codestoreList() ([]jen.Code, []jen.Code, *jen.Statement) {
	mList := "s.w.db.ListModel"
	if m.IsBsonable() {
		swdb = jen.Id("s.w.mdb")
		mList = "s.w.listModel"
	}
	return []jen.Code{jen.Id("spec").Op("*").Id(m.Name + "Spec")},
		[]jen.Code{jen.Id("data").Qual(m.getIPath(), m.GetPlural()),
			jen.Id("total").Int(), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			if cols, ok := m.HasTextSearch(); ok || len(cols) > 0 {
				if ok {
					g.Id("spec").Dot("SetTsConfig").Call(jen.Id("DbTsCheck").Call())
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
				jdataptr := jen.Op("&").Id("data")
				jspec := jen.Id("spec")
				var jcall jen.Code
				if isPG10 {
					jcall = jen.Dot("ModelContext").Call(jen.Id("ctx"), jdataptr)
				} else {
					jcall = jen.Dot("NewSelect").Call().Dot("Model").Call(jdataptr)
				}
				g.Id("q").Op(":=").Add(swdb, jcall) //.Dot("Apply").Call(jen.Id("spec").Dot("Sift"))
				g.If(jen.Err().Op("=").Id("s").Dot(hkBL).Call(jen.Id("ctx"), jspec, jen.Id("q")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("total").Op(",").Err().Op("=").Id("queryPager").Call(jen.Id("ctx"), jspec, jen.Id("q"))
			} else {
				g.Id("total").Op(",").Id("err").Op("=").Id(mList).Call(
					jen.Id("ctx"), jen.Id("spec"), jen.Op("&").Id("data"),
				)
			}

			if hkAL, okAL := m.hasStoreHook(afterList); okAL {
				g.If(jen.Err().Op("==").Nil().Op("&&").Len(jen.Id("data")).Op(">0")).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("spec"), jen.Id("data")),
				)
			}
			g.Return()
		})
}

func (mod *Model) codestoreGet() ([]jen.Code, []jen.Code, *jen.Statement) {
	utilsQual, _ := mod.doc.getQual("utils")
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
			jload := jen.Id("err").Op("=")
			fnGet := "dbGet"

			if mod.IsBsonable() {
				fnGet = "mgGet"
				jload.Id(fnGet).Call(jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"))
			} else {
				jload.Add(swdb).Dot("GetModel").Call(
					jen.Id("ctx"), jen.Id("obj"), jen.Id("id"))
			}
			if uf, isuniq := mod.UniqueOne(); isuniq {
				ukey := fmt.Sprintf("%s %s ?", uf.Column, uf.Op())
				if mod.IsBsonable() {
					ukey = uf.Column
				}
				g.If(jen.Err().Op("=").Id(fnGet).Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Lit(ukey), jen.Id("id"),
				).Op(";").Err().Op("!=").Nil()).Block(jload)
			} else {
				g.Add(jload)
			}

			jer := jen.Empty()
			if mod.doc.hasQualErrors() {
				jer.If(jen.Err().Op("==").Id("ErrNotFound")).Block(
					jen.Err().Op("=").Add(mod.doc.qual("errors.NewErrNotFound")).
						Call(jen.Lit(mod.getLabel()), jen.Id("id")),
				)
			}

			if hkEL, okEL := mod.hasStoreHook(errorLoad); okEL {
				g.If(jen.Err().Op("!=").Nil()).Block(
					jen.Err().Op("=").Id("s").Dot(hkEL).Call(jen.Id("ctx"), jen.Id("id"), jen.Err(), jen.Id("obj")),
				)
			}

			if hkAL, okAL := mod.hasStoreHook(afterLoad); okAL {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("obj")),
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
							jen.If(jen.Err().Op("=").Id("getModelWithPKID").Call(
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
}

func (mod *Model) codestoreCreate(mth Method) (arg []jen.Code, ret []jen.Code, acode jen.Code, bcode *jen.Statement) {
	tname := mod.Name + "Basic"

	hkBC, okBC := mod.hasStoreHook(beforeCreating)
	hkAC, okAC := mod.hasStoreHook(afterCreating)
	hkBS, okBS := mod.hasStoreHook(beforeSaving)
	hkAS, okAS := mod.hasStoreHook(afterSaving)
	isPG10 := mod.doc.IsPG10()
	unfd, isuniq := mod.UniqueOne()

	fnCreate := "dbInsert"
	if mod.IsBsonable() {
		fnCreate = "mgCreate"
	}

	arg = []jen.Code{jen.Id("in").Qual(mod.getIPath(), tname)}
	ret = []jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()}
	jaf := func(g *jen.Group, jdb jen.Code) {
		nname := "New" + mod.Name + "WithBasic"
		g.Id("obj").Op("=").Qual(mod.getIPath(), nname).Call(jen.Id("in"))

		targs := []jen.Code{jen.Id("ctx"), jdb, jen.Id("obj")}
		if isuniq && !mod.IsBsonable() {
			g.If(jen.Id("obj").Dot(unfd.Name).Op("==").Lit("")).Block(
				jen.Err().Op("=").Id("ErrEmptyKey"),
				jen.Return())
			targs = append(targs, jen.Lit(unfd.Column))
		} else if mod.ForceCreate {
			targs = append(targs, jen.Lit(true))
		}
		if jt, ok := mod.textSearchCodes("obj"); ok {
			g.Add(jt)
		}
		if okBC || okAC || okBS || okAS {
			if okBC {
				g.If(jen.Err().Op("=").Id(hkBC).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
					jen.Return(),
				)
			} else if okBS {
				g.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
					jen.Return(),
				)
			}
			if mod.hasMeta() && !mod.IsBsonable() {
				g.Id("dbOpModelMeta").Call(jen.Id("ctx"), jdb, jen.Id("obj"))
			}

			g.Err().Op("=").Id(fnCreate).Call(targs...)
			if okAC {
				g.If(jen.Err().Op("==")).Nil().Block(
					jen.Err().Op("=").Id(hkAC).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
				)
			} else if okAS {
				g.If(jen.Err().Op("==")).Nil().Block(
					jen.Err().Op("=").Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
				)
			}

		} else {
			if mod.hasMeta() && !mod.IsBsonable() {
				g.Id("dbOpModelMeta").Call(jen.Id("ctx"), swdb, jen.Id("obj"))
			}

			g.Err().Op("=").Id(fnCreate).Call(targs...)
		}
	}
	if mth.Export {
		args := []jen.Code{jactx, jadb}
		args = append(args, arg...)
		acode = jen.Func().Id(mth.Name).Params(args...).Params(ret...).BlockFunc(func(g *jen.Group) {
			jaf(g, jen.Id("db"))
			g.Return()
		}).Line()
	}

	bcode = jen.BlockFunc(func(g *jen.Group) {
		jbf := func(g2 *jen.Group, jdb jen.Code) {
			if mth.Export {
				args := []jen.Code{jen.Id("ctx"), jdb, jen.Id("in")}
				g2.Id("obj").Op(",").Err().Op("=").Id(mth.Name).Call(args...)
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
				jen.Err().Op("=").Id("s").Dot(hk).Call(jen.Id("ctx"), jen.Id("obj")),
			)
		}

		if fc, ok := mod.hasStoreHook(upsertES); ok {
			g.If(jen.Err().Op("==").Nil()).Block(
				jen.Err().Op("=").Id("s").Dot(fc).Call(jen.Id("ctx"), jen.Id("obj")),
			)
		}

		g.Return()
	})
	return
}

func (mod *Model) codestoreUpdate() ([]jen.Code, []jen.Code, *jen.Statement) {
	fnGet := "getModelWithPKID"
	fnUpdate := "dbUpdate"
	if mod.IsBsonable() {
		fnGet = "mgGet"
		fnUpdate = "mgUpdate"
	}
	tname := mod.Name + "Set"
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("exist").Op(":=").New(jen.Qual(mod.getIPath(), mod.Name))
			g.If(jen.Err().Op(":=").Id(fnGet).Call(
				jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("id"),
			).Op(";").Err().Op("!=").Nil()).Block(jen.Return(jen.Err()))

			if mod.IsBsonable() {
				g.Id("up").Op(":=").Id("exist").Dot("SetWith").Call(jen.Id("in"))
			} else {
				g.Id("exist").Dot("SetWith").Call(jen.Id("in"))
			}

			if jt, ok := mod.textSearchCodes("exist"); ok {
				g.Add(jt)
			}
			isPG10 := mod.doc.IsPG10()

			jfbd := jen.Empty()
			hkBU, okBU := mod.hasStoreHook(beforeUpdating)
			hkAU, okAU := mod.hasStoreHook(afterUpdating)
			hkBS, okBS := mod.hasStoreHook(beforeSaving)
			hkAS, okAS := mod.hasStoreHook(afterSaving)
			if okBU || okAU || okBS || okAS {
				jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						g2.Id("exist").Dot("SetIsUpdate").Call(jen.Lit(true))
						if okBU {
							g2.If(jen.Err().Op("=").Id(hkBU).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						} else if okBS {
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						}
						if mod.hasMeta() {
							g2.Id("dbOpModelMeta").Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("exist"))
						}

						jupArgs := []jen.Code{jen.Id("ctx"), jdb, jen.Id("exist")}
						if mod.IsBsonable() {
							jupArgs = append(jupArgs, jen.Id("up"))
						}

						jup := jen.Id(fnUpdate).Call(jupArgs...)

						if okAU {
							g2.If(jen.Err().Op("=").Add(jup).Op(";").Err().Op("==")).Nil().Block(
								jen.Return().Id(hkAU).Call(jen.Id("ctx"), jdb, jen.Id("exist")),
							)
							g2.Return()
						} else if okAS {
							g2.If(jen.Err().Op("=").Add(jup).Op(";").Err().Op("==")).Nil().Block(
								jen.Return().Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("exist")),
							)
							g2.Return()
						} else {
							g2.Return(jup)
						}
					}
					if isPG10 {
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					} else {
						g1.Nil()
						g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					}
				})

			} else {
				if mod.hasMeta() && !mod.IsBsonable() {
					g.Id("dbOpModelMeta").Call(jen.Id("ctx"), swdb, jen.Id("exist"))
				}
				jupArgs := []jen.Code{jen.Id("ctx"), swdb, jen.Id("exist")}
				if mod.IsBsonable() {
					jupArgs = append(jupArgs, jen.Id("up"))
				}
				jfbd.Id(fnUpdate).Call(jupArgs...)
			}

			hkau, okau := mod.hasStoreHook(afterUpdated)
			hkue, okue := mod.hasStoreHook(upsertES)
			if okau && okue {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callau := jen.Id("s").Dot(hkau).Call(jen.Id("ctx"), jen.Id("exist"))
				g.If(jen.Err().Op(":=").Add(callau).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callke := jen.Id("s").Dot(hkue).Call(jen.Id("ctx"), jen.Id("exist"))
				g.Return(callke)
			} else if okau {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callau := jen.Id("s").Dot(hkau).Call(jen.Id("ctx"), jen.Id("exist"))
				g.Return(callau)
			} else if okue {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callke := jen.Id("s").Dot(hkue).Call(jen.Id("ctx"), jen.Id("exist"))
				g.Return(callke)
			} else {
				g.Return(jfbd)
			}
		})
}

func (mod *Model) codestorePut(isSimp bool) ([]jen.Code, []jen.Code, *jen.Statement) {
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
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
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
								jen.Err().Op("=").Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
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

func (mod *Model) codestoreDelete() ([]jen.Code, []jen.Code, *jen.Statement) {
	mDelete := "s.w.db.DeleteModel"
	if mod.IsBsonable() {
		swdb = jen.Id("s.w.mdb")
		mDelete = "s.w.deleteModel"
	}
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
				g.If(jen.Id("err").Op(":=").Id("getModelWithPKID").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"),
				).Op(";").Id("err").Op("!=").Nil()).Block(jen.Return(jen.Err()))

				jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBD {
							g2.If(jen.Err().Op("=").Id(hkBD).Call(jen.Id("ctx"), jen.Id("tx"),
								jen.Id("obj")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
						}

						g2.Err().Op("=").Id("dbDeleteT").Call(jen.Id("ctx"), jen.Id("tx"),
							jen.Add(swdb).Dot("Schema").Call(),
							jen.Add(swdb).Dot("SchemaCrap").Call(),
							jtabl, jen.Id("obj").Dot("ID"))
						if okAD {
							g2.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
							g2.Return(jen.Id(hkAD).Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj")))
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

				callad := jen.Id("s").Dot(hkad).Call(jen.Id("ctx"), jen.Id("obj"))
				g.If(jen.Err().Op(":=").Add(callad).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)

				callde := jen.Id("s").Dot(hkde).Call(jen.Id("ctx"), jen.Id("obj"))
				g.Return(callde)

			} else if okad {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callad := jen.Id("s").Dot(hkad).Call(jen.Id("ctx"), jen.Id("obj"))
				g.Return(callad)
			} else if okde {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				callde := jen.Id("s").Dot(hkde).Call(jen.Id("ctx"), jen.Id("obj"))
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
		).Id("IdentityLabel").Params().String().Block(
			jen.Return(jen.Id(m.Name + "Label")),
		).Line()

		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityTable").Params().String().Block(
			jen.Return(jen.Id(m.Name + "Table")),
		).Line()

		st.Func().Params(
			jen.Id("_").Op("*").Id(m.Name),
		).Id("IdentityAlias").Params().String().Block(
			jen.Return(jen.Id(m.Name + "Alias")),
		).Line()
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
