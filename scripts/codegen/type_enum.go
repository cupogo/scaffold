// go:build codegen
package main

import (
	"fmt"

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

	Decodable       bool `yaml:"decodable,omitempty"`
	Stringer        bool `yaml:"stringer,omitempty"`
	TextMarshaler   bool `yaml:"textMarshaler,omitempty"`
	TextUnmarshaler bool `yaml:"textUnmarshaler,omitempty"`
	Multiple        bool `yaml:"multiple,omitempty"`
	Shorted         bool `yaml:"shorted,omitempty"`
}

type EnumVal struct {
	Label  string `yaml:"label,omitempty"`
	Suffix string `yaml:"suffix"`
}

func (ev EnumVal) getLabel(shorted bool) string {
	s := LcFirst(ev.Suffix)
	if shorted && len(s) > shortLen {
		return s[:shortLen]
	}
	return s
}

func (e *Enum) Code() jen.Code {
	st := jen.Comment(e.Comment).Line()
	st.Type().Id(e.Name).Id(e.Type).Line()

	if len(e.Values) > 1 {
		zeroStart := e.Start < 1
		var zeroValue *EnumVal
		vals := e.Values
		if e.Multiple && zeroStart {
			e.Start = 1
			zeroValue = &vals[0]
			vals = vals[1:]

		}
		st.Const().DefsFunc(func(g *jen.Group) {
			op := "+"
			if e.Multiple {
				op = "<<"
			}
			for i, ev := range vals {
				val := e.Start + i
				if e.Multiple {
					val = e.Start << i
				}
				name := e.Name + ev.Suffix
				cmt := fmt.Sprintf("%2d %s", val, ev.Label)
				if i == 0 {
					g.Id(name).Id(e.Name).Op("=").Lit(e.Start).Op(op).Id("iota").Comment(cmt)
				} else {
					g.Id(name).Comment(cmt)
				}
			}

			if zeroValue != nil {
				g.Line()
				g.Id(e.Name + zeroValue.Suffix).Id(e.Name).Op("=0").Comment(zeroValue.Label)
			}
		})

		if e.Decodable || e.TextUnmarshaler {
			st.Line()
			st.Func().Params(jen.Id("z").Op("*").Id(e.Name)).Id("Decode").Params(jen.Id("s").String()).Error()
			st.Block(jen.Switch(jen.Id("s")).BlockFunc(func(g *jen.Group) {
				if zeroValue != nil {
					label := zeroValue.getLabel(e.Shorted)
					cases := []jen.Code{jen.Lit("0"), jen.Lit(label)}
					if ss := zeroValue.getLabel(false); ss != label && e.Shorted {
						cases = append(cases, jen.Lit(ss))
					}
					g.Case(cases...)
					g.Op("*").Id("z").Op("=").Id(e.Name + zeroValue.Suffix)
				}
				for i, ev := range vals {
					val := e.Start + i
					if e.Multiple {
						val = e.Start << i
					}
					name := e.Name + ev.Suffix
					id := fmt.Sprint(val)
					label := ev.getLabel(e.Shorted)
					cases := []jen.Code{jen.Lit(id), jen.Lit(label)}
					if ss := ev.getLabel(false); ss != label && e.Shorted {
						cases = append(cases, jen.Lit(ss))
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
					g.Return(jen.Lit(ev.getLabel(e.Shorted)))
				}
				g.Default().Return(jen.Qual("fmt", "Sprintf").Call(jen.Lit(LcFirst(e.Name)+" %d"), jen.Id(e.Type).Call(jen.Id("z"))))
			}))

			if e.TextMarshaler {
				st.Line()
				st.Func().Params(jen.Id("z").Id(e.Name)).Id("MarshalText").Params().Params(jen.Index().Byte(), jen.Error())
				st.Block(jen.Return(jen.Index().Byte().Params(jen.Id("z").Dot("String").Call()), jen.Nil()))
			}
		}
	}

	return st
}
