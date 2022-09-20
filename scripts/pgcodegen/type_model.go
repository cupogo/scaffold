// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jinzhu/inflection"
)

func qual(args ...string) jen.Code {
	if len(args) == 0 {
		log.Fatal("empty args for qual")
	}
	if len(args) > 1 {
		return jen.Qual(args[0], args[1])
	}
	name := args[0]
	if pos := strings.Index(name, "."); pos > 0 {
		if qual, ok := getQual(name[0:pos]); ok {
			return jen.Qual(qual, name[pos+1:])
		} else {
			log.Printf("get qual %s fail", name)
		}
	}
	return jen.Id(name)
}

type Field struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type,omitempty"`
	Tags    Maps   `yaml:"tags,flow,omitempty"`
	Qual    string `yaml:"qual,omitempty"`
	IsBasic bool   `yaml:"basic,omitempty"`
	IsSet   bool   `yaml:"isset,omitempty"`
	Comment string `yaml:"comment,omitempty"`
	Query   string `yaml:"query,omitempty"` // '', 'equal', 'wildcard'

	IsChangeWith bool `yaml:"changeWith,omitempty"` // has ChangeWith method

	isOid    bool
	isDate   bool
	isIntDt  bool
	siftFn   string
	siftExt  string
	multable bool
	qtype    string
}

func (f *Field) isMeta() bool {
	return f.Name == metaField || f.Type == metaField
}

func (f *Field) isAudit() bool {
	return f.Name == auditField || f.Type == auditField
}

func (f *Field) isScalar() bool {
	if f.Type == "string" || f.Type == "bool" {
		return true
	}

	if strings.HasPrefix(f.Type, "int") || strings.HasPrefix(f.Type, "uint") {
		return true
	}

	if strings.Contains(f.Type, "Money") {
		return true
	}

	return false
}

func (f *Field) typeCode(pkgs ...string) *jen.Statement {
	typ := f.Type
	if len(typ) == 0 {
		typ = f.Name
	}
	if len(f.Qual) > 0 {
		return jen.Qual(f.Qual, typ)
	}
	if a, b, ok := strings.Cut(typ, "."); ok {
		if qual, ok := getQual(a); ok {
			return jen.Qual(qual, b)
		}
		return jen.Qual(a, b)
	}

	if len(pkgs) == 1 {
		if typ[0] == '*' {
			typ = typ[1:]
		}
		return jen.Qual(pkgs[0], typ)
	}
	return jen.Id(typ)
}

func (f *Field) isEmbed() bool {
	return len(f.Name) == 0 || len(f.Type) == 0
}

func (f *Field) preCode() (st *jen.Statement) {
	if len(f.Type) == 0 {
		f.Type = f.Name
		f.Name = ""
	}
	st = jen.Empty()
	if f.isEmbed() {
		st.Line()
	}
	if len(f.Comment) > 0 {
		st.Comment(f.Comment).Line()
	}
	if !f.isEmbed() {
		st.Id(f.Name)
	}

	return st
}

func (f *Field) defCode() jen.Code {
	st := jen.Empty()
	if len(f.Qual) > 0 {
		st.Qual(f.Qual, f.Type)
	} else if a, b, ok := strings.Cut(f.Type, "."); ok {
		if len(a) > 0 && a[0] == '*' {
			st.Op("*")
		}
		// log.Printf("field %s qual: %s", f.Name, f.Type)
		if qual, ok := getQual(a); ok {
			st.Qual(qual, b)
		} else {
			log.Printf("get qual %s fail", f.Type)
		}
	} else {
		st.Id(f.Type)
	}
	return st
}

func (f *Field) Code(idx int) jen.Code {
	st := f.preCode().Add(f.defCode())

	if len(f.Tags) > 0 {
		tags := f.Tags.Copy()
		if j, ok := tags["json"]; ok {
			if a, b, ok := strings.Cut(j, ","); ok {
				if f.isScalar() && !tags.Has("form") {
					tags["form"] = a
				}
				if b == "" && strings.HasSuffix(f.Type, "DateTime") {
					tags["json"] = a + ",omitempty"
				}
			} else if f.isScalar() && !tags.Has("form") {
				tags["form"] = j
			}

			tags.extOrder(idx)
		}

		st.Tag(tags)
	}

	if len(f.Name) == 0 {
		st.Line()
	}

	return st
}

