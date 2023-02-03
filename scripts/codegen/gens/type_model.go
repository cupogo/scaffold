// go:build codegen
package gens

import (
	"log"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Model struct {
	Comment    string   `yaml:"comment,omitempty"`
	Name       string   `yaml:"name"`
	TableTag   string   `yaml:"tableTag,omitempty"`
	Fields     Fields   `yaml:"fields"`
	Plural     string   `yaml:"plural,omitempty"`
	OIDCat     string   `yaml:"oidcat,omitempty"`
	StoHooks   Tags     `yaml:"hooks,omitempty"`
	SpecExtras Fields   `yaml:"specExtras,omitempty"`
	Sifters    []string `yaml:"sifters,omitempty"`
	SpecUp     string   `yaml:"specUp,omitempty"`

	DiscardUnknown bool `yaml:"discardUnknown,omitempty"` // 忽略未知的列
	WithColumnGet  bool `yaml:"withColumnGet,omitempty"`  // Get时允许定制列
	WithColumnList bool `yaml:"withColumnList,omitempty"` // List时允许定制列
	DbTriggerSave  bool `yaml:"dbTriggerSave,omitempty"`  // 已存在保存时生效的数据表触发器
	WithCreatedSet bool `yaml:"withCreatedSet,omitempty"` // 开放created的设置
	ForceCreate    bool `yaml:"forceCreate,omitempty"`    // 强行创建不报错
	PostNew        bool `yaml:"postNew,omitempty"`
	PreSet         bool `yaml:"preSet,omitempty"`
	PostSet        bool `yaml:"postSet,omitempty"`
	DisableLog     bool `yaml:"disableLog,omitempty"` // 不记录model的日志

	doc *Document
	pkg string
}

func (m *Model) String() string {
	return m.Name
}

func (m *Model) GetPlural() string {
	if m.Plural == "" {
		m.Plural = Plural(m.Name)
	}
	return m.Plural
}

func (m *Model) tableName() string {
	if m.TableTag == "" {
		return Underscore(m.GetPlural())
	}
	tt := m.TableTag
	if pos := strings.Index(tt, "table:"); pos > -1 {
		tt = tt[pos+6:]
	}
	if pos := strings.Index(tt, ","); pos > 0 {
		tt = tt[0:pos]
	}
	return tt
}

func (m *Model) tableAlias() string {
	if m.TableTag == "" {
		return Underscore(m.GetPlural())
	}
	tt := m.TableTag
	if pos := strings.Index(tt, "alias:"); pos > -1 {
		return tt[pos+6:]
	}
	if pos := strings.Index(tt, "table:"); pos > -1 {
		tt = tt[pos+6:]
	}
	return tt
}

func (m *Model) TableField() jen.Code {
	tt := m.TableTag
	if tt == "" {
		tt = Underscore(m.GetPlural())
	}
	if m.doc.IsPG10() {
		if m.DiscardUnknown && !strings.Contains(tt, "discard_unknown_columns") {
			tt += ",discard_unknown_columns"
		}
		return jen.Id("tableName").Add(jen.Struct()).Tag(Tags{"pg": tt}).Line()
	}
	return jen.Id("comm.BaseModel").Tag(Tags{"json": "-", "bun": "table:" + tt}).Line()
}

func (m *Model) UniqueOne() (name, col string, onlyOne bool) {
	var count int
	for _, field := range m.Fields {
		if cn, _, ok := field.ColName(); ok {
			count++
			name = field.Name
			col = cn
		}
	}
	onlyOne = count == 1
	return
}

func (m *Model) ChangablCodes() (ccs []jen.Code, scs []jen.Code) {
	if m.PreSet {
		scs = append(scs, jen.Id("z").Dot("PreSet").Call(jen.Op("&").Id("o")))
	}
	var hasMeta bool
	var hasOwner bool
	var idx int
	for _, field := range m.Fields {
		if !field.IsSet || field.isEmbed() {
			if field.isMeta() {
				hasMeta = true
			} else if field.isOwner() {
				hasOwner = true
			}
			continue
		}
		idx++
		var code *jen.Statement
		if len(field.Comment) > 0 {
			code = jen.Comment(field.Comment).Line()
		} else {
			code = jen.Empty()
		}
		code.Id(field.Name)
		cn, isInDb, _ := field.ColName()
		qn, tn, isptr := field.cutType()

		jcond := jen.Id("o").Dot(field.Name).Op("!=").Nil()
		if field.isScalar() {
			jcond.Op("&&").Id("z").Dot(field.Name).Op("!=").Op("*").Id("o").Dot(field.Name)
		} else if field.Compare == CompareEqualTo {
			jcond.Op("&&!").Id("z").Dot(field.Name).Dot("EqualTo").Call(jen.Id("o").Dot(field.Name))
		}

		if qn == "oid" && tn == "OID" {
			field.Type = "string"
			field.isOid = true
			field.Qual = ""
			tn = field.Type
		}
		if len(field.Qual) > 0 {
			code.Op("*").Qual(field.Qual, tn)
		} else {
			code.Op("*").Id(tn)
		}
		if s, ok := field.Tags["json"]; ok {
			tags := Tags{"json": s}
			tags.extOrder(idx)
			code.Tag(tags)
		}

		jcsa := jen.Id("z").Dot("SetChange").Call(jen.Lit(cn))

		ccs = append(ccs, code)
		scs = append(scs, jen.If(jcond).BlockFunc(func(g *jen.Group) {
			if isInDb && !m.DisableLog && !field.isOid {
				g.Id("z").Dot("LogChangeValue").Call(jen.Lit(cn), jen.Id("z").Dot(field.Name), jen.Id("o").Dot(field.Name))
			}
			if field.isOid {
				g.If(jen.Id("id").Op(":=").Id("oid").Dot("Cast").Call(jen.Op("*").Id("o").Dot(field.Name)).Op(";").
					Id("z").Dot(field.Name).Op("!=").Id("id")).BlockFunc(func(g1 *jen.Group) {
					if isInDb && !m.DisableLog {
						g1.Id("z").Dot("LogChangeValue").Call(jen.Lit(cn), jen.Id("z").Dot(field.Name), jen.Id("id"))
					}
					g1.Id("z").Dot(field.Name).Op("=").Id("id")
					if isInDb && m.DisableLog {
						g1.Add(jcsa)
					}
				})
			} else if field.IsChangeWith {
				g.Id("z").Dot(field.Name).Dot("ChangeWith").Call(jen.Id("o").Dot(field.Name))
			} else if isptr {
				g.Id("z").Dot(field.Name).Op("=").Id("o").Dot(field.Name)
			} else {
				g.Id("z").Dot(field.Name).Op("=").Op("*").Id("o").Dot(field.Name)
			}
			if isInDb && !field.isOid && m.DisableLog {
				g.Add(jcsa)
			}
		}))
	}

	if m.WithCreatedSet {
		idx++
		// CreatedAt time.Time `bson:"createdAt" json:"createdAt" form:"createdAt" pg:"created,notnull,default:now()" extensions:"x-order=["` // 创建时间
		ccs = append(ccs, createdUpCode(idx))
		scs = append(scs, jen.If(jen.Id("o").Dot(createdField).Op("!=").Nil().BlockFunc(func(g *jen.Group) {
			g.Id("z").Dot(createdField).Op("=").Op("*").Id("o").Dot(createdField)
			g.Id("z").Dot("SetChange").Call(jen.Lit(createdColumn))
		})))
	}

	if hasMeta {
		name := "MetaDiff"
		ccs = append(ccs, metaUpCode())
		scs = append(scs, jen.If(jen.Id("o").Dot(name).Op("!=").Nil().Op("&&").Id("z").Dot("MetaUp").Call(jen.Id("o").Dot(name))).Block(
			jen.Id("z").Dot("SetChange").Call(jen.Lit("meta")),
		))
	}
	if hasOwner {
		idx++
		name := "OwnerID"
		ccs = append(ccs, ownerUpCode(idx))
		scs = append(scs, jen.If(jen.Id("o").Dot(name).Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			g.If(jen.Id("id").Op(":=").Id("oid").Dot("Cast").Call(jen.Op("*").Id("o").Dot(name)).Op(";").
				Id("z").Dot(name).Op("!=").Id("id")).BlockFunc(func(g1 *jen.Group) {
				if !m.DisableLog {
					g1.Id("z").Dot("LogChangeValue").Call(jen.Lit("owner_id"), jen.Id("z").Dot(name), jen.Id("id"))
				}
				g1.Id("z").Dot("SetOwnerID").Call(jen.Id("id"))
				if m.DisableLog {
					g1.Id("z").Dot("SetChange").Call(jen.Lit("owner_id"))
				}
			})
		}))
	}

	if m.PostSet {
		scs = append(scs, jen.Id("z").Dot("PostSet").Call(jen.Op("&").Id("o")))
	}
	return
}

