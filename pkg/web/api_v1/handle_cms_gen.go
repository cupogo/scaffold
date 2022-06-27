// This file is generated - Do Not Edit.

package apiv1

import (
	gin "github.com/gin-gonic/gin"
	"hyyl.xyz/cupola/scaffold/pkg/models/cms1"
	"hyyl.xyz/cupola/scaffold/pkg/services/stores"
)

func init() {
	regHI(true, "GET", "/cms/clauses/:id", "", func(a *api) gin.HandlerFunc {
		return a.getCmsClause
	})
	regHI(true, "PUT", "/cms/clauses/:id", "v1-cms-clauses-id-put", func(a *api) gin.HandlerFunc {
		return a.putCmsClause
	})
	regHI(true, "GET", "/cms/clauses", "", func(a *api) gin.HandlerFunc {
		return a.getCmsClauses
	})
	regHI(true, "DELETE", "/cms/clauses/:id", "v1-cms-clauses-id-delete", func(a *api) gin.HandlerFunc {
		return a.deleteCmsClause
	})
}

// @Tags 默认 文档生成
// @Summary 获取内容条款
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Success 200 {object} resp.Done{result=cms1.Clause}
// @Failure 400 {object} resp.Failure "请求或参数错误"
// @Failure 401 {object} resp.Failure "未登录"
// @Failure 404 {object} resp.Failure "目标未找到"
// @Failure 503 {object} resp.Failure "服务端错误"
// @Router /api/v1/cms/clauses/{id} [get]
func (a *api) getCmsClause(c *gin.Context) {
	id := c.Param("id")
	obj, err := a.sto.Contant().GetClause(c.Request.Context(), id)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, obj)
}

// @Tags 默认 文档生成
// @ID v1-cms-clauses-id-put
// @Summary 录入内容条款
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Param   query  formData   cms1.ClauseSet  true   "Object"
// @Success 200 {object} resp.Done{result=string}
// @Failure 400 {object} resp.Failure "请求或参数错误"
// @Failure 401 {object} resp.Failure "未登录"
// @Failure 403 {object} resp.Failure "无权限"
// @Failure 503 {object} resp.Failure "服务端错误"
// @Router /api/v1/cms/clauses/{id} [put]
func (a *api) putCmsClause(c *gin.Context) {
	id := c.Param("id")
	var in cms1.ClauseSet
	if err := c.Bind(&in); err != nil {
		fail(c, 400, err)
		return
	}

	nid, err := a.sto.Contant().PutClause(c.Request.Context(), id, &in)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, idResult(nid))
}

// @Tags 默认 文档生成
// @Summary 列出内容条款
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   query  formData   stores.ClauseSpec  true   "Object"
// @Success 200 {object} resp.Done{result=resp.ResultData{cms1.Clauses}}
// @Failure 400 {object} resp.Failure "请求或参数错误"
// @Failure 401 {object} resp.Failure "未登录"
// @Failure 404 {object} resp.Failure "目标未找到"
// @Failure 503 {object} resp.Failure "服务端错误"
// @Router /api/v1/cms/clauses [get]
func (a *api) getCmsClauses(c *gin.Context) {
	var spec stores.ClauseSpec
	if err := c.Bind(&spec); err != nil {
		fail(c, 400, err)
		return
	}

	data, total, err := a.sto.Contant().ListClause(c.Request.Context(), &spec)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, dtResult(data, total))
}

// @Tags 默认 文档生成
// @ID v1-cms-clauses-id-delete
// @Summary 删除内容条款
// @Accept json
// @Produce json
// @Param token    header   string  true "登录票据凭证"
// @Param   id    path   string  true   "编号"
// @Success 200 {object} resp.Done
// @Failure 400 {object} resp.Failure "请求或参数错误"
// @Failure 401 {object} resp.Failure "未登录"
// @Failure 403 {object} resp.Failure "无权限"
// @Failure 503 {object} resp.Failure "服务端错误"
// @Router /api/v1/cms/clauses/{id} [delete]
func (a *api) deleteCmsClause(c *gin.Context) {
	id := c.Param("id")
	err := a.sto.Contant().DeleteClause(c.Request.Context(), id)
	if err != nil {
		fail(c, 503, err)
		return
	}

	success(c, "ok")
}