// return column name and is unquie
func (f *Field) ColName() (string, bool) {
	if s, ok := f.Tags["pg"]; ok && len(s) > 0 {
		if a, b, ok := strings.Cut(s, ","); ok {
			if len(a) == 0 {
				a = Underscore(f.Name)
			}
			return a, strings.Contains(b, "unique")
		}
	}
	return Underscore(f.Name), false
}

func (f *Field) relMode() (string, bool) {
	if s, ok := f.Tags["pg"]; ok && len(s) > 0 {
		if s == "rel:has-one" {
			return "has-one", true
		}
	}
	return "", false
}

func (f *Field) getArgTag() string {
	if s, ok := f.Tags["form"]; ok {
		return LcFirst(s)
	}

	if j, ok := f.Tags["json"]; ok {
		if a, _, ok := strings.Cut(j, ","); ok {
			return LcFirst(a)
		}
		return LcFirst(j)
	}

	return LcFirst(f.Name)
}

func (f *Field) queryTypeCode() jen.Code {
	if len(f.Type) > 0 {
		f.Type, _ = getModQual(f.Type)
	}
	return f.defCode()
}

func (f *Field) queryCode(idx int) jen.Code {

	if len(f.Comment) > 0 {
		if f.isDate {
			f.Comment += " + during"
		}
	}
	st := f.preCode()
	if len(f.qtype) > 0 {
		st.Id(f.qtype)
	} else {
		st.Add(f.queryTypeCode())
	}

	tags := f.Tags.Copy()
	if len(tags) > 0 {
		if !tags.Has("form") {
			if j, ok := tags["json"]; ok {
				if a, _, ok := strings.Cut(j, ","); ok {
					tags["form"] = a
				} else {
					tags["form"] = j
				}
			}
		}
		delete(tags, "pg")
		tags.extOrder(idx + 1)
		st.Tag(tags)
	}

	return st
}

type Fields []Field

// Codes return fields code of main and basic
func (z Fields) Codes(basicName string) (mcs, bcs []jen.Code) {
	var hasMeta bool
	var setBasic bool
	for i, field := range z {
		if field.IsSet || field.IsBasic {
			bcs = append(bcs, field.Code(i))
			if !setBasic {
				mcs = append(mcs, jen.Id(basicName).Line())
				setBasic = true
			}
		} else {
			mcs = append(mcs, field.Code(i))
		}
		if field.isMeta() {
			hasMeta = true
		}
	}
	if hasMeta {
		bcs = append(bcs, metaUpCode())
	}
	return
}

func (z Fields) relHasOne() (cols []string) {
	for i := range z {
		if _, ok := z[i].relMode(); ok && i > 0 {
			// 上一个字段必须指向关联的主键
			if z[i-1].Name == z[i].Name+"ID" {
				cols = append(cols, z[i].Name)
			}
		}
	}
	return
}

func (z Fields) withName(name string) (*Field, bool) {
	for _, field := range z {
		if field.Name == name {
			return &field, true
		}
	}
	return nil, false
}

type Model struct {
	Comment  string   `yaml:"comment,omitempty"`
	Name     string   `yaml:"name"`
	TableTag string   `yaml:"tableTag,omitempty"`
	Fields   Fields   `yaml:"fields"`
	Plural   string   `yaml:"plural,omitempty"`
	OIDCat   string   `yaml:"oidcat,omitempty"`
	Hooks    Maps     `yaml:"hooks,omitempty"`
	Sifters  []string `yaml:"sifters,omitempty"`
	SpecUp   string   `yaml:"specUp,omitempty"`

	DiscardUnknown bool `yaml:"discardUnknown,omitempty"` // 忽略未知的列
	WithColumnGet  bool `yaml:"withColumnGet,omitempty"`  // Get时允许定制列

	doc *Document
	pkg string
}

