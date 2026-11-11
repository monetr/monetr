package schemas

import (
	"github.com/monetr/validation"
)

var (
	IconSearch = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
	)
)
