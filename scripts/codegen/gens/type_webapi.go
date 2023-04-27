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

func (wa *WebAPI) getPkgName() string {
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
	BindObj  string   `yaml:"bindobj,omitempty"`
	Summary  string   `yaml:"summary,omitempty"`
	Accept   string   `yaml:"accept,omitempty"`
	Produce  string   `yaml:"produce,omitempty"`
	Name     string   `yaml:"name,omitempty"`
	Route    string   `yaml:"route,omitempty"`
	Params   []string `yaml:"params,omitempty"`
	Success  string   `yaml:"success,omitempty" `
	Failures []int    `yaml:"failures,flow,omitempty"`

	wa *WebAPI
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
	act, _, _ := cutMethod(h.Method)
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
					switch act {
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
			if act == "List" {
				st.Comment("@Success 200 {object} Done{result=ResultData{data=" + mth.Rets[0].Type + "}}").Line()
			} else if act == "Create" {
				st.Comment("@Success 200 {object} Done{result=ResultID}").Line()
			} else {
				st.Comment("@Success 200 {object} Done{result=" + mth.Rets[0].Type + "}").Line()
			}
		}
	}
	if !success {
		if act == "Put" || act == "Update" {
			st.Comment("@Success 200 {object} Done{result=string}").Line()
		} else {
			st.Comment("@Success 200 {object} Done").Line()
		}
	}
	for _, fi := range h.GetFails(act) {
		if s, ok := preFails[fi]; ok {
			st.Comment("@Failure " + s).Line()
		} else {
			log.Printf("invalid failure code: %d", fi)
		}
	}
	st.Comment("@Router " + h.Route).Line()

	return st
}

func (h *Handle) Codes(doc *Document) jen.Code {
	mth, ok := doc.getMethod(h.Method)
	if !ok {
		log.Printf("unknown method: %s", h.Method)
		return nil
	}
	act, mname, ok := cutMethod(h.Method)
	if !ok {
		log.Printf("cut method %s fail", h.Method)
		return nil
	}
	mod, modok := doc.modelWithName(mname)
	if !modok {
		panic("invalid model: " + mname)
	}

	jctx := jen.Id("c").Dot("Request").Dot("Context").Call()
	jmcc := jen.Op(":=").Id("a").Dot("sto").Dot(h.Store).Call().Dot(h.Method)
	jfail := func(st int) []jen.Code {
		return append([]jen.Code{}, jen.Id("fail").Call(jen.Id("c"), jen.Lit(st), jen.Err()), jen.Return())
	}
	jbind := func(id string) jen.Code {
		return jen.If(jen.Err().Op(":=").Id("c").Dot("Bind").Call(jen.Op("&").Id(id))).Op(";").Err().Op("!=").Nil().Block(
			jfail(400)...,
		).Line()
	}
	st := jen.Add(h.CommentCodes(doc))
	st.Func().Params(jen.Id("a").Op("*").Id("api")).Id(h.Name).Params(jen.Id("c").Op("*").Qual(ginQual, "Context"))
	st.BlockFunc(func(g *jen.Group) {

		if strings.Contains(h.Route, "{id}") { // Get, Put, Delete
			g.Id("id").Op(":=").Id("c").Dot("Param").Call(jen.Lit("id"))
			if act == "Get" || act == "Load" {
				if rels := mod.Fields.relHasOne(); len(rels) > 0 {
					g.Id("ctx").Op(":=").Add(jctx)
					g.If(
						jen.Id("rels").Op(",").Id("ok").Op(":=").Id("c").Dot("GetQueryArray").Call(jen.Lit("rel")).
							Op(";").Id("ok").Op("&&").Len(jen.Id("rels")).Op(">").Lit(0)).
						Block(
							jen.Id("ctx").Op("=").Id("stores").Dot("ContextWithRelation").Call(
								jen.Id("ctx"), jen.Id("rels").Op("..."),
							),
						)

					g.Id("obj").Op(",").Err().Add(jmcc).Call(jen.Id("ctx"), jen.Id("id"))

				} else {
					g.Id("obj").Op(",").Err().Add(jmcc).Call(
						jctx, jen.Id("id"),
					)
				}

				g.If(jen.Err().Op("!=").Nil()).Block(
					jfail(503)...,
				).Line()
				g.Id("success").Call(jen.Id("c"), jen.Id("obj"))
			} else if (act == "Put" || act == "Update") && len(mth.Args) > 2 {
				g.Var().Id("in").Add(doc.qual(mth.Args[2].Type))
				g.Add(jbind("in"))
				var retName string
				if act == "Put" {
					if mth.Simple {
						retName = "nid"
					} else {
						retName = "obj"
					}
					g.Id(retName).Op(",").Err().Add(jmcc).Call(
						jctx, jen.Id("id"), jen.Id("in"),
					)
				} else {
					g.Err().Add(jmcc).Call(
						jctx, jen.Id("id"), jen.Id("in"),
					)
				}
				g.If(jen.Err().Op("!=").Nil()).Block(
					jfail(503)...,
				).Line()

				if act == "Put" {
					g.Id("success").Call(jen.Id("c"), jen.Id(retName))
				} else {
					g.Id("success").Call(jen.Id("c"), jen.Lit("ok"))
				}
			} else if act == "Delete" {
				g.Err().Add(jmcc).Call(
					jctx, jen.Id("id"),
				)
				g.If(jen.Err().Op("!=").Nil()).Block(
					jfail(503)...,
				).Line()
				g.Id("success").Call(jen.Id("c"), jen.Lit("ok"))
			}
		} else if act == "List" && len(mth.Args) > 1 {
			g.Var().Id("spec").Add(doc.qual(mth.Args[1].Type))
			g.Add(jbind("spec"))
			g.Id("ctx").Op(":=").Add(jctx)
			if len(mod.SpecUp) > 0 {
				g.Id("spec").Dot(mod.SpecUp).Call(jen.Id("ctx"), jen.Lit(mname))
			}
			g.Id("data").Op(",").Id("total").Op(",").Err().Add(jmcc).Call(
				jen.Id("ctx"), jen.Op("&").Id("spec"),
			)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jfail(503)...,
			).Line()
			g.Id("success").Call(jen.Id("c"), jen.Id("dtResult").Call(jen.Id("data"), jen.Id("total")))
		} else if act == "Create" && len(mth.Args) > 1 {
			g.Var().Id("in").Add(doc.qual(mth.Args[1].Type))
			g.Add(jbind("in"))
			g.Id("obj").Op(",").Err().Add(jmcc).Call(
				jctx, jen.Id("in"),
			)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jfail(503)...,
			).Line()
			g.Id("success").Call(jen.Id("c"), jen.Id("idResult").Call(jen.Id("obj").Dot("ID")))

		}

	})

	// log.Printf("generate handle %s => %s done", h.Name, mth.Name)

	return st
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
