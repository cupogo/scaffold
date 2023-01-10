// go:build codegen
package main

import (
	"log"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/jennifer/jen"
)

var (
	hods = map[rune]string{
		'L': "List", 'G': "Get", 'P': "Put",
		'C': "Create", 'U': "Update", 'D': "Delete",
	}
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
	Hods     []Var    `yaml:"hods"`            // Customized

	allMM map[string]bool
	hodMn map[string]bool

	doc *Document
}

func (s *Store) prepareMethods() {
	s.allMM = make(map[string]bool)
	s.hodMn = make(map[string]bool)
	for _, m := range s.HodBread {
		s.hodMn[m] = true
		for _, a := range []string{"List", "Get", "Create", "Update", "Delete"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
		}
	}
	for _, m := range s.HodPrdb {
		s.hodMn[m] = true
		for _, a := range []string{"List", "Get", "Put", "Delete"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
		}
	}
	for _, m := range s.HodGL {
		s.hodMn[m] = true
		for _, a := range []string{"List", "Get"} {
			k := a + m
			if _, ok := s.allMM[k]; !ok {
				s.Methods = append(s.Methods, newMethod(a, m))
				s.allMM[k] = true
			}
		}
	}

	for _, hod := range s.Hods {
		m := hod.Name
		v := hod.Type
		s.hodMn[m] = true
		for _, c := range v {
			if a, ok := hods[c]; ok {
				k := a + m
				if _, ok := s.allMM[k]; !ok {
					s.Methods = append(s.Methods, newMethod(a, m))
					s.allMM[k] = true
				}
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

func (s *Store) hasModel(name string) bool {
	if _, ok := s.hodMn[name]; ok {
		return true
	}
	return false
}

func (s *Store) Interfaces(modelpkg string) (tcs, mcs []jen.Code, nap []bool, bcs []*jen.Statement) {
	// if _, ok := doc.getQual("comm"); !ok {
	// 	log.Print("get qual comm fail")
	// }

	for _, mth := range s.Methods {

		var args, rets []jen.Code
		var cs *jen.Statement

		mod, modok := doc.modelWithName(mth.model)
		if !modok {
			panic("invalid model: " + mth.model)
		}

		switch mth.action {
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
			log.Printf("unknown action: %s", mth.action)
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

func (s *Store) dstWrapField() *dst.Field {
	fd := newField(s.Name, s.Name, true)
	fd.Decorations().End.Append("// gened")
	return fd
}

func (s *Store) dstWrapFunc() *dst.FuncDecl {
	siname := s.ShortIName()
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{newField("w", storewn, true)}},
		Name: dst.NewIdent(siname),
		Type: &dst.FuncType{Results: &dst.FieldList{List: []*dst.Field{
			{Type: dst.NewIdent(s.GetIName())},
		}}},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{
			&dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent(s.Name)},
		}}}},
	}
	// f.Decorations().Start.Prepend("\n")
	f.Decorations().End.Append("// " + siname + " gened")

	return f
}

func newStoInterfaceMethod(name, ret string) *dst.Field {
	id := dst.NewIdent(name)
	id.Obj = dst.NewObj(dst.Fun, name)
	f := &dst.Field{
		Names: []*dst.Ident{id},
		Type: &dst.FuncType{
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent(ret)},
				},
			},
		},
	}
	f.Decorations().End.Append("// gened")

	return f
}

type storeHook struct {
	FunName string
	ObjName string

	k string
	m *Model
	s *Store
}

func (sh *storeHook) IsDB() bool {
	return len(sh.FunName) > 2 && sh.FunName[0:2] == "db"
}

func (sh *storeHook) dstFuncDecl(modipath string) *dst.FuncDecl {
	// log.Printf("dst FuncDecl: ObjName: %q, mod: %q", sh.ObjName, sh.m.Name)
	ctxIdent := dst.NewIdent("Context")
	ctxIdent.Path = "context"
	objIdent := dst.NewIdent(sh.ObjName)
	objIdent.Path = modipath
	bretst := &dst.ReturnStmt{Results: []dst.Expr{
		dst.NewIdent("nil"),
	}}
	bretst.Decs.Before = dst.NewLine
	bretst.Decs.Start.Append("// TODO: ")
	pars := []*dst.Field{newField("ctx", ctxIdent, false)}
	if strings.HasSuffix(sh.k, "ing") {
		pars = append(pars, newField("db", "ormDB", false), newField("obj", objIdent, true))
	} else if sh.k == beforeList {
		pars = append(pars, newField("spec", sh.ObjName+"Spec", true), newField("q", "ormQuery", true))
	} else if sh.k == afterList {
		dataIdent := dst.NewIdent(sh.m.GetPlural())
		dataIdent.Path = modipath
		pars = append(pars, newField("spec", sh.ObjName+"Spec", true), newField("data", dataIdent, false))
	} else {
		pars = append(pars, newField("obj", objIdent, true))
	}
	f := &dst.FuncDecl{
		Name: dst.NewIdent(sh.FunName),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: pars},
			Results: &dst.FieldList{List: []*dst.Field{
				{Type: dst.NewIdent("error")},
			}}},
		Body: &dst.BlockStmt{List: []dst.Stmt{bretst}},
	}
	if !sh.IsDB() {
		f.Recv = &dst.FieldList{List: []*dst.Field{newField("s", sh.s.Name, true)}}
	}
	// f.Decorations().Start.Prepend("\n")
	// f.Decorations().End.Append("// " + sh.FunName + " gened")

	return f
}
