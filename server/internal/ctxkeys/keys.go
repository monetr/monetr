package ctxkeys

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
)

type MonetrContextKey string

const (
	RequestID MonetrContextKey = "requestId"
	UserID    MonetrContextKey = "userId"
	AccountID MonetrContextKey = "accountId"
	LoginID   MonetrContextKey = "loginId"
	JobID     MonetrContextKey = "jobId"
)

var (
	keys = []MonetrContextKey{
		RequestID,
		UserID,
		AccountID,
		LoginID,
		JobID,
	}
)

// SlogAttrsFromContext extracts known fields from the Go context in order to
// enrich log records. This allows monetr to automatically include useful
// context (request ID, user ID, etc.) on every log entry.
func SlogAttrsFromContext(ctx context.Context) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(keys)+1)
	for _, key := range keys {
		if value := ctx.Value(key); value != nil {
			attrs = append(attrs, slog.Any(string(key), value))
		}
	}

	if span := sentry.SpanFromContext(ctx); span != nil {
		attrs = append(attrs, slog.Group("sentry",
			slog.Any("traceId", span.TraceID),
			slog.Any("spanId", span.SpanID),
			slog.Any("parentSpanId", span.ParentSpanID),
		))
	}

	return attrs
}
