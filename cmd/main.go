package main

import (
	"stmnplibrary/log"
	"stmnplibrary/cmd/wire"

	// "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	
	// "context"
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
	router, _, err := wiring.InitializeApp()
	if err != nil {
		panic(err)
	}
	router.Run()
}
