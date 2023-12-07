package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/honeycombio/otel-config-go/otelconfig"
)

func (s *Server) StartOtel() error {
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry(
		otelconfig.WithServiceName(fmt.Sprintf("%s-%s", s.cfg.Name, s.cfg.Env)),
		otelconfig.WithServiceVersion(s.cfg.Version),
		// otelconfig.WithHeaders(map[string]string{
		// 	"service-auth-key":     "value",
		// }),
		otelconfig.WithExporterEndpoint(
			fmt.Sprintf("%s:%d", s.cfg.Observability.Collector.Host, s.cfg.Observability.Collector.Port),
		),
		otelconfig.WithExporterInsecure(!s.cfg.Observability.Collector.IsSecure),
		otelconfig.WithMetricsEnabled(s.cfg.Observability.Collector.EnableMetrics),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize opentelemetry: %w", err)
	}

	s.shutdownHandler.Add("otel", func(context.Context) error {
		otelShutdown()

		return nil
	})

	s.logger.Info(
		"otel client started",
		slog.Int("collector_port", s.cfg.Observability.Collector.Port),
		slog.String("collector_host", s.cfg.Observability.Collector.Host),
	)

	return nil
}
