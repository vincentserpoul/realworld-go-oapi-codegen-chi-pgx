package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/arl/statsviz"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/induzo/gocom/http/health"
	"github.com/riandyrn/otelchi"
	otelchimetric "github.com/riandyrn/otelchi/metric"
)

type httpConfig struct {
	Port     int `koanf:"port"`
	Timeouts struct {
		ReadTimeout       time.Duration `koanf:"read_timeout"`
		ReadHeaderTimeout time.Duration `koanf:"read_header_timeout"`
		WriteTimeout      time.Duration `koanf:"write_timeout"`
		IdleTimeout       time.Duration `koanf:"idle_timeout"`
	} `koanf:"timeouts"`
	HealthEndpoint string `koanf:"health_endpoint"`
}

type httpManager struct {
	cfg           *httpConfig
	mux           *chi.Mux
	healthManager *health.Health
	withDebug     bool
	shutdownDebug func(context.Context) error
}

func newHTTPManager(
	serverName string,
	logger *slog.Logger,
	cfg *httpConfig,
	withDebug bool,
) (*httpManager, error) {
	mux := chi.NewMux()

	// hLogger
	hLoggerOptions := &httplog.Options{
		Skip: func(req *http.Request, respStatus int) bool {
			// skip health endpoint logging
			if req.URL.Path == cfg.HealthEndpoint && respStatus == http.StatusOK {
				return true
			}

			return false
		},
	}

	baseCfg := otelchimetric.NewBaseConfig(serverName)
	mux.Use(
		otelchi.Middleware(serverName, otelchi.WithChiRoutes(mux)),
		otelchimetric.NewRequestDurationMillis(baseCfg),
		otelchimetric.NewRequestInFlight(baseCfg),
		otelchimetric.NewResponseSizeBytes(baseCfg),
		httplog.RequestLogger(logger, hLoggerOptions),
	)

	// add all health checks to the health endpoint
	healthMgr := health.NewHealth()
	mux.Method(http.MethodGet, health.HealthEndpoint, healthMgr.Handler())

	// add debug endpoints if debug mode is enabled
	shutdownDebug := func(context.Context) error { return nil }

	if withDebug {
		var errSh error

		shutdownDebug, errSh = registerHTTPDebug(mux)
		if errSh != nil {
			return nil, fmt.Errorf("failed to register http debug: %w", errSh)
		}
	}

	return &httpManager{
		cfg:           cfg,
		mux:           mux,
		healthManager: healthMgr,
		withDebug:     withDebug,
		shutdownDebug: shutdownDebug,
	}, nil
}

func (hm *httpManager) startHTTPServer(logger *slog.Logger) func(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", hm.cfg.Port),
		ReadTimeout:       hm.cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: hm.cfg.Timeouts.ReadHeaderTimeout,
		// we don't want to set this, to be able to use sse without timeout
		// we ll use a middleware to manage timeouts on a per route basis
		// WriteTimeout:      hm.cfg.Timeouts.WriteTimeout,
		IdleTimeout: hm.cfg.Timeouts.IdleTimeout,
		Handler:     hm.mux,
	}

	// Start http server
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(context.Background(),
				"http server failed to listen and serve",
				slog.String("err", err.Error()),
			)
		}
	}()

	// serve router
	logger.InfoContext(
		context.Background(),
		"HTTP server listening",
		slog.Int("port", hm.cfg.Port),
		slog.String("readTimeout", hm.cfg.Timeouts.ReadTimeout.String()),
		slog.String("readHeaderTimeout", hm.cfg.Timeouts.ReadHeaderTimeout.String()),
		slog.String("writeTimeout", hm.cfg.Timeouts.WriteTimeout.String()),
		slog.String("idleTimeout", hm.cfg.Timeouts.IdleTimeout.String()),
		slog.Bool("debug", hm.withDebug),
	)

	return func(ctx context.Context) error {
		// shutdown debug server if enabled
		if err := hm.shutdownDebug(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown http debug server", "err", err)
		}

		err := httpServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("http server shutdown with err: %w", err)
		}

		return nil
	}
}

func registerHTTPDebug(rtr *chi.Mux) (func(context.Context) error, error) {
	rtr.Mount("/debug", middleware.Profiler())

	srvStatsviz, errViz := statsviz.NewServer()
	if errViz != nil {
		return nil, fmt.Errorf("failed to create statsviz server: %w", errViz)
	}

	rtr.Get("/debug/statsviz/ws", srvStatsviz.Ws())
	rtr.Get("/debug/statsviz", func(respW http.ResponseWriter, req *http.Request) {
		http.Redirect(respW, req, "/debug/statsviz/", http.StatusMovedPermanently)
	})
	rtr.Handle("/debug/statsviz/*", srvStatsviz.Index())

	return func(_ context.Context) error {
		return srvStatsviz.Close()
	}, nil
}

func (hm *httpManager) RegisterHTTPSvc(
	route string,
	handler http.Handler,
	healthChecks []health.CheckConfig,
) {
	hm.mux.Mount(route, handler)

	for _, hc := range healthChecks {
		hm.healthManager.RegisterCheck(hc)
	}
}
