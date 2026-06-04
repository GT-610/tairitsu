package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/GT-610/tairitsu/internal/app/bootstrap"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// main is the application entry point
func main() {
	fmt.Println("Tairitsu - ZeroTier Controller Interface")

	app, err := bootstrap.Build()
	if err != nil {
		logger.Fatal("application initialization failed", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := app.Listen(); err != nil {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	<-quit
	app.Shutdown()
	logger.Sync()
}
