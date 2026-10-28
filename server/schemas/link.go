package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	// CreateLink is a union of the three kinds of links that can be created
	// directly. A key is forbidden in a variant by leaving it out, so a manual
	// link can't have either upstream link Id and you can never send both.
	CreateLink = validation.OneOf(
		// Manual link schema
		validation.Map(
			validation.Key("institutionName",
				Name(),
			).Required(Require),
			validation.Key("description",
				validation.OneOf(
					validation.Nil,
					TextField(),
				),
			).Required(validators.Optional),
		),
		// Lunch Flow link schema
		validation.Map(
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
				ValidID[models.LunchFlowLink](),
				validation.Required,
			).Required(Require),
		),
		// SimpleFIN link schema
		validation.Map(
			validation.Key("institutionName",
				Name(),
			).Required(Require),
			validation.Key("description",
				validation.OneOf(
					validation.Nil,
					TextField(),
				),
			).Required(validators.Optional),
			validation.Key("simpleFinLinkId",
				ValidID[models.SimpleFinLink](),
				validation.Required,
			).Required(Require),
		),
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