func (m *Model) String() string {
	return m.Name
}

func (m *Model) GetPlural() string {
	if m.Plural == "" {
		m.Plural = inflection.Plural(m.Name)
	}
	return m.Plural
}

func (m *Model) tableName() string {
	tt := m.TableTag
	if tt == "" {
		tt = Underscore(m.GetPlural())
	} else if pos := strings.Index(tt, ","); pos > 0 {
		tt = tt[0:pos]
	}
	return tt
}

func (m *Model) TableField() jen.Code {
	tt := m.TableTag
	if tt == "" {
		tt = Underscore(m.GetPlural())
	}
	if m.DiscardUnknown && !strings.Contains(tt, "discard_unknown_columns") {
		tt += ",discard_unknown_columns"
	}
	return jen.Id("tableName").Add(jen.Struct()).Tag(Maps{"pg": tt}).Line()
}

func (m *Model) UniqueOne() (name, col string, onlyOne bool) {
	var count int
	for _, field := range m.Fields {
		if cn, ok := field.ColName(); ok {
			count++
			name = field.Name
			col = cn
		}
	}
	onlyOne = count == 1
	return
}

func (m *Model) ChangablCodes() (ccs []jen.Code, scs []jen.Code) {
	var hasMeta bool
	for idx, field := range m.Fields {
		if !field.IsSet || field.Type == "" || field.Name == "" {
			if field.isMeta() {
				hasMeta = true
			}
			continue
		}
		var code *jen.Statement
		if len(field.Comment) > 0 {
			code = jen.Comment(field.Comment).Line()
		} else {
			code = jen.Empty()
		}
		code.Id(field.Name)
		cn, _ := field.ColName()
		tn := field.Type
		if len(tn) == 0 {
			tn = field.Name
		} else if field.Type == "oid.OID" {
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
			tags := Maps{"json": s}
			tags.extOrder(idx)
			code.Tag(tags)
		}

		ccs = append(ccs, code)
		scs = append(scs, jen.If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			csst := jen.Id("cs").Op("=").Append(jen.Id("cs"), jen.Lit(cn))
			if field.isOid {
				g.If(jen.Id("id").Op(",").Err().Op(":=").Id("oid").Dot("CheckID").Call(jen.Op("*").Id("o").Dot(field.Name)).Op(";").Err().Op("==").Nil().Block(
					jen.Id("z").Dot(field.Name).Op("=").Id("id"), csst,
				))
			} else if field.IsChangeWith {
				g.If(jen.Id("z").Dot(field.Name).Dot("ChangeWith").Call(jen.Id("o").Dot(field.Name))).Block(csst)
			} else {
				g.Add(jen.Id("z").Dot(field.Name).Op("=").Op("*").Id("o").Dot(field.Name).Line(), csst)
			}
		}))
	}
	if hasMeta {
		name := "MetaDiff"
		ccs = append(ccs, metaUpCode())
		scs = append(scs, jen.If(jen.Id("o").Dot(name).Op("!=").Nil().Op("&&").Id("z").Dot("MetaUp").Call(jen.Id("o").Dot(name))).Block(
			jen.Id("cs").Op("=").Append(jen.Id("cs"), jen.Lit("meta")),
		))
	}
	scs = append(scs, jen.If(jen.Len(jen.Id("cs")).Op(">").Lit(0)).Block(
		jen.Id("z").Dot("SetChange").Call(jen.Id("cs").Op("...")),
	))
	return
}

