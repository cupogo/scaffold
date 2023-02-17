// go:build codegen
package gens

import (
	"go/token"
	"log"
	"strconv"
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

		mod, modok := s.doc.modelWithName(mth.model)
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
	modpkg, ok := s.doc.getQual(modelpkg)
	if !ok {
		log.Printf("get modpkg %s fail", modpkg)
	}
	tcs, mcs, nap, bcs := s.Interfaces(modelpkg)
	var ics []jen.Code
	if len(s.Embed) > 0 {
		ics = append(ics, jen.Id(s.Embed).Line())
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

func (sh *storeHook) esMainStmt(typ string) []dst.Stmt {
	var (
		arg dst.Expr
	)
	switch typ {
	case upsertES:
		typ = "UpsertESDoc"
		arg = dst.NewIdent("obj")
	case deleteES:
		typ = "DeleteESDoc"
		arg = &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   dst.NewIdent("obj"),
				Sel: dst.NewIdent("StringID"),
			}}
	}
	mainStmts := make([]dst.Stmt, 0)
	st := dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("obj"),
			Op: token.EQL,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{&dst.ReturnStmt{
				Results: []dst.Expr{dst.NewIdent("nil")},
			}},
		},
	}
	st.Decs.After = dst.NewLine

	st1 := dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.Ident{
					Name: typ,
				},
				Args: []dst.Expr{
					dst.NewIdent("ctx"),
					&dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   dst.NewIdent("obj"),
							Sel: dst.NewIdent("IdentityTable"),
						}},
					arg,
				},
			},
		},
		Decs: dst.AssignStmtDecorations{},
	}
	if typ == "UpsertESDoc" {
		st1.Decs.Start.Append("// TODO:")
	}
	st1.Decs.After = dst.NewLine

	st2 := dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("err"),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   &dst.CallExpr{Fun: dst.NewIdent("logger")},
							Sel: dst.NewIdent("Infow")},
						Args: []dst.Expr{
							&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(typ)},
							&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("index")},
							&dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X:   dst.NewIdent("obj"),
									Sel: dst.NewIdent("IdentityTable"),
								}},
							&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("error")},
							dst.NewIdent("err"),
						},
					},
				},
				&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}},
			},
		},
	}

	st2.Decs.After = dst.NewLine

	return append(mainStmts, &st, &st1, &st2)
}

func (sh *storeHook) dstFuncDecl(modipath string) *dst.FuncDecl {
	// log.Printf("dst FuncDecl: ObjName: %q, mod: %q", sh.ObjName, sh.m.Name)
	ctxIdent := dst.NewIdent("Context")
	ctxIdent.Path = "context"
	objIdent := dst.NewIdent(sh.ObjName)
	objIdent.Path = modipath

	bodyst := &dst.BlockStmt{List: make([]dst.Stmt, 0)}

	pars := []*dst.Field{newField("ctx", ctxIdent, false)}
	if strings.HasSuffix(sh.k, "ing") {
		pars = append(pars, newField("db", "ormDB", false), newField("obj", objIdent, true))
	} else if sh.k == beforeList {
		pars = append(pars, newField("spec", sh.ObjName+"Spec", true), newField("q", "ormQuery", true))
	} else if sh.k == afterList {
		dataIdent := dst.NewIdent(sh.m.GetPlural())
		dataIdent.Path = modipath
		pars = append(pars, newField("spec", sh.ObjName+"Spec", true), newField("data", dataIdent, false))
	} else if sh.k == upsertES || sh.k == deleteES {
		pars = append(pars, newField("obj", objIdent, true))
		bodyst.List = append(bodyst.List, sh.esMainStmt(sh.k)...)
	} else {
		pars = append(pars, newField("obj", objIdent, true))
	}

	bretst := &dst.ReturnStmt{Results: []dst.Expr{
		dst.NewIdent("nil"),
	}}
	bretst.Decs.Before = dst.NewLine
	if sh.k != deleteES && sh.k != upsertES {
		bretst.Decs.Start.Append("// TODO:")
	}
	bodyst.List = append(bodyst.List, bretst)

	f := &dst.FuncDecl{
		Name: dst.NewIdent(sh.FunName),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: pars},
			Results: &dst.FieldList{List: []*dst.Field{
				{Type: dst.NewIdent("error")},
			}}},
		Body: bodyst,
	}
	if !sh.IsDB() {
		f.Recv = &dst.FieldList{List: []*dst.Field{newField("s", sh.s.Name, true)}}
	}
	// f.Decorations().Start.Prepend("\n")
	// f.Decorations().End.Append("// " + sh.FunName + " gened")

	return f
}

