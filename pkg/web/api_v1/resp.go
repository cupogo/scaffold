package apiv1

import (
	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/web/resp"
)

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