func (m *Model) Codes() jen.Code {
	st := jen.Empty()
	basicName := m.Name + "Basic"
	var cs []jen.Code
	if m.IsTable() {
		st.Comment("consts of " + m.Name + " " + m.shortComment()).Line()
		st.Const().DefsFunc(func(g *jen.Group) {
			g.Id(m.Name + "Table").Op("=").Lit(m.tableName())
			g.Id(m.Name + "Alias").Op("=").Lit(m.tableAlias())
			g.Id(m.Name + "Label").Op("=").Lit(LcFirst(m.Name))
		}).Line()
		cs = append(cs, m.TableField())
	}
	mcs, bcs := m.Fields.Codes(basicName)
	cs = append(cs, mcs...)
	st.Comment(m.Name + " " + m.Comment).Line()
	var prefix string
	if m.doc != nil {
		prefix = m.doc.ModelPkg
	}

	st.Type().Id(m.Name).Struct(cs...).Add(jen.Comment("@name " + prefix + m.Name)).Line().Line()

	if len(bcs) > 0 {
		st.Type().Id(basicName).Struct(bcs...).Add(jen.Comment("@name " + prefix + basicName)).Line().Line()
	}

	st.Type().Id(m.GetPlural()).Index().Id(m.Name).Line().Line()

	if jhk := m.hookModelCodes(); jhk != nil {
		st.Add(jhk)
	}
	if m.DisableLog {
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("DisableLog").Params().Bool().Block(jen.Return(jen.Lit(true)))
		st.Line()
	}
	if jc := m.basicCodes(); jc != nil {
		st.Add(jc)
	}

	if ccs, scs := m.ChangablCodes(); len(ccs) > 0 {
		changeSetName := m.Name + "Set"
		st.Type().Id(changeSetName).Struct(ccs...).Add(jen.Comment("@name " + prefix + changeSetName)).Line().Line()
		// scs = append(scs, jen.Return(jen.Id("z").Dot("CountChange").Call().Op(">0")))
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("SetWith").Params(jen.Id("o").Id(changeSetName)).Params().Block(
			scs...,
		)
	}
	return st
}

