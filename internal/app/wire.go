//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"main/internal/middlewares"
	"main/internal/modules/permission"
	"main/internal/modules/role"
	"main/internal/modules/template"
	"main/internal/modules/user"
	"main/internal/pkgs"
)

// InitializeApp 初始化应用
func InitializeApp() (*App, func(), error) {
	panic(wire.Build(
		pkgs.ProviderSet,
		middlewares.ProviderSet,
		user.ProviderSet,
		permission.ProviderSet,
		role.ProviderSet,
		template.ProviderSet,
		NewApp,
	))
}