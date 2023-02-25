// go:build codegen
package gens

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v3"
)

type Unmarshaler = yaml.Unmarshaler

type Tags map[string]string

// Copy ...
func (m Tags) Copy() (out Tags) {
	out = Tags{}
	for k, v := range m {
		out[k] = v
	}
	return
}

func (m Tags) Has(k string) bool {
	_, ok := m[k]
	return ok
}

func (m Tags) GetAny(a ...string) (string, bool) {
	for _, k := range a {
		if v, ok := m[k]; ok {
			return v, ok
		}
	}
	return "", false
}

func (m Tags) extOrder(idx int) {
	if idx > 55 { // max ascii offset
		return
	}
	if _, ok := m[TagSwaggerIgnore]; !ok {
		if _, ok = m[TagExtensions]; !ok {
			m[TagExtensions] = fmt.Sprintf("x-order=%c", getRune(idx))
		}
	}
}

func getRune(idx int) rune {
	var offset int
	if idx > 26 {
		offset = 6
	}
	return rune(64 + idx + offset)
}

type Document struct {
	gened   string
	extern  string
	dirmod  string
	dirsto  string
	dirweb  string
	lock    sync.Mutex
	methods map[string]Method

	modipath string
	modtypes map[string]empty
	dbcode   string

	Enums     []Enum  `yaml:"enums"`
	ModelPkg  string  `yaml:"modelpkg"`
	Models    []Model `yaml:"models"`
	Qualified Tags    `yaml:"depends"` // imports name
	Stores    []Store `yaml:"stores"`
	WebAPI    WebAPI  `yaml:"webapi"`
}

func (doc *Document) Check() error {
	if 0 == len(doc.ModelPkg) {
		return errors.New("empty modelpkg")
	}

	if 0 == len(doc.Models) {
		return errors.New("empty models")
	}

	for i := 0; i < len(doc.Models); i++ {
		if 0 == len(doc.Models[i].Fields) {
			return errors.New("empty fields")
		}
	}

	return nil
}

func (doc *Document) Valid() bool {
	return len(doc.ModelPkg) > 0 && len(doc.Models) > 0 && len(doc.Models[0].Fields) > 0
}

func (doc *Document) getQual(k string) (qu string, ok bool) {
	if len(k) > 0 && k[0] == '*' {
		k = k[1:]
	}
	qu, ok = doc.Qualified[k]
	if !ok && k == "utils" {
		if qoid, _ok := doc.Qualified["oid"]; _ok {
			if pos := strings.LastIndex(qoid, "models/oid"); pos > 0 {
				return qoid[0:pos] + "utils", true
			}
		}
	}
	// log.Printf("get qual: k %s, v %s, ok %v", k, qu, ok)
	return
}

func (doc *Document) getModQual(k string) string {
	if _, ok := doc.modtypes[k]; ok {
		return doc.modipath
	}
	return ""
}

func (doc *Document) hasQualErrors() bool {
	if s, ok := doc.Qualified["errors"]; ok && s != "errors" {
		return true
	}
	return false
}

func (doc *Document) qual(args ...string) jen.Code {
	if len(args) == 0 {
		log.Fatal("empty args for qual")
	}
	if len(args) > 1 {
		return jen.Qual(args[0], args[1])
	}
	name := args[0]
	if pos := strings.Index(name, "."); pos > 0 {
		if qual, ok := doc.getQual(name[0:pos]); ok {
			return jen.Qual(qual, name[pos+1:])
		} else {
			log.Printf("get qual %s fail", name)
		}
	}
	return jen.Id(name)
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
	doc.gened, doc.extern = getOutName(docfile)
	doc.dirmod = path.Join("pkg", "models", doc.ModelPkg)
	doc.dirsto = path.Join("pkg", "services", storepkg)
	doc.dirweb = path.Join("pkg", "web", doc.WebAPI.Pkg)
	doc.methods = make(map[string]Method)
	doc.modtypes = make(map[string]empty)
	doc.dbcode = os.Getenv("SCAFFOLD_DB_CODE")

	log.Printf("loaded %d models, dbcode %s", len(doc.Models), doc.dbcode)
	// log.Printf("loaded webapi uris %+v", doc.WebAPI.URIs)

	return doc, nil
}

