package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
	"github.com/teambition/rrule-go"
)

func RuleSet() *z.PreprocessSchema[string, models.RuleSet] {
	return z.Preprocess[string, models.RuleSet](
		func(data string, ctx z.Ctx) (out models.RuleSet, err error) {
			ruleset, err := models.NewRuleSet(data)
			if err != nil {
				return models.RuleSet{},
					ctx.Issue().
						SetPath([]string{"ruleset"}).
						SetCode("invalid_ruleset").
						SetMessage("Invalid recurrence ruleset provided").
						SetError(err).
						SetValue(data)
			}

			if ruleset.GetDTStart().IsZero() {
				ctx.AddIssue(
					ctx.Issue().
						SetCode("ruleset_missing_dtstart").
						SetPath([]string{"ruleset"}).
						SetMessage("DTSTART is required for recurrence rulesets"),
				)
			}

			switch ruleset.GetRRule().Options.Freq {
			case rrule.HOURLY, rrule.MINUTELY, rrule.SECONDLY:
				ctx.AddIssue(
					ctx.Issue().
						SetCode("ruleset_bad_frequency").
						SetPath([]string{"ruleset"}).
						SetMessage("FREQ must not be more frequent than daily"),
				)
			}

			return *ruleset, nil
		},
		z.CustomFunc[models.RuleSet](func(ruleset *models.RuleSet, ctx z.Ctx) bool {
			return true
		}),
	)
}
