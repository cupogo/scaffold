// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"

	"hyyl.xyz/cupola/scaffold/pkg/utils"
)

const (
	afterSaving    = "afterSaving"
	beforeCreating = "beforeCreating"
	afterDeleting  = "afterDeleting"
)

type Var struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Method struct {
	Name   string `yaml:"name"`
	Simple bool   `yaml:"simple,omitempty"`
	Args   []Var  `yaml:"args,omitempty"`
	Rets   []Var  `yaml:"rets,omitempty"`

	action string
	model  string
}

func newMethod(act, mod string) Method {
	return Method{Name: act + mod, action: act, model: mod}
}

type Store struct {
	Name     string   `yaml:"name"`
	IName    string   `yaml:"iname,omitempty"`
	Methods  []Method `yaml:"methods"`
	Embed    string   `yaml:"embed,omitempty"`
	HodBread []string `yaml:"hodBread,omitempty"`
	HodPrdb  []string `yaml:"hodPrdb,omitempty"`

	mnames []string // TODO: aliases

	doc *Document
}

func (s *Store) prepareMethods() {
	for _, m := range s.HodBread {
		for _, a := range []string{"List", "Get", "Create", "Update", "Delete"} {
			s.Methods = append(s.Methods, newMethod(a, m))
		}
	}
	for _, m := range s.HodPrdb {
		for _, a := range []string{"List", "Get", "Put", "Delete"} {
			s.Methods = append(s.Methods, newMethod(a, m))
		}
	}

	for i := range s.Methods {
		if s.Methods[i].model == "" {
			s.Methods[i].action, s.Methods[i].model, _ = cutMethod(s.Methods[i].Name)
		}
	}
	log.Printf("inited store methods: %d", len(s.Methods))
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
		mod, modok := doc.modelWithName(mname)
		if !modok {
			panic("invalid model: " + mname)
		}
		swdb := jen.Id("s").Dot("w").Dot("db")
		// log.Printf("act %s, %s, %v", act, mname, ok)
		if act == "List" {
			tname := mname + "Spec"
			slicename := mod.GetPlural()
			tspec := mod.getSpecCodes()

			tcs = append(tcs, tspec)
			args = append(args, jen.Id("spec").Op("*").Id(tname))
			rets = append(rets, jen.Id("data").Qual(modpkg, slicename),
				jen.Id("total").Int(), jen.Id("err").Error())
			bcs = append(bcs, jen.Block(
				jen.Id("total").Op(",").Id("err").Op("=").Id("queryPager").Call(
					jen.Id("spec"),
					jen.Add(swdb).Dot("Model").
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
					swdb, jen.Id("obj"), jen.Id("id")), //err = getModelWithPKOID(s.w.db, obj, id)
				jen.Return(),
			))
			nap = append(nap, false)
		} else if act == "Create" {
			tname := mname + "Basic"
			args = append(args, jen.Id("in").Op("*").Qual(modpkg, tname))
			rets = append(rets, jen.Id("obj").Op("*").Qual(modpkg, mname), jen.Id("err").Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op("=&").Qual(modpkg, mname).Block(jen.Id(tname).Op(":").Op("*").Id("in").Op(","))

				if mod.hasMeta() {
					g.Id("s").Dot("w").Dot("opModelMeta").Call(jen.Id("ctx"),
						jen.Id("obj"), jen.Id("obj").Dot("MetaUp"))
				}
				targs := []jen.Code{jen.Id("ctx"), swdb, jen.Id("obj")}
				if fn, cn, isuniq := mod.Uniques(); isuniq {
					targs = append(targs, jen.Lit(cn), jen.Op("*").Id("in").Dot(fn))
				}

				if len(mod.Hooks) > 0 {
					g.Err().Op("=").Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
						g1.Id("ctx")
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {
							if hk, ok := mod.hasHook(beforeCreating); ok {
								g2.If(jen.Err().Op("=").Id(hk).Call(jen.Id("ctx"), swdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
									jen.Return(jen.Err()),
								)
							}
							g2.Id("err").Op("=").Id("dbInsert").Call(targs...)
							if hk, ok := mod.hasHook(afterSaving); ok {
								g2.If(jen.Err().Op("==")).Nil().Block(
									jen.Err().Op("=").Id(hk).Call(jen.Id("ctx"), swdb, jen.Id("obj")),
								)
							}

							g2.Return(jen.Err())
						})

					})

				} else {
					g.Id("err").Op("=").Id("dbInsert").Call(targs...)
				}

				g.Return()
			}))
			nap = append(nap, false)
		} else if act == "Update" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Op("*").Qual(modpkg, mname+"Set"))
			rets = append(rets, jen.Id("err").Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("exist").Op(":=").New(jen.Qual(modpkg, mname))
				g.Id("err").Op("=").Id("getModelWithPKOID").Call(
					swdb, jen.Id("exist"), jen.Id("id"),
				)
				g.If(jen.Id("err").Op("!=").Nil()).Block(jen.Return())
				g.Id("cs").Op(":=").Id("exist").Dot("SetWith").Call(jen.Id("in"))
				g.If(jen.Len(jen.Id("cs")).Op("==").Lit(0)).Block(
					jen.Return(),
				)

				if len(mod.Hooks) > 0 {
					g.Err().Op("=").Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
						g1.Id("ctx")
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Error()).BlockFunc(func(g2 *jen.Group) {
							g2.Err().Op(":=").Id("dbUpdate").Call(
								jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("cs..."),
							)

							if hk, ok := mod.hasHook(afterSaving); ok {
								g2.If(jen.Err().Op("==")).Nil().Block(
									jen.Err().Op("=").Id(hk).Call(jen.Id("ctx"), swdb, jen.Id("exist")),
								)
							}

							g2.Return(jen.Err())
						})

					})

				} else {
					g.Err().Op("=").Id("dbUpdate").Call(
						jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("cs..."),
					)
				}

				g.Return()
			}))
			nap = append(nap, false)
		} else if act == "Put" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Op("*").Qual(modpkg, mname+"Set"))
			if mth.Simple {
				rets = append(rets, jen.Id("nid").String())
			} else {
				rets = append(rets, jen.Id("isnew").Bool())
			}
			rets = append(rets, jen.Err().Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op(":=").New(jen.Qual(modpkg, mname))
				g.Id("obj").Dot("SetID").Call(jen.Id("id"))
				if mth.Simple {
					g.Id("cs").Op(":=").Id("obj").Dot("SetWith").Call(jen.Id("in"))
					g.Err().Op("=").Id("dbStoreSimple").Call(
						jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("cs..."),
					)
					g.Id("nid").Op("=").Id("obj").Dot("StringID").Call()
				} else {
					g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
					g.Id("exist").Op(":=").New(jen.Qual(modpkg, mname))
					cpms := []jen.Code{
						jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("obj"),
						jen.Func().Params().Index().String().Block(
							jen.Return(jen.Id("exist").Dot("SetWith").Call(jen.Id("in"))),
						),
					}
					if fn, cn, isuniq := mod.Uniques(); isuniq {
						cpms = append(cpms, jen.Lit(cn), jen.Op("*").Id("in").Dot(fn))
					}
					g.Id("isnew").Op(",").Err().Op("=").Id("dbStoreWithCall").Call(cpms...)
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
	if len(s.Embed) > 0 {
		ics = append(ics, jen.Id(s.Embed))
	}
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
