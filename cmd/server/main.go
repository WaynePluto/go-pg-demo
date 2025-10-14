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
