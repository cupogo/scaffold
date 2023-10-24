// go:build codegen
package gens

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"maps"
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

var (
	qualifications = map[string]string{
		"comm":  "github.com/cupogo/andvari/models/comm",
		"oid":   "github.com/cupogo/andvari/models/oid",
		"pgx":   "github.com/cupogo/andvari/stores/pgx",
		"utils": "github.com/cupogo/andvari/utils",
	}
)

func exQual(k string) (string, bool) {
	v, ok := qualifications[k]
	return v, ok
}

type Tags map[string]string

// Copy ...
func (m Tags) Copy() (out Tags) {
	return maps.Clone(m)
}

func (m Tags) CleanKeys(keys ...string) {
	maps.DeleteFunc(m, func(k, v string) bool {
		for _, s := range keys {
			if s == k {
				return true
			}
		}
		return false
	})
}

func (m Tags) GetVal(key string) (val string, ok bool) {
	if val, ok = m[key]; ok {
		var a string
		if a, _, ok = strings.Cut(val, ","); ok {
			val = a
		}
	}
	return
}

func (m Tags) FillKey(dst, src string) {
	if val, ok := m.GetVal(src); ok {
		m[dst] = val
	}
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

	Module string `yaml:"-"`

	DbCode    DbCode  `yaml:"dbcode"` //  default:"bun"
	Enums     []Enum  `yaml:"enums"`
	ModelPkg  string  `yaml:"modelpkg"`
	Models    []Model `yaml:"models"`
	Qualified Tags    `yaml:"depends"` // imports name
	Stores    []Store `yaml:"stores"`
	WebAPI    WebAPI  `yaml:"webapi"`
}

