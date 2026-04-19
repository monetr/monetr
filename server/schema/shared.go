package schema

import (
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/util"
	"github.com/monetr/monetr/server/zoneinfo"
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

func EmailAddress() *z.StringSchema[string] {
	return z.String().
		Email(z.Message("email address must be valid")).
		Trim().
		Transform(func(valPtr *string, ctx z.Ctx) error {
			*valPtr = strings.ToLower(*valPtr)
			return nil
		})
}

func Password() *z.StringSchema[string] {
	return z.String().
		Required().
		Trim().
		Min(8, z.Message("password must be at least 8 characters")).
		Max(71, z.Message("password cannot be longer than 71 characters"))
}

func Timezone() *z.StringSchema[string] {
	return z.String().
		Default("UTC").
		TestFunc(func(val *string, ctx z.Ctx) bool {
			_, err := zoneinfo.Timezone(*val)
			if err != nil {
				ctx.AddIssue(ctx.Issue().
					SetCode("timezone_invalid").
					SetPath([]string{"timezone"}).
					SetMessage("timezone not recognized by server").
					SetParams(map[string]any{
						"timezone": val,
					}),
				)
				return false
			}

			return true
		})
}

func Locale() *z.StringSchema[string] {
	return z.String().
		Default(consts.DefaultLocale).
		Transform(func(valPtr *string, ctx z.Ctx) error {
			if _, err := locale.GetLConv(*valPtr); err != nil {
				// If the provided locale is invalid, fallback to the default
				*valPtr = consts.DefaultLocale
			}

			return nil
		})
}

func Currency() *z.StringSchema[string] {
	return z.String().
		OneOf(
			locale.GetInstalledCurrencies(),
			z.Message("currency must be one supported by the server"),
		)
}
