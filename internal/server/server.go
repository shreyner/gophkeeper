package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/shreyner/gophkeeper/internal/pgk/stoken"
	"github.com/shreyner/gophkeeper/internal/server/auth"
	"github.com/shreyner/gophkeeper/internal/server/database"
	interceptor_auth "github.com/shreyner/gophkeeper/internal/server/interceptor/auth"
	"github.com/shreyner/gophkeeper/internal/server/rpchandlers"
	"github.com/shreyner/gophkeeper/internal/server/user"
	"github.com/shreyner/gophkeeper/internal/server/vault"
	pb "github.com/shreyner/gophkeeper/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/shreyner/gophkeeper/internal/pgk/grcserver"
)

func NewGophKeeperServer(logger *zap.Logger) error {
	ctxBase := context.Background()

	logger.Info("Start GophKeeper server...")

	logger.Info("Connect to database...")
	db, err := database.NewDataBase(ctxBase, "postgres://postgres:postgres@localhost:5432/develop?sslmode=disable")
	if err != nil {
		logger.Error("Error connection to database", zap.Error(err))
		return err
	}

	userRepository := user.NewRepository(db)
	vaultRepository := vault.NewRepository(db)

	vaultService := vault.NewService(vaultRepository)
	stokenService := stoken.NewService([]byte("123"))
	userService := user.NewService(userRepository)
	authService := auth.NewService(userService)

	logger.Info("Create grpc server")
	gserver, err := grcserver.NewGRPCServer(logger, ":3200", interceptor_auth.Interceptor(stokenService))

	if err != nil {
		logger.Error("Can't start grpc server", zap.Error(err))
		return err
	}

	rpcGophkeeperServer := rpchandlers.NewGophkeeperServer(logger, authService, stokenService, vaultService)

	pb.RegisterGophkeeperServer(gserver.Server, rpcGophkeeperServer)

	gserver.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case x := <-interrupt:
		logger.Info("Received a signal.", zap.String("signal", x.String()))
	case err := <-gserver.Notify():
		logger.Error("Received an error from the start grpc server", zap.Error(err))
	}

	logger.Info("Stopping server...")

	if err := gserver.Stop(context.Background()); err != nil {
		logger.Error("Got an error while stopping th grpc server", zap.Error(err))
	}

	logger.Info("The app is calling the last defers and will be stopped.")

	return nil
}
