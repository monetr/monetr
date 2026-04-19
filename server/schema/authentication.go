package schema

import (
	"regexp"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/pkgs/internals"
	"github.com/monetr/monetr/server/captcha"
)

var (
	AuthenticationLogin = z.Struct(z.Shape{
		"email":    EmailAddress().Required(),
		"password": Password().Required(),
	})

	AuthenticationRegister = z.Struct(z.Shape{
		"email":     EmailAddress().Required(),
		"password":  Password().Required(),
		"firstName": z.String().Required().Trim().Max(250),
		"lastName":  z.String().Optional().Trim().Max(250),
		"timezone":  Timezone().Required(),
		"locale":    Locale().Required(),
	})

	AuthenticationTOTP = z.Struct(z.Shape{
		"totp": z.String().
			Trim().
			Len(6).
			Required().
			Match(regexp.MustCompile(`\d{6}`)),
	})

	AuthenticationVerifyEmail = z.Struct(z.Shape{
		"token": z.String().Trim().Required().Max(2000),
	})

	AuthenticationResendVerifyEmail = z.Struct(z.Shape{
		"email": EmailAddress().Required(),
	})

	BetaCode = z.Struct(z.Shape{
		"betaCode": z.Ptr(
			z.String().
				Required(z.Message("beta code is required")).
				Max(100),
		).NotNil(z.Message("beta code is required")),
	})
)

// Captcha takes the captcha interface to verify the actual captcha provided as
// part of the schema testing process.
func Captcha(impl captcha.Verification) *z.StructSchema {
	return z.Struct(z.Shape{
		"captcha": z.String().
			Required().
			Max(3000).
			TestFunc(func(val *string, ctx internals.Ctx) bool {
				context := MustContext(ctx)
				if err := impl.VerifyCaptcha(context, *val); err != nil {
					ctx.AddIssue(ctx.Issue().
						SetCode("captcha_missing").
						SetPath([]string{"captcha"}).
						SetError(err).
						SetMessage("ReCAPTCHA is not valid").
						SetParams(map[string]any{
							"captcha": val,
						}),
					)
				}

				return true
			}),
	})
}
