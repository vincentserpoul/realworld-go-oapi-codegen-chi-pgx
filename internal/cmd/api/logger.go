package api

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func NewLogger(level string, isPretty bool) *slog.Logger {
	// log level
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelError
	}

	// create a new logger
	var handler slog.Handler

	if isPretty {
		handler = tint.NewHandler(
			os.Stderr,
			&tint.Options{
				AddSource:  true,
				Level:      lvl,
				TimeFormat: time.RFC3339,
			},
		)
	} else {
		handler = slog.NewJSONHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     lvl,
			},
		)
	}

	logger := slog.New(handler)

	// set global logger with custom options
	slog.SetDefault(logger)

	return logger
}