func (doc *Document) Check() error {
	module, err := GetModule("./")
	if err != nil {
		return err
	}
	doc.Module = module
	if 0 == len(doc.ModelPkg) {
		return errors.New("empty modelpkg")
	}

	if 0 == len(doc.Models) && 0 == len(doc.Enums) {
		return errors.New("empty models and enums")
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
	if !ok {
		if qoid, _ok := exQual(k); _ok {
			return qoid, true
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
	if len(doc.DbCode) == 0 {
		if s, ok := os.LookupEnv("SCAFFOLD_DB_CODE"); ok && len(s) > 0 {
			doc.DbCode = DbCode(s)
		}
	}

	log.Printf("loaded %d models, dbcode %s", len(doc.Models), doc.DbCode)
	// log.Printf("loaded webapi uris %+v", doc.WebAPI.URIs)

	return doc, nil
}

func (doc *Document) Init() {
	for i := range doc.Models {
		doc.Models[i].init(doc)
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
	return doc != nil && doc.DbCode == DbPgx
}

func (doc *Document) IsMongo() bool {
	return doc != nil && doc.DbCode == DbMgm
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

func (doc *Document) loadEsModels() (out []Model) {
	for i := 0; i < len(doc.Stores); i++ {
		for j := 0; j < len(doc.Models); j++ {
			if doc.Stores[i].hasModel(doc.Models[j].Name) {
				if doc.Models[j].hasAnyStoreHook(upsertES, deleteES) {
					out = append(out, doc.Models[j])
				}
			}
		}
	}
	return
}

func (doc *Document) ModelIPath() string {
	return doc.modipath
}

type oidKey struct {
	name string
	code string
}

func (doc *Document) genModels(dropfirst bool) error {
	mgf := jen.NewFile(doc.ModelPkg)
	mgf.HeaderComment(headerComment)
	// mgf.ImportNames(doc.Qualified.Copy())

	var oidKeys []oidKey
	for _, mod := range doc.Models {
		if len(mod.OIDKey) >= 2 {
			oidKeys = append(oidKeys, oidKey{mod.Name + "Label", mod.OIDKey[0:2]})
		}
	}
	if len(oidKeys) > 0 {
		oidQual, _ := doc.getQual("oid")
		mgf.Func().Id("init").Params().BlockFunc(func(g *jen.Group) {
			for _, ok := range oidKeys {
				g.Qual(oidQual, "RegistCate").Call(jen.Id(ok.name), jen.Lit(ok.code))
			}
		})
	}

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
	if len(mods) > 0 {
		mgf.Line()
	}

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

func (doc *Document) allModelAliases() (exports, aliases []string) {
	for _, m := range doc.Models {
		if m.ExportSingle {
			exports = append(exports, m.Name)
		} else {
			aliases = append(aliases, m.Name)
		}

		if mns := m.GetPlural(); mns != m.Name && (len(m.Plural) > 0 || m.WithPlural) {
			if m.ExportPlural {
				exports = append(exports, mns)
			} else {
				aliases = append(aliases, mns)
			}
		}
	}
	return
}

func (doc *Document) loadModPkg() (ipath string) {
	mpkg := loadPackage(doc.dirmod)
	// log.Printf("loaded mpkg: %s name %q: files %q,%q", mpkg.ID, mpkg.Types.Name(), mpkg.GoFiles, mpkg.CompiledGoFiles)
	// log.Printf("types: %+v, ", mpkg.Types)
	doc.modipath = mpkg.ID
	ipath = mpkg.ID

	doc.lock.Lock()
	doc.Qualified[doc.ModelPkg] = mpkg.ID
	for _, f := range mpkg.Syntax {
		for k, o := range f.Scope.Objects {
			if o.Kind == ast.Typ {
				doc.modtypes[k] = empty{}
			}

		}
	}
	doc.lock.Unlock()
	return
}

func (doc *Document) genStores(dropfirst bool) (err error) {
	ipath := doc.loadModPkg()

	sgf := jen.NewFile(storepkg)
	sgf.HeaderComment(headerComment)

	sgf.ImportName(ipath, doc.ModelPkg)

	exports, aliases := doc.allModelAliases()
	sort.Strings(aliases)
	sort.Strings(exports)
	for _, k := range exports {
		sgf.Type().Id(k).Op("=").Qual(ipath, k)
	}
	for _, k := range aliases {
		jal := jen.Type().Id(k).Op("=").Qual(ipath, k)
		sgf.Comment(jal.GoString())
	}
	sgf.Line()

	var tables []jen.Code
	var cloads []jPair
	for _, model := range doc.Models {
		ctable, cload := model.codeRegSto()
		if ctable != nil {
			tables = append(tables, ctable)
		}
		if cload.p1 != nil {
			cloads = append(cloads, cload)
		}
	}
	if len(tables) > 0 || len(cloads) > 0 {
		sgf.Func().Id("init").Params().BlockFunc(func(g *jen.Group) {
			if len(tables) > 0 {
				g.Id("RegisterModel").Call(tables...)
			}
			pgxQual, _ := doc.getQual("pgx")
			for _, cload := range cloads {
				g.Id("RegisterLoader").Call(cload.p1, jen.Qual(pgxQual, "GetModelByID").Index(cload.p2))
			}
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
	if !hasAnyStore {
		log.Print("no store found, skip wrap")
		return nil
	}
	gfile := path.Join(doc.dirsto, doc.extern)
	var svd *vdst
	if CheckFile(gfile) {
		svd, err = newDST(gfile, storepkg)
		if err != nil {
			return
		}
	}

	for _, store := range doc.Stores {
		if svd != nil {
			if svd.existFunc("new" + store.GetIName()) {
				store.extInit = true
			}
			if _, ok := svd.existMethod("strap", store.Name); ok {
				store.extStrap = true
			}
		}

		sgf.Add(store.Codes(doc.ModelPkg)).Line()
	}

	err = sgf.Save(outname)
	if err != nil {
		log.Fatalf("generate stores fail: %s", err)
		return err
	}

	log.Printf("generated '%s/%s' ok", doc.dirsto, doc.gened)

	_ = goImports(outname)

	if doc.hasStoreHooks() {
		ensureGoFile(gfile, "stores/doc_x", doc)
		if svd == nil {
			svd, err = newDST(gfile, storepkg)
			if err != nil {
				return err
			}
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
	ensureGoFile(sfile, "stores/wrap", doc)
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
					nst := store.dstWrapVarAsstmt()
					if idx, ok := existBlockAssign(cn, store.Name); ok {
						cn.List[idx] = nst
						continue
					}
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

	ensureGoFile(path.Join(doc.dirweb, "api.go"), "web/api", map[string]string{
		"Module": doc.Module,
		"WebPkg": doc.WebAPI.GetPkgName(),
	})

	outname := path.Join(doc.dirweb, "handle_"+doc.gened)
	if dropfirst && CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	mpkg := loadPackage(doc.dirmod)
	spkg := loadPackage(doc.dirsto)

	wgf := jen.NewFile(doc.WebAPI.GetPkgName())
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
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	return err
}

func (doc *Document) getEnumDoc(name string) (ed EnumDoc, ok bool) {
	for _, e := range doc.Enums {
		if e.Name == name {
			return e.docComments()
		}
	}
	return
}
