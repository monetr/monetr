package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/zconst"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"
)

const (
	IssueCodeRRuleInvalid zconst.ZogIssueCode = "invalid_rrule"
)

func RRule() *z.PreprocessSchema[string, models.RuleSet] {
	return z.Preprocess(
		func(raw string, ctx z.Ctx) (models.RuleSet, error) {
			set, err := models.NewRuleSet(raw)
			if err != nil {
				return models.RuleSet{}, ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("invalid RRule").
					SetError(errors.WithStack(err)).
					SetParams(map[string]any{
						"ruleset": raw,
					})
			}

			if set.GetDTStart().IsZero() {
				ctx.AddIssue(ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("DTSTART is required on rulesets").
					SetParams(map[string]any{
						"ruleset": raw,
					}))
			}

			switch set.GetRRule().OrigOptions.Freq {
			case rrule.DAILY, rrule.WEEKLY, rrule.MONTHLY, rrule.YEARLY:
			default:
				ctx.AddIssue(ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("FREQ must be one of DAILY, WEEKLY, MONTHLY, YEARLY").
					SetParams(map[string]any{
						"ruleset": raw,
					}))
			}

			if set.GetRRule().OrigOptions.Byhour != nil {
				ctx.AddIssue(ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("BYHOUR is not supported").
					SetParams(map[string]any{
						"ruleset": raw,
					}))
			}

			if set.GetRRule().OrigOptions.Byminute != nil {
				ctx.AddIssue(ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("BYMINUTE is not supported").
					SetParams(map[string]any{
						"ruleset": raw,
					}))
			}

			if set.GetRRule().OrigOptions.Bysecond != nil {
				ctx.AddIssue(ctx.Issue().
					SetPath([]string{"ruleset"}).
					SetCode(IssueCodeRRuleInvalid).
					SetMessage("BYSECOND is not supported").
					SetParams(map[string]any{
						"ruleset": raw,
					}))
			}

			return *set, nil
		},
		// Writes the pre-parsed rrule.Set into the destination (*rrule.Set).
		z.CustomFunc[models.RuleSet](func(_ *models.RuleSet, _ z.Ctx) bool {
			return true
		}),
	)
}
