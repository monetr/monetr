package email_templates

import (
	"embed"
	"html/template"

	"github.com/pkg/errors"
)

const (
	VerifyEmailTemplate     = "templates/verify.html"
	ForgotPasswordTemplate  = "templates/forgot.html"
	PasswordChangedTemplate = "templates/password_changed.html"
)

//go:embed templates/*.html
var templates embed.FS

func GetEmailTemplate(name string) (*template.Template, error) {
	data, err := templates.ReadFile(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open email template (%s)", name)
	}

	emailTemplate := template.New(name)
	emailTemplate, err = emailTemplate.Parse(string(data))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse email template (%s)", name)
	}

	return emailTemplate, nil
}