func (m *Model) hasBasic() bool {
	for i := range m.Fields {
		if m.Fields[i].IsBasic || m.Fields[i].IsSet {
			return true
		}
	}
	return false
}

func (m *Model) hasMeta() bool {
	for i := range m.Fields {
		if m.Fields[i].isMeta() {
			return true
		}
	}
	return false
}

func (m *Model) hasAudit() bool {
	for i := range m.Fields {
		if m.Fields[i].isAudit() {
			return true
		}
	}
	return false
}

func (m *Model) hasModHook() (ok bool, idf string, dtf string) {
	for _, field := range m.Fields {
		typ := field.getType()
		if strings.HasSuffix(typ, modelDefault) {
			return true, modelDefault, modelDefault
		}
		if strings.HasSuffix(typ, modelDunce) {
			return true, modelDunce, modelDunce
		}
		if strings.HasSuffix(typ, modelSerial) {
			return false, modelSerial, modelSerial
		}

		if strings.Contains(typ, "IDField") { // .IDField, .IDFieldStr
			idf = "IDField"
		} else if strings.HasSuffix(typ, "SerialField") {
			idf = "SerialField"
		} else if strings.HasSuffix(typ, "DateFields") {
			dtf = "DateFields"
		}
	}

	if len(idf) > 0 && len(dtf) > 0 {
		ok = idf == "IDField"
	}
	return
}

func (mod *Model) IsTable() bool {
	if _, idf, _ := mod.hasModHook(); len(idf) > 0 && len(mod.TableTag) > 0 {
		return true
	}
	return false
}

func (m *Model) hookModelCodes() (st *jen.Statement) {
	if hasHooks, idF, dtF := m.hasModHook(); hasHooks {
		st = new(jen.Statement)
		st.Comment("Creating function call to it's inner fields defined hooks").Line()
		st.Func().Params(
			jen.Id("z").Op("*").Id(m.Name),
		).Id("Creating").Params().Error().Block(
			jen.If(jen.Id("z").Dot("IsZeroID").Call()).BlockFunc(func(g *jen.Group) {
				oidcat := CamelCased(m.OIDCat)
				if len(oidcat) > 0 && (idF == modelDefault || idF == "IDField") {
					oidQual, _ := m.doc.getQual("oid")
					g.Id("z").Dot("SetID").Call(
						jen.Qual(oidQual, "NewID").Call(jen.Qual(oidQual, "Ot"+oidcat)),
					)
				} else {
					g.Return(jen.Id("comm").Dot("ErrEmptyID"))
				}
			}).Line(),
			jen.Return(jen.Id("z").Dot(dtF).Dot("Creating").Call()),
		).Line()

	}
	return st
}

func (m *Model) basicCodes() (st *jen.Statement) {
	if !m.hasBasic() {
		return
	}
	st = new(jen.Statement)
	basicName := m.Name + "Basic"
	st.Func().Id("New" + m.Name + "WithBasic").Params(jen.Id("in").Id(basicName)).Op("*").Id(m.Name).BlockFunc(func(g *jen.Group) {
		g.Id("obj").Op(":=&").Id(m.Name).Block(jen.Id(basicName).Op(":").Id("in").Op(","))
		if m.hasMeta() {
			g.Op("_=").Id("obj").Dot("MetaUp").Call(jen.Id("in").Dot("MetaDiff"))
		}
		if m.PostNew {
			g.Id("obj").Dot("PostNew").Call(jen.Op("&").Id("in"))
		}
		g.Return(jen.Id("obj"))
	})
	st.Line()
	return st
}

func (m *Model) specFields() (out Fields) {
	for _, f := range m.Fields {
		if sfn, ext, ok := f.parseQuery(); ok {
			// log.Printf("name: %s, sfn: %q, ext: %q", f.Name, sfn, ext)
			f.siftExt = ext
			if ext == "ints" || ext == "strs" || ext == "oids" {
				ftyp := "string"
				if ext == "oids" {
					ftyp = "oid.OIDsStr"
				}
				argTag := Plural(f.getArgTag())
				f0 := Field{
					Comment:  f.Comment + " (多值逗号分隔)",
					Type:     ftyp,
					Name:     Plural(f.Name),
					Tags:     Tags{"form": argTag, "json": argTag + ",omitempty"},
					siftExt:  ext,
					multable: true,
				}
				// log.Printf("f0: %+v", f0)
				out = append(out, f0)
			} else if ext == "decode" {
				f.qtype = "string"
				f.Comment += " (支持混合解码)"
			} else if ext == "hasVals" {
				f.Comment += " (多值数字相加)"
			}
			if f.Type == "oid.OID" {
				f.Type = "string"
				f.isOid = true
				if sfn == "siftOIDs" {
					f.siftFn = sfn
				} else {
					f.siftFn = "siftOID"
				}
			} else if f.Type == "oid.OIDs" {
				f.Type = "string"
				f.isOid = true
				f.siftFn = "sift"
				f.siftOp = "any"
				f.Comment += " (多值逗号分隔)"
			} else if strings.HasSuffix(f.Type, "DateTime") {
				f.Type = "string"
				f.isDate = true
				f.isIntDt = true
				f.siftFn = "siftDate"
			} else if strings.HasSuffix(f.Type, "Time") {
				f.Type = "string"
				f.isDate = true
				f.siftFn = "siftDate"
			} else {
				f.siftFn = sfn
			}

			out = append(out, f)
		} else if f.isOwner() {
			f0 := Field{
				Comment: "所有者编号 (多值使用逗号分隔)",
				Type:    "string",
				Name:    "OwnerID",
				Tags:    Tags{"form": "owner", "json": "owner,omitempty"},
				siftFn:  "siftOIDs",
				colname: "owner_id",
			}
			out = append(out, f0)
		}
	}
	return
}

