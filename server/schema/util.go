package schema

import (
	"context"
	"time"

	"github.com/Oudwins/zog"
	z "github.com/Oudwins/zog"
	"github.com/benbjohnson/clock"
)

func MustTimezone(ctx z.Ctx) *time.Location {
	timezone, ok := ctx.Get("timezone").(*time.Location)
	if !ok {
		panic("timzone is not present on schema context")
	}

	return timezone
}

func MustClock(ctx z.Ctx) clock.Clock {
	clock, ok := ctx.Get("clock").(clock.Clock)
	if !ok {
		panic("clock is not present on schema context")
	}

	return clock
}

func MustContext(ctx zog.Ctx) context.Context {
	context, ok := ctx.Get("context").(context.Context)
	if !ok {
		panic("context.Context is not present on schema context")
	}

	return context
}

func WithContext(ctx context.Context) zog.ExecOption {
	return zog.WithCtxValue("context", ctx)
}
