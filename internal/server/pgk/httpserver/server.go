package httpserver

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/shreyner/gophkeeper/pkg/server"
)

var (
	_ server.Server = (*HTTPServer)(nil)
)

type HTTPServer struct {
	log    *zap.Logger
	errors chan error
	server http.Server
}

func NewHTTPServer(log *zap.Logger, address string, router http.Handler) *HTTPServer {
	return &HTTPServer{
		server: http.Server{
			Addr:    address,
			Handler: router,
		},
		log:    log,
		errors: make(chan error),
	}
}

func (s *HTTPServer) Start() error {
	go func() {
		s.log.Info("Http HTTPServer listening on ", zap.String("addr", s.server.Addr))
		defer close(s.errors)

		s.errors <- s.server.ListenAndServe()
	}()

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) Notify() <-chan error {
	return s.errors
}
