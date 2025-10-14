package template

import (
	"github.com/gin-gonic/gin"
)

// 注册模板路由
func RegisterRoutesV1(router *gin.RouterGroup, handler *TemplateHandler) {
	templates := router.Group("/template")
	{
		templates.POST("", handler.CreateTemplate)
		templates.GET("/:id", handler.GetTemplateByID)
		templates.PUT("/:id", handler.UpdateTemplate)
		templates.DELETE("/:id", handler.DeleteTemplate)
		templates.POST("/batch-create", handler.CreateTemplateBatch)
		templates.GET("/list", handler.QueryTemplateList)
		templates.POST("/batch-delete", handler.DeleteTemplateBatch)
	}
}
