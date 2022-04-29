// go:build codegen
package main

import (
	"fmt"
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

type Document struct {
	name   string
	dirmod string
	dirsto string
	lock   sync.Mutex

	Models    []Model `yaml:"models"`
	ModelPkg  string  `yaml:"modelpkg"`
	Qualified Maps    `yaml:"qualified"` // imports name
	Stores    []Store `yaml:"stores"`
}

func (d *Document) getQual(k string) (string, bool) {
	v, ok := d.Qualified[k]
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

func (d *Document) genModels(dropfirst bool) error {
	mgf := jen.NewFile(d.ModelPkg)
	mgf.HeaderComment(headerComment)

	for _, model := range doc.Models {
		log.Printf("found model %s", model.Name)
		mgf.Add(model.Codes())
	}
	mgf.Line()

	if !osutil.IsDir(d.dirmod) {
		if err := os.Mkdir(d.dirmod, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", d.dirmod, err)
			return err
		}
	}

	outname := path.Join(d.dirmod, d.name)
	log.Printf("%s: %s", d.ModelPkg, outname)
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
	return nil
}

func (d *Document) genStores(dropfirst bool) error {
	mpkg := loadPackage(d.dirmod)
	// if err != nil {
	// 	log.Printf("get package fail: %s", err)
	// }
	log.Printf("loaded mpkg: %s name %q: path %q", mpkg.ID, mpkg.Types.Name(), mpkg.Types.Path())
	log.Printf("types: %+v, ", mpkg.Types)

	sgf := jen.NewFile(storepkg)
	sgf.HeaderComment(headerComment)

	sgf.ImportName(mpkg.ID, d.ModelPkg)
	d.lock.Lock()
	doc.Qualified[d.ModelPkg] = mpkg.ID
	d.lock.Unlock()

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

	if !osutil.IsDir(d.dirsto) {
		if err := os.Mkdir(d.dirsto, 0755); err != nil {
			log.Printf("mkdir %s fail: %s", d.dirsto, err)
			return err
		}
	}
	outname := path.Join(d.dirsto, d.name)
	if dropfirst && osutil.CheckFile(outname) {
		if err := os.Remove(outname); err != nil {
			log.Printf("drop %s fail: %s", outname, err)
			return err
		}
	}

	spkg := loadPackage(d.dirsto)
	log.Printf("loaded spkg: %s name %q", spkg.ID, spkg.Types.Name())

	for _, store := range d.Stores {
		sgf.Add(store.Codes(d.ModelPkg)).Line()
	}

	err := sgf.Save(outname)
	if err != nil {
		log.Fatalf("generate stores fail: %s", err)
		return err
	}
	return nil
}
