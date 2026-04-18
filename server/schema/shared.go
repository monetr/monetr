package schema

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/util"
)

func Name() *z.StringSchema[string] {
	return z.String().Min(1).Max(300).Required().Trim()
}

func Description() *z.StringSchema[string] {
	return z.String().Min(1).Max(300).Optional().Trim()
}

func isMidnight(val *time.Time, ctx z.Ctx) bool {
	tz := MustTimezone(ctx)
	truncated := util.Midnight(*val, tz)
	if !truncated.Equal(*val) {
		ctx.AddIssue(ctx.Issue().
			SetCode("invalid_date").
			SetMessage("must be at midnight in the user's timezone").
			SetParams(map[string]any{
				"input":    val,
				"timezone": tz,
			}),
		)
		return false
	}

	return true
}

func isFuture(val *time.Time, ctx z.Ctx) bool {
	now := MustClock(ctx).Now()
	if val.Before(now) {
		ctx.AddIssue(ctx.Issue().
			SetCode("invalid_date").
			SetMessage("must be in the future").
			SetParams(map[string]any{
				"input": val,
				"now":   now,
			}),
		)
		return false
	}

	return true
}

func Date() *z.TimeSchema {
	return z.Time(z.Time.Format(time.RFC3339)).
		TestFunc(isMidnight)
}

// FutureDate is the same as [Date] but requires that the timestamp by greater
// than the current timestamp relative to the clock on the context.
func FutureDate() *z.TimeSchema {
	return z.Time(z.Time.Format(time.RFC3339)).
		TestFunc(isMidnight).
		TestFunc(isFuture)
}
