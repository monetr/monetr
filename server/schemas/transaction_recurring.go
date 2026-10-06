package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
)

var (
	PatchTransactionRecurring = validation.Map(
		validation.Key("spendingId",
			validation.OneOf(
				validation.Nil.Error("must be nil"),
				ValidID[models.Spending](),
			),
		).Required(Optional),
		validation.Key("fundingScheduleId",
			validation.OneOf(
				validation.Nil.Error("must be nil"),
				ValidID[models.FundingSchedule](),
			),
		).Required(Optional),
	)
)
