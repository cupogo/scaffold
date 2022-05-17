// go:build codegen
package main

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v3"

	"daxv.cn/gopak/lib/osutil"
)

const (
	headerComment = "This file is generated - Do Not Edit."

	storepkg = "stores"
)

type Unmarshaler = yaml.Unmarshaler

type Maps map[string]string

// Copy ...
func (m Maps) Copy() (out Maps) {
	out = Maps{}
	for k, v := range m {
		out[k] = v
	}
	return
}

type Document struct {
	name    string
	dirmod  string
	dirsto  string
	dirweb  string
	lock    sync.Mutex
	methods map[string]Method

	ModelPkg  string  `yaml:"modelpkg"`
	Models    []Model `yaml:"models"`
	Qualified Maps    `yaml:"depends"` // imports name
	Stores    []Store `yaml:"stores"`
	WebAPI    WebAPI  `yaml:"webapi"`
}

func (doc *Document) getQual(k string) (string, bool) {
	v, ok := doc.Qualified[k]
	// log.Printf("get qual: k %s, v %s, ok %v", k, v, ok)
	return v, ok
}

func NewDoc(docfile string) (*Document, error) {
	yf, err := os.Open(docfile)
	if err != nil {
		return nil, err
	}
	doc := new(Document)
	err = yaml.NewDecoder(yf).Decode(doc)
	if err != nil {
		return nil, err
	}
	if doc.ModelPkg == "" {
		return nil, fmt.Errorf("modelpkg is empty")
	}
	doc.name = getOutName(docfile)
	doc.dirmod = path.Join("pkg", "models", doc.ModelPkg)
	doc.dirsto = path.Join("pkg", "services", "stores")
	doc.dirweb = path.Join("pkg", "web", doc.WebAPI.Pkg)
	doc.methods = make(map[string]Method)

	log.Printf("loaded %d models, out name %q", len(doc.Models), doc.name)

	return doc, nil
}

func getOutName(docfile string) string {
	name := path.Base(docfile)
	if pos := strings.LastIndex(name, "."); pos > 1 {
		name = name[0:pos]
	}
	name = name + "_gen.go"

	return name
}

func (doc *Document) genModels(dropfirst bool) error {
	mgf := jen.NewFile(doc.ModelPkg)
	mgf.HeaderComment(headerComment)

	for _, model := range doc.Models {
		log.Printf("found model %s", model.Name)
		mgf.Add(model.Codes())
	}
	mgf.Line()

	if !osutil.IsDir(doc.dirmod) {
		if err := os.Mkdir(doc.dirmod, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirmod, err)
			return err
		}
	}

	outname := path.Join(doc.dirmod, doc.name)
	log.Printf("%s: %s", doc.ModelPkg, outname)
	if dropfirst && osutil.CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	if err := mgf.Save(outname); err != nil {
		log.Fatalf("generate models fail: %s", err)
		return err
	}
	log.Printf("generated for %s ok", doc.dirmod)
	return nil
}

func (doc *Document) modelWithName(name string) (*Model, bool) {
	for _, m := range doc.Models {
		if m.Name == name {
			return &m, true
		}
	}
	return nil, false
}

