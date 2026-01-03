package cmd

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    string
		isPretty bool
	}{
		{
			name:     "pretty logger",
			level:    "info",
			isPretty: true,
		},
		{
			name:     "json logger",
			level:    "debug",
			isPretty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testWriter := bytes.NewBuffer(nil)

			logger := NewLogger(testWriter, tt.level, tt.isPretty)
			logger.InfoContext(t.Context(), "test log message")

			out := make([]byte, 1024)
			n, _ := testWriter.Read(out)
			output := string(out[:n])

			if !strings.Contains(output, "test log message") {
				t.Errorf("expected log output to contain message, got: %s", output)
			}
		})
	}
}

func TestTraceHandler_Handle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{AddSource: false})
	th := contextHandler{handler}

	ctx := t.Context()
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, `test message`, 0)

	// No span in context
	if err := th.Handle(ctx, rec); err != nil {
		t.Errorf("Handle() without span returned error: %v", err)
	}

	if got := buf.String(); !strings.Contains(got, `,"level":"INFO","msg":"test message"}`) {
		t.Errorf("expected log output to contain message, got: %s", got)
	}

	buf.Reset()

	// With valid span in context
	tracer := sdktrace.NewTracerProvider().Tracer("test-tracer")
	_, span := tracer.Start(ctx, "test-span")
	ctxWithSpan := trace.ContextWithSpan(ctx, span)

	rec2 := slog.NewRecord(time.Now(), slog.LevelInfo, "span message", 0)

	if err := th.Handle(ctxWithSpan, rec2); err != nil {
		t.Errorf("Handle() with span returned error: %v", err)
	}

	out := buf.String()

	// {"time":"2025-06-24T22:58:34.414187+08:00","level":"INFO","msg":"span message","trace_id":"837a71ca7cef495adca48a4c3452513b","span_id":"e34fd70ca53c1475"}
	// test if the output contains the expected fields
	if !strings.Contains(out, `"level":"INFO"`) || !strings.Contains(out, `"msg":"span message"`) {
		t.Errorf("expected log output to contain message, got: %s", out)
	}

	if !strings.Contains(out, `"trace_id":"`) || !strings.Contains(out, `"span_id":"`) {
		t.Errorf("expected log output to contain trace and span IDs, got: %s", out)
	}

	span.End()

	sdktrace.NewTracerProvider().Shutdown(ctx) // clean up tracer provider
}