func (m *Model) Codes() jen.Code {
	basicName := m.Name + "Basic"
	var cs []jen.Code
	cs = append(cs, m.TableField())
	mcs, bcs := m.Fields.Codes(basicName)
	cs = append(cs, mcs...)
	st := jen.Comment(m.Name + " " + m.Comment).Line()

	st.Type().Id(m.Name).Struct(cs...).Add(jen.Comment("@name " + m.Name)).Line().Line()

	if len(bcs) > 0 {
		st.Type().Id(basicName).Struct(bcs...).Add(jen.Comment("@name " + basicName)).Line().Line()
	}

	st.Type().Id(m.GetPlural()).Index().Id(m.Name).Line().Line()

	if jhk := m.hookModelCodes(); jhk != nil {
		st.Add(jhk)
	}

	if ccs, scs := m.ChangablCodes(); len(ccs) > 0 {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(ccs...).Add(jen.Comment("@name " + changeSetName)).Line().Line()
		scs = append(scs, jen.Return())
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("SetWith").Params(jen.Id("o").Id(changeSetName)).Params(
			jen.Id("cs").Index().String(), // TODO: return bool or nil
		).Block(
			scs...,
		)
	}
	return st
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

func (m *Model) hasHooks() (bool, string) {
	var hasIDField bool
	var hasDateFields bool
	for _, field := range m.Fields {
		if strings.HasSuffix(field.Name, modelDefault) {
			return true, modelDefault
		}
		if strings.HasSuffix(field.Name, modelDunce) {
			return true, modelDunce
		}

		if strings.Contains(field.Name, "IDField") {
			hasIDField = true
		} else if strings.HasSuffix(field.Name, "DateFields") {
			hasDateFields = true
		}
	}

	if hasIDField && hasDateFields {
		return true, "DateFields"
	}
	return false, ""
}

func (m *Model) hookModelCodes() jen.Code {
	var st *jen.Statement
	if hasHooks, field := m.hasHooks(); hasHooks {
		log.Printf("model %s has hooks", m.Name)
		st = new(jen.Statement)
		st.Comment("Creating function call to it's inner fields defined hooks").Line()
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("Creating").Params().Error().Block(
			jen.If(jen.Id("z").Dot("IsZeroID").Call()).BlockFunc(func(g *jen.Group) {
				switch field {
				case modelDefault:
					oidcat := CamelCased(m.OIDCat)
					if oidcat == "" {
						oidcat = "Default"
					}
					g.Id("z").Dot("SetID").Call(
						jen.Qual(oidQual, "NewID").Call(jen.Qual(oidQual, "Ot"+oidcat)),
					)
				default:
					g.Return(jen.Id("comm").Dot("ErrEmptyID"))
				}
			}).Line(),
			jen.Return(jen.Id("z").Dot(field).Dot("Creating").Call()),
		).Line()

		// st.Comment("Saving function call to it's inner fields defined hooks").Line()
		// st.Func().Params(
		// 	jen.Id("z").Op("*").Id(m.Name),
		// ).Id("Saving").Params().Error().Block(
		// 	jen.Return(jen.Id("z").Dot(field).Dot("Saving").Call()),
		// ).Line()
	}
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
					Tags:     Maps{"form": argTag, "json": argTag},
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
			} else if strings.HasSuffix(f.Type, "DateTime") {
				f.Type = "string"
				f.isDate = true
				f.isIntDt = true
				f.siftFn = "siftDate"
			} else if strings.HasSuffix(f.Type, "Time") {
				f.Type = "string"
				f.isDate = true
				f.siftFn = "siftDate"
			} else {
				f.siftFn = sfn
			}

			out = append(out, f)
		}
	}
	return
}

