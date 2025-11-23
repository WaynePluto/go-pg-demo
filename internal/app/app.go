package app

import (
	"fmt"

	v1 "go-pg-demo/api/v1"
	_ "go-pg-demo/docs" // Swagger docs
	"go-pg-demo/migration"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewGin() *gin.Engine {
	return gin.New()
}

type App struct {
	Server    *gin.Engine
	Logger    *zap.Logger
	Conf      *pkgs.Config
	DB        *sqlx.DB
	V1Router  *v1.Router
	Scheduler *pkgs.Scheduler
}

func NewApp(
	server *gin.Engine,
	logger *zap.Logger,
	conf *pkgs.Config,
	db *sqlx.DB,
	middlewares []gin.HandlerFunc,
	v1Router *v1.Router,
	scheduler *pkgs.Scheduler,
) (*App, error) {

	// 数据库迁移
	err := migration.RunMigrations(db, conf)
	if err != nil {
		return nil, err
	}

	// 应用中间件
	for _, middleware := range middlewares {
		server.Use(middleware)
	}

	v1Router.Register()

	// 添加Swagger路由，仅在非生产环境启用
	if conf.Server.Mode != "release" {
		server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return &App{
		Server:    server,
		Logger:    logger,
		Conf:      conf,
		DB:        db,
		V1Router:  v1Router,
		Scheduler: scheduler,
	}, nil
}

// Run 启动 http 服务
func (a *App) Run() error {
	host := "localhost"
	port := a.Conf.Server.Port
	addr := fmt.Sprintf("%s:%d", host, port)

	a.Logger.Info("HTTP server is starting...", zap.String("addr", addr))

	// 启动定时任务
	if a.Scheduler != nil {
		a.Scheduler.Start()
	}

	return a.Server.Run(addr)
}
