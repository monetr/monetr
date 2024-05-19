package communication

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	htmlTemplate "html/template"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

const (
	VerifyEmailTemplate     = "VerifyEmailAddress"
	ForgotPasswordTemplate  = "ForgotPassword"
	PasswordChangedTemplate = "PasswordChanged"
)

//go:embed email_templates/*
var templates embed.FS

type (
	VerifyEmailParams struct {
		BaseURL      string
		Email        string
		FirstName    string
		LastName     string
		SupportEmail string
		VerifyURL    string
	}

	PasswordResetParams struct {
		BaseURL      string
		Email        string
		FirstName    string
		LastName     string
		SupportEmail string
		ResetURL     string
	}

	PasswordChangedParams struct {
		BaseURL      string
		Email        string
		FirstName    string
		LastName     string
		SupportEmail string
	}
)

//go:generate go run go.uber.org/mock/mockgen@v0.4.0 -source=email.go -package=mockgen -destination=../internal/mockgen/email.go EmailCommunication
type EmailCommunication interface {
	SendVerification(ctx context.Context, params VerifyEmailParams) error
	SendPasswordReset(ctx context.Context, params PasswordResetParams) error
	SendPasswordChanged(ctx context.Context, params PasswordChangedParams) error
}

func NewEmailCommunication(log *logrus.Entry, configuration config.Configuration) EmailCommunication {
	return &emailCommunicationBase{
		log:    log,
		config: configuration,
	}
}

type emailCommunicationBase struct {
	log    *logrus.Entry
	config config.Configuration
}

func (e *emailCommunicationBase) fromAddress(msg *mail.Msg) error {
	return msg.FromFormat("monetr", fmt.Sprintf("no-reply@%s", e.config.Email.Domain))
}

func (e *emailCommunicationBase) toAddress(msg *mail.Msg, firstName, lastName, emailAddress string) error {
	// Clean up the things that I **know** will cause problems with SMTP.
	firstName = strings.ReplaceAll(firstName, "\n", "")
	firstName = strings.ReplaceAll(firstName, "\r", "")
	lastName = strings.ReplaceAll(lastName, "\n", "")
	lastName = strings.ReplaceAll(lastName, "\r", "")

	name := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))

	return msg.AddToFormat(name, emailAddress)
}

func (e *emailCommunicationBase) SendVerification(ctx context.Context, params VerifyEmailParams) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	html, text := e.getTemplates(VerifyEmailTemplate)
	m := mail.NewMsg()
	m.Subject("Verify Your Email Address")
	e.toAddress(m, params.FirstName, params.LastName, params.Email)
	e.fromAddress(m)
	m.SetBodyTextTemplate(text, params)
	m.AddAlternativeHTMLTemplate(html, params)

	return e.sendMessage(span.Context(), m)
}

func (e *emailCommunicationBase) SendPasswordReset(ctx context.Context, params PasswordResetParams) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	html, text := e.getTemplates(ForgotPasswordTemplate)
	m := mail.NewMsg()
	m.Subject("Reset Your Password")
	e.toAddress(m, params.FirstName, params.LastName, params.Email)
	e.fromAddress(m)
	m.SetBodyTextTemplate(text, params)
	m.AddAlternativeHTMLTemplate(html, params)

	return e.sendMessage(span.Context(), m)
}

func (e *emailCommunicationBase) SendPasswordChanged(ctx context.Context, params PasswordChangedParams) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	html, text := e.getTemplates(PasswordChangedTemplate)
	m := mail.NewMsg()
	m.Subject("Password Updated")
	e.toAddress(m, params.FirstName, params.LastName, params.Email)
	e.fromAddress(m)
	m.SetBodyTextTemplate(text, params)
	m.AddAlternativeHTMLTemplate(html, params)

	return e.sendMessage(span.Context(), m)
}

func (e *emailCommunicationBase) getTemplates(name string) (html *htmlTemplate.Template, text *textTemplate.Template) {
	{ // HTML template
		data, err := templates.ReadFile(fmt.Sprintf("email_templates/%s.html", name))
		if err != nil {
			panic(fmt.Sprintf("failed to load embedded email template %s: %+v", name, err))
		}
		data = bytes.Join(bytes.Split(data, []byte("\n")), []byte{})

		html = htmlTemplate.New(name)
		html, err = html.Parse(string(data))
		if err != nil {
			panic(fmt.Sprintf("failed to parse embedded email template %s: %+v", name, err))
		}
	}

	{ // Plain text template
		data, err := templates.ReadFile(fmt.Sprintf("email_templates/%s.txt", name))
		if err != nil {
			panic(fmt.Sprintf("failed to load embedded email template %s: %+v", name, err))
		}

		text = textTemplate.New(name)
		text, err = text.Parse(string(data))
		if err != nil {
			panic(fmt.Sprintf("failed to parse embedded email template %s: %+v", name, err))
		}
	}

	return html, text
}

type messagePayload struct {
	From        string
	To          string
	Subject     string
	HTMLContent string
	TextContent string
}

func (e *emailCommunicationBase) sendMessage(ctx context.Context, payload *mail.Msg) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	c, err := mail.NewClient(
		e.config.Email.SMTP.Host,
		mail.WithPort(e.config.Email.SMTP.Port),
		mail.WithSMTPAuthCustom(PlainAuth(
			e.config.Email.SMTP.Identity,
			e.config.Email.SMTP.Username,
			e.config.Email.SMTP.Password,
			e.config.Email.SMTP.Host,
		)),
		mail.WithTimeout(5*time.Second),
		mail.WithTLSPolicy(TLSPolicy),
		// Move to this once we are no longer using mailhog? It overwrites the port
		// in a weird way.
		// mail.WithTLSPortPolicy(TLSPolicy),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create mail client")
	}
	defer c.Close()

	payload.SetDate()
	payload.SetUserAgent(strings.TrimSpace(fmt.Sprintf("monetr %s", build.Release)))

	if err := c.DialAndSendWithContext(span.Context(), payload); err != nil {
		return errors.Wrap(err, "failed to send email")
	}

	return nil
}