func (m *Model) getSpecCodes() jen.Code {
	comm, _ := doc.getQual("comm")
	var fcs []jen.Code
	fcs = append(fcs, jen.Qual(comm, "PageSpec"), jen.Id("ModelSpec"))
	if m.hasAudit() {
		fcs = append(fcs, jen.Id("AuditSpec"))
	}
	for _, sifter := range m.Sifters {
		fcs = append(fcs, jen.Id(sifter))
	}
	_, okTS := m.HasTextSearch()
	if okTS {
		fcs = append(fcs, jen.Id("TextSearchSpec"))
	}

	specFields := m.specFields()
	if len(specFields) > 0 {
		// log.Printf("specFields: %+v", specFields)
		fcs = append(fcs, jen.Empty())
		for i, field := range specFields {
			delete(field.Tags, "binding")
			delete(field.Tags, "extensions")
			fcs = append(fcs, field.queryCode(i))
		}

	}

	var withRel string
	_, okAL := m.hasHook(afterList)
	relNames := m.Fields.relHasOne()
	if len(relNames) > 0 || okAL {
		ftyp := "bool"
		if okAL {
			ftyp = "string"
		}
		withRel = "WithRel"
		jtag := "rel"
		field := &Field{
			Name: withRel,
			Type: ftyp, Tags: Maps{"json": jtag},
			Comment: "include relation column"}
		fcs = append(fcs, jen.Empty(), field.queryCode(len(specFields)))
	}

	tname := m.Name + "Spec"
	st := jen.Type().Id(tname).Struct(fcs...).Line()
	if len(fcs) > 2 {
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("Sift").Params(jen.Id("q").Op("*").Id("ormQuery")).
			Params(jen.Op("*").Id("ormQuery"), jen.Error())
		st.BlockFunc(func(g *jen.Group) {
			if len(relNames) > 0 && !okAL {
				log.Printf("%s relNames %+v", m.Name, relNames)
				g.If(jen.Id("spec").Dot(withRel)).BlockFunc(func(g *jen.Group) {
					for _, relName := range relNames {
						g.Id("q").Dot("Relation").Call(jen.Lit(relName))
					}
				}).Line()
			}
			g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("ModelSpec").Dot("Sift").Call(jen.Id("q"))
			if m.hasAudit() {
				g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("AuditSpec").Dot("sift").Call(jen.Id("q"))
			}
			for _, sifter := range m.Sifters {
				g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot(sifter).Dot("Sift").Call(jen.Id("q"))
			}

			for i := 0; i < len(specFields); i++ {
				field := specFields[i]
				fieldM := field
				if field.multable { // ints, strs, oids
					field = specFields[i+1]
					i++
				}
				cn, _ := field.ColName()
				params := []jen.Code{jen.Id("q"), jen.Lit(cn), jen.Id("spec").Dot(field.Name)}
				cfn := field.siftFn
				if field.isDate && field.isIntDt {
					params = append(params, jen.True())
				}
				params = append(params, jen.False())
				jq := jen.Id("q").Op(",").Id("_").Op("=").Id(cfn).Call(params...)
				if field.siftExt == "decode" {
					g.If(jen.Len(jen.Id("spec").Dot(field.Name)).Op(">0")).Block(
						jen.Var().Id("v").Add(field.queryTypeCode()),
						jen.If(jen.Err().Op(":=").Id("v").Dot("Decode").Call(jen.Id("spec").Dot(field.Name)).Op(";").Err().Op("==").Nil()).Block(
							jen.Id("q").Op("=").Id("q").Dot("Where").Call(jen.Lit(cn+" = ?"), jen.Id("v")),
						),
					)
				} else if field.siftExt == "hasVals" {
					g.If(jen.Id("vals").Op(":=").Id("spec").Dot(field.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
						jen.Id("q").Op("=").Id("q").Dot("WhereIn").Call(jen.Lit(cn+" IN(?)"), jen.Id("vals")),
					).Else().Block(jq)
				} else if field.siftExt == "ints" {
					g.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseInts").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
						jen.Id("q").Op("=").Id("q").Dot("WhereIn").Call(jen.Lit(cn+" IN(?)"), jen.Id("vals")),
					).Else().Block(jq)
				} else if field.siftExt == "strs" {
					g.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseStrs").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
						jen.Id("q").Op("=").Id("q").Dot("WhereIn").Call(jen.Lit(cn+" IN(?)"), jen.Id("vals")),
					).Else().Block(jq)
				} else if field.siftExt == "oids" {
					g.If(jen.Id("vals").Op(":=").Id("spec").Dot(fieldM.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
						jen.Id("q").Op("=").Id("q").Dot("WhereIn").Call(jen.Lit(cn+" IN(?)"), jen.Id("vals")),
					).Else().Block(jq)
				} else {
					g.Add(jq)
				}

				// TODO: set text wildcard
			}
			if okTS {
				g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("TextSearchSpec").Dot("Sift").Call(jen.Id("q"))
			}
			g.Line()
			g.Return(jen.Id("q"), jen.Nil())
		}).Line()
	}

	return st
}

