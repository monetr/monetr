package schemas

import "github.com/monetr/validation"

var (
	CreateApiKey = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
	)
)
