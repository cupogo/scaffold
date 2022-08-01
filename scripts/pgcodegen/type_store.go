// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	beforeSaving   = "beforeSaving"
	afterSaving    = "afterSaving"
	beforeCreating = "beforeCreating"
	beforeUpdating = "beforeUpdating"
	afterDeleting  = "afterDeleting"
	afterLoad      = "afterLoad"
	afterCreated   = "afterCreated"
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
	if _, ok := doc.getQual("comm"); !ok {
		log.Print("get qual comm fail")
	}
	modpkg, ok := doc.getQual(modelpkg)
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
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				jq := jen.Add(swdb).Dot("Model").Call(
					jen.Op("&").Id("data")).Dot("Apply").Call(
					jen.Id("spec").Dot("Sift"))
				if cols, ok := mod.HasTextSearch(); ok {
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
			}))
			nap = append(nap, false)
		} else if act == "Get" {
			args = append(args, jen.Id("id").String())
			rets = append(rets, jen.Id("obj").Op("*").Qual(modpkg, mname), jen.Id("err").Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op("=").New(jen.Qual(modpkg, mname))
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
							g2.If(jen.Id("rn").Op("==").Lit(rn)).Block(
								jen.Id("ro").Op(":=").New(field.typeCode(modpkg)),
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
			}))
			nap = append(nap, false)
		} else if act == "Create" {
			tname := mname + "Basic"
			args = append(args, jen.Id("in").Qual(modpkg, tname))
			rets = append(rets, jen.Id("obj").Op("*").Qual(modpkg, mname), jen.Id("err").Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op("=&").Qual(modpkg, mname).Block(jen.Id(tname).Op(":").Id("in").Op(","))

				if mod.hasMeta() {
					g.Id("s").Dot("w").Dot("opModelMeta").Call(jen.Id("ctx"),
						jen.Id("obj"), jen.Id("obj").Dot("MetaUp"))
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
							g2.Id("err").Op("=").Id("dbInsert").Call(targs...)
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
			}))
			nap = append(nap, false)
		} else if act == "Update" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Qual(modpkg, mname+"Set"))
			rets = append(rets, jen.Id("err").Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("exist").Op(":=").New(jen.Qual(modpkg, mname))
				g.Id("err").Op("=").Id("getModelWithPKID").Call(
					jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("id"),
				)
				g.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("cs").Op(":=").Id("exist").Dot("SetWith").Call(jen.Id("in"))
				g.If(jen.Len(jen.Id("cs")).Op("==").Lit(0)).Block(
					jen.Return(),
				)

				hkBU, okBU := mod.hasHook(beforeUpdating)
				hkBS, okBS := mod.hasHook(beforeSaving)
				hkAS, okAS := mod.hasHook(afterSaving)
				if okBU || okBS || okAS {
					g.Return().Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
						jdb := jen.Id("tx")
						g1.Id("ctx")
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {
							if okBU {
								g2.If(jen.Id("cs").Op(",").Err().Op("=").Id(hkBU).Call(jen.Id("ctx"), jdb, jen.Id("exist"), jen.Id("cs")).Op(";").Err().Op("!=")).Nil().Block(
									jen.Return(),
								)
							} else if okBS {
								g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
									jen.Return(),
								)
							}
							jup := jen.Id("dbUpdate").Call(
								jen.Id("ctx"), jdb, jen.Id("exist"), jen.Id("cs..."),
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
						jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("cs..."),
					)
				}

			}))
			nap = append(nap, false)
		} else if act == "Put" {
			args = append(args, jen.Id("id").String(), jen.Id("in").Qual(modpkg, mname+"Set"))
			if mth.Simple {
				rets = append(rets, jen.Id("nid").String())
			} else {
				rets = append(rets, jen.Id("isnew").Bool())
			}
			rets = append(rets, jen.Err().Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op(":=").New(jen.Qual(modpkg, mname))
				g.Id("_").Op("=").Id("obj").Dot("SetID").Call(jen.Id("id"))
				// g.If(jen.Op("!").Id("obj").Dot("SetID").Call(jen.Id("id"))).Block(
				// 	jen.Err().Op("=").Qual(errsQual, "NewErrInvalidID").Call(jen.Id("id")),
				// 	jen.Return(),
				// )
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
					if fn, cn, isuniq := mod.UniqueOne(); isuniq {
						cpms = append(cpms, jen.Lit(cn), jen.Op("*").Id("in").Dot(fn))
					}
					g.Id("isnew").Op(",").Err().Op("=").Id("dbStoreWithCall").Call(cpms...)
				}
				g.Return()
			}))
			nap = append(nap, false)

		} else if act == "Delete" {
			args = append(args, jen.Id("id").String())
			rets = append(rets, jen.Err().Error())
			bcs = append(bcs, jen.BlockFunc(func(g *jen.Group) {
				g.Id("obj").Op(":=").New(jen.Qual(modpkg, mname))
				if hk, ok := mod.hasHook(afterDeleting); ok {
					g.If(jen.Id("err").Op("=").Id("getModelWithPKID").Call(
						jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"),
					).Op(";").Id("err").Op("!=").Nil()).Block(jen.Return())

					g.Err().Op("=").Add(swdb).Dot("RunInTransaction").CallFunc(func(g1 *jen.Group) {
						g1.Id("ctx")
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(func(g2 *jen.Group) {

							g2.If(jen.Err().Op("=").Id("dbDeleteT").Call(jen.Id("ctx"), jen.Id("tx"),
								jen.Add(swdb).Dot("Schema").Call(),
								jen.Add(swdb).Dot("SchemaCrap").Call(),
								jen.Lit(mod.tableName()), jen.Id("obj").Dot("ID")).Op(";").Err().Op("!=").Nil()).Block(
								jen.Return(),
							)
							g2.Return(jen.Id(hk).Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj")))
						})

					})
					g.Return()
				} else {
					g.If(jen.Op("!").Id("obj").Dot("SetID").Call(jen.Id("id"))).Block(
						jen.Return().Qual(errsQual, "NewErrInvalidID").Call(jen.Id("id")),
					)
					g.Return(jen.Id("s").Dot("w").Dot("db").Dot("OpDeleteAny").Call(
						jen.Id("ctx"), jen.Lit(mod.tableName()), jen.Id("obj").Dot("ID"),
					))
				}

			}).Line())
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
	return CamelCased(s.Name)
}

func (s *Store) ShortIName() string {
	in := s.GetIName()
	if strings.HasSuffix(in, "Store") {
		in = in[0 : len(in)-5]
	}
	return in
}

func (s *Store) Codes(modelpkg string) jen.Code {
	modpkg, ok := doc.getQual(modelpkg)
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
