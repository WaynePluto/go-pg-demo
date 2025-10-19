package middlewares

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-pg-demo/pkgs"
)

// 恢复中间件，捕获任何panic并记录错误日志，防止服务崩溃
type RecoveryMiddleware gin.HandlerFunc

func NewRecoveryMiddleware(logger *zap.Logger) RecoveryMiddleware {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := debug.Stack()

				// 记录错误日志
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				)

				// 返回500错误
				pkgs.Error(c, 500, "服务器内部错误")
			}
		}()
		c.Next()
	}
}
