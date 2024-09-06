// go:build codegen
package gens

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	shortLen = 3
)

type Enum struct {
	Comment string    `yaml:"comment"`
	Name    string    `yaml:"name"`
	Type    string    `yaml:"type,omitempty"`
	Start   int       `yaml:"start,omitempty"`
	Values  []EnumVal `yaml:"values,omitempty"`

	Decodable       bool   `yaml:"decodable,omitempty"`
	Stringer        bool   `yaml:"stringer,omitempty"`
	TextMarshaler   bool   `yaml:"textMarshaler,omitempty"`
	TextUnmarshaler bool   `yaml:"textUnmarshaler,omitempty"`
	Multiple        bool   `yaml:"multiple,omitempty"`
	Shorted         bool   `yaml:"shorted,omitempty"`
	ValStr          bool   `yaml:"valstr,omitempty"` // return a value as string
	Labeled         bool   `yaml:"labeled,omitempty"`
	FuncAll         string `yaml:"funcAll,omitempty"`
	SpecNs          string `yaml:"specNs,omitempty"` // prefix of Enum in api

	doc *Document
}

type EnumVal struct {
	Label  string `yaml:"label,omitempty"`
	Suffix string `yaml:"suffix"`
	Value  int    `yaml:"value,omitempty"`
	Lower  bool   `yaml:"lower,omitempty"`
	Descr  string `yaml:"descr,omitempty"`

	realVal int
}

func (ev EnumVal) getCode(shorted bool) (s string) {
	if ev.Lower || isUpperString(ev.Suffix) {
		s = strings.ToLower(ev.Suffix)
	} else {
		s = LcFirst(ev.Suffix)
	}

	if shorted && len(s) > shortLen {
		return s[:shortLen]
	}
	return s
}

func (e *Enum) prepare() (vals []EnumVal, zv *EnumVal) {
	zeroStart := e.Start < 1
	vals = e.Values
	if e.Multiple && zeroStart {
		e.Start = 1
		zv = &vals[0]
		vals = vals[1:]
	}
	for i, ev := range vals {
		val := e.Start + i
		if e.Multiple {
			val = e.Start << i
		} else if ev.Value > 0 {
			val = ev.Value
		}
		vals[i].realVal = val
	}

	return vals, zv
}

