package schema

import z "github.com/Oudwins/zog"

var (
	CreateFundingSchedule = z.Struct(z.Shape{
		"name":             Name(),
		"description":      Description(),
		"ruleset":          z.Ptr(RRule()).NotNil(),
		"excludeWeekends":  z.Bool().Default(false).Required(),
		"estimatedDeposit": z.Ptr(z.Int64().GT(0).Optional()),
		"nextRecurrence":   Date().Required(),
	})

	PatchFundingSchedule = z.Struct(z.Shape{
		"name":             Name().Optional(),
		"description":      Description().Optional(),
		"ruleset":          z.Ptr(RRule()),
		"excludeWeekends":  z.Bool().Optional(),
		"estimatedDeposit": z.Ptr(z.Int64().GT(0).Optional()),
		"nextRecurrence":   Date().Optional(),
	})
)
