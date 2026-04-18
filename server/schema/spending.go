package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/internals"
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
	}).Test(z.Test[any]{
		Func: func(val any, ctx internals.Ctx) {
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
		},
	})
)
