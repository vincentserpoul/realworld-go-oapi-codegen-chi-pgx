package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/induzo/gocom/shutdown"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	otelTraceIDKey = "trace_id"
	otelSpanIDKey  = "span_id"
)

type Header struct {
	Key   string `koanf:"key"`
	Value string `koanf:"value"`
}

type observabilityConfig struct {
	Collector struct {
		Host       string   `koanf:"host"`
		Port       int      `koanf:"port"`
		Headers    []Header `koanf:"headers"`
		IsInsecure bool     `koanf:"is_insecure"`
	} `koanf:"collector"`
}

func StartOtel(
	ctx context.Context,
	bCfg *BasicConfig,
	shutdownHandler *shutdown.Shutdown,
	logger *slog.Logger,
) error {
	headers := make(map[string]string, len(bCfg.Observability.Collector.Headers))
	for _, header := range bCfg.Observability.Collector.Headers {
		headers[header.Key] = header.Value
	}

	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(
			fmt.Sprintf(
				"%s:%d",
				bCfg.Observability.Collector.Host,
				bCfg.Observability.Collector.Port,
			),
		),
		otlptracegrpc.WithHeaders(headers),
	}

	if bCfg.Observability.Collector.IsInsecure {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithInsecure())
	}

	exporter, errT := otlptracegrpc.New(ctx, exporterOptions...)
	if errT != nil {
		return fmt.Errorf("failed to create cli: %w", errT)
	}

	// Ensure default SDK resources and the required service name are set.
	rsrc, errR := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(fmt.Sprintf("%s-%s", bCfg.Name, bCfg.Env)),
			semconv.ServiceVersion(bCfg.Version),
		),
	)
	if errR != nil {
		return fmt.Errorf("failed to merge resource otel: %w", errR)
	}

	prov := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(rsrc),
	)

	otel.SetTracerProvider(prov)

	// Set up metrics
	metricExporterOptions := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(
			fmt.Sprintf(
				"%s:%d",
				bCfg.Observability.Collector.Host,
				bCfg.Observability.Collector.Port,
			),
		),
		otlpmetricgrpc.WithHeaders(headers),
	}

	if bCfg.Observability.Collector.IsInsecure {
		metricExporterOptions = append(metricExporterOptions, otlpmetricgrpc.WithInsecure())
	}

	metricExporter, errM := otlpmetricgrpc.New(ctx, metricExporterOptions...)
	if errM != nil {
		return fmt.Errorf("failed to create metric exporter: %w", errM)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(rsrc),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)

	otel.SetMeterProvider(meterProvider)

	shutdownHandler.Add("otel-metrics", func(context.Context) error {
		if errS := meterProvider.Shutdown(ctx); errS != nil {
			return fmt.Errorf("failed to shutdown otel metrics: %w", errS)
		}

		return nil
	})

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdownHandler.Add("otel", func(context.Context) error {
		if errS := prov.Shutdown(ctx); errS != nil {
			return fmt.Errorf("failed to shutdown otel: %w", errS)
		}

		return nil
	})

	logger.InfoContext(
		ctx,
		"otel client started",
		slog.Int("collector_port", bCfg.Observability.Collector.Port),
		slog.String("collector_host", bCfg.Observability.Collector.Host),
	)

	return nil
}
