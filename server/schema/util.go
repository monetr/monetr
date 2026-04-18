package schema

import (
	"time"

	"github.com/Oudwins/zog/internals"
	"github.com/benbjohnson/clock"
)

func MustTimezone(ctx internals.Ctx) *time.Location {
	timezone, ok := ctx.Get("timezone").(*time.Location)
	if !ok {
		panic("timzone is not present on schema context")
	}

	return timezone
}

func MustClock(ctx internals.Ctx) *clock.Clock {
	clock, ok := ctx.Get("clock").(*clock.Clock)
	if !ok {
		panic("clock is not present on schema context")
	}

	return clock
}