func (doc *Document) genStores(dropfirst bool) error {
	mpkg := loadPackage(doc.dirmod)
	// if err != nil {
	// 	log.Printf("get package fail: %s", err)
	// }
	log.Printf("loaded mpkg: %s name %q: path %q", mpkg.ID, mpkg.Types.Name(), mpkg.Types.Path())
	log.Printf("types: %+v, ", mpkg.Types)

	sgf := jen.NewFile(storepkg)
	sgf.HeaderComment(headerComment)

	sgf.ImportName(mpkg.ID, doc.ModelPkg)
	doc.lock.Lock()
	doc.Qualified[doc.ModelPkg] = mpkg.ID
	doc.lock.Unlock()

	var aliases []string
	for _, f := range mpkg.Syntax {
		for k := range f.Scope.Objects {
			if strings.HasSuffix(k, "Basic") {
				continue
			}
			aliases = append(aliases, k)
		}
	}
	sort.Strings(aliases)
	for _, k := range aliases {
		sgf.Type().Id(k).Op("=").Qual(mpkg.ID, k)
	}
	sgf.Line()

	if !osutil.IsDir(doc.dirsto) {
		if err := os.Mkdir(doc.dirsto, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirsto, err)
			return err
		}
	}
	outname := path.Join(doc.dirsto, doc.name)
	if dropfirst && osutil.CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	// spkg := loadPackage(doc.dirsto)
	// log.Printf("loaded spkg: %s name %q", spkg.ID, spkg.Types.Name())

	for _, store := range doc.Stores {
		sgf.Add(store.Codes(doc.ModelPkg)).Line()
	}

	err := sgf.Save(outname)
	if err != nil {
		log.Fatalf("generate stores fail: %s", err)
		return err
	}
	log.Printf("generated for %s ok", doc.dirsto)
	return nil
}

func (doc *Document) getMethod(name string) (m Method, ok bool) {
	m, ok = doc.methods[name]
	return
}

func (doc *Document) genWebAPI() error {

	if !osutil.IsDir(doc.dirweb) {
		if err := os.Mkdir(doc.dirweb, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirweb, err)
			return err
		}
	}
	outname := path.Join(doc.dirweb, "handle_"+doc.name)
	if dropfirst && osutil.CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	mpkg := loadPackage(doc.dirmod)
	spkg := loadPackage(doc.dirsto)

	wgf := jen.NewFile(doc.WebAPI.getPkgName())
	wgf.HeaderComment(headerComment)

	wgf.ImportName(mpkg.ID, doc.ModelPkg)
	wgf.ImportName(spkg.ID, storepkg)
	doc.lock.Lock()
	doc.Qualified[doc.ModelPkg] = mpkg.ID
	doc.Qualified[storepkg] = spkg.ID
	doc.lock.Unlock()

	// log.Printf("loaded spkg: %+v", spkg.Types.Scope().Names())
	stoName := doc.Stores[0].GetIName()
	obj := spkg.Types.Scope().Lookup(stoName)
	if obj == nil {
		log.Fatalf("%s not found in declared types of %s", stoName, spkg)
	}
	log.Printf("lookuped: %+v", obj)
	if _, ok := obj.(*types.TypeName); !ok {
		log.Fatalf("%v is not a named type", obj)
	}
	objType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		log.Fatalf("type %v is not a struct", obj)
	}
	log.Printf("NumMethods: %d", objType.NumMethods())
	doc.lock.Lock()
	for i := 0; i < objType.NumMethods(); i++ {
		smt := objType.Method(i)
		if sig, ok := smt.Type().(*types.Signature); ok {
			log.Printf("method sign: params %s %+v, result: %+v", smt.Name(), sig.Params(), sig.Results())
			var args []Var
			var rets []Var
			for j := 0; j < sig.Params().Len(); j++ {
				args = append(args, getVarFromTypesVar(sig.Params().At(j)))
			}
			for j := 0; j < sig.Results().Len(); j++ {
				rets = append(rets, getVarFromTypesVar(sig.Results().At(j)))
			}
			doc.methods[smt.Name()] = Method{Name: smt.Name(), Args: args, Rets: rets}
		}
	}
	doc.lock.Unlock()
	// TODO: put spkg methods into webapi

	wgf.Add(doc.WebAPI.Codes(doc))

	// err := wgf.Render(os.Stdout)

	err := wgf.Save(outname)
	if err != nil {
		log.Fatalf("generate stores fail: %s", err)
		return err
	}
	log.Printf("generated for %s ok", doc.dirweb)
	return nil
}

func getVarFromTypesVar(v *types.Var) Var {
	typs := v.Type().String()
	if pos := strings.LastIndex(typs, "/"); pos > 0 {
		typs = typs[pos+1:]
	}
	return Var{Name: v.Name(), Type: typs}
}
