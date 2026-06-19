package schemas

import (
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	PatchLink = validation.Map(
		validation.Key(
			"institutionName",
			Name(),
		).Required(validators.Optional),
		validation.Key(
			"description",
			validation.OneOf(
				validation.Nil,
				TextField(),
			),
		).Required(validators.Optional),
	)
)
