package resp

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/web/i18n"
)

// Ok ...
func Ok(c *gin.Context, args ...any) {
	Out(c, 200, args...)
}

// Out ...
func Out(c *gin.Context, code int, args ...any) {
	res := &Done{Time: getTime()}
	if len(args) > 0 {
		if v, ok := args[0].(ViewPatcher); ok {
			v.PatchView()
			res.Result = v
		} else {
			res.Result = args[0]
		}

		if len(args) > 1 {
			res.Extra = args[1]
		}
	}
	c.JSON(code, res)
}

// Fail response fail with code, err, field
func Fail(c *gin.Context, code int, args ...any) {
	if len(args) == 0 {
		c.AbortWithStatus(code)
		return
	}
	res := &Failure{Time: getTime()}
	er := GetError(c.Request, code, args[0], args[1:]...)
	logger().Infow("request fail", "code", code, "args", args,
		"er", er, c.Request.Method, c.Request.RequestURI, "ip", c.ClientIP())
	if code >= 500 {
		// TODO: report error
	}

	res.Code = er.Code
	res.Message = er.Message
	res.Field = er.Field
	c.AbortWithStatusJSON(code, res)
}

// Done 操作成功返回的结构
type Done struct {
	Code   int   `json:"status" example:"0"` // 状态值，0=ok
	Time   int64 `json:"t,omitempty"`        // 时间戳
	Result any   `json:"result,omitempty" `  // 主体数据,可选
	Extra  any   `json:"extra,omitempty" `   // 附加数据,可选
} // @name Response

// Failure 出现错误，返回相关的错误码和消息文本
type Failure struct {
	Code    int    `json:"status" example:"1"`             // 状态值
	Time    int64  `json:"t,omitempty"`                    // 时间戳
	Message string `json:"message" example:"错误信息"`         // 错误信息
	Field   string `json:"field,omitempty" example:"错误字段"` // 错误字段,可选,多用于表单校验
} // @name Failure

// Error ...
type Error struct {
	Code    int    `json:"status" example:"1"`             // 错误代码,可选
	Message string `json:"message" example:"错误信息"`         // 错误信息
	Field   string `json:"field,omitempty" example:"错误字段"` // 错误字段,可选,多用于表单校验
}

// CoderMessager get message and code with http request
type CoderMessager interface {
	Code() int
	GetMessage(r *http.Request) string
}

// Messager get message with http request
type Messager interface {
	GetMessage(r *http.Request) string
}

// FieldError a error with field, for validator
type FieldError interface {
	error
	Field() string
}

// FieldMessager a error with field
type FieldMessager interface {
	Field() string
	GetMessage(r *http.Request) string
}

// GetError ...
func GetError(r *http.Request, code int, err any, args ...any) Error {
	var field string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			field = v
		}
	}
	switch e := err.(type) {
	case Error:
		return e
	case *Error:
		return *e
	case *json.UnmarshalTypeError:
		return Error{
			Code:    code,
			Message: i18n.GetDecodeValueError(r, e.Value, "json"),
			Field:   e.Field,
		}
	case CoderMessager:
		return Error{Code: e.Code(), Message: e.GetMessage(r), Field: field}
	case FieldMessager:
		return Error{Code: code, Message: e.GetMessage(r), Field: e.Field()}
	case Messager:
		return Error{Code: code, Message: e.GetMessage(r), Field: field}
	case []FieldError:
		if len(e) > 0 {
			return Error{Code: code, Message: i18n.Field(e[0].Field()).GetMessage(r), Field: e[0].Field()}
		}
	case FieldError:
		return Error{Code: code, Message: i18n.Field(e.Field()).GetMessage(r), Field: e.Field()}
	case interface{ GetMessage() string }:
		return Error{Code: code, Message: e.GetMessage(), Field: field}
	case string:
		return Error{Code: code, Message: e, Field: field}
	case error:
		return Error{Code: code, Message: e.Error(), Field: field}
	}
	if code >= 100 && code < 600 {
		return Error{Code: code, Message: http.StatusText(code), Field: field}
	}
	return Error{Code: code, Message: "unkown error", Field: field}
}
