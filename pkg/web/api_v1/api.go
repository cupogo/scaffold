package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/services/stores"
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