func (doc *Document) Init() {
	for i := range doc.Models {
		doc.Models[i].doc = doc
		doc.Models[i].pkg = doc.ModelPkg
		for j := range doc.Models[i].Fields {
			f := doc.Models[i].Fields[j]
			if k, _, _ := f.cutType(); len(k) > 0 && len(f.Qual) == 0 {
				if p, ok := doc.Qualified[k]; ok {
					doc.Models[i].Fields[j].Qual = p
				}
			}
		}
	}
	for i := range doc.Stores {
		doc.Stores[i].doc = doc
		doc.Stores[i].prepareMethods()
	}
	doc.WebAPI.doc = doc
	doc.WebAPI.prepareHandles()
}

func getOutName(docfile string) (gened string, extern string) {
	name := path.Base(docfile)
	if pos := strings.LastIndex(name, "."); pos > 1 {
		name = name[0:pos]
	}
	gened = name + "_gen.go"
	extern = name + "_x.go"

	return
}

func (doc *Document) IsPG10() bool {
	if doc != nil && doc.dbcode == "pg10" {
		return true
	}
	return false
}

func (doc *Document) hasStoreHooks() bool {
	for _, m := range doc.Models {
		if len(m.StoHooks) > 0 {
			return true
		}
	}
	return false
}

func (doc *Document) storeHooks() (out []storeHook) {
	for i := 0; i < len(doc.Stores); i++ {
		for j := 0; j < len(doc.Models); j++ {
			if doc.Stores[i].hasModel(doc.Models[j].Name) {
				for _, sh := range doc.Models[j].StoreHooks() {
					sh.s = &doc.Stores[i]
					out = append(out, sh)
				}
			}
		}
	}
	return
}

func (doc *Document) ModelIPath() string {
	return doc.modipath
}

func (doc *Document) genModels(dropfirst bool) error {
	mgf := jen.NewFile(doc.ModelPkg)
	mgf.HeaderComment(headerComment)
	// mgf.ImportNames(doc.Qualified.Copy())

	for _, enum := range doc.Enums {
		mgf.Add(enum.Code())
	}

	var mods []string
	for _, model := range doc.Models {
		mods = append(mods, model.Name)
		// log.Printf("found model %s", model.Name)
		mgf.Add(model.Codes())
	}
	log.Printf("found models %v", mods)
	mgf.Line()

	if !IsDir(doc.dirmod) {
		if err := os.Mkdir(doc.dirmod, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirmod, err)
			return err
		}
	}

	outname := path.Join(doc.dirmod, doc.gened)
	// log.Printf("%s: %s", doc.ModelPkg, outname)
	if dropfirst && CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	if err := mgf.Save(outname); err != nil {
		log.Fatalf("generate models fail: %s", err)
		return err
	}

	_ = goImports(outname)

	log.Printf("generated '%s/%s' ok", doc.dirmod, doc.gened)
	return nil
}

func (doc *Document) modelWithName(name string) (*Model, bool) {
	for _, m := range doc.Models {
		if m.Name == name {
			return &m, true
		}
	}
	return &Model{}, false
}

func (doc *Document) modelAliasable(name string) bool {
	for _, m := range doc.Models {
		if m.Name == name || strings.HasPrefix(name, m.Name) {
			return true
		}
	}
	return false
}

func (doc *Document) loadModPkg() (ipath string, aliases []string) {
	mpkg := loadPackage(doc.dirmod)
	// log.Printf("loaded mpkg: %s name %q: files %q,%q", mpkg.ID, mpkg.Types.Name(), mpkg.GoFiles, mpkg.CompiledGoFiles)
	// log.Printf("types: %+v, ", mpkg.Types)
	doc.modipath = mpkg.ID
	ipath = mpkg.ID

	doc.lock.Lock()
	doc.Qualified[doc.ModelPkg] = mpkg.ID
	for i, f := range mpkg.Syntax {
		// log.Printf("gofile: %s,", mpkg.CompiledGoFiles[i])
		var ismods bool
		if path.Base(mpkg.CompiledGoFiles[i]) == doc.gened {
			ismods = true
		}
		for k, o := range f.Scope.Objects {
			if o.Kind == ast.Typ {
				doc.modtypes[k] = empty{}
				if ismods && doc.modelAliasable(k) {
					aliases = append(aliases, k)
				}
			}

		}
	}
	doc.lock.Unlock()

	sort.Strings(aliases)

	return
}

