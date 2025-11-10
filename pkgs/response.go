package pkgs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构体
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// 处理错误响应
func HandleError(c *gin.Context, err error) {
	if apiErr, ok := err.(*ApiError); ok {
		Error(c, apiErr.Code, apiErr.Message)
	} else {
		Error(c, http.StatusInternalServerError, "服务器内部错误")
	}
}
