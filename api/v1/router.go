package v1

import (
	"go-pg-demo/internal/modules/template"

	"github.com/gin-gonic/gin"
)

type RegisterRoutesFunc func()

// 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	templateHandler *template.TemplateHandler,
) RegisterRoutesFunc {
	return func() {
		api := router.Group("/v1")
		// 注册模板路由
		template.RegisterRoutesV1(api, templateHandler)
	}

}
