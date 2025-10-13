package middlewares

import (
	"github.com/google/wire"
)

// MiddlewareSet 中间件集合，用于依赖注入
var MiddlewareSet = wire.NewSet(
	LoggerMiddleware,
	RecoveryMiddleware,
	AuthMiddleware,
)