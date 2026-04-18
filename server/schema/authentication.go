package schema

import (
	"regexp"

	z "github.com/Oudwins/zog"
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

	Captcha = z.Struct(z.Shape{
		"captcha": z.String().Required().Max(250),
	})

	BetaCode = z.Struct(z.Shape{
		"betaCode": z.Ptr(z.String().Required().Max(100)).NotNil(),
	})
)
