// Package main
//
// @title           Go-PG Demo API
// @version         1.0
// @description     This is a sample server for go-pg-demo server.
// @termsOfService  https://swagger.io/terms/
//
// @contact.name   API Support
// @contact.url    https://www.swagger.io/support
// @contact.email  support@swagger.io
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host      localhost:8080
// @BasePath  /v1
//
// @securityDefinitions.apikey JWT
// @in                          header
// @name                        Authorization
// @description                 JWT token for authentication
package main

import (
	"go-pg-demo/internal/app"
	"log"
)

func main() {
	// 初始化应用
	application, cleanup, err := app.InitializeApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer cleanup()

	// 启动服务
	if err := application.Run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
