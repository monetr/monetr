package schema

import (
	"github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateFundingSchedule = zog.Struct(zog.Shape{
		"name":                  Name(),
		"description":           Description(),
		"ruleset":               zog.Ptr(RRule()).NotNil(),
		"excludeWeekends":       zog.Bool().Default(false).Required(),
		"estimatedDeposit":      zog.Ptr(zog.Int64().GT(0).Optional()),
		"nextRecurrence":        Date().Required(),
		"autoCreateTransaction": zog.Bool().Default(false).Required(),
	})

	PatchFundingSchedule = zog.Struct(zog.Shape{
		"name":                  Name().Optional(),
		"description":           Description().Optional(),
		"ruleset":               zog.Ptr(RRule()),
		"excludeWeekends":       zog.Bool().Optional(),
		"estimatedDeposit":      zog.Ptr(zog.Int64().GT(0).Optional()),
		"nextRecurrence":        Date().Optional(),
		"autoCreateTransaction": zog.Bool().Optional(),
	})
)

func isValidFundingSchedule(val any, ctx zog.Ctx) bool {
	switch val := val.(type) {
	case *models.FundingSchedule:
		if val.AutoCreateTransaction &&
			(val.EstimatedDeposit == nil || *val.EstimatedDeposit <= 0) {
			ctx.AddIssue(ctx.Issue().
				SetCode("invalid_autoCreateTransaction").
				SetPath([]string{"autoCreateTransaction"}).
				SetMessage("auto create transaction requires that an estimated deposit is specified").
				SetParams(map[string]any{
					"estimatedDeposit":      val.EstimatedDeposit,
					"autoCreateTransaction": val.AutoCreateTransaction,
				}),
			)
		}
	}

	return true
}
