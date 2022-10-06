// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
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
	SIName   string   `yaml:"siname,omitempty"`
	Methods  []Method `yaml:"methods"`
	Embed    string   `yaml:"embed,omitempty"`
	HodBread []string `yaml:"hodBread,omitempty"`
	HodPrdb  []string `yaml:"hodPrdb,omitempty"`
	HodGL    []string `yaml:"hodGL,omitempty"` // Get and List 只读（含列表）

	allMM map[string]bool

	doc *Document
}

func (s *Store) prepareMethods() {
	s.allMM = make(map[string]bool)
	for _, m := range s.HodBread {
		for _, a := range []string{"List", "Get", "Create", "Update", "Delete"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
		}
	}
	for _, m := range s.HodPrdb {
		for _, a := range []string{"List", "Get", "Put", "Delete"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
		}
	}
	for _, m := range s.HodGL {
		for _, a := range []string{"List", "Get"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
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
	// if _, ok := doc.getQual("comm"); !ok {
	// 	log.Print("get qual comm fail")
	// }

	for _, mth := range s.Methods {
		var args, rets []jen.Code
		var cs *jen.Statement
		act, mname, ok := cutMethod(mth.Name)
		if !ok {
			log.Fatalf("inalid method: %s", mth.Name)
			return
		}
		mod, modok := doc.modelWithName(mname)
		if !modok {
			panic("invalid model: " + mname)
		}

		switch act {
		case "List":
			tcs = append(tcs, mod.getSpecCodes())
			args, rets, cs = mod.codestoreList()
			bcs = append(bcs, cs)
			nap = append(nap, false)
		case "Get":
			args, rets, cs = mod.codestoreGet()
			bcs = append(bcs, cs)
			nap = append(nap, false)
		case "Create":
			args, rets, cs = mod.codestoreCreate()
			bcs = append(bcs, cs)
			nap = append(nap, false)
		case "Update":
			args, rets, cs = mod.codestoreUpdate()
			bcs = append(bcs, cs)
			nap = append(nap, false)
		case "Put":
			args, rets, cs = mod.codestorePut(mth.Simple)
			bcs = append(bcs, cs)
			nap = append(nap, false)

		case "Delete":
			args, rets, cs = mod.codestoreDelete()
			bcs = append(bcs, cs.Line())
			nap = append(nap, true)
		default:
			log.Printf("unknown action: %s", act)
			bcs = append(bcs, jen.Block())
			nap = append(nap, false)
		}
		args = append([]jen.Code{jen.Id("ctx").Qual("context", "Context")}, args...)
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
	if len(s.SIName) > 0 {
		return s.SIName
	}
	in := s.GetIName()
	return strings.TrimSuffix(in, "Store")
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

type storeHook struct {
	FunName string
	ObjName string
}
