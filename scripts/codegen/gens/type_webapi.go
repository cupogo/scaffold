// go:build codegen
package gens

import (
	"fmt"
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

var preFails = map[int]string{
	400: `400 {object} Failure "请求或参数错误"`,
	401: `401 {object} Failure "未登录"`,
	403: `403 {object} Failure "无权限"`,
	404: `404 {object} Failure "目标未找到"`,
	503: `503 {object} Failure "服务端错误"`,
}

var msmethods = map[string]string{
	"List":   "GET",
	"Get":    "GET",
	"Create": "POST",
	"Update": "PUT",
	"Put":    "PUT",
	"Delete": "DELETE",
}

var mslabels = map[string]string{
	"List":   "列出",
	"Get":    "获取",
	"Create": "录入",
	"Update": "更新",
	"Put":    "录入/更新",
	"Delete": "删除",
}

const paramAuth = `token    header   string  true "登录票据凭证"`
const swagTags = "默认 文档生成"

var replPkg = strings.NewReplacer("_", "", "-", "")
var replPath = strings.NewReplacer("{", ":", "}", "")
var replRoute = strings.NewReplacer("[", "", "]", "", "{", "", "}", "", "/", "-", " ", "-")

func GenRouteID(s string) string {
	if n := strings.Index(s, "/api/"); n >= 0 {
		s = s[0:n] + s[n+4:]
	}
	s = strings.TrimPrefix(s, "/")
	s = replRoute.Replace(s)

	return strings.TrimSpace(s)
}

type UriSpot struct {
	Model  string `yaml:"model"`
	Prefix string `yaml:"prefix,omitempty"`
	URI    string `yaml:"uri,omitempty"`
	Ignore string `yaml:"ignore,omitempty"`
	Batch  string `yaml:"batch,omitempty"`

	HandReg  bool `yaml:"handReg,omitempty"`
	NeedAuth bool `yaml:"needAuth,omitempty"`
	NeedPerm bool `yaml:"needPerm,omitempty"`
	NoPost   bool `yaml:"noPost,omitempty"`
	Auth     bool `yaml:"auth,omitempty"` // old
	Perm     bool `yaml:"perm,omitempty"` // old
}

type WebAPI struct {
	Pkg       string    `yaml:"pkg"`
	Handles   []Handle  `yaml:"handles,omitempty"`
	URIs      []UriSpot `yaml:"uris,omitempty"`
	HandReg   bool      `yaml:"handReg,omitempty"`
	NeedAuth  bool      `yaml:"needAuth,omitempty"`
	NeedPerm  bool      `yaml:"needPerm,omitempty"`
	TagLabel  string    `yaml:"tagLabel,omitempty"`
	UriPrefix string    `yaml:"uriPrefix,omitempty"`

	doc *Document
}

func (wa *WebAPI) genHandle(us UriSpot, mth Method, stoName string) (hdl Handle, match bool) {
	if us.Model != mth.model {
		return
	}
	for _, c := range us.Ignore {
		if a, ok := hods[c]; ok && strings.HasPrefix(mth.Name, a) {
			// log.Printf("ignore: method: %+v, %v", mth.Name, a)
			return
		}
	}
	var mod *Model
	mod, match = wa.doc.modelWithName(us.Model)
	if !match {
		return
	}
	plural := mod.GetPlural()
	if len(plural) == 0 {
		log.Printf("WARN: empty name of %s[%s]", us.Model, mod.Name)
	}
	uri := us.URI
	if len(uri) == 0 {
		prefix := us.Prefix
		if len(prefix) == 0 {
			prefix = wa.UriPrefix
		}
		uri = prefix + "/" + strings.ToLower(plural)
	}

	method := msmethods[mth.action]
	fct := strings.ToLower(method)
	var cat string
	if stoName != mod.Name {
		cat = stoName
	}
	name := fct + cat + mod.Name
	switch mth.action {
	case "Get", "Update", "Put", "Delete":
		uri = uri + "/{id}"
	case "List":
		name = fct + cat + plural
	}
	// log.Printf("uri: %s [%s]", uri, method)

	hdl = Handle{
		UriSpot: us,
		Name:    name,
		Method:  mth.Name,
		Store:   stoName,
		Route:   fmt.Sprintf("%s [%s]", uri, strings.ToLower(method)),
		Summary: mslabels[mth.action] + mod.shortComment(),
		wa:      wa,
	}
	hdl.NeedPerm = mth.action == "Create" || mth.action == "Update" ||
		mth.action == "Put" || mth.action == "Delete" || wa.NeedPerm || us.NeedPerm || us.Perm
	hdl.NeedAuth = hdl.NeedPerm || wa.NeedAuth || us.NeedPerm || us.NeedAuth || us.Perm || us.Auth
	hdl.NoPost = us.NoPost
	if len(wa.TagLabel) > 0 {
		hdl.Tags = wa.TagLabel
	}

	return
}

func (wa *WebAPI) GetPkgName() string {
	return replPkg.Replace(wa.Pkg)
}

func (wa *WebAPI) prepareHandles() {
	if wa.doc == nil {
		log.Printf("doc is nil")
		return
	}
	for i := range wa.Handles {
		if wa.Handles[i].wa == nil {
			wa.Handles[i].wa = wa
		}
	}
	for _, u := range wa.URIs {
		for _, sto := range wa.doc.Stores {
			iname := sto.ShortIName()
			for _, mth := range sto.Methods {
				if hdl, ok := wa.genHandle(u, mth, iname); ok {
					wa.Handles = append(wa.Handles, hdl)
				}
			}
		}
	}
	log.Printf("inited webapi handles: %d", len(wa.Handles))
}

type Handle struct {
	UriSpot `yaml:",inline"`

	ID       string   `yaml:"id,omitempty"`
	Tags     string   `yaml:"tags,omitempty"`
	Store    string   `yaml:"store,omitempty"`
	Method   string   `yaml:"method,omitempty"`
	Summary  string   `yaml:"summary,omitempty"`
	Accept   string   `yaml:"accept,omitempty"`
	Produce  string   `yaml:"produce,omitempty"`
	Name     string   `yaml:"name,omitempty"`
	Route    string   `yaml:"route,omitempty"`
	Params   []string `yaml:"params,omitempty"`
	Success  string   `yaml:"success,omitempty" `
	Failures []int    `yaml:"failures,flow,omitempty"`

	act  string // action
	mona string // model name

	wa *WebAPI
}

func (h *Handle) cuted() (ok bool) {
	h.act, h.mona, ok = cutMethod(h.Method)
	return
}

func (h *Handle) GetAccept() string {
	if len(h.Accept) > 0 {
		return h.Accept
	}
	if _, b, ok := strings.Cut(h.Route, " "); ok {
		b = strings.Trim(b, "[]")
		if b == "post" || b == "put" {
			return "json,mpfd"
		}
	}
	return "json"
}

func (h *Handle) GetProduce() string {
	if len(h.Produce) > 0 {
		return h.Produce
	}
	return "json"
}

func (h *Handle) GenID() string {
	return GenRouteID(h.Route)
}

func (h *Handle) GetPermID() string {
	if len(h.ID) > 0 {
		return h.ID
	}
	if h.NeedPerm {
		return h.GenID()
	}
	return ""
}

func (h *Handle) GenPathMethod() (string, string) {
	s := h.Route
	if h.wa != nil && len(h.wa.UriPrefix) > 0 {
		s = strings.TrimPrefix(s, h.wa.UriPrefix)
	}
	if n := strings.Index(s, "/api/"); n >= 0 {
		if len(s) > n+7 {
			if s[n+6] == '1' && s[n+7] == '/' { // '/api/v1/x [xxx]'
				n += 7
			} else {
				n += 4
			}
		}
		s = s[n:]
	}
	if a, b, ok := strings.Cut(s, " "); ok {
		return replPath.Replace(a), strings.ToUpper(strings.Trim(b, "[]"))
	}
	panic("invalid route: " + s)
}

func (h *Handle) GetTags() string {
	if len(h.Tags) > 0 {
		return h.Tags
	}
	return swagTags
}

func (h *Handle) GetFails(act string) []int {
	if h.Failures == nil {
		return getDftFails(act)
	}
	return h.Failures
}

func (us UriSpot) IsBatchCreate() bool {
	return strings.ContainsRune(us.Batch, 'C')
}

func (us UriSpot) IsBatchUpdate() bool {
	return strings.ContainsRune(us.Batch, 'U')
}

func (h *Handle) CommentCodes(doc *Document) jen.Code {
	if len(h.Summary) == 0 {
		log.Printf("WARN: empty handle summary of %s", h.Name)
		return nil
	}
	if len(h.Route) == 0 {
		log.Printf("WARN: empty handle route of %s", h.Name)
		return nil
	}
	st := jen.Empty()
	st.Comment("@Tags " + h.GetTags()).Line()

	if hid := h.GetPermID(); len(hid) > 0 {
		st.Comment("@ID " + hid).Line()
	}
	st.Comment("@Summary " + h.Summary).Line()
	st.Comment("@Accept " + h.GetAccept()).Line()
	st.Comment("@Produce " + h.GetProduce()).Line()
	if h.NeedAuth || h.NeedPerm {
		st.Comment("@Param " + paramAuth).Line()
	}
	var paramed bool
	// log.Printf("params %+v", h.Params)
	for _, param := range h.Params {
		if len(param) > 0 {
			paramed = true
			st.Comment("@Param  " + param).Line()
		}
	}
	if !paramed {
		if mth, ok := doc.getMethod(h.Method); ok {
			for _, arg := range mth.Args {
				if arg.Name == "ctx" {
					continue
				}
				if arg.Name == "id" {
					st.Comment("@Param   id    path   string  true   \"编号\"").Line()
				} else if arg.Type == "string" && strings.Contains(h.Route, "{"+arg.Name+"}") {
					st.Comment("@Param   " + arg.Name + "  path  " + arg.Type + "  true  \"\"").Line()
				} else if strings.Contains(arg.Type, ".") {
					ppos := "formData"
					switch h.act {
					case "List":
						ppos = "query"
					case "Create", "Update", "Put":
						ppos = "body"
					}
					st.Comment("@Param   query  " + ppos + "   " + arg.Type + "  true   \"Object\"").Line()
				} else {
					log.Printf("unknown arg: %s(%s)", arg.Name, arg.Type)
				}
			}
		}
	}
	var success bool
	if len(h.Success) > 0 {
		success = true
		st.Comment("@Success " + h.Success).Line()
	} else if mth, ok := doc.getMethod(h.Method); ok {
		if len(mth.Rets) > 0 && mth.Rets[0].Type != "error" {
			success = true
			if h.act == "List" {
				st.Comment("@Success 200 {object} Done{result=ResultData{data=" + mth.Rets[0].Type + "}}").Line()
			} else if h.act == "Create" {
				st.Comment("@Success 200 {object} Done{result=ResultID}").Line()
			} else {
				st.Comment("@Success 200 {object} Done{result=" + mth.Rets[0].Type + "}").Line()
			}
		}
	}
	if !success {
		if h.act == "Put" || h.act == "Update" {
			st.Comment("@Success 200 {object} Done{result=string}").Line()
		} else {
			st.Comment("@Success 200 {object} Done").Line()
		}
	}
	for _, fi := range h.GetFails(h.act) {
		if s, ok := preFails[fi]; ok {
			st.Comment("@Failure " + s).Line()
		} else {
			log.Printf("invalid failure code: %d", fi)
		}
	}
	st.Comment("@Router " + h.Route).Line()

	return st
}

func (h *Handle) jcall() jen.Code {
	return jen.Id("a").Dot("sto").Dot(h.Store).Call().Dot(h.Method)
}

func (h *Handle) Codes(doc *Document) jen.Code {
	mth, ok := doc.getMethod(h.Method)
	if !ok {
		log.Printf("unknown method: %s", h.Method)
		return nil
	}
	if !h.cuted() {
		log.Printf("cut method %s fail", h.Method)
		return nil
	}
	mod, modok := doc.modelWithName(h.mona)
	if !modok {
		panic("invalid model: " + h.mona)
	}

	st := jen.Add(h.CommentCodes(doc))
	st.Func().Params(jen.Id("a").Op("*").Id("api")).Id(h.Name).Params(jen.Id("c").Op("*").Qual(ginQual, "Context"))
	st.BlockFunc(func(g *jen.Group) {

		if strings.Contains(h.Route, "{id}") { // Get, Put, Delete
			g.Id("id").Op(":=").Id("c").Dot("Param").Call(jen.Lit("id"))
			if h.act == "Get" || h.act == "Load" {
				rels := mod.Fields.relHasOne()
				h.codeLoad(g, rels, doc.qual(mth.Rets[0].Type))
				return
			}
			if (h.act == "Put" || h.act == "Update") && len(mth.Args) > 2 {
				h.codeUpdate(g, doc.qual(mth.Args[2].Type), mth.Simple)
				return
			}
			if h.act == "Delete" {
				g.Add(h.codeDelete())
				return
			}
			log.Printf("invalid act: %s", h.act)
			return
		}
		if h.act == "List" && len(mth.Args) > 1 {
			h.codeList(g, doc.qual(mth.Args[1].Type), mod)
			return
		}
		if h.act == "Create" && len(mth.Args) > 1 {
			h.codeCreate(g, doc.qual(mth.Args[1].Type))
			return
		}
		log.Printf("invalid act: %s", h.act)

	})

	// log.Printf("generate handle %s => %s done", h.Name, mth.Name)

	return st
}

func (h *Handle) codeLoad(g *jen.Group, rels Fields, jarg jen.Code) {
	op := ":="
	needDef := strings.ContainsAny(h.Ignore, "CU")
	if needDef {
		op = "="
		g.Var().Id("obj").Op("*").Add(jarg)
		g.Var().Err().Error()
	}
	if len(rels) > 0 {
		g.Id("ctx").Op(":=").Add(jrctx)
		g.If(
			jen.Id("rels").Op(",").Id("ok").Op(":=").Id("c").Dot("GetQueryArray").Call(jen.Lit("rel")).
				Op(";").Id("ok").Op("&&").Len(jen.Id("rels")).Op(">").Lit(0)).
			Block(
				jen.Id("ctx").Op("=").Id("stores").Dot("ContextWithRelation").Call(
					jen.Id("ctx"), jen.Id("rels").Op("..."),
				),
			)

		g.Id("obj").Op(",").Err().Op(op).Add(h.jcall()).Call(jen.Id("ctx"), jen.Id("id"))

	} else {
		g.Id("obj").Op(",").Err().Op(op).Add(h.jcall()).Call(
			jrctx, jen.Id("id"),
		)
	}

	g.If(jen.Err().Op("!=").Nil()).Block(
		jfails(503)...,
	).Line()
	g.Id("success").Call(jen.Id("c"), jen.Id("obj"))
}

func (h *Handle) codeUpdate(g *jen.Group, in jen.Code, simple bool) {
	g.Var().Id("in").Add(in)
	g.Add(jbind("in"))
	var retName string
	if h.act == "Put" {
		if simple {
			retName = "nid"
		} else {
			retName = "obj"
		}
		g.Id(retName).Op(",").Err().Op(":=").Add(h.jcall()).Call(
			jrctx, jen.Id("id"), jen.Id("in"),
		)
	} else {
		g.Err().Op(":=").Add(h.jcall()).Call(
			jrctx, jen.Id("id"), jen.Id("in"),
		)
	}
	g.If(jen.Err().Op("!=").Nil()).Block(
		jfails(503)...,
	).Line()

	if h.act == "Put" {
		g.Id("success").Call(jen.Id("c"), jen.Id(retName))
	} else {
		g.Id("success").Call(jen.Id("c"), jen.Lit("ok"))
	}
}

func (h *Handle) codeDelete() jen.Code {
	return jen.Err().Op(":=").Add(h.jcall()).Call(
		jrctx, jen.Id("id"),
	).Line().
		If(jen.Err().Op("!=").Nil()).Block(
		jfails(503)...,
	).Line().Line().
		Id("success").Call(jen.Id("c"), jen.Lit("ok"))
}

func (h *Handle) codeList(g *jen.Group, spec jen.Code, mod *Model) {
	g.Var().Id("spec").Add(spec)
	g.Add(jbind("spec"))
	g.Id("ctx").Op(":=").Add(jrctx)
	if len(mod.SpecUp) > 0 {
		g.Id("spec").Dot(mod.SpecUp).Call(jen.Id("ctx"), jen.Lit(mod.Name))
	}
	g.Id("data").Op(",").Id("total").Op(",").Err().Op(":=").Add(h.jcall()).Call(
		jen.Id("ctx"), jen.Op("&").Id("spec"),
	)
	g.If(jen.Err().Op("!=").Nil()).Block(
		jfails(503)...,
	).Line()
	g.Id("success").Call(jen.Id("c"), jen.Id("dtResult").Call(jen.Id("data"), jen.Id("total")))
}

func (h *Handle) jstomb() jen.Code {
	return jen.Id("obj").Op(",").Err().Op(":=").Add(h.jcall()).Call(
		jrctx, jen.Id("in"),
	).Line().
		If(jen.Err().Op("!=").Nil()).Block(
		jfails(503)...,
	).Line().Line().
		Id("success").Call(jen.Id("c"), jen.Id("idResult").Call(jen.Id("obj").Dot("ID")))
}

func (h *Handle) codeCreate(g *jen.Group, jarg jen.Code) {
	if h.IsBatchCreate() {
		g.Id("bd").Op(":=").Qual("github.com/gin-gonic/gin/binding", "Default").Call(jen.Id("c.Request.Method"), jen.Id("c").Dot("ContentType").Call())
		g.Id("bb,ok:=bd.").Call(jen.Id("binding.BindingBody"))
		g.If(jen.Op("!ok")).Block(
			jfails(400, jen.Lit("bad request"))...)
		g.Var().Id("ain").Index().Add(jarg)
		g.Add(jbindWith("ain", true,
			jen.Var().Id("in").Add(jarg),
			jen.Add(jbindWith("in", true, jfails(400)...)),
			h.jstomb(),
			jen.Return(),
		))
		g.Var().Id("ret").Index().Any()
		g.For(jen.Id("_,in").Op(":=").Range().Id("ain")).Block(jen.Id("obj").Op(",").Err().Op(":=").Add(h.jcall()).Call(
			jrctx, jen.Id("in"),
		), jen.If(jen.Err().Op("!=").Nil()).Block(jen.Id("ret").Op("=").Append(jen.Id("ret"), jen.Id("getError").Call(jen.Id("c"), jen.Lit(0), jen.Err()))).
			Else().Block(jen.Id("ret").Op("=").Append(jen.Id("ret"), jen.Id("idResult").Call(jen.Id("obj").Dot("ID")))),
		)
		g.Id("success").Call(jen.Id("c"), jen.Id("dtResult").Call(jen.Id("ret"), jen.Len(jen.Id("ret"))))
	} else {
		g.Var().Id("in").Add(jarg)
		g.Add(jbind("in"))
		g.Add(h.jstomb())
	}
}

func (wa *WebAPI) initRegCodes() jen.Code {
	st := jen.Empty()
	if wa.HandReg {
		return st
	}
	st.Func().Id("init").Params().BlockFunc(func(g *jen.Group) {
		for _, h := range wa.Handles {
			if h.HandReg {
				continue
			}
			uri, method := h.GenPathMethod()
			g.Id("regHI").Call(
				jen.Lit(h.NeedAuth), jen.Lit(method), jen.Lit(uri), jen.Lit(h.GetPermID()),
				jen.Func().Params(jen.Id("a").Op("*").Id("api")).Id("gin.HandlerFunc").Block(
					jen.Return(jen.Id("a."+h.Name)),
				),
			)

			if !h.NoPost && strings.HasPrefix(h.Method, "Put") && strings.HasSuffix(uri, "/:id") {
				g.Id("regHI").Call(
					jen.Lit(h.NeedAuth), jen.Lit("POST"), jen.Lit(uri[0:len(uri)-4]), jen.Lit(h.GetPermID()),
					jen.Func().Params(jen.Id("a").Op("*").Id("api")).Id("gin.HandlerFunc").Block(
						jen.Return(jen.Id("a."+h.Name)),
					),
				)
			}
		}
	}).Line()

	return st
}

func (wa *WebAPI) Codes(doc *Document) jen.Code {
	st := jen.Empty()
	st.Add(wa.initRegCodes())

	for _, h := range wa.Handles {
		if len(h.Tags) == 0 {
			h.Tags = wa.TagLabel
		}
		st.Add(h.Codes(doc)).Line()
	}

	return st
}

func getDftFails(act string) []int {
	if act == "List" || act == "Get" {
		return []int{400, 401, 404, 503}
	}
	return []int{400, 401, 403, 503}
}

func jfails(sc int, ae ...jen.Code) []jen.Code {
	if len(ae) == 0 {
		ae = append(ae, jen.Err())
	}
	return append([]jen.Code{}, jen.Id("fail").Call(jen.Id("c"), jen.Lit(sc), ae[0]), jen.Return())
}

func jbind(id string) jen.Code {
	return jbindWith(id, false, jfails(400)...)
}

func jbindWith(id string, useBody bool, blocks ...jen.Code) jen.Code {
	st := jen.Empty()
	if len(blocks) == 0 || len(id) == 0 {
		return st
	}
	bind := "Bind"
	args := []jen.Code{jen.Op("&").Id(id)}
	if useBody {
		bind = "ShouldBindBodyWith"
		args = append(args, jen.Id("bb"))
	}
	st.If(jen.Err().Op(":=").Id("c").Dot(bind).Call(args...)).Op(";").Err().Op("!=").Nil().Block(
		blocks...,
	).Line()
	return st
}
