package api

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
	"github.com/go-chi/httplog/v2"
	"github.com/induzo/gocom/http/health"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type httpManager struct {
	cfg           *HTTPConfig
	mux           *chi.Mux
	healthManager *health.Health
	withDebug     bool
}

const healthQuietDownPeriod = 15 * time.Minute

func newHTTPManager(cfg *HTTPConfig, withDebug bool) (*httpManager, error) {
	mux := chi.NewMux()

	// Logger
	logger := &httplog.Logger{
		Logger: slog.Default(),
		Options: httplog.Options{
			Concise:          true,
			RequestHeaders:   true,
			MessageFieldName: "message",
			QuietDownRoutes: []string{
				"/",
				cfg.HealthEndpoint,
			},
			QuietDownPeriod: healthQuietDownPeriod,
			SourceFieldName: "source",
		},
	}

	mux.Use(httplog.RequestLogger(logger))
	// add otel middleware
	mux.Use(otelHandler)

	// add all health checks to the health endpoint
	healthMgr := health.NewHealth()
	mux.Method(http.MethodGet, health.HealthEndpoint, healthMgr.Handler())

	// add debug endpoints if debug mode is enabled
	if withDebug {
		if err := registerHTTPDebug(mux); err != nil {
			return nil, fmt.Errorf("failed to register http debug: %w", err)
		}
	}

	return &httpManager{
		cfg:           cfg,
		mux:           mux,
		healthManager: healthMgr,
		withDebug:     withDebug,
	}, nil
}

func (hm *httpManager) startHTTPServer(logger *slog.Logger) func(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", hm.cfg.Port),
		ReadTimeout:       hm.cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: hm.cfg.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      hm.cfg.Timeouts.WriteTimeout,
		IdleTimeout:       hm.cfg.Timeouts.IdleTimeout,
		Handler:           hm.mux,
	}

	// Start http server
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"http server failed to listen and serve",
				slog.String("err", err.Error()),
			)
		}
	}()

	// serve router
	logger.Info(
		"HTTP server listening",
		slog.Int("port", hm.cfg.Port),
		slog.String("readTimeout", hm.cfg.Timeouts.ReadTimeout.String()),
		slog.String("readHeaderTimeout", hm.cfg.Timeouts.ReadHeaderTimeout.String()),
		slog.String("writeTimeout", hm.cfg.Timeouts.WriteTimeout.String()),
		slog.String("idleTimeout", hm.cfg.Timeouts.IdleTimeout.String()),
		slog.Bool("debug", hm.withDebug),
	)

	return func(ctx context.Context) error {
		err := httpServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("http server shutdown with err: %w", err)
		}

		return nil
	}
}

func registerHTTPDebug(rtr *chi.Mux) error {
	rtr.Mount("/debug", middleware.Profiler())

	srvStatsviz, errViz := statsviz.NewServer()
	if errViz != nil {
		return fmt.Errorf("failed to create statsviz server: %w", errViz)
	}

	rtr.Get("/debug/statsviz/ws", srvStatsviz.Ws())
	rtr.Get("/debug/statsviz", func(respW http.ResponseWriter, req *http.Request) {
		http.Redirect(respW, req, "/debug/statsviz/", http.StatusMovedPermanently)
	})
	rtr.Handle("/debug/statsviz/*", srvStatsviz.Index())

	return nil
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

func otelHandler(h http.Handler) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(resp, req)

			routePattern := chi.RouteContext(req.Context()).RoutePattern()

			span := trace.SpanFromContext(req.Context())
			span.SetName(routePattern)
			span.SetAttributes(semconv.HTTPTarget(req.URL.String()), semconv.HTTPRoute(routePattern))

			labeler, ok := otelhttp.LabelerFromContext(req.Context())
			if ok {
				labeler.Add(semconv.HTTPRoute(routePattern))
			}
		}),
		"",
		otelhttp.WithMeterProvider(otel.GetMeterProvider()),
	)
}
