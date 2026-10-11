package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	CreateLink = validation.Map(
		validation.Key("institutionName",
			Name(),
		).Required(Require),
		validation.Key("description",
			validation.OneOf(
				validation.Nil,
				TextField(),
			),
		).Required(validators.Optional),
		validation.Key("lunchFlowLinkId",
			validation.OneOf(
				validation.Nil,
				ValidID[models.LunchFlowLink](),
			),
		).Required(validators.Optional),
	)

	PatchLink = validation.Map(
		validation.Key("institutionName",
			Name(),
		).Required(validators.Optional),
		validation.Key("description",
			validation.OneOf(
				validation.Nil,
				TextField(),
			),
		).Required(validators.Optional),
		validation.Key("bankAccountOrder",
			validation.Each(
				ValidID[models.BankAccount](),
				validation.Required,
			),
			validation.Length(0, 100),
			validators.Unique[string](),
		).Required(validators.Optional),
	)
)
