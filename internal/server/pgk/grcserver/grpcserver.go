package grcserver

import (
	"context"
	"crypto/tls"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/shreyner/gophkeeper/internal/server/config"
	"github.com/shreyner/gophkeeper/pkg/server"
)

var (
	_ server.Server = (*GRPCServer)(nil)
)

type GRPCServer struct {
	log    *zap.Logger
	errors chan error

	Server *grpc.Server
	listen net.Listener
}

func NewGRPCServer(log *zap.Logger, cfg *config.Config, address string, interceptors ...grpc.UnaryServerInterceptor) (*GRPCServer, error) {
	cert, err := tls.X509KeyPair([]byte(cfg.CertFile), []byte(cfg.KetFile))

	if err != nil {
		return nil, err
	}

	creds := credentials.NewServerTLSFromCert(&cert)

	grcServer := GRPCServer{
		log:    log,
		errors: make(chan error),

		Server: grpc.NewServer(grpc.Creds(creds), grpc.ChainUnaryInterceptor(interceptors...)),
	}

	listen, err := net.Listen("tcp", address)

	if err != nil {
		return nil, err
	}

	grcServer.listen = listen

	return &grcServer, nil
}

func (s *GRPCServer) Start() error {
	go func() {
		s.log.Info("gRPC server listen on", zap.String("addr", s.listen.Addr().String()))
		defer close(s.errors)

		s.errors <- s.Server.Serve(s.listen)
	}()

	return nil
}

func (s *GRPCServer) Stop(_ context.Context) error {
	s.log.Info("gRPC server stopping ...")
	s.Server.GracefulStop()

	return s.listen.Close()
}

func (s *GRPCServer) Notify() <-chan error {
	return s.errors
}
