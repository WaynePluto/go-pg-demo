//go:build wireinject

package app

import (
	v1 "go-pg-demo/api/v1"
	"go-pg-demo/internal/modules/template"
	"go-pg-demo/internal/pkgs"

	"github.com/google/wire"
)

func InitializeApp() (*App, func(), error) {
	wire.Build(
		NewGin,
		pkgs.ProviderSet,
		template.NewTemplateHandler,
		v1.NewRegisterRouter,
		NewApp,
	)
	return nil, nil, nil
}
