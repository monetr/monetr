package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var CreateFundingScheduleSchema = z.Struct(z.Shape{
	"bankAccountId":    ID[models.BankAccount](z.IssuePath([]string{"bankAccountId"})).Required(),
	"name":             Name().Required(),
	"description":      z.String().Min(1).Max(300).Optional(),
	"ruleSet":          RuleSet(),
	"excludeWeekends":  z.Bool().Default(false).Required(),
	"estimatedDeposit": z.Int64().GTE(0).Optional(),
	"nextRecurrence":   z.Time().Required(),
})

var PatchFundingScheduleSchema = z.Struct(z.Shape{
	"name":             Name().Optional(),
	"description":      z.String().Min(1).Max(300).Optional(),
	"ruleSet":          z.Ptr(RuleSet()),
	"excludeWeekends":  z.Bool().Default(false).Optional(),
	"estimatedDeposit": z.Int64().GTE(0).Optional(),
	"nextRecurrence":   z.Time().Optional(),
})
