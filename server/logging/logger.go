package logging

import (
	"log/slog"
	"os"
	"strings"

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
		if configuration.StackDriver.Enabled {
			opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
				// Stackdriver expects the message field to be named "message".
				if len(groups) == 0 && a.Key == slog.MessageKey {
					a.Key = "message"
				}
				return a
			}
		}
		inner = slog.NewJSONHandler(os.Stderr, opts)
	default: // "text"
		inner = slog.NewTextHandler(os.Stderr, opts)
	}

	if configuration.StackDriver.Enabled {
		inner = NewStackDriverHandler(inner)
	}

	inner = NewContextHandler(inner)

	return slog.New(inner)
}

func NewLogger() *slog.Logger {
	return NewLoggerWithLevel("info")
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
