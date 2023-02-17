package httpserver

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/shreyner/gophkeeper/internal/server/config"
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

func NewHTTPServer(log *zap.Logger, cfg *config.Config, address string, router http.Handler) (*HTTPServer, error) {
	cert, err := tls.X509KeyPair([]byte(cfg.CertFile), []byte(cfg.KeyFile))

	if err != nil {
		return nil, err
	}

	tlcCfg := tls.Config{}
	tlcCfg.Certificates = append(tlcCfg.Certificates, cert)

	return &HTTPServer{
		server: http.Server{
			Addr:      address,
			Handler:   router,
			TLSConfig: &tlcCfg,
		},
		log:    log,
		errors: make(chan error),
	}, nil
}

func (s *HTTPServer) Start() error {
	go func() {
		s.log.Info("Http HTTPServer listening on ", zap.String("addr", s.server.Addr))
		defer close(s.errors)

		s.errors <- s.server.ListenAndServeTLS("", "")
	}()

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) Notify() <-chan error {
	return s.errors
}
