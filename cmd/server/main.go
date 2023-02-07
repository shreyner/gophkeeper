package main

import (
	logStd "log"

	"go.uber.org/zap"

	"github.com/shreyner/gophkeeper/internal/server"
	"github.com/shreyner/gophkeeper/internal/server/pgk/logger"
)

func main() {
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
