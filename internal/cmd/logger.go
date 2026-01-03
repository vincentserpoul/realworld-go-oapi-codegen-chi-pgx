package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel/trace"
)

func NewLogger(writer io.Writer, level string, isPretty bool) *slog.Logger {
	// log level
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelError
	}

	// create a new logger
	var handler slog.Handler

	if isPretty {
		handler = tint.NewHandler(
			writer,
			&tint.Options{
				AddSource:  true,
				Level:      lvl,
				TimeFormat: time.RFC3339,
			},
		)
	} else {
		handler = slog.NewJSONHandler(
			writer,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     lvl,
			},
		)
	}

	traceHandler := contextHandler{handler}

	logger := slog.New(traceHandler)

	// set global logger with custom options
	slog.SetDefault(logger)

	return logger
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(
	ctx context.Context,
	record slog.Record, //nolint:gocritic // contextHandler implements slog.Handler interface
) error {
	if h.Handler == nil {
		panic("contextHandler - handler is nil")
	}

	record.AddAttrs(h.addTraceFromContext(ctx)...)

	if err := h.Handler.Handle(ctx, record); err != nil {
		return fmt.Errorf("contextHandler - failed to handle log record: %w", err)
	}

	return nil
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.Handler == nil {
		panic("contextHandler - handler is nil")
	}

	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	if h.Handler == nil {
		panic("contextHandler - handler is nil")
	}

	return contextHandler{h.Handler.WithGroup(name)}
}

func (h contextHandler) addTraceFromContext(ctx context.Context) []slog.Attr {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()

	var attrs []slog.Attr

	// if span is not valid, return without adding trace attributes
	if spanCtx.IsValid() {
		if spanCtx.HasTraceID() {
			attrs = append(attrs, slog.String(otelTraceIDKey, spanCtx.TraceID().String()))
		}

		if spanCtx.HasSpanID() {
			attrs = append(attrs, slog.String(otelSpanIDKey, spanCtx.SpanID().String()))
		}
	}

	return attrs
}