func (m *Model) sortableColumns() (cs []string) {
	for _, f := range m.Fields {
		if f.isEmbed() {
			continue
		}

		if cn, ok, _ := f.ColName(); ok && len(cn) > 0 && f.Sortable {
			cs = append(cs, cn)
		}
	}
	return
}

func (m *Model) getSpecCodes() jen.Code {
	var fcs []jen.Code
	fcs = append(fcs, jen.Id("PageSpec"), jen.Id("ModelSpec"))
	if m.hasAudit() {
		fcs = append(fcs, jen.Id("AuditSpec"))
	}
	for _, sifter := range m.Sifters {
		fcs = append(fcs, jen.Id(sifter))
	}
	_, okTS := m.HasTextSearch()
	if okTS {
		fcs = append(fcs, jen.Id("TextSearchSpec"))
	}

	var idx int
	specFields := m.specFields()
	if len(specFields) > 0 {
		// log.Printf("specFields: %+v", specFields)
		fcs = append(fcs, jen.Empty())
		for _, field := range specFields {
			fcs = append(fcs, field.queryCode(idx, m.doc.getModQual(field.getType())))
			idx++
		}

	}

	var withRel string
	var wrTyp string
	_, okAL := m.hasStoreHook(afterList)
	belonNames := m.Fields.relHasOne()
	relations := m.Fields.Relations()
	if len(belonNames) > 0 || len(relations) > 0 || okAL {
		wrTyp = "bool"
		if okAL || len(relations) > 1 {
			wrTyp = "string"
		}
		withRel = "WithRel"
		jtag := "rel"
		field := &Field{
			Name: withRel,
			Type: wrTyp, Tags: Tags{"json": jtag},
			Comment: "include relation column"}
		fcs = append(fcs, jen.Empty(), field.queryCode(idx))
		idx++
	}
	for _, field := range m.SpecExtras {
		fcs = append(fcs, field.queryCode(idx, m.doc.getModQual(field.getType())))
		idx++
	}

	tname := m.Name + "Spec"
	st := jen.Type().Id(tname).Struct(fcs...).Line()
	if len(fcs) > 2 {
		isPG10 := m.doc.IsPG10()
		jfSiftCall := func(name string) jen.Code {
			if isPG10 {
				return jen.Id("q").Op(",").Id("_").Op("=").Id("spec").Dot(name).Dot("Sift").Call(jen.Id("q"))
			}
			return jen.Id("q").Op("=").Id("spec").Dot(name).Dot("Sift").Call(jen.Id("q"))
		}
		args := []jen.Code{jen.Op("*").Id("ormQuery")}
		if isPG10 {
			args = append(args, jen.Error())
		}
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("Sift").Params(jen.Id("q").Op("*").Id("ormQuery")).
			Params(args...)
		st.BlockFunc(func(g *jen.Group) {
			if len(belonNames) > 0 && !okAL {
				log.Printf("%s belongsTo Names %+v", m.Name, belonNames)
				// g.Var().Id("pre").String()
				var jcond jen.Code
				if wrTyp == "bool" {
					jcond = jen.Id("spec").Dot(withRel)
				} else {
					jcond = jen.Len(jen.Id("spec").Dot(withRel)).Op(">0")
				}
				g.If(jcond).BlockFunc(func(g *jen.Group) {
					for _, relName := range belonNames {
						g.Id("q").Dot("Relation").Call(jen.Lit(relName))
					}
					// g.Id("pre").Op("=").Lit("?TableAlias.")
				}).Line()
			}
			g.Add(jfSiftCall("ModelSpec"))

			if m.hasAudit() {
				g.Add(jfSiftCall("AuditSpec"))
			}
			for _, sifter := range m.Sifters {
				g.Add(jfSiftCall(sifter))
			}
			utilsQual, _ := m.doc.getQual("utils")

			for i := 0; i < len(specFields); i++ {
				field := specFields[i]
				fieldM := field
				if field.multable { // ints, strs, oids
					field = specFields[i+1]
					i++
				}
				cn, indb, _ := field.ColName()
				if !indb && len(field.siftFn) == 0 {
					continue
				}
				acn := cn
				if !isPG10 && len(withRel) > 0 {
					// cn = m.tableAlias() + "." + cn
					acn = "?TableAlias." + cn
				}
				jSV := jen.Id("spec").Dot(field.Name)
				params := []jen.Code{jen.Id("q"), jen.Lit(cn), jSV}
				cfn := field.siftFn
				if field.isDate && field.isIntDt {
					params = append(params, jen.True())
				}
				params = append(params, jen.False())
				jq := jen.Id("q").Op(",").Id("_").Op("=").Id(cfn).Call(params...)
				jSiftVals := jen.Id("q").Op(",").Id("_").Op("=").Id("sift").Call(jen.Id("q"), jen.Lit(cn), jen.Lit("IN"), jen.Id("vals"), jen.Lit(false))
				if field.siftExt == "decode" {
					g.If(jen.Len(jSV).Op(">0")).Block(
						jen.Var().Id("v").Add(field.typeCode(m.doc.getModQual(field.getType()))),
						jen.If(jen.Err().Op(":=").Id("v").Dot("Decode").Call(jSV).Op(";").Err().Op("==").Nil()).Block(
							jen.Id("q").Op("=").Id("q").Dot("Where").Call(jen.Lit(acn+" = ?"), jen.Id("v")),
						),
					)
				} else if field.siftOp == "any" {
					g.If(jen.Id("vals").Op(":=").Qual("strings", "Split").Call(jSV, jen.Lit(",")).Op(";").Len(jSV).Op(">0").Op("&&").Len(jen.Id("vals")).Op(">0")).Block(
						// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
						jen.Id("q").Op(",").Id("_").Op("=").Id("sift").Call(jen.Id("q"), jen.Lit(cn), jen.Lit(field.siftOp), jen.Id("vals"), jen.Lit(false)),
					)
				} else if field.siftExt == "hasVals" {
					g.If(jen.Id("vals").Op(":=").Id("spec").Dot(field.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
						// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
						jSiftVals,
					).Else().Block(jq)
				} else if field.siftExt == "ints" {
					g.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseInts").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
						// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
						jSiftVals,
					).Else().Block(jq)
				} else if field.siftExt == "strs" {
					g.If(jen.Id("vals").Op(",").Id("ok").Op(":=").Qual(utilsQual, "ParseStrs").Call(jen.Id("spec").Dot(fieldM.Name)).Op(";").Id("ok")).Block(
						// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
						jSiftVals,
					).Else().Block(jq)
				} else if field.siftExt == "oids" {
					g.If(jen.Id("vals").Op(":=").Id("spec").Dot(fieldM.Name).Dot("Vals").Call().Op(";").Len(jen.Id("vals")).Op(">0")).Block(
						// jen.Id("q").Dot("Where").Call(jen.Lit(cn+" IN(?)"), jen.Id("pgIn").Call(jen.Id("vals"))),
						jSiftVals,
					).Else().Block(jq)
				} else {
					g.Add(jq)
				}

			}
			if okTS {
				g.Add(jfSiftCall("TextSearchSpec"))
			}
			g.Line()

			if isPG10 {
				g.Return(jen.Id("q"), jen.Nil())
			} else {
				g.Return(jen.Id("q"))
			}

		}).Line()
	}

	if cols := m.sortableColumns(); len(cols) > 0 {
		log.Printf("sortable: %+v", cols)
		st.Func().Params(jen.Id("spec").Op("*").Id(tname)).Id("CanSort").Params(jen.Id("k").Id("string")).Bool()
		st.BlockFunc(func(g *jen.Group) {
			g.Switch(jen.Id("k")).BlockFunc(func(g1 *jen.Group) {
				g1.Case(jen.ListFunc(func(g2 *jen.Group) {
					for _, s := range cols {
						g2.Lit(s)
					}
				})).Return(jen.True())
				g1.Default().Return(jen.Id("spec").Dot("ModelSpec").Dot("CanSort").Call(jen.Id("k")))
			})
		}).Line()
	}

	return st
}

