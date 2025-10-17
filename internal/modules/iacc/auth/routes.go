package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutesV1 注册认证路由
func (h *Handler) RegisterRoutesV1(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}
}