func (e *Enum) Code() jen.Code {
	st := jen.Comment(e.Comment).Line()
	st.Type().Id(e.Name).Id(e.Type).Line()

	if len(e.Values) <= 1 { // At least 2 vales are required
		return st
	}

	vals, zv := e.prepare()

	st.Const().DefsFunc(func(g *jen.Group) {
		op := "+"
		if e.Multiple {
			op = "<<"
		}
		for i, ev := range vals {
			var cmt string
			if e.Multiple {
				cmt = fmt.Sprintf("%3d %s", ev.realVal, ev.Label)
			} else {
				cmt = fmt.Sprintf("%2d %s", ev.realVal, ev.Label)
			}
			name := e.Name + ev.Suffix
			if i == 0 && (ev.Value == 0 || ev.Value == e.Start) {
				g.Id(name).Id(e.Name).Op("=").Lit(e.Start).Op(op).Id("iota").Comment(cmt)
			} else if ev.Value > 0 {
				g.Id(name).Id(e.Name).Op("=").Lit(ev.Value).Comment(cmt)
			} else {
				g.Id(name).Comment(cmt)
			}
		}

		if zv != nil {
			g.Line()
			g.Id(e.Name + zv.Suffix).Id(e.Name).Op("=0").Comment(zv.Label)
		}
	})

	if e.Decodable || e.TextUnmarshaler {
		st.Line()
		st.Func().Params(jen.Id("z").Op("*").Id(e.Name)).Id("Decode").Params(jen.Id("s").String()).Error()
		st.Block(jen.Switch(jen.Id("s")).BlockFunc(func(g *jen.Group) {
			if zv != nil {
				code := zv.getCode(e.Shorted)
				cases := []jen.Code{jen.Lit("0"), jen.Lit(code)}
				if ss := zv.getCode(false); ss != code && e.Shorted {
					cases = append(cases, jen.Lit(ss))
				}
				g.Case(cases...)
				g.Op("*").Id("z").Op("=").Id(e.Name + zv.Suffix)
			}
			for _, ev := range vals {
				name := e.Name + ev.Suffix
				id := fmt.Sprint(ev.realVal)
				code := ev.getCode(e.Shorted)
				cases := []jen.Code{jen.Lit(id), jen.Lit(code)}
				if ss := ev.getCode(false); ss != code && e.Shorted {
					cases = append(cases, jen.Lit(ss))
				}
				if code != ev.Suffix {
					cases = append(cases, jen.Lit(ev.Suffix))
				}
				if IsAlphaOnly(ev.Label) && code != ev.Label && ev.Suffix != ev.Label {
					cases = append(cases, jen.Lit(ev.Label))
				}
				g.Case(cases...)
				g.Op("*").Id("z").Op("=").Id(name)
			}
			g.Default().Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("invalid "+LcFirst(e.Name)+": %q"), jen.Id("s")))

		}), jen.Return(jen.Nil()))

		if e.TextUnmarshaler {
			st.Line()
			st.Func().Params(jen.Id("z").Op("*").Id(e.Name)).Id("UnmarshalText").Params(jen.Id("b").Index().Byte()).Error()
			st.Block(jen.Return(jen.Id("z").Dot("Decode").Call(jen.String().Call(jen.Id("b")))))
		}
	}

	if e.Stringer || e.TextMarshaler {
		st.Line()
		st.Func().Params(jen.Id("z").Id(e.Name)).Id("String").Params().String()
		st.Block(jen.Switch(jen.Id("z")).BlockFunc(func(g *jen.Group) {
			for _, ev := range e.Values {
				name := e.Name + ev.Suffix
				g.Case(jen.Id(name))
				g.Return(jen.Lit(ev.getCode(e.Shorted)))
			}
			g.Default().Return(jen.Qual("fmt", "Sprintf").Call(jen.Lit(LcFirst(e.Name)+" %d"), jen.Id(e.Type).Call(jen.Id("z"))))
		}))

		if e.TextMarshaler {
			st.Line()
			st.Func().Params(jen.Id("z").Id(e.Name)).Id("MarshalText").Params().Params(jen.Index().Byte(), jen.Error())
			st.Block(jen.Return(jen.Index().Byte().Params(jen.Id("z").Dot("String").Call()), jen.Nil()))
		}
	}

	if e.ValStr {
		st.Line()
		st.Func().Params(jen.Id("z").Id(e.Name)).Id("ValStr").Params().String()
		st.Block(jen.Return().Qual("strconv", "Itoa").Call(jen.Int().Call(jen.Id("z"))))
	}

	if e.Labeled {
		st.Line()
		st.Func().Params(jen.Id("z").Id(e.Name)).Id("Label").Params().String()
		st.Block(jen.Switch(jen.Id("z")).BlockFunc(func(g *jen.Group) {
			for _, ev := range e.Values {
				name := e.Name + ev.Suffix
				g.Case(jen.Id(name))
				label, _, _ := strings.Cut(ev.Label, " ")
				g.Return(jen.Lit(label))
			}
			g.Default().Return(jen.Qual("fmt", "Sprintf").Call(jen.Lit(LcFirst(e.Name)+"#%d"), jen.Id(e.Type).Call(jen.Id("z"))))
		}))
	}

	if len(e.FuncAll) > 0 {
		wizhZero := e.isFuncWithZero()
		if e.isFuncValues() {
			st.Line()
			st.Func().Id(e.funcAllName()).Params().Params(jen.Index().Id(e.Name))
			st.BlockFunc(func(g *jen.Group) {
				g.Return().Index().Id(e.Name).Op("{")
				if wizhZero && zv != nil {
					g.Id(e.Name + zv.Suffix).Op(",")
				}
				for _, ev := range vals {
					name := e.Name + ev.Suffix
					g.Id(name).Op(",")
				}
				g.Op("}")
			})
		}
		if e.isFuncOptions() {
			jitem := e.jQualItem()
			st.Line()
			st.Type().Id(e.Name + "Item").Op("=").Add(jitem).Line()
			st.Func().Id(e.funcAllOptionsName()).Params().Params(jen.Index().Add(jitem))
			st.BlockFunc(func(g *jen.Group) {
				g.Return().Index().Add(jitem).Op("{")
				if wizhZero && zv != nil {
					g.Add(zv.jItem(e)).Op(",")
				}
				for _, ev := range vals {
					g.Add(ev.jItem(e)).Op(",")
				}
				g.Op("}")
			})

		}

	}

	return st
}

