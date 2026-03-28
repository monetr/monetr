package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/monetr/monetr/server/config"
)

func NewLoggerWithConfig(configuration config.Logging) *slog.Logger {
	level := parseLevel(configuration.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	var inner slog.Handler
	switch strings.ToLower(configuration.Format) {
	case "json":
		inner = slog.NewJSONHandler(os.Stderr, opts)
	default: // "text"
		inner = devslog.NewHandler(os.Stderr, &devslog.Options{
			HandlerOptions:      opts,
			MaxSlicePrintSize:   0,
			SortKeys:            true,
			TimeFormat:          time.RFC3339,
			NewLineAfterLog:     false,
			StringIndentation:   true,
			StringerFormatter:   true,
			NoColor:             false,
			SameSourceInfoColor: false,
		})
	}

	inner = NewContextHandler(inner)

	return slog.New(inner)
}

func NewLoggerWithLevel(levelString string) *slog.Logger {
	return NewLoggerWithConfig(config.Logging{
		Level: levelString,
	})
}

func parseLevel(levelString string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelString)) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