func (m *Model) hasHook(k string) (v string, ok bool) {
	v, ok = m.Hooks[k]
	return
}

func metaUpCode() jen.Code {
	code := jen.Comment("for meta update").Line()
	code.Id("MetaDiff").Op("*").Add(qual("comm.MetaDiff"))
	code.Tag(Maps{"bson": "-", "json": "metaUp,omitempty", "pg": "-", "swaggerignore": "true"})
	return code
}

func (m *Model) HasTextSearch() (cols []string, ok bool) {
	for _, field := range m.Fields {
		if field.Query == "fts" {
			cn, _ := field.ColName()
			cols = append(cols, cn)
		}
		if field.Name == textSearchField || field.Type == textSearchField {
			ok = true
		}
	}

	return
}

func (f *Field) parseQuery() (fn, ext string, ok bool) {
	var a string
	a, ext, _ = strings.Cut(f.Query, ",")
	switch a {
	case "oids":
		fn, ok = "siftOIDs", f.Type == "oid.OID"
	case "equal":
		fn, ok = "siftEquel", true
	case "ice", "ilike":
		fn, ok = "siftILike", true
	case "match":
		fn, ok = "siftMatch", true
	case "date":
		fn, ok = "siftDate", true
	case "great":
		fn, ok = "siftGreat", true
	case "less":
		fn, ok = "siftLess", true
	}
	return
}

var (
	swdb = jen.Id("s").Dot("w").Dot("db")
)

func (m *Model) getIPath() string {
	if m.doc != nil {
		return m.doc.modipath
	}
	return m.pkg
}

func (m *Model) codestoreList() ([]jen.Code, []jen.Code, *jen.Statement) {
	return []jen.Code{jen.Id("spec").Op("*").Id(m.Name + "Spec")},
		[]jen.Code{jen.Id("data").Qual(m.getIPath(), m.GetPlural()),
			jen.Id("total").Int(), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			if cols, ok := m.HasTextSearch(); ok {
				g.Id("spec").Dot("SetTsConfig").Call(jen.Add(swdb).Dot("GetTsCfg").Call())
				if len(cols) > 0 {
					g.Id("spec").Dot("SetTsFallback").Call(jen.ListFunc(func(g1 *jen.Group) {
						for _, s := range cols {
							g1.Lit(s)
						}
					}))
				}
			}
			jq := jen.Add(swdb).Dot("Model").Call(
				jen.Op("&").Id("data")).Dot("Apply").Call(
				jen.Id("spec").Dot("Sift"))

			g.Id("total").Op(",").Id("err").Op("=").Id("queryPager").Call(
				jen.Id("spec"), jq,
			)

			if hkAL, okAL := m.hasHook(afterList); okAL {
				g.If(jen.Err().Op("==").Nil().Op("&&").Len(jen.Id("data")).Op(">0")).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("spec"), jen.Id("data")),
				)
			}
			g.Return()
		})
}