func (doc *Document) genStores(dropfirst bool) error {
	ipath, aliases := doc.loadModPkg()

	sgf := jen.NewFile(storepkg)
	sgf.HeaderComment(headerComment)

	sgf.ImportName(ipath, doc.ModelPkg)

	for _, k := range aliases {
		// sgf.Type().Id(k).Op("=").Qual(ipath, k)
		sgf.Comment(jen.Type().Id(k).Op("=").Qual(ipath, k).GoString())
	}
	sgf.Line()

	var tables []jen.Code
	for _, model := range doc.Models {
		if model.IsTable() {
			// name, _ := doc.getModQual(model.Name)
			tables = append(tables, jen.Op("(*").Qual(ipath, model.Name).Op(")(nil)"))
		}
	}
	if len(tables) > 0 {
		sgf.Func().Id("init").Params().BlockFunc(func(g *jen.Group) {
			g.Id("RegisterModel").Call(tables...)
		})
	}

	if !IsDir(doc.dirsto) {
		if err := os.Mkdir(doc.dirsto, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirsto, err)
			return err
		}
	}
	outname := path.Join(doc.dirsto, doc.gened)
	if dropfirst && CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	hasAnyStore := len(doc.Stores) > 0
	for _, store := range doc.Stores {
		sgf.Add(store.Codes(doc.ModelPkg)).Line()
	}

	err := sgf.Save(outname)
	if err != nil {
		log.Fatalf("generate stores fail: %s", err)
		return err
	}

	log.Printf("generated '%s/%s' ok", doc.dirsto, doc.gened)

	if !hasAnyStore {
		log.Print("no store found, skip wrap")
		return nil
	}

	_ = goImports(outname)

	if doc.hasStoreHooks() {
		gfile := path.Join(doc.dirsto, doc.extern)
		ensureGoFile(gfile, "stores/doc_x", doc)
		svd, err := newDST(gfile, storepkg)
		if err != nil {
			return err
		}
		for _, sh := range doc.storeHooks() {
			// log.Printf("check storeHook: %s, %+v", sh.FunName, svd.existFunc(sh.FunName))
			svd.ensureFunc(sh.FunName, sh.dstFuncDecl(doc.modipath))
			if sh.k == upsertES || sh.k == deleteES {
				mesh := storeHook{
					FunName: MigrateES + sh.m.Name,
					ObjName: sh.ObjName,
					k:       MigrateES,
					m:       sh.m,
					s:       sh.s,
				}
				svd.ensureFunc(mesh.FunName, mesh.dstMEFuncDecl(doc.modipath))
				index, decl := mesh.dstMEGenDecl(svd)
				if decl != nil {
					svd.file.Decls[index] = decl
				}

			}
		}
		_ = svd.overwrite()
	}

	_ = doc.ensureWrapPatch()

	_ = doc.encureStoMethod()

	return err
}

