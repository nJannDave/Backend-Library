package main

import (
	"context"
	"os/signal"
	"stmnplibrary/cmd/wire"
	"stmnplibrary/log"
	"syscall"
	"time"

	// "github.com/gin-gonic/gin"
	"net/http"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"os"
	// "fmt"
)

func initLog() *zap.Logger {
	zl := zap.NewDevelopmentConfig()
	zl.DisableStacktrace = true
	zapLog, err := zl.Build()
	if err != nil {
		panic(err)
	}
	log.LogInit(zapLog)
	return zapLog
}

func main() {
	zapLog := initLog()
	defer zapLog.Sync()
	if err := godotenv.Load(".env"); err != nil {
		zapLog.Error("failed open .env file")
		return
	}	
	router, stop, err := wiring.InitializeApp()
	if err != nil {
		zapLog.Sugar().Fatalf("Error while initialize app: %v", err)
	}

	srv := &http.Server{
		Addr: os.Getenv("SERVER_PORT"),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed  {
			zapLog.Sugar().Fatalf("Error while start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<- quit

	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLog.Sugar().Fatalf("Error while shutdown server: %v", err)
	}

	stop()
}
