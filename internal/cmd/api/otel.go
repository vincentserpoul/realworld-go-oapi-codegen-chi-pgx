package api

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/induzo/gocom/monitoring/otelinit/v3"
)

func (s *Server) startOtel(ctx context.Context) error {
	var err error

	otelShutdowns, err := otelinit.Start(
		ctx,
		&otelinit.Config{
			AppName:       fmt.Sprintf("%s-%s", s.cfg.Name, s.cfg.Env),
			Host:          s.cfg.Observability.Collector.Host,
			Port:          s.cfg.Observability.Collector.Port,
			APIKey:        "",
			IsSecure:      s.cfg.Observability.Collector.IsSecure,
			EnableMetrics: s.cfg.Observability.Collector.EnableMetrics,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to initialize opentelemetry: %w", err)
	}

	for i, sdF := range otelShutdowns {
		s.shutdownHandler.Add("otel_"+strconv.Itoa(i), sdF)
	}

	s.logger.Info(
		"otel client started",
		slog.Int("collector_port", s.cfg.Observability.Collector.Port),
		slog.String("collector_host", s.cfg.Observability.Collector.Host),
	)

	return nil
}
