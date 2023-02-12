package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shreyner/gophkeeper/internal/server/config"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/shreyner/gophkeeper/internal/server/auth"
	"github.com/shreyner/gophkeeper/internal/server/httphandlers"
	interceptor_auth "github.com/shreyner/gophkeeper/internal/server/interceptor/auth"
	"github.com/shreyner/gophkeeper/internal/server/pgk/database"
	"github.com/shreyner/gophkeeper/internal/server/pgk/grcserver"
	"github.com/shreyner/gophkeeper/internal/server/pgk/httpserver"
	"github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
	"github.com/shreyner/gophkeeper/internal/server/rpchandlers"
	"github.com/shreyner/gophkeeper/internal/server/user"
	"github.com/shreyner/gophkeeper/internal/server/vault"
	pb "github.com/shreyner/gophkeeper/proto"
)

func NewGophKeeperServer(logger *zap.Logger, cfg *config.Config) error {
	ctxBase := context.Background()

	logger.Info("Start GophKeeper server...")

	logger.Info("Connect to database...")
	db, err := database.NewDataBase(ctxBase, cfg.DBDSN)
	if err != nil {
		logger.Error("Error connection to database", zap.Error(err))
		return err
	}

	logger.Info("Initialize S3 Minio client ...")
	s3minioCLient, err := minio.New(
		cfg.S3MinioEndpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(cfg.S3MinioAccessKeyID, cfg.S3MinioSecretAccessKey, ""),
		},
	)

	if err != nil {
		logger.Error("Can't create S3 Minio client", zap.Error(err))
		return err
	}

	userRepository := user.NewRepository(db)
	vaultRepository := vault.NewRepository(db)

	vaultService := vault.NewService(vaultRepository)
	stokenService := stoken.NewService([]byte(cfg.JWTSign))
	userService := user.NewService(userRepository)
	authService := auth.NewService(userService)

	logger.Info("Create http router...")
	router := httphandlers.NewRouter(logger, stokenService, s3minioCLient)

	logger.Info("Create http server...")
	hserver := httpserver.NewHTTPServer(
		logger,
		fmt.Sprintf("0.0.0.0:%v", cfg.Port),
		router,
	)

	logger.Info("Create grpc server...")
	gserver, err := grcserver.NewGRPCServer(
		logger,
		fmt.Sprintf(":%v", cfg.GRPCServerPort),
		interceptor_auth.Interceptor(stokenService),
	)

	if err != nil {
		logger.Error("Can't start grpc server", zap.Error(err))
		return err
	}

	rpcGophkeeperServer := rpchandlers.NewGophkeeperServer(logger, authService, stokenService, vaultService)

	pb.RegisterGophkeeperServer(gserver.Server, rpcGophkeeperServer)

	logger.Info("Start services ...")
	_ = hserver.Start()
	_ = gserver.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	logger.Info("Listen interrupt or errors from service ...")
	select {
	case x := <-interrupt:
		logger.Info("Received a signal.", zap.String("signal", x.String()))
	case err := <-gserver.Notify():
		logger.Error("Received an error from the start grpc server", zap.Error(err))
	case err := <-hserver.Notify():
		logger.Error("Received an error from the start http server", zap.Error(err))
	}

	logger.Info("Stopping server...")

	err = gserver.Stop(context.Background())
	if err != nil {
		logger.Error("Got an error when stopping the grpc server", zap.Error(err))
	}

	err = hserver.Stop(context.Background())
	if err != nil {
		logger.Error("Got an error when stopping the http server", zap.Error(err))
	}

	logger.Info("The app is calling the last defers and will be stopped.")

	return nil
}
