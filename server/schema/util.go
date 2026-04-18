package schema

import (
	"time"

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
