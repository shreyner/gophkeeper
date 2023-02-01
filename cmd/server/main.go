package main

import (
	"fmt"
	logStd "log"

	"github.com/shreyner/gophkeeper/internal/pgk/logger"
	"github.com/shreyner/gophkeeper/internal/server"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Hello server")

	log, err := logger.InitLogger()

	if err != nil {
		logStd.Fatal("Error initialization logger: %w", err)
		return
	}

	defer log.Sync()

	if err := server.NewGophKeeperServer(log); err != nil {
		log.Error("GophKeeper Server return error", zap.Error(err))
		return
	}

}
