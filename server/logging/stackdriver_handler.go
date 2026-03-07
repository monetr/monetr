package logging

import (
	"context"
	"fmt"
	"log/slog"
)

var (
	_ slog.Handler = &stackDriverHandler{}

	stackdriverFieldLabels = []string{
		"accountId",
		"userId",
		"loginId",
		"requestId",
		"jobId",
	}

	// Stackdriver log levels documented here:
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
	levelsToStackdriver = map[slog.Level]string{
		LevelTrace:      "DEBUG",
		slog.LevelDebug: "INFO",
		slog.LevelInfo:  "NOTICE",
		slog.LevelWarn:  "WARNING",
		slog.LevelError: "ERROR",
	}
)

type stackDriverHandler struct {
	inner slog.Handler
}

func NewStackDriverHandler(inner slog.Handler) slog.Handler {
	return &stackDriverHandler{inner: inner}
}

func (h *stackDriverHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *stackDriverHandler) Handle(ctx context.Context, r slog.Record) error {
	severity, ok := levelsToStackdriver[r.Level]
	if !ok {
		severity = "DEFAULT"
	}
	r.AddAttrs(slog.String("severity", severity))

	// Promote well-known fields into the Stackdriver labels map.
	// https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
	labels := map[string]string{}
	r.Attrs(func(a slog.Attr) bool {
		for _, label := range stackdriverFieldLabels {
			if a.Key == label {
				labels[label] = fmt.Sprint(a.Value.Any())
			}
		}
		return true
	})

	if len(labels) > 0 {
		labelAttrs := make([]any, 0, len(labels)*2)
		for k, v := range labels {
			labelAttrs = append(labelAttrs, k, v)
		}
		r.AddAttrs(slog.Group("logging.googleapis.com/labels", labelAttrs...))
	}

	return h.inner.Handle(ctx, r)
}

func (h *stackDriverHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &stackDriverHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *stackDriverHandler) WithGroup(name string) slog.Handler {
	return &stackDriverHandler{inner: h.inner.WithGroup(name)}
}
