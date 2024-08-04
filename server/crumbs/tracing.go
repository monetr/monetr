package crumbs

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/getsentry/sentry-go"
)

func StartFnTrace(ctx context.Context) *sentry.Span {
	span := sentry.StartSpan(ctx, "function")
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		name := details.Name()
		span.Description = strings.TrimPrefix(name, "github.com/monetr/monetr/server/")
		file, line := details.FileLine(pc)
		span.SetTag("code.filepath", file)
		span.SetTag("code.lineno", fmt.Sprint(line))
		span.SetTag("code.function", name)
	}

	return span
}
