package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/honeycombio/otel-config-go/otelconfig"
)

func (s *Server) StartOtel() error {
	headers := make(map[string]string, len(s.cfg.Observability.Collector.Headers))
	for _, header := range s.cfg.Observability.Collector.Headers {
		headers[header.Key] = header.Value
	}

	otelShutdown, err := otelconfig.ConfigureOpenTelemetry(
		otelconfig.WithServiceName(fmt.Sprintf("%s-%s", s.cfg.Name, s.cfg.Env)),
		otelconfig.WithServiceVersion(s.cfg.Version),
		otelconfig.WithHeaders(headers),
		otelconfig.WithExporterEndpoint(
			fmt.Sprintf("%s:%d", s.cfg.Observability.Collector.Host, s.cfg.Observability.Collector.Port),
		),
		otelconfig.WithExporterInsecure(s.cfg.Observability.Collector.IsInsecure),
		otelconfig.WithMetricsEnabled(s.cfg.Observability.Collector.WithMetricsEnabled),
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
