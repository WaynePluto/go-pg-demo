package user

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutesV1 注册用户路由
func (h *Handler) RegisterRoutesV1(router *gin.RouterGroup) {
	users := router.Group("/user")
	{
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUserByID)
		users.PUT("/:id", h.UpdateUser)
		users.GET("/list", h.ListUsers)
		users.POST("/:id/role", h.AssignRole)
		users.DELETE("/:id/role/:role_id", h.RemoveRole)
		users.GET("/:id/permissions", h.GetUserPermissions)
	}
}