package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/services/stores"
	"github.com/cupogo/scaffold/pkg/web/resp"
)

var handles = []handleIn{}

type haFunc func(a *api) gin.HandlerFunc

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
func route(r gin.IRoutes, method, path string, hs ...gin.HandlerFunc) {
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

// TODO: strap router

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
