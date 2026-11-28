package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/teambition/rrule-go"
)

func Ruleset() validation.Rule {
	return validation.AllOf(
		is.String,
		validation.NewStringRule(func(input string) bool {
			ruleset, err := models.NewRuleSet(input)
			if err != nil {
				return false
			}

			// We require that the caller always specify a DTSTART for these rulesets.
			// The rrule library will happily parse a rule that has no DTSTART but then
			// it falls back to a zero value start time, and all of our relative date
			// math downstream goes sideways when that happens. So if there is not an
			// explicit start we reject the whole thing and make them be specific about
			// when the schedule actually begins.
			if ruleset.GetDTStart().IsZero() {
				return false
			}

			// All of these are just a denial of service waiting to happen. rrule-go
			// walks every occurrence from DTSTART, so anything more often than daily
			// can mean millions of occurrences per call.
			switch ruleset.GetRRule().Options.Freq {
			case rrule.HOURLY, rrule.MINUTELY, rrule.SECONDLY:
				return false
			}

			if len(ruleset.GetRRule().Options.Byhour) > 0 {
				return false
			}

			if len(ruleset.GetRRule().Options.Byminute) > 0 {
				return false
			}

			if len(ruleset.GetRRule().Options.Bysecond) > 0 {
				return false
			}

			return true
		}, "Ruleset must be valid"),
	)
}