func (m *Model) hasStoreHook(k string) (v string, ok bool) {
	if v, ok = m.StoHooks[k]; ok {
		v, ok = m.storeHookName(k, v)
	}
	return
}

func (m *Model) storeHookName(k, v string) (string, bool) {
	return HookMethod(m.Name, k, v)
}

func (m *Model) StoreHooks() (out []storeHook) {
	for k, v := range m.StoHooks {
		if len(v) == 0 {
			continue
		}
		fn, ok := m.storeHookName(k, v)
		if !ok {
			continue
		}

		out = append(out, storeHook{
			FunName: fn,
			ObjName: m.Name,
			k:       k,
			m:       m,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].FunName > out[j].FunName
	})

	return out
}

func (m *Model) shortComment() string {
	s := m.Comment
	if a, _, ok := strings.Cut(s, " "); ok {
		return a
	}
	return s
}

func metaUpCode(a ...bool) jen.Code {
	tags := Tags{"json": "metaUp,omitempty", TagSwaggerIgnore: "true"}
	if len(a) > 0 && a[0] {
		tags["bson"] = "-"
		tags["bun"] = "-"
		tags["pg"] = "-"
	}
	code := jen.Comment("for meta update").Line()
	code.Id("MetaDiff").Op("*").Add(jen.Id("comm.MetaDiff"))
	code.Tag(tags)
	return code
}

func ownerUpCode(idx int) jen.Code {
	tags := Tags{"json": "ownerID,omitempty"}
	tags.extOrder(idx)
	code := jen.Comment("仅用于更新所有者(负责人)").Line()
	code.Id("OwnerID").Op("*").Id("string")
	code.Tag(tags)
	return code
}

