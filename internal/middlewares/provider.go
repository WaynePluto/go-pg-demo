package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 注册中间件函数
type UseMiddlewares func()

// 注册中间件函数的构造函数
func NewUseMiddlewares(
	server *gin.Engine,
	logger *zap.Logger,
) UseMiddlewares {
	return func() {
		server.Use(LoggerMiddleware(logger))
		server.Use(RecoveryMiddleware(logger))
	}
}
