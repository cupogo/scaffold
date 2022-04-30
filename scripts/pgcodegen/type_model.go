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
	oidQual = "hyyl.xyz/cupola/aurora/pkg/models/oid"
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
	IsSet   bool   `yaml:"isset,omitempty"`
	Comment string `yaml:"comment,omitempty"`
}

func (f *Field) Code() jen.Code {
	var st *jen.Statement
	if len(f.Qual) > 0 {
		st = jen.Qual(f.Qual, f.Name)
	} else if pos := strings.Index(f.Name, "."); pos > 0 {
		if qual, ok := getQual(f.Name[0:pos]); ok {
			st = jen.Qual(qual, f.Name[pos+1:])
		} else {
			log.Printf("get qual %s fail", f.Name)
		}
	} else {
		st = jen.Id(f.Name)
	}

	switch f.Type {
	case "":
		// embed field
		st.Line()
	default:
		st.Id(f.Type)
	}
	if len(f.Tags) > 0 {
		// log.Printf("%s: %+v", f.Name, f.Tags)
		st.Tag(f.Tags)
	}

	if len(f.Comment) > 0 {
		st.Comment(f.Comment)
	}

	return st
}

func (f *Field) ColName() string {
	if s, ok := f.Tags["pg"]; ok && len(s) > 0 {
		if a, _, ok := strings.Cut(s, ","); ok && len(a) > 0 {
			return a
		}
	}
	return utils.Underscore(f.Name)
}

type Fields []Field

// Codes return fields code of main and basic
func (f Fields) Codes() (mcs, bcs []jen.Code) {
	for _, field := range f {
		if field.IsSet {
			bcs = append(bcs, field.Code())
		} else {
			mcs = append(mcs, field.Code())
		}

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

func (m *Model) ChangablCodes() (ccs []jen.Code, scs []jen.Code) {
	for _, field := range m.Fields {
		if !field.IsSet || field.Type == "" {
			continue
		}
		code := jen.Id(field.Name)
		tn := field.Type
		if len(tn) == 0 {
			tn = field.Name
		}
		if len(field.Qual) > 0 {
			code.Op("*").Qual(field.Qual, tn)
		} else {
			code.Op("*").Id(tn)
		}
		if s, ok := field.Tags["json"]; ok {
			code.Tag(Maps{"json": s})
		}
		ccs = append(ccs, code)
		scs = append(scs, jen.If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).Block(
			jen.Id("z").Dot(field.Name).Op("=").Op("*").Id("o").Dot(field.Name),
			jen.Id("cs").Op("=").Append(jen.Id("cs"), jen.Lit(field.ColName())),
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

	st.Type().Id(m.Name).Struct(cs...).Line().Line()

	if len(bcs) > 0 {
		st.Type().Id(basicName).Struct(bcs...).Line().Line()
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
			jen.Return(jen.Id("z").Dot(field).Dot("Creating").Call()),
		).Line()
	}

	if ccs, scs := m.ChangablCodes(); len(ccs) > 0 {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(ccs...).Line().Line()
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