func createdUpCode(idx int) jen.Code {
	code := jen.Comment("创建时间").Line()

	tags := Tags{"json": "created,omitempty"}
	tags.extOrder(idx)
	code.Id(createdField).Op("*").Qual("time", "Time")
	code.Tag(tags)
	return code
}

func (m *Model) HasTextSearch() (cols []string, ok bool) {
	var hasTs bool
	for _, field := range m.Fields {
		if strings.HasSuffix(field.Query, "fts") {
			cn, _, _ := field.ColName()
			cols = append(cols, cn)
		}
		if matchs(textSearchField, field.Name, field.Type) {
			hasTs = true
		}
	}
	ok = hasTs

	return
}

func (mod *Model) textSearchCodes(id string) (jen.Code, bool) {
	st := jen.Empty()
	if cols, ok := mod.HasTextSearch(); ok {
		st.If(jen.Id("tscfg").Op(",").Id("ok").Op(":=").Add(swdb).Dot("GetTsCfg").Call().Op(";").Id("ok")).BlockFunc(func(g *jen.Group) {
			g.Id(id).Dot("TsCfgName").Op("=").Id("tscfg")
			if !mod.DbTriggerSave && len(cols) > 0 {
				g.Id(id).Dot("SetTsColumns").Call(jen.ListFunc(func(g1 *jen.Group) {
					for _, s := range cols {
						g1.Lit(s)
					}
				}))
			}
			g.Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
		})
		// if id == "exist" {
		// 	// st.Else().Block(jen.Id(id).Dot("TsCfgName").Op("=").Lit(""))
		// 	st.Line().Id(id).Dot("SetChange").Call(jen.Lit("ts_cfg"))
		// }

		return st, true
	}

	return st, false
}

var (
	swdb  jen.Code = jen.Id("s").Dot("w").Dot("db")
	jactx jen.Code = jen.Id("ctx").Id("context.Context")
)

func (m *Model) getIPath() string {
	if m.doc != nil {
		return m.doc.modipath
	}
	return m.pkg
}

func (m *Model) dbTxFn() string {
	if m.doc.IsPG10() {
		return "RunInTransaction"
	}
	return "RunInTx"
}

func (m *Model) codestoreList() ([]jen.Code, []jen.Code, *jen.Statement) {
	return []jen.Code{jen.Id("spec").Op("*").Id(m.Name + "Spec")},
		[]jen.Code{jen.Id("data").Qual(m.getIPath(), m.GetPlural()),
			jen.Id("total").Int(), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			if cols, ok := m.HasTextSearch(); ok {
				g.Id("spec").Dot("SetTsConfig").Call(jen.Add(swdb).Dot("GetTsCfg").Call())
				if len(cols) > 0 {
					g.Id("spec").Dot("SetTsFallback").Call(jen.ListFunc(func(g1 *jen.Group) {
						for _, s := range cols {
							g1.Lit(s)
						}
					}))
				}
			}

			isPG10 := m.doc.IsPG10()
			if hkBL, okBL := m.hasStoreHook(beforeList); okBL {
				jdataptr := jen.Op("&").Id("data")
				jspec := jen.Id("spec")
				var jcall jen.Code
				if isPG10 {
					jcall = jen.Dot("ModelContext").Call(jen.Id("ctx"), jdataptr)
				} else {
					jcall = jen.Dot("NewSelect").Call().Dot("Model").Call(jdataptr)
				}
				g.Id("q").Op(":=").Add(swdb, jcall) //.Dot("Apply").Call(jen.Id("spec").Dot("Sift"))
				g.If(jen.Err().Op("=").Id("s").Dot(hkBL).Call(jen.Id("ctx"), jspec, jen.Id("q")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("total").Op(",").Err().Op("=").Id("queryPager").Call(jen.Id("ctx"), jspec, jen.Id("q"))
			} else {
				g.Id("total").Op(",").Id("err").Op("=").Add(swdb).Dot("ListModel").Call(
					jen.Id("ctx"), jen.Id("spec"), jen.Op("&").Id("data"),
				)
			}

			if hkAL, okAL := m.hasStoreHook(afterList); okAL {
				g.If(jen.Err().Op("==").Nil().Op("&&").Len(jen.Id("data")).Op(">0")).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("spec"), jen.Id("data")),
				)
			}
			g.Return()
		})
}

