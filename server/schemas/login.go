package schemas

import (
	"net/mail"
	"strings"

	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Challenge and Nonce are only present when proof of work is enabled.
	Challenge string `json:"challenge"`
	Nonce     uint64 `json:"nonce"`
}

var (
	LoginSchema = validation.Map(
		validation.Key("email",
			EmailAddress(),
		).Required(Require),
		validation.Key("password",
			Password(),
		).Required(Require),
	)

	LoginChallengeSchema = validation.Map(
		validation.Key("email",
			EmailAddress(),
		).Required(Require),
		validation.Key("password",
			Password(),
		).Required(Require),
		validation.Key("challenge",
			Challenge(),
		).Required(Require),
		validation.Key("nonce",
			Nonce(),
		).Required(Require),
	)
)

func EmailAddress() validation.Rule {
	return validation.AllOf(
		is.EmailFormat.Error("Email address is not valid"),
		is.LowerCase.Error("Email address must be lower case"),
		validation.NewStringRule(func(input string) bool {
			address, err := mail.ParseAddress(input)
			return err == nil && strings.EqualFold(input, address.Address)
		}, "Email address is not valid"),
		validation.Required,
	)
}

func Password() validation.Rule {
	return validation.AllOf(
		is.PrintableUnicode.Error("Password must be printable characters"),
		validation.Length(8, 72).Error("Password must be between 8 and 72 characters"),
		validation.Required,
	)
}
