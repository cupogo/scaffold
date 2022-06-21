// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"

	"hyyl.xyz/cupola/scaffold/pkg/utils"
)

type Var struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Method struct {
	Name   string `yaml:"name"`
	Simple bool   `yaml:"simple"`
	Args   []Var  `yaml:"args,omitempty"`
	Rets   []Var  `yaml:"rets,omitempty"`
}

type Store struct {
	Name    string   `yaml:"name"`
	IName   string   `yaml:"iname,omitempty"`
	Methods []Method `yaml:"methods"`

	mnames []string // TODO: aliases
}

func (s *Store) Interfaces(modelpkg string) (tcs, mcs []jen.Code, nap []bool, bcs []*jen.Statement) {
	if _, ok := getQual("comm"); !ok {
		log.Print("get qual comm fail")
	}
	modpkg, ok := getQual(modelpkg)
	if !ok {
		log.Printf("get modpkg %s fail", modpkg)
	}

	_ctx := jen.Id("ctx").Qual("context", "Context")
	for _, mth := range s.Methods {
		var args, rets []jen.Code
		args = append(args, _ctx)
		act, mname, ok := cutMethod(mth.Name)
		if !ok {
			log.Fatalf("inalid method: %s", mth.Name)
			return
		}
		// log.Printf("act %s, %s, %v", act, mname, ok)
		if act == "List" {
			tname := mname + "Spec"
			var tspec jen.Code
			if m, ok := doc.modelWithName(mname); ok {
				tspec = m.getSpecCodes()
			} else {
				comm, _ := getQual("comm")
				tspec = jen.Type().Id(tname).Struct(jen.Qual(comm, "PageSpec"), jen.Id("MDftSpec")).Line()
			}

			tcs = append(tcs, tspec)
			args = append(args, jen.Id("spec").Op("*").Id(tname))
			rets = append(rets, jen.Id("data").Index().Qual(modpkg, mname),
				jen.Id("total").Int(), jen.Id("err").Error())
			bcs = append(bcs, jen.Block(
				jen.Id("total").Op(",").Id("err").Op("=").Id("queryPager").Call(
					jen.Id("spec"),
					jen.Id("s").Dot("w").Dot("db").Dot("Model").
						Call(jen.Op("&").Id("data")).Dot("Apply").
						Call(jen.Id("spec").Dot("Sift")),
				),
				jen.Return(),
			))
			nap = append(nap, false)
		} else if act == "Get" {
			args = append(args, jen.Id("id").String())
			rets = append(rets, jen.Id("obj").Op("*").Qual(modpkg, mname), jen.Id("err").Error())
			bcs = append(bcs, jen.Block(
				jen.Id("obj").Op("=").New(jen.Qual(modpkg, mname)),
				jen.Id("err").Op("=").Id("getModelWithPKOID").Call(
					jen.Id("s").Dot("w").Dot("db"), jen.Id("obj"), jen.Id("id")), //err = getModelWithPKOID(s.w.db, obj, id)
				jen.Return(),
			))
			nap = append(nap, false)
		} else if act == "Create" {
			tname := mname + "Basic"
			args = append(args, jen.Id("in").Op("*").Qual(modpkg, tname))
			rets = append(rets, jen.Id("obj").Op("*").Qual(modpkg, mname), jen.Id("err").Error())
			bcs = append(bcs, jen.Block(
				jen.Id("obj").Op("=&").Qual(modpkg, mname).Block(jen.Id(tname).Op(":").Op("*").Id("in").Op(",")),
				jen.Id("err").Op("=").Id("dbInsert").Call(
					jen.Id("ctx"), jen.Id("s").Dot("w").Dot("db"), jen.Id("obj"),
				),
				jen.Return(),
			))
			nap = append(nap, false)
		} else if act == "Update" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Op("*").Qual(modpkg, mname+"Set"))
			rets = append(rets, jen.Id("err").Error())
			bcs = append(bcs, jen.Block(
				jen.Id("exist").Op(":=").New(jen.Qual(modpkg, mname)),
				jen.Id("err").Op("=").Id("getModelWithPKOID").Call(
					jen.Id("s").Dot("w").Dot("db"), jen.Id("exist"), jen.Id("id"),
				),
				jen.If(jen.Id("err").Op("!=").Nil()).Block(jen.Return()),
				jen.Id("cs").Op(":=").Id("exist").Dot("SetWith").Call(jen.Id("in")),
				jen.If(jen.Len(jen.Id("cs")).Op("==").Lit(0)).Block(
					jen.Return(),
				),
				jen.Err().Op("=").Id("dbUpdate").Call(
					jen.Id("ctx"), jen.Id("s").Dot("w").Dot("db"), jen.Id("exist"), jen.Id("cs..."),
				),
				jen.Return(),
			))
			nap = append(nap, false)
		} else if act == "Put" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Op("*").Qual(modpkg, mname+"Set"))
			if !mth.Simple {
				rets = append(rets, jen.Id("isnew").Bool())
			}
			rets = append(rets, jen.Err().Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op(":=").New(jen.Qual(modpkg, mname))
				g.Id("obj").Dot("SetID").Call(jen.Id("id"))
				if mth.Simple {
					g.Id("cs").Op(":=").Id("obj").Dot("SetWith").Call(jen.Id("in"))
					g.Err().Op("=").Id("dbStoreSimple").Call(
						jen.Id("ctx"), jen.Id("s").Dot("w").Dot("db"), jen.Id("obj"), jen.Id("cs..."),
					)
				} else {
					g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
					g.Id("exist").Op(":=").New(jen.Qual(modpkg, mname))
					g.Id("isnew").Op(",").Id("err").Op("=").Id("dbStoreWithCall").Call(
						jen.Id("ctx"), jen.Id("s").Dot("w").Dot("db"), jen.Id("exist"), jen.Id("obj"),
						jen.Func().Params().Index().String().Block(
							jen.Return(jen.Id("exist").Dot("SetWith").Call(jen.Id("in"))),
						),
					)
				}
				g.Return()
			}))
			nap = append(nap, false)

		} else if act == "Delete" {
			args = append(args, jen.Id("id").String())
			rets = append(rets, jen.Error())
			bcs = append(bcs, jen.Block(
				jen.Return(jen.Id("s").Dot("w").Dot("db").Dot("OpDelete").Call(
					jen.Id("ctx"), jen.Lit(getTableName(mname)), jen.Id("id"),
				)), // dbOpDelete(ctx, tableClause, id)
			).Line())
			nap = append(nap, true)
		} else {
			log.Printf("unknown action: %s", act)
			bcs = append(bcs, jen.Block())
			nap = append(nap, false)

		}
		mcs = append(mcs, jen.Id(mth.Name).Params(args...).Params(rets...))
	}

	return
}

func (s *Store) GetIName() string {
	if len(s.IName) > 0 {
		return s.IName
	}
	return utils.CamelCased(s.Name)
}

func (s *Store) ShortIName() string {
	in := s.GetIName()
	if strings.HasSuffix(in, "Store") {
		in = in[0 : len(in)-5]
	}
	return in
}

func (s *Store) Codes(modelpkg string) jen.Code {
	modpkg, ok := getQual(modelpkg)
	if !ok {
		log.Printf("get modpkg %s fail", modpkg)
	}
	tcs, mcs, nap, bcs := s.Interfaces(modelpkg)
	var ics []jen.Code
	for i := range mcs {
		ics = append(ics, mcs[i])
		if nap[i] {
			ics = append(ics, jen.Line())
		}
	}

	st := jen.Type().Id(s.GetIName()).Interface(ics...).Line().Line()
	st.Add(tcs...).Line()

	st.Type().Id(s.Name).Struct(jen.Id("w").Op("*").Id("Wrap")).Line()

	for i := range mcs {
		st.Func().Params(jen.Id("s").Op("*").Id(s.Name)).Add(mcs[i], bcs[i]).Line()
	}

	return st
}
