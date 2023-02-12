package main

import (
	logStd "log"

	"github.com/shreyner/gophkeeper/internal/server/config"
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

	cfg := config.New()
	err = cfg.Parse()
	if err != nil {
		log.Error("can't parsed config", zap.Error(err))
		return
	}

	err = server.NewGophKeeperServer(log, cfg)

	if err != nil {
		log.Error("GophKeeper Server return error", zap.Error(err))
		return
	}

}
