package role

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutesV1 注册角色路由
func (h *Handler) RegisterRoutesV1(router *gin.RouterGroup) {
	roles := router.Group("/role")
	{
		roles.POST("", h.CreateRole)
		roles.GET("/:id", h.GetRoleByID)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.GET("/list", h.ListRoles)
	}
}