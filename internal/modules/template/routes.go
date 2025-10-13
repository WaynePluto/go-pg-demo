package template

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册模板路由
func RegisterRoutes(router *gin.RouterGroup, handler *TemplateHandler) {
	templates := router.Group("/templates")
	{
		templates.POST("", handler.CreateTemplate)
		templates.GET("/:id", handler.GetTemplate)
		templates.PUT("/:id", handler.UpdateTemplate)
		templates.DELETE("/:id", handler.DeleteTemplate)
		templates.GET("", handler.ListTemplates)
	}
}