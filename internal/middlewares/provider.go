package middlewares

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewLoggerMiddleware,
	NewRecoveryMiddleware,
	NewAuthMiddleware,
)
