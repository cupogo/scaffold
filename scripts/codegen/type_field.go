// go:build codegen
package main

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

type Field struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type,omitempty"`
	Tags     Tags   `yaml:"tags,flow,omitempty"`
	Qual     string `yaml:"qual,omitempty"`
	IsBasic  bool   `yaml:"basic,omitempty"`
	IsSet    bool   `yaml:"isset,omitempty"`
	Sortable bool   `yaml:"sortable,omitempty"`
	Comment  string `yaml:"comment,omitempty"`
	Query    string `yaml:"query,omitempty"` // '', 'equal', 'wildcard'

	IsChangeWith bool `yaml:"changeWith,omitempty"` // has ChangeWith method

	isOid    bool
	isDate   bool
	isIntDt  bool
	siftFn   string
	siftExt  string
	multable bool
	qtype    string
	colname  string
}

func (f *Field) isMeta() bool {
	return f.Name == metaField || f.Type == metaField
}

func (f *Field) isOwner() bool {
	return f.Name == ownerField || f.Type == ownerField
}

func (f *Field) isAudit() bool {
	return f.Name == auditField || f.Type == auditField
}

func (f *Field) getType() string {
	if len(f.Type) == 0 && len(f.Name) > 0 {
		return f.Name
	}
	return f.Type
}

func (f *Field) cutType() (qn string, typ string, isptr bool) {
	typ = f.getType()
	if len(typ) > 0 && typ[0] == '*' {
		isptr = true
		typ = typ[1:]
	}
	if a, b, ok := strings.Cut(typ, "."); ok {
		qn = a
		typ = b
	}
	return
}

func (f *Field) isScalar() bool {
	if f.Type == "string" || f.Type == "bool" {
		return true
	}

	if strings.HasPrefix(f.Type, "int") || strings.HasPrefix(f.Type, "uint") {
		return true
	}

	if strings.Contains(f.Type, "Money") {
		return true
	}

	if strings.HasSuffix(f.Type, "Status") || strings.HasSuffix(f.Type, "Type") {
		return true
	}

	return false
}

var replTrimUseZero = strings.NewReplacer(",use_zero", "")

func (f *Field) bunPatchTags() (out Tags) {
	out = f.Tags.Copy()
	if !out.Has("bun") && out.Has("pg") {
		v := out["pg"]
		out["bun"] = replTrimUseZero.Replace(v)
	}
	return out
}

func (f *Field) typeCode(pkgs ...string) *jen.Statement {
	st := jen.Empty()
	qn, typ, _ := f.cutType()
	if len(qn) > 0 {
		if len(f.Qual) > 0 {
			return st.Qual(f.Qual, typ)
		}
		if qual, ok := getQual(qn); ok {
			return st.Qual(qual, typ)
		}
		return st.Qual(qn, typ)
	}
	if len(pkgs) == 1 && len(pkgs[0]) > 0 {
		return st.Qual(pkgs[0], typ)
	}
	return st.Id(typ)
}

func (f *Field) isEmbed() bool {
	return len(f.Name) == 0 || len(f.Type) == 0
}

func (f *Field) preCode() (st *jen.Statement) {
	isEmbed := f.isEmbed()
	st = jen.Empty()
	if isEmbed {
		st.Line()
	}
	if len(f.Comment) > 0 {
		st.Comment(f.Comment).Line()
	}
	if !isEmbed {
		st.Id(f.Name)
	}

	return st
}

func (f *Field) defCode() jen.Code {
	qn, typ, isptr := f.cutType()
	st := jen.Empty()
	if isptr {
		st.Op("*")
	}
	if len(qn) > 0 {
		if len(f.Qual) > 0 {
			return st.Qual(f.Qual, typ)
		}
		if qual, ok := getQual(qn); ok {
			return st.Qual(qual, typ)
		}
		return st.Qual(qn, typ)
	}
	return st.Id(typ)
}

func (f *Field) Code(idx int) jen.Code {
	st := f.preCode().Add(f.defCode())

	if len(f.Tags) > 0 {
		tags := f.bunPatchTags()
		if j, ok := tags["json"]; ok {
			if a, b, ok := strings.Cut(j, ","); ok {
				if f.isScalar() && !tags.Has("form") {
					tags["form"] = a
				}
				if b == "" && strings.HasSuffix(f.Type, "DateTime") {
					tags["json"] = a + ",omitempty"
				}
			} else if f.isScalar() && !tags.Has("form") {
				tags["form"] = j
			}

			tags.extOrder(idx)
		}

		st.Tag(tags)
	}

	if f.isEmbed() {
		st.Line()
	}

	return st
}

