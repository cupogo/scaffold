// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jinzhu/inflection"

	"hyyl.xyz/cupola/scaffold/pkg/utils"
)

const (
	oidQual   = "hyyl.xyz/cupola/aurora/pkg/models/oid"
	metaField = "comm.MetaField"
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

	isOid bool
}

func (f *Field) isMeta() bool {
	return f.Name == metaField || f.Type == metaField
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
	} else if pos := strings.Index(f.Type, "."); pos > 0 {
		if qual, ok := getQual(f.Type[0:pos]); ok {
			st.Qual(qual, f.Type[pos+1:])
		} else {
			log.Printf("get qual %s fail", f.Type)
		}
	} else {
		st.Id(f.Type)
	}

	return st
}

func (f *Field) Code() jen.Code {
	var st *jen.Statement
	if len(f.Comment) > 0 {
		st = jen.Comment(f.Comment).Line()
	} else {
		st = jen.Empty()
	}

	st.Add(f.preCode())

	if len(f.Tags) > 0 {
		// log.Printf("%s: %+v", f.Name, f.Tags)
		st.Tag(f.Tags)
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
				a = utils.Underscore(f.Name)
			}
			return a, strings.Contains(b, "unique")
		}
	}
	return utils.Underscore(f.Name), false
}

func (f *Field) queryCode() jen.Code {
	if len(f.Type) > 0 {
		f.Type, _ = getModQual(f.Type)
	}
	st := f.preCode()

	tags := f.Tags.Copy()
	if len(tags) > 0 {
		if _, ok := tags["form"]; !ok {
			if v, ok := tags["json"]; ok {
				tags["form"] = v
			}
		}
		delete(tags, "pg")
		st.Tag(tags)
	}

	if len(f.Comment) > 0 {
		st.Comment(f.Comment)
	}

	return st
}

type Fields []Field

// Codes return fields code of main and basic
func (f Fields) Codes() (mcs, bcs []jen.Code) {
	var hasMeta bool
	for _, field := range f {
		if field.IsSet || field.IsBasic {
			bcs = append(bcs, field.Code())
		} else {
			mcs = append(mcs, field.Code())
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

type Model struct {
	Comment  string `yaml:"comment,omitempty"`
	Name     string `yaml:"name"`
	TableTag string `yaml:"tableTag"`
	Fields   Fields `yaml:"fields"`
	Plural   string `json:"plural"`
	OIDCat   string `json:"oidcat,omitempty"`
}

func (m *Model) GetPlural() string {
	if m.Plural != "" {
		return m.Plural
	}
	return inflection.Plural(m.Name)
}

func (m *Model) tableName() string {
	tt := m.TableTag
	if tt == "" {
		tt = utils.Underscore(m.GetPlural())
	} else if pos := strings.Index(tt, ","); pos > 0 {
		tt = tt[0:pos]
	}
	return tt
}

func (m *Model) TableField() jen.Code {
	tt := m.TableTag
	if tt == "" {
		tt = utils.Underscore(m.GetPlural())
	}
	return jen.Id("tableName").Add(jen.Struct()).Tag(Maps{"pg": tt}).Line()
}

func (m *Model) Uniques() (name, col string, ok bool) {
	var count int
	for _, field := range m.Fields {
		name = field.Name
		if col, ok = field.ColName(); ok {
			count++
		}
	}
	ok = count == 1
	return
}

func (m *Model) ChangablCodes() (ccs []jen.Code, scs []jen.Code) {
	var hasMeta bool
	for _, field := range m.Fields {
		if !field.IsSet || field.Type == "" || field.Name == "" {
			if field.isMeta() {
				hasMeta = true
			}
			continue
		}
		code := jen.Id(field.Name)
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
			code.Tag(Maps{"json": s})
		}
		if len(field.Comment) > 0 {
			code.Comment(field.Comment)
		}
		ccs = append(ccs, code)
		scs = append(scs, jen.If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			csst := jen.Id("cs").Op("=").Append(jen.Id("cs"), jen.Lit(cn))
			if field.isOid {
				g.If(jen.Id("id").Op(",").Err().Op(":=").Id("oid").Dot("CheckID").Call(jen.Op("*").Id("o").Dot(field.Name)).Op(";").Err().Op("==").Nil().Block(
					jen.Id("z").Dot(field.Name).Op("=").Id("id"), csst,
				))
			} else {
				g.Add(jen.Id("z").Dot(field.Name).Op("=").Op("*").Id("o").Dot(field.Name).Line(), csst)
			}
		}))
	}
	if hasMeta {
		name := "MetaUp"
		ccs = append(ccs, metaUpCode())
		scs = append(scs, jen.If(jen.Id("o").Dot(name).Op("!=").Nil().Op("&&").Id("z").Dot("UpMeta").Call(jen.Id("o").Dot(name))).Block(
			jen.Id("cs").Op("=").Append(jen.Id("cs"), jen.Lit("meta")),
		))
	}
	return
}

