package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateSpending = z.Struct(z.Shape{
		"name":              Name(),
		"description":       Description(),
		"fundingScheduleId": ID[models.FundingSchedule]().Required(),
		"spendingType": z.StringLike[models.SpendingType]().
			OneOf([]models.SpendingType{
				models.SpendingTypeExpense,
				models.SpendingTypeGoal,
			}).
			Default(models.SpendingTypeExpense).
			Required(),
		"ruleset":        z.Ptr(RRule()),
		"targetAmount":   z.Int64().GT(0).Required(),
		"currentAmount":  z.Int64().GTE(0).Default(0).Required(),
		"usedAmount":     z.Int64().GTE(0).Default(0).Required(),
		"nextRecurrence": Date().Required(),
		"isPaused":       z.Bool().Default(false).Optional(),
	}).Test(z.Test[any]{Func: validateSpending})

	PatchSpending = z.Struct(z.Shape{
		"name":              Name().Optional(),
		"description":       Description().Optional(),
		"fundingScheduleId": ID[models.FundingSchedule]().Optional(),
		"ruleset":           z.Ptr(RRule()),
		"targetAmount":      z.Int64().GT(0).Optional(),
		"currentAmount":     z.Int64().GTE(0).Optional(),
		"usedAmount":        z.Int64().GTE(0).Optional(),
		"nextRecurrence":    Date().Optional(),
		"isPaused":          z.Bool().Optional(),
	}).Test(z.Test[any]{Func: validateSpending})
)

// validateSpending enforces basic spending rules that are not specific to a
// single field. In this case it enforces that expenses must have a ruleset and
// goals must not.
func validateSpending(val any, ctx z.Ctx) {
	switch val := val.(type) {
	case *models.Spending:
		switch val.SpendingType {
		case models.SpendingTypeExpense:
			if val.Ruleset == nil {
				ctx.AddIssue(ctx.Issue().
					SetCode("rulset_required").
					SetPath([]string{"ruleset"}).
					SetMessage("expenses must have a ruleset").
					SetParams(map[string]any{
						"spendingType": val.SpendingType,
						"ruleset":      nil,
					}),
				)
			}
		case models.SpendingTypeGoal:
			if val.Ruleset != nil {
				ctx.AddIssue(ctx.Issue().
					SetCode("rulset_not_allowed").
					SetPath([]string{"ruleset"}).
					SetMessage("goals cannot have a ruleset").
					SetParams(map[string]any{
						"spendingType": val.SpendingType,
						"ruleset":      val.Ruleset,
					}),
				)
			}
		}
	}
}
