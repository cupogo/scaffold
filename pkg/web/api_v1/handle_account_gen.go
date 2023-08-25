// This file is generated - Do Not Edit.

package apiv1

import (
	"github.com/cupogo/scaffold/pkg/models/accounts"
	"github.com/cupogo/scaffold/pkg/services/stores"
	gin "github.com/gin-gonic/gin"
)

func init() {
	regHI(true, "GET", "/accounts", "v1-accounts-get", func(a *api) HandlerFunc {
		return a.getAccounts
	})
	regHI(true, "GET", "/accounts/:id", "v1-accounts-id-get", func(a *api) HandlerFunc {
		return a.getAccount
	})
	regHI(true, "POST", "/accounts", "v1-accounts-post", func(a *api) HandlerFunc {
		return a.postAccount
	})
	regHI(true, "PUT", "/accounts/:id", "v1-accounts-id-put", func(a *api) HandlerFunc {
		return a.putAccount
	})
	regHI(true, "DELETE", "/accounts/:id", "v1-accounts-id-delete", func(a *api) HandlerFunc {
		return a.deleteAccount
	})
}

// @Tags Cupola-accounts
// @ID v1-accounts-get
// @Summary 列出账号
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   query  query   stores.AccountSpec  true   "Object"
// @Success 200 {object} Done{result=ResultData{data=accounts.Accounts}}
// @Failure 400 {object} Failure "请求或参数错误"
// @Failure 401 {object} Failure "未登录"
// @Failure 404 {object} Failure "目标未找到"
// @Failure 503 {object} Failure "服务端错误"
// @Router /api/v1/accounts [get]
func (a *api) getAccounts(c *gin.Context) {
	var spec stores.AccountSpec
	if err := c.Bind(&spec); err != nil {
		fail(c, 400, err)
		return
	}

	ctx := c.Request.Context()
	data, total, err := a.sto.Account().ListAccount(ctx, &spec)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, dtResult(data, total))
}

// @Tags Cupola-accounts
// @ID v1-accounts-id-get
// @Summary 获取账号
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Success 200 {object} Done{result=accounts.Account}
// @Failure 400 {object} Failure "请求或参数错误"
// @Failure 401 {object} Failure "未登录"
// @Failure 404 {object} Failure "目标未找到"
// @Failure 503 {object} Failure "服务端错误"
// @Router /api/v1/accounts/{id} [get]
func (a *api) getAccount(c *gin.Context) {
	id := c.Param("id")
	obj, err := a.sto.Account().GetAccount(c.Request.Context(), id)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, obj)
}

// @Tags Cupola-accounts
// @ID v1-accounts-post
// @Summary 录入账号
// @Accept json,mpfd
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   query  body   accounts.AccountBasic  true   "Object"
// @Success 200 {object} Done{result=ResultID}
// @Failure 400 {object} Failure "请求或参数错误"
// @Failure 401 {object} Failure "未登录"
// @Failure 403 {object} Failure "无权限"
// @Failure 503 {object} Failure "服务端错误"
// @Router /api/v1/accounts [post]
func (a *api) postAccount(c *gin.Context) {
	var in accounts.AccountBasic
	if err := c.Bind(&in); err != nil {
		fail(c, 400, err)
		return
	}

	obj, err := a.sto.Account().CreateAccount(c.Request.Context(), in)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, idResult(obj.ID))
}

// @Tags Cupola-accounts
// @ID v1-accounts-id-put
// @Summary 更新账号
// @Accept json,mpfd
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Param   query  body   accounts.AccountSet  true   "Object"
// @Success 200 {object} Done{result=string}
// @Failure 400 {object} Failure "请求或参数错误"
// @Failure 401 {object} Failure "未登录"
// @Failure 403 {object} Failure "无权限"
// @Failure 503 {object} Failure "服务端错误"
// @Router /api/v1/accounts/{id} [put]
func (a *api) putAccount(c *gin.Context) {
	id := c.Param("id")
	var in accounts.AccountSet
	if err := c.Bind(&in); err != nil {
		fail(c, 400, err)
		return
	}

	err := a.sto.Account().UpdateAccount(c.Request.Context(), id, in)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, "ok")
}

// @Tags Cupola-accounts
// @ID v1-accounts-id-delete
// @Summary 删除账号
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Success 200 {object} Done
// @Failure 400 {object} Failure "请求或参数错误"
// @Failure 401 {object} Failure "未登录"
// @Failure 403 {object} Failure "无权限"
// @Failure 503 {object} Failure "服务端错误"
// @Router /api/v1/accounts/{id} [delete]
func (a *api) deleteAccount(c *gin.Context) {
	id := c.Param("id")
	err := a.sto.Account().DeleteAccount(c.Request.Context(), id)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, "ok")
}
