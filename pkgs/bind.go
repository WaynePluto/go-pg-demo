package pkgs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/mo"
)

// 绑定并返回路径参数。
func BindUri[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindUri(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

func BindQuery[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindQuery(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

// 绑定并返回请求体的JSON数据。
func BindJSON[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

// 同时绑定路径参数和请求体的JSON数据。
func BindUriAndJSON[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindUri(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}

	return mo.Ok(&req)
}
