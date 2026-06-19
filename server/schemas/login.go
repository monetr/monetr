package schemas

import (
	"context"
	"encoding/json"
	"net/mail"
	"strings"

	"github.com/monetr/monetr/server/powchallenge"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
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
			validation.AllOf(
				is.Base32,
				// Our tokens are a fixed size so they always encode to the exact same
				// length. This is just a cheap fail-fast, powchallenge.Verify is what
				// actually proves the token is one we issued.
				validation.Length(powchallenge.EncodedTokenLength, powchallenge.EncodedTokenLength),
				validation.Required,
			).Error("Challenge must be valid"),
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

// Nonce validates the proof of work solution that is submitted alongside a
// challenge. It must be a non negative integer. We deliberately do NOT require
// it to be greater than zero, a nonce of 0 is a perfectly valid solution. The
// solver starts counting at 0 so roughly one in every 2^difficulty challenges
// is solved by 0 and we must not reject it. We still parse it by hand here so
// that a negative value comes back as a clean validation error rather than
// blowing up later when merge tries to fit it into the uint64 field.
func Nonce() validation.Rule {
	return validation.AllOf(
		is.Integer,
		validation.Required,
		validators.By(func(ctx context.Context, value *any) error {
			if value == nil || *value == nil {
				return errors.New("nonce must be a non negative integer")
			}

			// TODO Pretty sure this should be a uint64 and should be handled the same
			// way we do it in the merging code. BUT then i think that makes it hard
			// to handle negative numbers? Does that just manifest as an overflow?
			// Realistically its unlikely we'll get a nonce high enough for the
			// difference between uint64 and int64's upper bounds to matter.
			var nonce int64
			switch value := (*value).(type) {
			case json.Number:
				// When the request comes in through Parse the body is decoded with
				// UseNumber so the nonce lands here as a json.Number.
				parsed, err := value.Int64()
				if err != nil {
					return errors.Wrap(err, "nonce must be a non negative integer")
				}
				nonce = parsed
			case int64:
				nonce = value
			case int:
				nonce = int64(value)
			default:
				return errors.New("nonce must be a non negative integer")
			}

			if nonce < 0 {
				return errors.New("nonce must be a non negative integer")
			}

			return nil
		}),
	).Error("Challenge nonce must be valid")
}