// return column name, is in db and is unquie
func (f *Field) ColName() (cn string, hascol bool, unique bool) {
	if s, ok := f.Tags.GetAny("pg", "bun"); ok && len(s) > 0 && s != "-" {
		hascol = true
		if a, b, ok := strings.Cut(s, ","); ok {
			cn = a
			unique = strings.Contains(b, "unique")
		}
		if len(cn) == 0 {
			cn = Underscore(f.Name)
		}
	} else if len(f.colname) > 0 {
		cn = f.colname
	}
	return
}

func (f *Field) relMode() (string, bool) {
	if s, ok := f.Tags.GetAny("pg", "bun"); ok && len(s) > 4 {
		if strings.HasPrefix(s, "rel:has-one") {
			return "has-one", true
		}
		if strings.HasPrefix(s, "rel:has-many") {
			return "has-many", true
		}
	}
	return "", false
}

func (f *Field) getArgTag() string {
	if s, ok := f.Tags["form"]; ok {
		return LcFirst(s)
	}

	if j, ok := f.Tags["json"]; ok {
		if a, _, ok := strings.Cut(j, ","); ok {
			return LcFirst(a)
		}
		return LcFirst(j)
	}

	return LcFirst(f.Name)
}

func (f *Field) queryCode(idx int, pkgs ...string) jen.Code {

	if len(f.Comment) > 0 {
		if f.isDate {
			f.Comment += " + during"
		}
	}
	st := f.preCode()
	if len(f.qtype) > 0 {
		st.Id(f.qtype)
	} else {
		st.Add(f.typeCode(pkgs...))
	}

	if json, jok := f.Tags["json"]; jok {
		tags := Tags{"json": json}
		if !f.Tags.Has("form") {
			if a, _, ok := strings.Cut(json, ","); ok {
				tags["form"] = a
			} else {
				tags["form"] = json
			}
		} else {
			tags["form"] = f.Tags["form"]
		}
		tags.extOrder(idx + 1)
		st.Tag(tags)
	}

	return st
}

type Fields []Field

// Codes return fields code of main and basic
func (z Fields) Codes(basicName string) (mcs, bcs []jen.Code) {
	var hasMeta bool
	var setBasic bool
	var idx int
	for _, field := range z {
		if !field.isEmbed() {
			idx++
		}
		if field.IsSet || field.IsBasic {
			bcs = append(bcs, field.Code(idx))
			if !setBasic {
				mcs = append(mcs, jen.Id(basicName).Line())
				setBasic = true
			}
		} else {
			mcs = append(mcs, field.Code(idx))
		}
		if field.isMeta() {
			hasMeta = true
		}
	}
	if hasMeta {
		bcs = append(bcs, metaUpCode(true))
	}
	return
}

func (z Fields) relHasOne() (cols []string) {
	for i := range z {
		if n, ok := z[i].relMode(); ok && i > 0 {
			// 上一个字段必须指向关联的主键
			if n == "has-one" && z[i-1].Name == z[i].Name+"ID" {
				cols = append(cols, z[i].Name)
			}
		}
	}
	return
}

func (z Fields) Relations() (cols []string) {
	for i := range z {
		if _, ok := z[i].relMode(); ok && i > 0 {
			cols = append(cols, z[i].Name)
		}
	}
	return
}

func (z Fields) withName(name string) (*Field, bool) {
	for _, field := range z {
		if field.Name == name {
			return &field, true
		}
	}
	return nil, false
}

func (f *Field) parseQuery() (fn, ext string, ok bool) {
	var a string
	a, ext, _ = strings.Cut(f.Query, ",")
	switch a {
	case "oids":
		fn, ok = "siftOIDs", f.Type == "oid.OID"
	case "equal":
		fn, ok = "siftEquel", true
	case "ice", "ilike":
		fn, ok = "siftICE", true
	case "match":
		fn, ok = "siftMatch", true
	case "date":
		fn, ok = "siftDate", true
	case "great":
		fn, ok = "siftGreat", true
	case "less":
		fn, ok = "siftLess", true
	}
	return
}