func (mod *Model) codestoreGet() ([]jen.Code, []jen.Code, *jen.Statement) {
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			params := []jen.Code{jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id")}
			if mod.WithColumnGet {
				params = append(params, jen.Id("ColumnsFromContext").Call(jen.Id("ctx")).Op("..."))
			}
			g.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
			jload := jen.Id("err").Op("=").Id("getModelWithPKID").Call(params...)
			if _, cn, isuniq := mod.UniqueOne(); isuniq {
				g.If(jen.Err().Op("=").Id("getModelWithUnique").Call(
					swdb, jen.Id("obj"), jen.Lit(cn), jen.Id("id"),
				).Op(";").Err().Op("!=").Nil()).Block(jload)
			} else {
				g.Add(jload)
			}

			if hkAL, okAL := mod.hasHook(afterLoad); okAL {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("obj")),
				)
			} else if rels := mod.Fields.relHasOne(); len(rels) > 0 {
				g.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
				g.For().Op("_,").Id("rn").Op(":=").Range().Id("RelationFromContext").Call(jen.Id("ctx")).BlockFunc(func(g2 *jen.Group) {
					for _, rn := range rels {
						field, _ := mod.Fields.withName(rn)
						g2.If(jen.Id("rn").Op("==").Lit(rn).Op("&&!").Qual(utilsQual, "IsZero").Call(jen.Id("obj."+rn+"ID"))).Block(
							jen.Id("ro").Op(":=").New(field.typeCode(mod.getIPath())),
							jen.If(jen.Err().Op("=").Id("getModelWithPKID").Call(
								jen.Id("ctx"), swdb, jen.Id("ro"), jen.Id("obj").Dot(rn+"ID")).Op(";").Err().Op("==").Nil()).Block(
								jen.Id("obj").Dot(rn).Op("=").Id("ro"),
								jen.Continue(),
							),
						)
					}

				})
			}
			g.Return()
		})
}

func (mod *Model) codestoreCreate() ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Basic"
	return []jen.Code{jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op("=&").Qual(mod.getIPath(), mod.Name).Block(jen.Id(tname).Op(":").Id("in").Op(","))

			if mod.hasMeta() {
				g.Id("s").Dot("w").Dot("opModelMeta").Call(jen.Id("ctx"),
					jen.Id("obj"), jen.Id("obj").Dot("MetaDiff"))
			}
			if _, ok := mod.HasTextSearch(); ok {
				g.If(jen.Id("tscfg").Op(",").Id("ok").Op(":=").Add(swdb).Dot("GetTsCfg").Call().Op(";").Id("ok")).Block(
					jen.Id("obj").Dot("TsCfgName").Op("=").Id("tscfg"),
				)
			}
			targs := []jen.Code{jen.Id("ctx"), swdb, jen.Id("obj")}
			if _, cn, isuniq := mod.UniqueOne(); isuniq {
				targs = append(targs, jen.Lit(cn))
			}

			hkBC, okBC := mod.hasHook(beforeCreating)
			hkBS, okBS := mod.hasHook(beforeSaving)
			hkAS, okAS := mod.hasHook(afterSaving)
			if okBC || okBS || okAS {
				g.Err().Op("=").Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					targs[1] = jdb
					g1.Id("ctx")
					g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {
						if okBC {
							g2.If(jen.Err().Op("=").Id(hkBC).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(jen.Err()),
							)
						} else if okBS {
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(jen.Err()),
							)
						}
						g2.Err().Op("=").Id("dbInsert").Call(targs...)
						if okAS {
							g2.If(jen.Err().Op("==")).Nil().Block(
								jen.Err().Op("=").Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
							)
						}

						g2.Return(jen.Err())
					})

				})

			} else {
				g.Err().Op("=").Id("dbInsert").Call(targs...)
			}

			if hk, ok := mod.hasHook(afterCreated); ok {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id(hk).Call(jen.Id("ctx"), jen.Id("s").Dot("w"), jen.Id("obj")),
				)
			}

			g.Return()
		})
}