func (doc *Document) ensureWrapPatch() bool {
	sfile := path.Join(doc.dirsto, storewf)
	ensureGoFile(sfile, "stores/wrap", nil)
	vd, err := newDST(sfile, storepkg)
	if err != nil {
		return false
	}
	var lastWM string
	foundWM := make(map[string]bool)
	_ = vd.Apply(func(c *dstutil.Cursor) bool {
		return true
	}, func(c *dstutil.Cursor) bool {
		for _, store := range doc.Stores {
			if pn, ok := c.Parent().(*dst.TypeSpec); ok && pn.Name.Obj.Name == storewn {
				if cn, ok := c.Node().(*dst.StructType); ok {
					if isFieldInList(cn.Fields, store.Name) {
						continue
					}
					fd := store.dstWrapField()
					if len(cn.Fields.List) < 3 {
						fd.Decs.Before = dst.EmptyLine
					}
					cn.Fields.List = append(cn.Fields.List, fd)
				}
				continue
			}
			if pn, ok := c.Parent().(*dst.FuncDecl); ok && pn.Name.Name == "NewWithDB" {
				if cn, ok := c.Node().(*dst.BlockStmt); ok {
					if existBlockAssign(cn, store.Name) {
						continue
					}
					nst := store.dstWrapVarAsstmt()
					var arr []dst.Stmt
					n := len(cn.List)
					arr = append(arr, cn.List[0:n-1]...)
					arr = append(arr, nst, cn.List[n-1])
					shimNode(arr[n-2])
					cn.List = arr
				}
				continue
			}

			if cn, ok := c.Node().(*dst.FuncDecl); ok {
				siname := store.ShortIName()
				lastWM = cn.Name.Name
				if lastWM == siname {
					foundWM[siname] = true
				}
			}

		}
		return true
	})
	// log.Printf("found %+v,last wrap method: %s", foundWM, lastWM)
	if len(foundWM) == 0 {
		vd.Apply(nil, func(c *dstutil.Cursor) bool {
			if cn, ok := c.Node().(*dst.FuncDecl); ok && cn.Name.Name == lastWM {
				for _, store := range doc.Stores {
					c.InsertAfter(store.dstWrapFunc())
					log.Printf("insert func %s", store.GetIName())
				}
			}
			return true
		})
	}
	_ = vd.overwrite()
	return true
}

func (doc *Document) encureStoMethod() bool {

	iffile := path.Join(doc.dirsto, "interfaces.go")
	ensureGoFile(iffile, "stores/interfaces", nil)
	vd, err := newDST(iffile, storepkg)
	if err != nil {
		// TODO: new file
		return false
	}

	_ = vd.Apply(func(c *dstutil.Cursor) bool { return true }, func(c *dstutil.Cursor) bool {
		for _, sto := range doc.Stores {
			if pn, ok := c.Parent().(*dst.TypeSpec); ok && pn.Name.Obj.Name == storein {
				if cn, ok := c.Node().(*dst.InterfaceType); ok {
					siname := sto.ShortIName()
					if !isFieldInList(cn.Methods, siname) {
						log.Printf("generate interface method: %q", siname)
						cn.Methods.List = append(cn.Methods.List, newStoInterfaceMethod(siname, sto.GetIName()))
					}
				}
			}
		}
		return true
	})

	_ = vd.overwrite()
	return true
}

func (doc *Document) getMethod(name string) (m Method, ok bool) {
	m, ok = doc.methods[name]
	return
}

func (doc *Document) genWebAPI(dropfirst bool) error {

	if len(doc.WebAPI.Handles) == 0 {
		log.Print("no handle found, skip api")
		return nil
	}

	if !IsDir(doc.dirweb) {
		if err := os.Mkdir(doc.dirweb, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", doc.dirweb, err)
			return err
		}
	}

	afile := path.Join(doc.dirweb, "api.go")
	if !CheckFile(afile) {
		data := map[string]string{"webpkg": doc.WebAPI.getPkgName()}
		if err := renderTmpl("web/api", afile, data); err != nil {
			return err
		}
	}

	outname := path.Join(doc.dirweb, "handle_"+doc.gened)
	if dropfirst && CheckFile(outname) {
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
	// log.Printf("lookuped: %+v", obj)
	if _, ok := obj.(*types.TypeName); !ok {
		log.Fatalf("%v is not a named type", obj)
	}
	objType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		log.Fatalf("type %v is not a struct", obj)
	}
	// log.Printf("NumMethods: %d", objType.NumMethods())
	doc.lock.Lock()
	for i := 0; i < objType.NumMethods(); i++ {
		smt := objType.Method(i)
		if sig, ok := smt.Type().(*types.Signature); ok {
			// log.Printf("method sign: params %s %+v, result: %+v", smt.Name(), sig.Params(), sig.Results())
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

	_ = goImports(outname)

	log.Printf("generated '%s/%s' ok", doc.dirweb, "handle_"+doc.gened)
	return nil
}

func getVarFromTypesVar(v *types.Var) Var {
	typs := v.Type().String()
	if pos := strings.LastIndex(typs, "/"); pos > 0 {
		typs = typs[pos+1:]
	}
	return Var{Name: v.Name(), Type: typs}
}

func goImports(path string) (err error) {
	cmd := exec.Command("goimports", "-w", path)
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	return err
}
