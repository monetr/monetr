package ctxkeys

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
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

// LogrusFieldsFromContext will extract known fields from the Go context in
// order to better support logging. This allows monetr to more easily provide
// useful context for all of its log entries for every action that occurs.
func LogrusFieldsFromContext(
	ctx context.Context,
	existingFields logrus.Fields,
) logrus.Fields {
	fields := logrus.Fields{}
	for _, key := range keys {
		// If the field is already on the log entry then do not overwrite it.
		if _, ok := existingFields[string(key)]; ok {
			continue
		}

		if value := ctx.Value(key); value != nil {
			fields[string(key)] = value
		}
	}

	// Add tracing details to our log messages to make it easier to go from a
	// trace in Sentry to logs elsewhere.
	if span := sentry.SpanFromContext(ctx); span != nil {
		fields["sentry"] = logrus.Fields{
			"traceId":      span.TraceID,
			"spanId":       span.SpanID,
			"parentSpanId": span.ParentSpanID,
		}
	}

	return fields
}
