package main

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/bootstrap"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// main is the application entry point
func main() {
	fmt.Println("Tairitsu - ZeroTier Controller Interface")

	app, err := bootstrap.Build()
	if err != nil {
		logger.Fatal("应用初始化失败", zap.Error(err))
	}

	if err := app.Listen(); err != nil {
		logger.Fatal("启动服务器失败", zap.Error(err))
	}
}
