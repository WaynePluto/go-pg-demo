package v1

import (
	"go-pg-demo/internal/modules/template"

	"github.com/gin-gonic/gin"
)

type RegisterRouter func()

// 注册所有路由
func NewRegisterRouter(
	router *gin.Engine,
	templateHandler *template.TemplateHandler,
) RegisterRouter {
	return func() {
		api := router.Group("/v1")
		// 注册模板路由
		template.RegisterRoutesV1(api, templateHandler)
	}

}
