package schemas

import (
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

var (
	CreateFundingSchedule = validation.Map(
		validation.Key("name",
			validation.Required.Error("Name is required"),
			Name(),
		).Required(Require),
		validation.Key("description",
			TextField(),
		).Required(validators.Optional),
		validation.Key("ruleset",
			validation.Required.Error("Ruleset must be specified for funding schedules"),
			Ruleset(),
		).Required(validators.Require),
		validation.Key("excludeWeekends",
			Boolean().Error("Exclude weekends must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key("autoCreateTransaction",
			Boolean().Error("Auto create transaction must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key("estimatedDeposit",
			validation.OneOf(
				validation.Nil,
				validation.AllOf(
					PositiveAmount("Estimated deposit"),
					// TODO This might be redundant?
					validation.Min(float64(0)).Error("Estimated deposit cannot be less than 0"),
				),
			),
		).Required(validators.Optional),
		validation.Key("nextRecurrence",
			Timestamp().Error("Next recurrence must be a valid date"),
		).Required(validators.Optional),
	)

	PatchFundingSchedule = validation.Map(
		validation.Key("name",
			validation.Required.Error("Name is required"),
			Name(),
		).Required(Optional),
		validation.Key("description",
			TextField(),
		).Required(validators.Optional),
		validation.Key("ruleset",
			validation.Required.Error("Ruleset cannot be blank when specified"),
			is.String,
			Ruleset(),
		).Required(validators.Optional),
		validation.Key("excludeWeekends",
			is.Boolean,
			validation.In(true, false).Error("Exclude weekends must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key("autoCreateTransaction",
			is.Boolean,
			validation.In(true, false).Error("Auto create transaction must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key("estimatedDeposit",
			is.Integer,
			// TODO [PositiveAmount] here instead?
			validation.Min(float64(0)).Error("Estimated deposit cannot be less than 0"),
		).Required(validators.Optional),
		validation.Key("nextRecurrence",
			Timestamp().Error("Next recurrence must be a valid date"),
			validation.Required.Error("Next recurrence cannot be blank when specified"),
		).Required(validators.Optional),
	)
)
