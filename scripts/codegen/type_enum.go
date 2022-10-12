package main

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type Enum struct {
	Comment string    `yaml:"comment"`
	Name    string    `yaml:"name"`
	Type    string    `yaml:"type,omitempty"`
	Start   int       `yaml:"start,omitempty"`
	Values  []EnumVal `yaml:"values,omitempty"`

	Decodable       bool `yaml:"decodable,omitempty"`
	TextUnmarshaler bool `yaml:"textUnmarshaler,omitempty"`
}

type EnumVal struct {
	Label  string `yaml:"label,omitempty"`
	Suffix string `yaml:"suffix"`
}

func (e *Enum) Code() jen.Code {
	st := jen.Comment(e.Comment).Line()
	st.Type().Id(e.Name).Id(e.Type).Line()

	if len(e.Values) > 0 {
		st.Const().DefsFunc(func(g *jen.Group) {
			for i, ev := range e.Values {
				name := e.Name + ev.Suffix
				if i == 0 {
					g.Id(name).Id(e.Name).Op("=").Lit(e.Start).Op("+").Id("iota").Comment(ev.Label)
				} else {
					g.Id(name).Comment(ev.Label)
				}
			}
		})

		if e.Decodable || e.TextUnmarshaler {
			st.Line()
			st.Func().Params(jen.Id("z").Op("*").Id(e.Name)).Id("Decode").Params(jen.Id("s").String()).Error()
			st.Block(jen.Switch(jen.Id("s")).BlockFunc(func(g *jen.Group) {
				for i, ev := range e.Values {
					name := e.Name + ev.Suffix
					id := fmt.Sprint(e.Start + i)
					g.Case(jen.Lit(id), jen.Lit(LcFirst(ev.Suffix)))
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
	}

	return st
}
