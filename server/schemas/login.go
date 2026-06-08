package schemas

import (
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

var (
	LoginSchema = validation.Map(
		validation.Key("email",
			is.EmailFormat.Error("Email address is not valid"),
			validation.Required,
		),
		validation.Key("password",
			is.PrintableUnicode.Error("Password must be printable characters"),
			validation.Length(8, 72).Error("Password must be between 8 and 72 characters"),
			validation.Required,
		),
	)
)
