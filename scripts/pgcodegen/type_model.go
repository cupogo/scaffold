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

	isOid   bool
	isDate  bool
	isIntDt bool
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

func (f *Field) preCode() (st *jen.Statement) {
	if len(f.Type) == 0 {
		f.Type = f.Name
		f.Name = ""
	}
	switch f.Name {
	case "":
		// embed field
		st = jen.Empty()
	default:
		st = jen.Id(f.Name)
	}

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
	var st *jen.Statement
	if len(f.Comment) > 0 {
		st = jen.Comment(f.Comment).Line()
	} else {
		st = jen.Empty()
	}

	st.Add(f.preCode())

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

func (f *Field) queryCode(idx int) jen.Code {
	if len(f.Type) > 0 {
		f.Type, _ = getModQual(f.Type)
	}
	st := f.preCode()

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

	if len(f.Comment) > 0 {
		if f.isDate {
			f.Comment += " + during"
		}
		st.Comment(f.Comment)
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

	if hasHooks, field := m.hasHooks(); hasHooks {
		log.Printf("model %s has hooks", m.Name)
		oidcat := CamelCased(m.OIDCat)
		if oidcat == "" {
			oidcat = "Default"
		}
		st.Comment("Creating function call to it's inner fields defined hooks").Line()
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("Creating").Params().Error().Block(
			jen.If(jen.Id("z").Dot("ID")).Dot("IsZero").Call().Block(
				jen.Id("z").Dot("SetID").Call(
					jen.Qual(oidQual, "NewID").Call(jen.Qual(oidQual, "Ot"+oidcat)),
				),
			).Line(),
			jen.Return(jen.Id("z").Dot(field).Dot("Creating").Call()),
		).Line()

		// st.Comment("Saving function call to it's inner fields defined hooks").Line()
		// st.Func().Params(
		// 	jen.Id("z").Op("*").Id(m.Name),
		// ).Id("Saving").Params().Error().Block(
		// 	jen.Return(jen.Id("z").Dot(field).Dot("Saving").Call()),
		// ).Line()
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
	var hasDefaultModel bool
	var hasIDField bool
	var hasDateFields bool
	for _, field := range m.Fields {
		if strings.HasSuffix(field.Name, "DefaultModel") {
			hasDefaultModel = true
		} else if strings.Contains(field.Name, "IDField") {
			hasIDField = true
		} else if strings.HasSuffix(field.Name, "DateFields") {
			hasDateFields = true
		}
	}
	if hasDefaultModel {
		return true, "DefaultModel"
	}
	if hasIDField && hasDateFields {
		return true, "DateFields"
	}
	return false, ""
}

func (m *Model) specFields() (out Fields) {
	for _, f := range m.Fields {
		if validQuery(f.Query) {
			if f.Type == "oid.OID" {
				f.Type = "string"
				f.isOid = true
			} else if strings.HasSuffix(f.Type, "DateTime") {
				f.Type = "string"
				f.isDate = true
				f.isIntDt = true
			} else if strings.HasSuffix(f.Type, "Time") {
				f.Type = "string"
				f.isDate = true
			}

			out = append(out, f)
		}
	}
	return
}

func (m *Model) getSpecCodes() jen.Code {
	comm, _ := doc.getQual("comm")
	var fcs []jen.Code
	fcs = append(fcs, jen.Qual(comm, "PageSpec"), jen.Id("MDftSpec"))
	if m.hasAudit() {
		fcs = append(fcs, jen.Id("AuditSpec"))
	}
	for _, sifter := range m.Sifters {
		fcs = append(fcs, jen.Id(sifter))
	}
	specFields := m.specFields()
	if len(specFields) > 0 {
		fcs = append(fcs, jen.Empty())
		for i, field := range specFields {
			delete(field.Tags, "binding")
			delete(field.Tags, "extensions")
			fcs = append(fcs, field.queryCode(i))
		}
	}

	var withRel string
	relNames := m.Fields.relHasOne()
	if len(relNames) > 0 {
		withRel = "WithRel"
		jtag := "rel"
		field := &Field{
			Name: withRel,
			Type: "bool", Tags: Maps{"json": jtag},
			Comment: "include relation column"}
		fcs = append(fcs, jen.Empty(), field.queryCode(len(specFields)))
	}

	tname := m.Name + "Spec"
	st := jen.Type().Id(tname).Struct(fcs...).Line()
	if len(fcs) > 2 {
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("Sift").Params(jen.Id("q").Op("*").Id("ormQuery")).
			Params(jen.Op("*").Id("ormQuery"), jen.Error())
		st.BlockFunc(func(g *jen.Group) {
			if len(relNames) > 0 {
				log.Printf("%s relNames %+v", m.Name, relNames)
				g.If(jen.Id("spec").Dot(withRel)).BlockFunc(func(g *jen.Group) {
					for _, relName := range relNames {
						g.Id("q").Dot("Relation").Call(jen.Lit(relName))
					}
				}).Line()
			}
			g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("MDftSpec").Dot("Sift").Call(jen.Id("q"))
			if m.hasAudit() {
				g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("AuditSpec").Dot("sift").Call(jen.Id("q"))
			}
			for _, sifter := range m.Sifters {
				g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot(sifter).Dot("Sift").Call(jen.Id("q"))
			}
			for _, field := range specFields {
				cn, _ := field.ColName()
				params := []jen.Code{jen.Id("q"), jen.Lit(cn), jen.Id("spec").Dot(field.Name)}
				cfn := "siftEquel"
				if field.isOid {
					cfn = "siftOID"
				} else if field.isDate {
					cfn = "siftDate"
					if field.isIntDt {
						params = append(params, jen.True())
					}
				}
				params = append(params, jen.False())
				// TODO: set text wildcard
				g.Id("q").Op(",").Id("_").Op("=").Id(cfn).Call(params...)
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

func validQuery(s string) bool {
	switch s {
	case "eq", "equal": // TODO: more query support
		return true
	default:
		return false
	}
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
			jq := jen.Add(swdb).Dot("Model").Call(
				jen.Op("&").Id("data")).Dot("Apply").Call(
				jen.Id("spec").Dot("Sift"))
			if cols, ok := m.HasTextSearch(); ok {
				g.Id("q").Op(":=").Add(jq)
				g.Id("tss").Op(":=").Add(swdb).Dot("GetTsSpec").Call()
				if len(cols) > 0 {
					g.Id("tss").Dot("SetFallback").Call(jen.ListFunc(func(g1 *jen.Group) {
						for _, s := range cols {
							g1.Lit(s)
						}
					}))
				}
				g.Id("total").Op(",").Id("err").Op("=").Id("queryPager").Call(
					jen.Id("spec"), jen.Id("q").Dot("Apply").Call(jen.Id("tss").Dot("Sift")),
				)
			} else {
				g.Id("total").Op(",").Id("err").Op("=").Id("queryPager").Call(
					jen.Id("spec"), jq,
				)
			}
			g.Return()
		})
}

func (mod *Model) codestoreGet() ([]jen.Code, []jen.Code, *jen.Statement) {
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
			jload := jen.Id("err").Op("=").Id("getModelWithPKID").Call(
				jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"))
			if _, cn, isuniq := mod.UniqueOne(); isuniq {
				g.If(jen.Err().Op("=").Id("getModelWithUnique").Call(
					swdb, jen.Id("obj"), jen.Lit(cn), jen.Id("id"),
				).Op(";").Err().Op("!=").Nil()).Block(jload)
			} else {
				g.Add(jload)
			}

			if hkAL, okAL := mod.hasHook(afterLoad); okAL {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id(hkAL).Call(jen.Id("ctx"), jen.Id("s").Dot("w"), jen.Id("obj")),
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
			if fn, cn, isuniq := mod.UniqueOne(); isuniq {
				targs = append(targs, jen.Lit(cn), jen.Id("in").Dot(fn))
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
