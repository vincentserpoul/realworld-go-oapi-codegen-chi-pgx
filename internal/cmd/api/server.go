package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/induzo/gocom/http/health"
	"github.com/induzo/gocom/shutdown"
)

type ShutdownErrors []error

func (e ShutdownErrors) Error() string {
	errStr := make([]string, 0, len(e))

	for _, err := range e {
		if err != nil {
			errStr = append(errStr, err.Error())
		}
	}

	return fmt.Sprintf("error cause: %s", strings.Join(errStr, ";"))
}

func (e ShutdownErrors) IsNil() bool {
	return len(e) == 0
}

type Server struct {
	cfg             *Config
	logger          *slog.Logger
	httpMgr         *httpManager
	shutdownHandler *shutdown.Shutdown
}

// NewServer in default, the server will serve grpc, swagger, http server
func NewServer(cfg *Config, shutdownHandler *shutdown.Shutdown, logger *slog.Logger) (*Server, error) {
	httpMgr, err := newHTTPManager(&cfg.HTTP, cfg.WithDebugProfiler)
	if err != nil {
		return nil, fmt.Errorf("failed to create http manager: %w", err)
	}

	return &Server{
		cfg:             cfg,
		logger:          logger,
		httpMgr:         httpMgr,
		shutdownHandler: shutdownHandler,
	}, nil
}

type HTTPSvcRegisterer interface {
	RegisterHTTPSvc(
		route string,
		handler http.Handler,
		healthChecks []health.CheckConfig,
		shutdownFuncs map[string]func(ctx context.Context) error,
	) error
}

func (s *Server) RegisterHTTPSvc(
	route string,
	handler http.Handler,
	healthChecks []health.CheckConfig,
	shutdownFuncs map[string]func(ctx context.Context) error,
) error {
	s.httpMgr.RegisterHTTPSvc(route, handler, healthChecks)

	for svc, shutdownFunc := range shutdownFuncs {
		s.shutdownHandler.Add(svc, shutdownFunc)
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	if err := s.startOtel(ctx); err != nil {
		return fmt.Errorf("failed to start otel: %w", err)
	}

	if s.httpMgr != nil {
		shutdownHTTP := s.httpMgr.startHTTPServer(s.logger)

		s.shutdownHandler.Add("http server", shutdownHTTP)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context, signals ...os.Signal) error {
	if err := s.shutdownHandler.Listen(ctx, signals...); err != nil {
		return fmt.Errorf("failed to listen shutdown signals: %w", err)
	}

	return nil
}
