package app

import (
	"fmt"

	v1 "go-pg-demo/internal/api/v1"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/internal/pkgs"
	_ "go-pg-demo/docs" // Swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// App 是一个包含了所有应用组件的结构体
type App struct {
	Server *gin.Engine
	Logger *zap.Logger
	Conf   *pkgs.Config
	DB     *sqlx.DB
}

func NewGin() *gin.Engine {
	return gin.New()
}

// NewApp 是 App 的构造函数，用于 wire 注入
func NewApp(
	server *gin.Engine,
	logger *zap.Logger,
	conf *pkgs.Config,
	db *sqlx.DB,
	useRouterV1 v1.RegisterRouter,
	useMiddlewares middlewares.UseMiddlewares,
) (*App, error) {

	err := pkgs.RunMigrations(db, conf)
	if err != nil {
		return nil, err
	}
	useMiddlewares()
	useRouterV1()

	// 添加Swagger路由，仅在非生产环境启用
	if conf.Server.Mode != "release" {
		server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return &App{
		Server: server,
		Logger: logger,
		Conf:   conf,
		DB:     db,
	}, nil
}

// Run 启动 http 服务
func (a *App) Run() error {
	host := "localhost"
	port := a.Conf.Server.Port
	addr := fmt.Sprintf("%s:%d", host, port)

	a.Logger.Info("HTTP server is starting...", zap.String("addr", addr))

	return a.Server.Run(addr)
}