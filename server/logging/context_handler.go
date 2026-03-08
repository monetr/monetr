package logging

import (
	"context"
	"log/slog"

	"github.com/monetr/monetr/server/internal/ctxkeys"
)

var _ slog.Handler = &contextHandler{}

type contextHandler struct {
	inner slog.Handler
}

func NewContextHandler(inner slog.Handler) slog.Handler {
	return &contextHandler{inner: inner}
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		r.AddAttrs(ctxkeys.SlogAttrsFromContext(ctx)...)
	}
	return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{inner: h.inner.WithGroup(name)}
}
