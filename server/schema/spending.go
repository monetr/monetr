package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

func SpendingType(options ...z.TestOption) *z.StringSchema[models.SpendingType] {
	return z.StringLike[models.SpendingType]().
		OneOf(
			[]models.SpendingType{
				models.SpendingTypeExpense,
				models.SpendingTypeGoal,
			},
			// Merge the options that we default to with the options the caller
			// provides.
			append(
				[]z.TestOption{
					z.IssueCode("invalid_spending_type"),
					z.IssuePath([]string{"spendingType"}),
					z.Message("Spending type must be an expense or a goal"),
				},
				options...,
			)...,
		)
}

var CreateSpendingSchema = z.Struct(z.Shape{
	"bankAccountId":     ID[models.BankAccount](z.IssuePath([]string{"bankAccountId"})).Required(),
	"fundingScheduleId": ID[models.FundingSchedule](z.IssuePath([]string{"fundingScheduleId"})).Required(),
	"spendingType":      SpendingType().Default(models.SpendingTypeExpense).Required(),
	"name":              Name(),
	"description":       z.String().Min(1).Max(300).Optional(),
	"targetAmount":      z.Int64().GT(0).Required(),
	"currentAmount":     z.Int64().GTE(0).Default(0).Optional(),
	"ruleSet":           z.Ptr(RuleSet()),
	"nextRecurrence":    z.Time().Required(),
	"isPaused":          z.Bool().Default(false).Required(),
})

var PatchSpendingSchema = z.Struct(z.Shape{
	"fundingScheduleId": ID[models.FundingSchedule](z.IssuePath([]string{"fundingScheduleId"})).Optional(),
	"name":              Name().Optional(),
	"description":       z.String().Min(1).Max(300).Optional(),
	"targetAmount":      z.Int64().GT(0).Optional(),
	"ruleSet":           RuleSet(), // Needs to be called ruleSet
	"nextRecurrence":    z.Time().Optional(),
	"isPaused":          z.Bool().Default(false).Optional(),
})
