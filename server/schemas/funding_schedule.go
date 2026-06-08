package schemas

import (
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	CreateFundingSchedule = validation.Map(
		validators.Name(validators.Require),
		validators.Description(),
		validation.Key(
			"ruleset",
			validation.Required.Error("Ruleset must be specified for funding schedules"),
			validation.NewStringRule(func(input string) bool {
				_, err := models.NewRuleSet(input)
				return err == nil
			}, "Ruleset must be valid"),
		).Required(validators.Require),
		validation.Key(
			"excludeWeekends",
			validation.In(true, false).Error("Exclude weekends must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key(
			"autoCreateTransaction",
			validation.In(true, false).Error("Auto create transaction must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key(
			"estimatedDeposit",
			PositiveAmount("Estimated deposit"),
			validation.Min(float64(0)).Error("Estimated deposit cannot be less than 0"),
		).Required(validators.Optional),
		validation.Key(
			"nextRecurrence",
			validation.Date(time.RFC3339).Error("Next recurrence must be in the future"),
		).Required(validators.Optional),
	)
)