func (mod *Model) codestoreGet() ([]jen.Code, []jen.Code, *jen.Statement) {
	utilsQual, _ := mod.doc.getQual("utils")
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op("=").New(jen.Qual(mod.getIPath(), mod.Name))
			jload := jen.Id("err").Op("=").Add(swdb).Dot("GetModel").Call(
				jen.Id("ctx"), jen.Id("obj"), jen.Id("id"))
			if _, cn, isuniq := mod.UniqueOne(); isuniq {
				g.If(jen.Err().Op("=").Id("getModelWithUnique").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Lit(cn), jen.Id("id"),
				).Op(";").Err().Op("!=").Nil()).Block(jload)
			} else {
				g.Add(jload)
			}

			if hkAL, okAL := mod.hasStoreHook(afterLoad); okAL {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id("s").Dot(hkAL).Call(jen.Id("ctx"), jen.Id("obj")),
				)
			} else if rels := mod.Fields.relHasOne(); len(rels) > 0 {
				g.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
				g.For().Op("_,").Id("rn").Op(":=").Range().Id("RelationFromContext").Call(jen.Id("ctx")).BlockFunc(func(g2 *jen.Group) {
					for _, rn := range rels {
						field, _ := mod.Fields.withName(rn)
						g2.If(jen.Id("rn").Op("==").Lit(rn).Op("&&!").Qual(utilsQual, "IsZero").Call(jen.Id("obj."+rn+"ID"))).Block(
							jen.Id("ro").Op(":=").New(field.typeCode(mod.getIPath())),
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
		})
}

func (mod *Model) codestoreCreate() ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Basic"
	return []jen.Code{jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jen.Id("obj").Op("*").Qual(mod.getIPath(), mod.Name), jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			nname := "New" + mod.Name + "WithBasic"
			g.Id("obj").Op("=").Qual(mod.getIPath(), nname).Call(jen.Id("in"))

			targs := []jen.Code{jen.Id("ctx"), swdb, jen.Id("obj")}
			if fn, cn, isuniq := mod.UniqueOne(); isuniq {
				g.If(jen.Id("obj").Dot(fn).Op("==").Lit("")).Block(
					jen.Err().Op("=").Id("ErrEmptyKey"),
					jen.Return())
				targs = append(targs, jen.Lit(cn))
			} else if mod.ForceCreate {
				targs = append(targs, jen.Lit(true))
			}

			if jt, ok := mod.textSearchCodes("obj"); ok {
				g.Add(jt)
			}
			isPG10 := mod.doc.IsPG10()

			hkBC, okBC := mod.hasStoreHook(beforeCreating)
			hkBS, okBS := mod.hasStoreHook(beforeSaving)
			hkAS, okAS := mod.hasStoreHook(afterSaving)
			if okBC || okBS || okAS {
				g.Err().Op("=").Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					targs[1] = jdb
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBC {
							g2.If(jen.Err().Op("=").Id(hkBC).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(jen.Err()),
							)
						} else if okBS {
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("obj")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(jen.Err()),
							)
						}
						if mod.hasMeta() {
							g2.Id("dbOpModelMeta").Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj"))
						}

						g2.Err().Op("=").Id("dbInsert").Call(targs...)
						if okAS {
							g2.If(jen.Err().Op("==")).Nil().Block(
								jen.Err().Op("=").Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("obj")),
							)
						}

						g2.Return(jen.Err())
					}
					if isPG10 {
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					} else {
						g1.Nil()
						g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					}

				})

			} else {
				if mod.hasMeta() {
					g.Id("dbOpModelMeta").Call(jen.Id("ctx"), swdb, jen.Id("obj"))
				}

				g.Err().Op("=").Id("dbInsert").Call(targs...)
			}

			if hk, ok := mod.hasStoreHook(afterCreated); ok {
				g.If(jen.Err().Op("==").Nil()).Block(
					jen.Err().Op("=").Id("s").Dot(hk).Call(jen.Id("ctx"), jen.Id("obj")),
				)
			}

			g.Return()
		})
}

func (mod *Model) codestoreUpdate() ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Set"
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("exist").Op(":=").New(jen.Qual(mod.getIPath(), mod.Name))
			g.If(jen.Err().Op(":=").Id("getModelWithPKID").Call(
				jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("id"),
			).Op(";").Err().Op("!=").Nil()).Block(jen.Return(jen.Err()))

			g.Id("exist").Dot("SetWith").Call(jen.Id("in"))

			if jt, ok := mod.textSearchCodes("exist"); ok {
				g.Add(jt)
			}
			isPG10 := mod.doc.IsPG10()

			jfbd := jen.Empty()
			hkBU, okBU := mod.hasStoreHook(beforeUpdating)
			hkBS, okBS := mod.hasStoreHook(beforeSaving)
			hkAS, okAS := mod.hasStoreHook(afterSaving)
			if okBU || okBS || okAS {
				jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					jdb := jen.Id("tx")
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBU {
							g2.If(jen.Err().Op("=").Id(hkBU).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						} else if okBS {
							g2.If(jen.Err().Op("=").Id(hkBS).Call(jen.Id("ctx"), jdb, jen.Id("exist")).Op(";").Err().Op("!=")).Nil().Block(
								jen.Return(),
							)
						}
						if mod.hasMeta() {
							g2.Id("dbOpModelMeta").Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("exist"))
						}

						jup := jen.Id("dbUpdate").Call(
							jen.Id("ctx"), jdb, jen.Id("exist"),
						)

						if okAS {
							g2.If(jen.Err().Op("=").Add(jup).Op(";").Err().Op("==")).Nil().Block(
								jen.Return().Id(hkAS).Call(jen.Id("ctx"), jdb, jen.Id("exist")),
							)
							g2.Return()
						} else {
							g2.Return(jup)
						}
					}
					if isPG10 {
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					} else {
						g1.Nil()
						g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					}
				})

			} else {
				if mod.hasMeta() {
					g.Id("dbOpModelMeta").Call(jen.Id("ctx"), swdb, jen.Id("exist"))
				}
				jfbd.Id("dbUpdate").Call(
					jen.Id("ctx"), swdb, jen.Id("exist"),
				)
			}

			if hk, ok := mod.hasStoreHook(afterUpdated); ok {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				g.Return(jen.Id("s").Dot(hk).Call(jen.Id("ctx"), jen.Id("exist")))
			} else {
				g.Return(jfbd)
			}

		})
}

