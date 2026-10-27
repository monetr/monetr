package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
)

var (
	ForecastInitialFunding = validation.Map(
		validation.Key("transactionRecurringId",
			ValidID[models.TransactionRecurring](),
			validation.Required,
		).Required(Require),
		validation.Key("fundingScheduleId",
			ValidID[models.FundingSchedule](),
			validation.Required,
		).Required(Require),
	)
)