func (mod *Model) codestoreUpdate() ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Set"
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("exist").Op(":=").New(jen.Qual(mod.getIPath(), mod.Name))
			g.If(jen.Id("err").Op(":=").Id("getModelWithPKID").Call(
				jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("id"),
			).Op(";").Err().Op("!=").Nil()).Block(jen.Return(jen.Err()))

			g.Id("_").Op("=").Id("exist").Dot("SetWith").Call(jen.Id("in"))

			hkBU, okBU := mod.hasHook(beforeUpdating)
			hkBS, okBS := mod.hasHook(beforeSaving)
			hkAS, okAS := mod.hasHook(afterSaving)
			if okBU || okBS || okAS {
				g.Return().Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					g1.Id("ctx")
					g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {
						if okBU {
							g2.If(jen.Err().Op("=").Id(hkBU).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						} else if okBS {
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						}
						jup := jen.Id("dbUpdate").Call(
							jen.Id("ctx"), jdb, jen.Id("exist"),
						)

						if okAS {
							g2.If(jen.Err().Op("=").Add(jup).Op(";").Err().Op("==")).Nil().Block(
								jen.Return().Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("exist")),
							)
							g2.Return()
						} else {
							g2.Return(jup)
						}

					})

				})

			} else {
				g.Return().Id("dbUpdate").Call(
					jen.Id("ctx"), swdb, jen.Id("exist"),
				)
			}
		})
}

func (mod *Model) codestorePut(isSimp bool) ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Set"
	jqual := jen.Qual(mod.getIPath(), mod.Name)
	var jret *jen.Statement
	if isSimp {
		jret = jen.Id("nid").String()
	} else {
		jret = jen.Id("isnew").Bool()
	}
	// log.Printf("jret: %s, %+v", mod.Name, jret)
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jret, jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op(":=").New(jqual)
			g.Id("_").Op("=").Id("obj").Dot("SetID").Call(jen.Id("id"))

			if isSimp {
				g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
				g.Err().Op("=").Id("dbStoreSimple").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"),
				)
				g.Id("nid").Op("=").Id("obj").Dot("StringID").Call()
			} else {
				g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
				g.Id("exist").Op(":=").New(jqual)
				cpms := []jen.Code{
					jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("obj"),
					jen.Func().Params().Index().String().Block(
						jen.Return(jen.Id("exist").Dot("SetWith").Call(jen.Id("in"))),
					),
				}
				if fn, cn, isuniq := mod.UniqueOne(); isuniq {
					cpms = append(cpms, jen.Lit(cn), jen.Op("*").Id("in").Dot(fn))
				}
				g.Id("isnew").Op(",").Err().Op("=").Id("dbStoreWithCall").Call(cpms...)
			}
			g.Return()
		})
}

func (mod *Model) codestoreDelete() ([]jen.Code, []jen.Code, *jen.Statement) {
	jqual := jen.Qual(mod.getIPath(), mod.Name)
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op(":=").New(jqual)
			hkBD, okBD := mod.hasHook(beforeDeleting)
			hkAD, okAD := mod.hasHook(afterDeleting)
			if okBD || okAD {
				g.If(jen.Id("err").Op(":=").Id("getModelWithPKID").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"),
				).Op(";").Id("err").Op("!=").Nil()).Block(jen.Return(jen.Err()))

				g.Return().Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
					g1.Id("ctx")
					g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {
						if okBD {
							g2.If(jen.Err().Op("=").Id(hkBD).Call(jen.Id("ctx"), jen.Id("tx"),
								jen.Id("obj")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
						}

						g2.Err().Op("=").Id("dbDeleteT").Call(jen.Id("ctx"), jen.Id("tx"),
							jen.Add(swdb).Dot("Schema").Call(),
							jen.Add(swdb).Dot("SchemaCrap").Call(),
							jen.Lit(mod.tableName()), jen.Id("obj").Dot("ID"))
						if okAD {
							g2.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
							g2.Return(jen.Id(hkAD).Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj")))
						} else {
							g2.Return()
						}

					})
				})
			} else {
				g.If(jen.Op("!").Id("obj").Dot("SetID").Call(jen.Id("id"))).Block(
					jen.Return().Qual("fmt", "Errorf").Call(jen.Lit("id: '%s' is invalid"), jen.Id("id")),
				)
				g.Return(jen.Id("s").Dot("w").Dot("db").Dot("OpDeleteAny").Call(
					jen.Id("ctx"), jen.Lit(mod.tableName()), jen.Id("obj").Dot("ID"),
				))
			}

		})
}
