package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// NewUseMiddlewares 创建并返回按正确顺序排列的中间件切片
// 顺序：logger -> auth -> permission -> recovery
func NewUseMiddlewares(
	loggerMiddleware LoggerMiddleware,
	authMiddleware AuthMiddleware,
	permissionMiddleware PermissionMiddleware,
	recoveryMiddleware RecoveryMiddleware,
) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.HandlerFunc(loggerMiddleware),
		gin.HandlerFunc(authMiddleware),
		gin.HandlerFunc(permissionMiddleware),
		gin.HandlerFunc(recoveryMiddleware),
	}
}

var ProviderSet = wire.NewSet(
	NewLoggerMiddleware,
	NewRecoveryMiddleware,
	NewAuthMiddleware,
	NewPermissionMiddleware,
	NewUseMiddlewares,
)