func (sh *storeHook) dstMEGenDecl(vd *vdst) (index int, xdecl dst.Decl) {
loop:
	for i, decl := range vd.file.Decls {
		if d, ok := decl.(*dst.GenDecl); ok && d.Tok == token.TYPE && len(d.Specs) > 0 {
			for j, spec := range d.Specs {
				if ins, ok := spec.(*dst.TypeSpec); ok && ins.Name.Name == sh.s.Embed && ins.Type != nil {
					if inf, ok := ins.Type.(*dst.InterfaceType); ok {
						if inf.Methods != nil {
							find := false
						loop1:
							for _, fd := range inf.Methods.List {
								for _, fn := range fd.Names {
									if fn.Name == sh.FunName {
										find = true
										break loop1
									}
								}
							}
							if !find {
								inf.Methods.List = append(inf.Methods.List, &dst.Field{
									Names: []*dst.Ident{dst.NewIdent(sh.FunName)},
									Type: &dst.FuncType{
										Params: &dst.FieldList{
											List: []*dst.Field{
												&dst.Field{
													Names: []*dst.Ident{dst.NewIdent("ctx")},
													Type: &dst.Ident{
														Name: "Context",
														Path: "context",
													},
												}},
										},
										Results: &dst.FieldList{
											List: []*dst.Field{
												&dst.Field{
													Names: []*dst.Ident{dst.NewIdent("err")},
													Type: &dst.Ident{
														Name: "error",
													},
												}},
										},
									},
								})
								ins.Type = inf
								d.Specs[j] = ins
							}
						}
						index = i
						xdecl = d
						break loop
					}

				}
			}
		}
	}
	return
}

