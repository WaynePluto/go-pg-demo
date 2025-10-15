package v1

import (
	"go-pg-demo/internal/modules/template"

	"github.com/gin-gonic/gin"
)

// 提供一个注册路由的函数
type RegisterRouter func()

// 构造路由注册函数
func NewRegisterRouter(
	router *gin.Engine,
	templateHandler *template.Handler,
) RegisterRouter {
	return func() {
		routerGroup := router.Group("/v1")
		// 注册模板路由
		templateHandler.RegisterRoutesV1(routerGroup)
	}

}
