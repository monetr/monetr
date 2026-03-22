package schema

import (
	"github.com/Oudwins/zog"
	p "github.com/Oudwins/zog/internals"
	"github.com/benbjohnson/clock"
)

func WithClock(clock clock.Clock) zog.ExecOption {
	return func(ctx *p.ExecCtx) {
		ctx.Set("clock", clock)
	}
}

func GetClock(ctx zog.Ctx) clock.Clock {
	return ctx.Get("clock").(clock.Clock)
}