func (sh *storeHook) dstMEFuncDecl(modipath string) *dst.FuncDecl {
	ctxIdent := dst.NewIdent("Context")
	ctxIdent.Path = "context"

	// var
	varStmt := &dst.DeclStmt{
		Decl: &dst.GenDecl{
			Tok: token.VAR,
			Specs: []dst.Spec{
				&dst.ValueSpec{
					Names: []*dst.Ident{
						dst.NewIdent("ms")},
					Type: &dst.Ident{Path: modipath, Name: sh.m.GetPlural()},
				},
				&dst.ValueSpec{
					Names: []*dst.Ident{
						dst.NewIdent("limit"),
						dst.NewIdent("page"),
					},
					Values: []dst.Expr{
						&dst.BasicLit{Kind: token.INT, Value: "1000"},
						&dst.BasicLit{Kind: token.INT, Value: "1"},
					},
				},
			},
		},
	}
	varStmt.Decs.After = dst.NewLine

	specStmt := dst.DeclStmt{
		Decl: &dst.GenDecl{
			Tok: token.VAR,
			Specs: []dst.Spec{
				&dst.ValueSpec{
					Names: []*dst.Ident{dst.NewIdent("spec")},
					Type:  dst.NewIdent(sh.m.Name + "Spec"),
				},
			},
		},
	}
	limitStmt := dst.AssignStmt{
		Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("spec"), Sel: dst.NewIdent("Limit")}},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{dst.NewIdent("limit")},
	}
	pageStmt := dst.AssignStmt{
		Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("spec"), Sel: dst.NewIdent("Page")}},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{dst.NewIdent("page")},
	}
	sortStmt := dst.AssignStmt{
		Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("spec"), Sel: dst.NewIdent("Sort")}},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("created desc")}},
	}

	// migrate data
	upsertFunc, _ := sh.m.storeHookName(upsertES, "yes")
	forStmt := &dst.ForStmt{
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&specStmt,
				&limitStmt,
				&pageStmt,
				&sortStmt,
				&dst.AssignStmt{
					Lhs: []dst.Expr{
						dst.NewIdent("ms"),
						dst.NewIdent("_"),
						dst.NewIdent("err"),
					},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{
								X: &dst.CallExpr{
									Fun: &dst.SelectorExpr{
										X: &dst.SelectorExpr{
											X:   dst.NewIdent("s"),
											Sel: dst.NewIdent("w"),
										},
										Sel: dst.NewIdent(sh.s.ShortIName()),
									}},
								Sel: dst.NewIdent("List" + sh.m.Name),
							},
							Args: []dst.Expr{
								dst.NewIdent("ctx"),
								dst.NewIdent("&spec"),
							},
						},
					},
					Decs: dst.AssignStmtDecorations{
						NodeDecs: dst.NodeDecs{
							Before: dst.NewLine,
							Start:  dst.Decorations{"// TODO:"},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{
						X: &dst.BinaryExpr{
							X:  dst.NewIdent("err"),
							Op: token.NEQ,
							Y:  dst.NewIdent("nil"),
						},
						Op: token.LAND,
						Y: &dst.BinaryExpr{
							X:  dst.NewIdent("err"),
							Op: token.NEQ,
							Y:  dst.NewIdent("ErrNoRows"),
						},
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun: &dst.SelectorExpr{
										X:   &dst.CallExpr{Fun: dst.NewIdent("logger")},
										Sel: dst.NewIdent("Infow")},
									Args: []dst.Expr{
										&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("get " + sh.m.Name)},
										&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("error")},
										dst.NewIdent("err"),
									},
								},
							},
							&dst.ReturnStmt{},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{
						X: &dst.CallExpr{
							Fun:  dst.NewIdent("len"),
							Args: []dst.Expr{dst.NewIdent("ms")},
						},
						Op: token.EQL,
						Y:  dst.NewIdent("0"),
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{&dst.BranchStmt{Tok: token.BREAK}},
					},
				},
				&dst.RangeStmt{
					Key: dst.NewIdent("i"),
					Tok: token.DEFINE,
					X:   dst.NewIdent("ms"),
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.AssignStmt{
								Lhs: []dst.Expr{dst.NewIdent("err")},
								Tok: token.ASSIGN,
								Rhs: []dst.Expr{
									&dst.CallExpr{
										Fun: &dst.SelectorExpr{
											X:   dst.NewIdent("s"),
											Sel: dst.NewIdent(upsertFunc),
										},
										Args: []dst.Expr{
											dst.NewIdent("ctx"),
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.IndexExpr{
													X:     dst.NewIdent("ms"),
													Index: dst.NewIdent("i"),
												},
											},
										},
									},
								},
							},
							&dst.IfStmt{
								Cond: &dst.BinaryExpr{
									X:  dst.NewIdent("err"),
									Op: token.NEQ,
									Y:  dst.NewIdent("nil"),
								},
								Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{}}},
							},
						},
					},
				},
				&dst.IncDecStmt{
					X:   dst.NewIdent("page"),
					Tok: token.INC,
				},
			},
		},
	}
	forStmt.Decs.After = dst.NewLine

	f := &dst.FuncDecl{
		Name: dst.NewIdent(sh.FunName),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: ctxIdent}}},
			Results: &dst.FieldList{List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("err")}, Type: dst.NewIdent("error")},
			}}},
		Body: &dst.BlockStmt{List: []dst.Stmt{varStmt, forStmt, &dst.ReturnStmt{}}},
	}
	if !sh.IsDB() {
		f.Recv = &dst.FieldList{List: []*dst.Field{newField("s", sh.s.Name, true)}}
	}

	return f
}
