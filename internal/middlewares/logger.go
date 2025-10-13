package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 日志中间件，记录每个请求的路径、方法、IP和耗时
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		
		// 获取状态码
		statusCode := c.Writer.Status()

		// 记录日志
		if len(c.Errors) > 0 {
			// 如果有错误，记录错误日志
			logger.Error("Request",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", ip),
				zap.Duration("latency", latency),
				zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			)
		} else {
			// 记录普通日志
			logger.Info("Request",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", ip),
				zap.Duration("latency", latency),
			)
		}
	}
}