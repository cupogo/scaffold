// go:build codegen
package gens

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
	Descr    string `yaml:"descr,omitempty"`
	Query    string `yaml:"query,omitempty"` // '', 'equal', 'wildcard'

	Compare CompareType `yaml:"compare,omitempty"` // scalar, equalTo

	IsChangeWith bool `yaml:"changeWith,omitempty"` // has ChangeWith method
	IgnoreCase   bool `yaml:"icse,omitempty"`       // Ignore case sensitivity equality

	isOid    bool
	isDate   bool
	isIntDt  bool
	siftFn   string
	siftOp   string
	siftExt  string
	multable bool
	qtype    string
	colname  string
	bson     bool

	mod *Model
	qer *Query
}

func (f *Field) isMeta() bool {
	return matchs(metaField, f.Name, f.Type)
}

func (f *Field) isOwner() bool {
	return matchs(ownerField, f.Name, f.Type)
}

func (f *Field) isAudit() bool {
	return matchs(auditField, f.Name, f.Type)
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

func (f *Field) isOID() bool {
	qn, typ, _ := f.cutType()
	return qn == "oid" && typ == "OID"
}

func (f *Field) isScalar() bool {
	if f.Type == "string" || f.Type == "bool" ||
		strings.HasPrefix(f.Type, "int") || strings.HasPrefix(f.Type, "uint") ||
		strings.Contains(f.Type, "Money") || strings.HasSuffix(f.Type, "DateTime") || strings.HasSuffix(f.Type, "Date") ||
		strings.HasSuffix(f.Type, "Status") || strings.HasSuffix(f.Type, "Type") {
		return true
	}

	return f.Compare == CompareScalar
}

func (f *Field) maybeEnum() bool {
	if f.Type == "string" || f.Type == "bool" {
		return false
	}

	if strings.HasPrefix(f.Type, "*") || strings.HasPrefix(f.Type, "[]") ||
		strings.HasPrefix(f.Type, "int") || strings.HasPrefix(f.Type, "uint") ||
		strings.HasSuffix(f.Type, "DateTime") || strings.Contains(f.Type, "Money") {
		return false
	}

	return true
}

var replTrimUseZero = strings.NewReplacer(",use_zero", "")

func (f *Field) bunPatchTags() (out Tags) {
	out = f.Tags.Clone()
	if !out.Has("bun") && out.Has("pg") {
		v := out["pg"]
		out["bun"] = replTrimUseZero.Replace(v)
	}
	if f.isOID() && !out.Has(TagSwaggerType) {
		out[TagSwaggerType] = "string"
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

func (f *Field) commentCode() (st *jen.Statement) {
	if len(f.Comment) > 0 {
		st = jen.Empty()
		st.Comment(f.Comment).Line()
		if f.mod != nil && f.maybeEnum() {
			if ed, ok := f.mod.doc.getEnumDoc(f.Type); ok {
				if _, b, ok := strings.Cut(f.Query, ","); ok && b == "decode" {
					f.qtype = "string"
				}
				// log.Printf("field %s.%s has enums %v", f.Name, f.Type, ed)
				for _, line := range ed.Lines {
					st.Comment(" * " + line).Line()
				}
				if !f.Tags.Has("enums") {
					f.Tags["enums"] = strings.Join(ed.Codes, ",")
				}
				if len(ed.SwaggerT) > 0 && !f.Tags.Has(TagSwaggerType) {
					f.Tags[TagSwaggerType] = ed.SwaggerT
				}
			}
		}
		jcodeDesc(st, f.Descr, "")
	}
	return
}

func (f *Field) preCode() (st *jen.Statement) {
	isEmbed := f.isEmbed()
	st = jen.Empty()
	if isEmbed {
		st.Line()
	}

	if c := f.commentCode(); c != nil {
		st.Add(c)
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

func (f *Field) Code(idx, max int) jen.Code {
	st := f.preCode().Add(f.defCode())

	tags := f.bunPatchTags()
	if f.isEmbed() {
		if f.bson || f.inMgm() {
			if tags == nil {
				tags = make(Tags)
			}
			tags["bson"] = ",inline"
		}
	}
	if len(tags) > 0 {
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

			tags.extOrder(idx, max)
		}

		st.Tag(tags)
	}

	if f.isEmbed() {
		st.Line()
	}

	return st
}

func (f *Field) dbCode() DbCode {
	if f.mod != nil && f.mod.doc != nil {
		return f.mod.doc.DbCode
	}
	return ""
}

func (f *Field) inMgm() bool {
	return f.dbCode() == DbMgm
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

func (f *Field) BsonName() (name string, has bool) {
	if s, ok := f.Tags.GetAny("bson", "json"); ok && len(s) > 0 && s != "-" {
		if a, _, ok := strings.Cut(s, ","); ok {
			name = a
		} else {
			name = s
		}
		has = len(name) > 0
	}
	return
}

func (f *Field) isTagJsonString() bool {
	if s, ok := f.Tags["json"]; ok && strings.HasSuffix(s, ",string") {
		return true
	}
	return false
}

func (f *Field) relMode(dbtag ...string) (string, bool) {
	if len(dbtag) == 0 {
		dbtag = []string{"pg", "bun"}
	}
	if s, ok := f.Tags.GetAny(dbtag...); ok && len(s) > 4 {
		if strings.HasPrefix(s, "rel:belongs-to") {
			return relBelongsTo, true
		}
		if strings.HasPrefix(s, "rel:has-one") {
			return relHasOne, true
		}
		if strings.HasPrefix(s, "rel:has-many") {
			return relMasMany, true
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
		// TODO: insert comments of enum
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
		if f.Tags.Has(TagSwaggerIgnore) {
			tags[TagSwaggerIgnore] = f.Tags[TagSwaggerIgnore]
		} else if f.Tags.Has(TagSwaggerType) {
			tags[TagSwaggerType] = f.Tags[TagSwaggerType]
		}
		if f.Tags.Has(TagSpecBinding) {
			tags["binding"] = f.Tags[TagSpecBinding]
		}
		tags.extOrder(idx+1, 0)
		st.Tag(tags)
	}

	return st
}

type Fields []Field

// Codes return fields code of main and basic
func (z Fields) Codes(basicName string, isTable, bsonable bool) (mcs, bcs []jen.Code) {
	var hasMeta bool
	var setBasic bool
	var idx int
	for _, field := range z {
		field.bson = bsonable
		if !field.isEmbed() && !field.Tags.Has(TagSwaggerIgnore) {
			idx++
		}
		if (isTable || bsonable) && (field.IsSet || field.IsBasic) {
			bcs = append(bcs, field.Code(idx, len(z)))
			if !setBasic {
				stb := jen.Id(basicName)
				if bsonable {
					stb.Tag(Tags{"bson": ",inline"})
				}
				mcs = append(mcs, stb.Line())
				setBasic = true
			}
		} else {
			mcs = append(mcs, field.Code(idx, len(z)))
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

func (z Fields) relHasOne() (out Fields) {
	for i := range z {
		if n, ok := z[i].relMode(); ok && i > 0 {
			// 上一个字段必须指向关联的主键
			if (n == relBelongsTo || n == relHasOne) && z[i-1].Name == z[i].Name+"ID" {
				out = append(out, z[i])
			}
		}
	}
	return
}

func (z Fields) Relations() (out []string) {
	for i := range z {
		if _, ok := z[i].relMode(); ok && i > 0 {
			out = append(out, z[i].Name)
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

type Query struct {
	sift   string
	ext    string
	add    bool
	custom bool
	both   bool
}

func (f *Field) parseQuery() (q Query, ok bool) {
	var a string
	a, q.ext, _ = strings.Cut(f.Query, ",")
	switch a {
	case "oids":
		q.sift, ok = "siftOIDs", f.Type == "oid.OID" || f.Type == "oid.OIDs"
	case "equal":
		q.sift, ok = "siftEqual", true
	case "ice", "ilike", "ice2":
		q.sift, ok = "siftICE", true
		q.both = a == "ice2"
	case "match", "match2":
		q.sift, ok = "siftMatch", true
		q.both = a == "match2"
	case "date":
		q.sift, ok = "siftDate", true
	case "great":
		q.sift, ok = "siftGreat", true
	case "less":
		q.sift, ok = "siftLess", true
	default:
		q.custom = a == "custom"
		ok = len(a) > 0 && len(q.ext) > 0
	}
	q.add = q.ext == "ints" || q.ext == "strs" || q.ext == "oids"
	if ok {
		f.qer = &q
	}
	return
}

func (f Field) siftCode(ismg bool) jen.Code {
	if len(f.siftFn) == 0 {
		return nil
	}
	jsv := jen.Id("spec").Dot(f.Name)
	if ismg {
		cn, ok := f.BsonName()
		if !ok {
			return nil
		}
		return jen.Id("q").Op("=").Id("mg"+ToExported(f.siftFn)).Call(
			jen.Id("q"), jen.Lit(cn), jsv)
	}
	cn, indb, _ := f.ColName()
	if !indb && len(f.siftFn) == 0 {
		return nil
	}
	params := []jen.Code{jen.Id("q"), jen.Lit(cn), jsv}
	cfn := f.siftFn
	if f.isDate && f.isIntDt {
		params = append(params, jen.True())
	}
	params = append(params, jen.False())
	if f.qer != nil && f.qer.both {
		params = append(params, jen.True())
	}
	return jen.Id("q").Op(",").Id("_").Op("=").Id(cfn).Call(params...)
}
