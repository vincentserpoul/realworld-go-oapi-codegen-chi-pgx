package cmd

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"syscall"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/induzo/gocom/http/health"
	"github.com/induzo/gocom/shutdown"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type testConfig struct {
	BasicConfig
}

func (c testConfig) GetBasicConfig() BasicConfig {
	return c.BasicConfig
}

//nolint:tparallel // port conflict.
func TestAPI_Start(
	t *testing.T,
) {
	t.Parallel()

	cfg := testConfig{}
	cfg.BasicConfig.Version = "v1"
	cfg.BasicConfig.Log.IsPretty = true
	cfg.BasicConfig.Log.Level = "error"
	cfg.BasicConfig.WithDebugProfiler = true
	cfg.BasicConfig.HTTP.Port = 8000
	cfg.BasicConfig.HTTP.Timeouts.ReadTimeout = 2 * time.Second
	cfg.BasicConfig.HTTP.Timeouts.ReadHeaderTimeout = 1 * time.Second
	cfg.BasicConfig.HTTP.Timeouts.WriteTimeout = 2 * time.Second
	cfg.BasicConfig.HTTP.Timeouts.IdleTimeout = 1 * time.Minute
	cfg.BasicConfig.Observability.Collector.Host = "opentelemetry-collector.otel-collector"
	cfg.BasicConfig.Observability.Collector.Port = 4317

	tests := []struct {
		name          string
		cfg           testConfig
		shutdownFuncs map[string]func(ctx context.Context) error
		wantErr       bool
	}{
		{
			name: "shutdown error",
			cfg:  cfg,
			shutdownFuncs: map[string]func(ctx context.Context) error{
				"mock err": func(_ context.Context) error {
					return errors.New("mock err")
				},
			},
			wantErr: true,
		},
		{
			name: "happy",
			cfg:  cfg,
			shutdownFuncs: map[string]func(ctx context.Context) error{
				"shutdown": func(_ context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests { //nolint: paralleltest // disable because port conflict on host machine
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			logger := slog.New(slog.DiscardHandler)
			shutdownHandler := shutdown.New(logger)

			srv, err := NewServer(cfg, shutdownHandler, logger)
			if err != nil {
				t.Errorf("failed to create api server: %v", err)

				return
			}

			for name, shutdownFunc := range tt.shutdownFuncs {
				shutdownHandler.Add(name, shutdownFunc)
			}

			// register api router
			if errR := srv.RegisterHTTPSvc(
				"/",
				chi.NewRouter(),
				[]health.CheckConfig{},
				map[string]func(ctx context.Context) error{},
			); errR != nil {
				log.Fatalf("failed to register api service: %v", errR)
			}

			if errS := srv.Serve(); errS != nil {
				t.Errorf("api failed to start: %v", errS)

				return
			}

			go func() {
				time.Sleep(10 * time.Millisecond)

				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()

			errS := srv.Shutdown(ctx, syscall.SIGINT)
			if (errS != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v, wantedErr: %v", errS, tt.wantErr)
			}
		})
	}
}

func TestShutdownErrors_Error(t *testing.T) {
	t.Parallel()

	err := ShutdownErrors{
		errors.New("error 1"),
		errors.New("error 2"),
	}

	want := "error cause: error 1;error 2"

	if err.Error() != want {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestShutdownErrors_IsNil(t *testing.T) {
	t.Parallel()

	err := ShutdownErrors{}
	if !err.IsNil() {
		t.Errorf("unexpected error: %v", err)
	}
}