func (mod *Model) codestorePut(isSimp bool) ([]jen.Code, []jen.Code, *jen.Statement) {
	tname := mod.Name + "Set"
	jqual := jen.Qual(mod.getIPath(), mod.Name)
	var jret *jen.Statement
	if isSimp {
		jret = jen.Id("nid").String()
	} else {
		jret = jen.Id("isnew").Bool()
	}
	// log.Printf("jret: %s, %+v", mod.Name, jret)
	return []jen.Code{jen.Id("id").String(), jen.Id("in").Qual(mod.getIPath(), tname)},
		[]jen.Code{jret, jen.Err().Error()},
		jen.BlockFunc(func(g *jen.Group) {
			g.Id("obj").Op(":=").New(jqual)
			g.Id("_").Op("=").Id("obj").Dot("SetID").Call(jen.Id("id"))

			if isSimp {
				g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
				g.Err().Op("=").Id("dbStoreSimple").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"),
				)
				g.Id("nid").Op("=").Id("obj").Dot("StringID").Call()
			} else {
				g.Id("obj").Dot("SetWith").Call(jen.Id("in"))
				g.Id("exist").Op(":=").New(jqual)

				cpms := []jen.Code{
					jen.Id("ctx"), swdb, jen.Id("exist"), jen.Id("obj"),
					jen.Func().Params().Index().String().Block(
						jen.Id("exist").Dot("SetWith").Call(jen.Id("in")),
						jen.Return(jen.Id("exist").Dot("GetChanges").Call()),
					),
				}
				if fn, cn, isuniq := mod.UniqueOne(); isuniq {
					g.If(jen.Id("isZero").Call(jen.Id("obj").Dot(fn))).Block(
						jen.Err().Op("=").Qual("fmt", "Errorf").Call(jen.Lit("need "+cn)),
						jen.Return())
					cpms = append(cpms, jen.Lit(cn), jen.Id("obj").Dot(fn))
				}
				g.Id("isnew").Op(",").Err().Op("=").Id("dbStoreWithCall").Call(cpms...)
			}
			g.Return()
		})
}

func (mod *Model) codestoreDelete() ([]jen.Code, []jen.Code, *jen.Statement) {
	jqual := jen.Qual(mod.getIPath(), mod.Name)
	jtabl := jen.Qual(mod.getIPath(), mod.Name+"Table")
	return []jen.Code{jen.Id("id").String()},
		[]jen.Code{jen.Error()},
		jen.BlockFunc(func(g *jen.Group) {
			jfbd := jen.Empty()
			g.Id("obj").Op(":=").New(jqual)
			hkBD, okBD := mod.hasStoreHook(beforeDeleting)
			hkAD, okAD := mod.hasStoreHook(afterDeleting)
			if okBD || okAD {
				g.If(jen.Id("err").Op(":=").Id("getModelWithPKID").Call(
					jen.Id("ctx"), swdb, jen.Id("obj"), jen.Id("id"),
				).Op(";").Id("err").Op("!=").Nil()).Block(jen.Return(jen.Err()))

				jfbd.Add(swdb).Dot(mod.dbTxFn()).CallFunc(func(g1 *jen.Group) {
					g1.Id("ctx")
					jbf := func(g2 *jen.Group) {
						if okBD {
							g2.If(jen.Err().Op("=").Id(hkBD).Call(jen.Id("ctx"), jen.Id("tx"),
								jen.Id("obj")).Op(";").Err().Op("!=").Nil()).Block(jen.Return())
						}

						g2.Err().Op("=").Id("dbDeleteT").Call(jen.Id("ctx"), jen.Id("tx"),
							jen.Add(swdb).Dot("Schema").Call(),
							jen.Add(swdb).Dot("SchemaCrap").Call(),
							jtabl, jen.Id("obj").Dot("ID"))
						if okAD {
							g2.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
							g2.Return(jen.Id(hkAD).Call(jen.Id("ctx"), jen.Id("tx"), jen.Id("obj")))
						} else {
							g2.Return()
						}
					}
					if mod.doc.IsPG10() {
						g1.Func().Params(jen.Id("tx").Op("*").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					} else {
						g1.Nil()
						g1.Func().Params(jactx, jen.Id("tx").Id("pgTx")).Params(jen.Err().Error()).BlockFunc(jbf)
					}

				})
			} else {
				if mod.doc.IsPG10() {
					g.If(jen.Op("!").Id("obj").Dot("SetID").Call(jen.Id("id"))).Block(
						jen.Return().Qual("fmt", "Errorf").Call(jen.Lit("id: '%s' is invalid"), jen.Id("id")),
					)
					jfbd.Id("s").Dot("w").Dot("db").Dot("OpDeleteAny").Call(
						jen.Id("ctx"), jtabl, jen.Id("obj").Dot("ID"),
					)
				} else {
					jfbd.Add(swdb).Dot("DeleteModel").Call(
						jen.Id("ctx"), jen.Id("obj"), jen.Id("id"),
					)
				}

			}

			if hk, ok := mod.hasStoreHook(afterDeleted); ok {
				g.If(jen.Err().Op(":=").Add(jfbd).Op(";").Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				g.Return(jen.Id("s").Dot(hk).Call(jen.Id("ctx"), jen.Id("obj")))
			} else {
				g.Return(jfbd)
			}

		})
}
