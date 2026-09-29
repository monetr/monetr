package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	PatchUserSchema = validation.Map(
		validation.Key("linkOrder",
			validation.Each(
				ValidID[models.Link](),
				validation.Required,
			),
			validation.Length(0, 100),
			validators.Unique[string](),
		).Required(Optional),
	)
)
