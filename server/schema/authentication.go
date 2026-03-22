package schema

import z "github.com/Oudwins/zog"

var LoginSchema = z.Struct(z.Shape{
	"email":    z.String().Email().Trim().Required(),
	"password": z.String().Min(8).Trim().Required(),
})

var RegisterSchema = z.Struct(z.Shape{
	"email":     z.String().Email().Trim().Required(),
	"password":  z.String().Min(8).Trim().Required(),
	"firstName": z.String().Min(1).Trim().Required(),
	"lastName":  z.String().Trim(),
	"locale":    z.String().Required(),
	"timezone":  z.String().Required(),
	"betaCode":  z.String(),
})

// TODO Emails should be lower case
// TODO Validate locale
// TODO Validate timezone
var ForgotPasswordSchema = z.Struct(z.Shape{
	"email": z.String().Email().Trim().Required(),
})

var VerifyEmailSchema = z.Struct(z.Shape{
	"token": z.String().Trim().Required(),
})