func (m *Model) Codes() jen.Code {
	var cs []jen.Code
	cs = append(cs, m.TableField())
	mcs, bcs := m.Fields.Codes()
	cs = append(cs, mcs...)
	st := jen.Comment(m.Name + " " + m.Comment).Line()

	basicName := m.Name + "Basic"
	if len(bcs) > 0 {
		cs = append(cs, jen.Id(basicName))
	}

	st.Type().Id(m.Name).Struct(cs...).Add(jen.Comment("@name " + m.Name)).Line().Line()

	if len(bcs) > 0 {
		st.Type().Id(basicName).Struct(bcs...).Add(jen.Comment("@name " + basicName)).Line().Line()
	}

	st.Type().Id(m.GetPlural()).Index().Id(m.Name).Line().Line()

	if hasHooks, field := m.hasHooks(); hasHooks {
		log.Print("has hooks")
		oidcat := utils.CamelCased(m.OIDCat)
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

		st.Comment("Saving function call to it's inner fields defined hooks").Line()
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("Saving").Params().Error().Block(
			jen.Return(jen.Id("z").Dot(field).Dot("Saving").Call()),
		).Line()
	}

	if ccs, scs := m.ChangablCodes(); len(ccs) > 0 {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(ccs...).Add(jen.Comment("@name " + changeSetName)).Line().Line()
		scs = append(scs, jen.Return())
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("SetWith").Params(jen.Id("o").Op("*").Id(changeSetName)).Params(
			jen.Id("cs").Index().String(),
		).Block(
			scs...,
		)
	}
	return st
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
		if f.Query != "" {
			if f.Type == "oid.OID" {
				f.Type = "string"
				f.isOid = true
			}
			out = append(out, f)
		}
	}
	return
}

func (m *Model) getSpecCodes() jen.Code {
	comm, _ := getQual("comm")
	var fcs []jen.Code
	fcs = append(fcs, jen.Qual(comm, "PageSpec"), jen.Id("MDftSpec"))
	specFields := m.specFields()
	if len(specFields) > 0 {
		fcs = append(fcs, jen.Empty())
		for _, field := range specFields {
			delete(field.Tags, "binding")
			delete(field.Tags, "extensions")
			fcs = append(fcs, field.queryCode())
		}
	}

	tname := m.Name + "Spec"
	st := jen.Type().Id(tname).Struct(fcs...).Line()
	if len(fcs) > 2 {
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("Sift").Params(jen.Id("q").Op("*").Id("ormQuery")).
			Params(jen.Op("*").Id("ormQuery"), jen.Error())
		st.BlockFunc(func(g *jen.Group) {
			g.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot("MDftSpec").Dot("Sift").Call(jen.Id("q"))
			for _, field := range specFields {
				cfn := "siftEquel"
				if field.isOid {
					cfn = "siftOID"
				}
				// TODO: set text wildcard
				cn, _ := field.ColName()
				g.Id("q").Op(",").Id("_").Op("=").Id(cfn).Call(
					jen.Id("q"), jen.Lit(cn), jen.Id("spec").Dot(field.Name), jen.False(),
				)
			}
			g.Line()
			g.Return(jen.Id("q"), jen.Nil())
		}).Line()
	}

	return st
}

func metaUpCode() jen.Code {
	code := jen.Comment("for meta update").Line()
	code.Id("MetaUp").Op("*").Add(qual("comm.MetaUp"))
	code.Tag(Maps{"bson": "-", "json": "metaUp,omitempty", "pg": "-", "swaggerignore": "true"})
	return code
}
