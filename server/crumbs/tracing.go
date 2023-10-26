package crumbs

import (
	"context"
	"runtime"
	"strings"

	"github.com/getsentry/sentry-go"
)

func StartFnTrace(ctx context.Context) *sentry.Span {
	span := sentry.StartSpan(ctx, "function")
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		span.Description = strings.TrimPrefix(details.Name(), "github.com/monetr/monetr/server/")
	}

	return span
}
