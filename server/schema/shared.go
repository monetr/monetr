package schema

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/internals"
	"github.com/monetr/monetr/server/util"
)

func Name() *z.StringSchema[string] {
	return z.String().Min(1).Max(300).Required().Trim()
}

func Description() *z.StringSchema[string] {
	return z.String().Min(1).Max(300).Optional().Trim()
}

func Date() *z.TimeSchema {
	return z.Time(z.Time.Format(time.RFC3339)).
		TestFunc(func(val *time.Time, ctx internals.Ctx) bool {
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
			}

			return true
		})
}
