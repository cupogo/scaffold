package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/services/stores"
	"github.com/cupogo/scaffold/pkg/web/resp"
	"github.com/cupogo/scaffold/pkg/web/routes"
)

type Done = resp.Done
type Failure = resp.Failure
type ResultData = resp.ResultData
type ResultID = resp.ResultID
type ResultOk = resp.ResultOk

type HandlerFunc = gin.HandlerFunc

var handles = []handleIn{}

type haFunc func(a *api) HandlerFunc

type handleIn struct {
	auth   bool
	method string
	path   string
	rid    string
	hafn   haFunc
}

func regHI(auth bool, method string, path string, rid string, hafn haFunc) {
	handles = append(handles, handleIn{auth, method, path, rid, hafn})
}

// nolint
func route(r gin.IRoutes, method, path string, hs ...HandlerFunc) {
	switch method {
	case http.MethodPost:
		r.Handle(http.MethodPost, path, hs...)
	case http.MethodPut:
		r.Handle(http.MethodPut, path, hs...)
	case http.MethodDelete:
		r.Handle(http.MethodDelete, path, hs...)
	default:
		r.Handle(http.MethodGet, path, hs...)
	}
}

// nolint
type api struct {
	sto stores.Storage
}

func init() {
	routes.Register("apiv1", routes.StrapFunc(strap))
}

func strap(router gin.IRouter) {
	a := newapi(stores.Sgt())
	a.Strap(router)
}

func newapi(sto stores.Storage) *api {
	return &api{sto: sto}
}

func (a *api) Strap(r gin.IRouter) {

	vr := r.Group("/api/v1")
	vr.GET("/ping", ping)

	privater := vr.Group("", a.authSignedIn())

	for _, hi := range handles {
		if hi.auth {
			if len(hi.rid) > 0 {
				route(privater, hi.method, hi.path, a.authPerm(hi.rid), hi.hafn(a))
			} else {
				route(privater, hi.method, hi.path, hi.hafn(a))
			}
		} else {
			route(vr, hi.method, hi.path, hi.hafn(a))
		}
	}
}

// authSignedIn 验证登录中间件
func (a *api) authSignedIn() gin.HandlerFunc {
	// TODO:
	return func(c *gin.Context) {}
}

func (a *api) authPerm(permID string) gin.HandlerFunc {
	// TODO:
	return func(c *gin.Context) {}
}

// @Summary API health check
// @Description API health check
// @Produce plain
// @Success 200 {string} pong
// @Router /api/v1/ping [get]
func ping(c *gin.Context) {
	c.String(200, "pong")
}

// nolint
func success(c *gin.Context, result interface{}) {
	resp.Ok(c, result)
}

// nolint
func fail(c *gin.Context, code int, args ...interface{}) {
	resp.Fail(c, code, args...)
}

// nolint
func dtResult(data any, total int) *resp.ResultData {
	return &resp.ResultData{
		Data:  data,
		Total: total,
	}
}

// nolint
func idResult(id any) *resp.ResultID {
	return &resp.ResultID{ID: id}
}

// nolint
func getError(c *gin.Context, code int, err error, args ...any) resp.Error {
	return resp.GetError(c.Request, code, err, args...)
}