func (e *Enum) funcAllName(ss ...string) string {
	out := "All" + e.Name
	if len(ss) > 0 && len(ss[0]) > 0 {
		out += ss[0]
	}
	return out
}

func (e *Enum) funcAllOptionsName() string {
	return e.funcAllName("Options")
}

func (e *Enum) isFuncValues() bool {
	return strings.Contains(e.FuncAll, "value")
}

func (e *Enum) isFuncOptions() bool {
	return strings.Contains(e.FuncAll, "option")
}

func (e *Enum) isFuncWithZero() bool {
	return strings.Contains(e.FuncAll, "zero")
}

func (e *Enum) getItemName() string {
	if e.doc != nil {
		if e.doc.EnumItem != "" {
			return e.doc.EnumItem
		}
	}
	return "EnumItem"
}

func (e *Enum) jQualItem() jen.Code {
	if e.doc != nil && e.doc.EnumCore != "" {
		return jen.Qual(e.doc.EnumCore, e.getItemName())
	}
	return jen.Id(e.getItemName())
}

func (e *Enum) Label() string {
	a, _, _ := strings.Cut(e.Comment, " ")
	return a
}

func (ev EnumVal) jItem(e *Enum) jen.Code {
	st := new(jen.Statement)
	st.Op("{")
	if !e.TextMarshaler {
		st.Id("ID:").Lit(ev.realVal).Op(",")
	}
	st.Id("Code:").Lit(ev.getCode(e.Shorted)).Op(",")
	name := strings.TrimSpace(ev.Label)
	descr := strings.TrimSpace(ev.Descr)
	if len(descr) == 0 {
		a, b, ok := strings.Cut(name, "  ")
		if !ok || len(b) == 0 {
			a, b, ok = strings.Cut(name, "\n")
		}
		name = a
		descr = b
	}
	st.Id("Name:").Lit(name).Op(",")
	if len(descr) > 1 {
		st.Id("Descr:").Lit(descr).Op(",")
	}
	st.Op("}")
	return st
}

type EnumDoc struct {
	Lines    []string
	Codes    []string
	SwaggerT string
}

func (e *Enum) docComments() (ed EnumDoc, ok bool) {
	vals, _ := e.prepare()

	for _, ev := range vals {
		code := ev.getCode(e.Shorted)
		label := ev.Label
		if a, _, ok := strings.Cut(label, " "); ok && len(a) > 0 {
			label = a
		}
		var suf string
		if code != label && ev.Suffix != label {
			suf = " - " + label
		}
		var cs string
		if e.TextMarshaler {
			cs = fmt.Sprintf("`%s`%s", code, suf)
			ed.SwaggerT = "string"
			ed.Codes = append(ed.Codes, code)
		} else {
			cs = fmt.Sprintf("%d=`%s`%s", ev.realVal, code, suf)
			ed.Codes = append(ed.Codes, fmt.Sprintf("%d", ev.realVal))
		}

		ed.Lines = append(ed.Lines, cs)
	}
	ok = len(ed.Lines) > 0

	return
}
