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

//go:embed email_templates/*
var templates embed.FS

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=email.go -package=mockgen -destination=../internal/mockgen/email.go EmailCommunication
type EmailCommunication interface {
	SendEmail(ctx context.Context, email Email) error
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

func (e *emailCommunicationBase) SendEmail(ctx context.Context, params Email) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	crumbs.AddTag(span.Context(), "email.template", params.Template())

	html, text := e.getTemplates(params.Template())
	m := mail.NewMsg()
	m.Subject(params.Subject())
	first, last := params.Name()
	e.toAddress(m, first, last, params.EmailAddress())
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
		mail.WithUsername(e.config.Email.SMTP.Username),
		mail.WithPassword(e.config.Email.SMTP.Password),
		mail.WithSMTPAuth(SMTPAuth),
		mail.WithTLSPolicy(TLSPolicy),
		mail.WithTimeout(5*time.Second),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create mail client")
	}
	defer c.Close()

	payload.SetDate()
	payload.SetUserAgent(strings.TrimSpace(fmt.Sprintf("monetr %s", build.Release)))

	e.log.WithContext(span.Context()).Debug("sending email message")
	if err := c.DialAndSendWithContext(span.Context(), payload); err != nil {
		return errors.Wrap(err, "failed to send email")
	}

	return nil
}
