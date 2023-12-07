package api

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/induzo/gocom/shutdown"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "use leak detector")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m)

		return
	}

	os.Exit(m.Run())
}

func TestAPI_Start(t *testing.T) { //nolint: tparallel // disable because port conflict on host machine
	t.Parallel()

	cfg := Config{}
	cfg.Version = "v1"
	cfg.Log.IsPretty = true
	cfg.Log.Level = "error"
	cfg.WithDebugProfiler = false
	cfg.HTTP.Port = 8000
	cfg.HTTP.Timeouts.ReadTimeout = 2 * time.Second
	cfg.HTTP.Timeouts.ReadHeaderTimeout = 1 * time.Second
	cfg.HTTP.Timeouts.WriteTimeout = 2 * time.Second
	cfg.HTTP.Timeouts.IdleTimeout = 1 * time.Minute
	cfg.Observability.Collector.Host = "opentelemetry-collector.otel-collector"
	cfg.Observability.Collector.Port = 4317

	tests := []struct {
		name          string
		cfg           Config
		shutdownFuncs map[string]func(ctx context.Context) error
		wantErr       bool
	}{
		{
			name: "shutdown error",
			cfg:  cfg,
			shutdownFuncs: map[string]func(ctx context.Context) error{
				"mock err": func(ctx context.Context) error {
					return fmt.Errorf("mock err")
				},
			},
			wantErr: true,
		},
		{
			name: "happy",
			cfg:  cfg,
			shutdownFuncs: map[string]func(ctx context.Context) error{
				"shutdown": func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests { //nolint: paralleltest // disable because port conflict on host machine
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			shutdownHandler := shutdown.New(logger)
			srv, err := NewServer(&cfg, shutdownHandler, logger)
			if err != nil {
				t.Errorf("failed to create api server: %v", err)

				return
			}

			for name, shutdownFunc := range tt.shutdownFuncs {
				shutdownHandler.Add(name, shutdownFunc)
			}

			if errS := srv.Serve(ctx); errS != nil {
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
