//go:build wireinject

package app

import (
	v1 "go-pg-demo/api/v1"
	"go-pg-demo/api/v1/intf"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/internal/modules/iacc/auth"
	"go-pg-demo/internal/modules/iacc/role"
	iacc_service "go-pg-demo/internal/modules/iacc/service"
	"go-pg-demo/internal/modules/iacc/user"
	"go-pg-demo/internal/modules/template"
	"go-pg-demo/pkgs"

	"github.com/google/wire"
)

func InitializeApp() (*App, func(), error) {
	wire.Build(
		NewGin,
		pkgs.ProviderSet,
		middlewares.ProviderSet,
		template.NewTemplateHandler,
		iacc_service.NewPermissionService,
		user.NewUserHandler,
		role.NewRoleHandler,
		auth.NewAuthHandler,
		v1.NewRouter,
		NewApp,
		// 绑定接口实现
		wire.Bind(new(intf.ITemplateHandler), new(*template.Handler)),
		wire.Bind(new(intf.UserHandler), new(*user.Handler)),
		wire.Bind(new(intf.RoleHandler), new(*role.Handler)),
		wire.Bind(new(intf.AuthHandler), new(*auth.Handler)),
	)
	return nil, nil, nil
}
