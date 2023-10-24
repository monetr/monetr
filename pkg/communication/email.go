package communication

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	VerifyEmailTemplate    = "VerifyEmailAddress"
	ForgotPasswordTemplate = "ForgotPassword"
	// PasswordChangedTemplate = "email_templates/password_changed.html"
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
)

//go:generate mockgen -source=email.go -package=mockgen -destination=../internal/mockgen/email.go EmailCommunication
type EmailCommunication interface {
	SendVerification(ctx context.Context, params VerifyEmailParams) error
	SendPasswordReset(ctx context.Context, params PasswordResetParams) error
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

func (e *emailCommunicationBase) fromAddress() string {
	return fmt.Sprintf("monetr <no-reply@%s>", e.config.Email.Domain)
}

func (e *emailCommunicationBase) toAddress(firstName, lastName, emailAddress string) string {
	// Clean up the things that I **know** will cause problems with SMTP.
	firstName = strings.ReplaceAll(firstName, "\n", "")
	firstName = strings.ReplaceAll(firstName, "\r", "")
	lastName = strings.ReplaceAll(lastName, "\n", "")
	lastName = strings.ReplaceAll(lastName, "\r", "")

	return fmt.Sprintf("%s %s <%s>", firstName, lastName, emailAddress)
}

func (e *emailCommunicationBase) SendVerification(ctx context.Context, params VerifyEmailParams) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	html, text := e.getTemplates(VerifyEmailTemplate)

	htmlBuffer := bytes.NewBuffer(nil)
	if err := html.Execute(htmlBuffer, params); err != nil {
		return errors.Wrap(err, "failed to execute verification email html template")
	}

	textBuffer := bytes.NewBuffer(nil)
	if err := text.Execute(textBuffer, params); err != nil {
		return errors.Wrap(err, "failed to execute verification email text template")
	}

	payload := messagePayload{
		From:        e.fromAddress(),
		To:          e.toAddress(params.FirstName, params.LastName, params.Email),
		Subject:     "Verify Your Email Address",
		HTMLContent: htmlBuffer.String(),
		TextContent: textBuffer.String(),
	}

	return e.sendMessage(span.Context(), payload)
}

func (e *emailCommunicationBase) SendPasswordReset(ctx context.Context, params PasswordResetParams) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	html, text := e.getTemplates(ForgotPasswordTemplate)

	htmlBuffer := bytes.NewBuffer(nil)
	if err := html.Execute(htmlBuffer, params); err != nil {
		return errors.Wrap(err, "failed to execute password reset email html template")
	}

	textBuffer := bytes.NewBuffer(nil)
	if err := text.Execute(textBuffer, params); err != nil {
		return errors.Wrap(err, "failed to execute password reset email text template")
	}

	payload := messagePayload{
		From:        e.fromAddress(),
		To:          e.toAddress(params.FirstName, params.LastName, params.Email),
		Subject:     "Reset Your Password",
		HTMLContent: htmlBuffer.String(),
		TextContent: textBuffer.String(),
	}

	return e.sendMessage(span.Context(), payload)
}

func (e *emailCommunicationBase) getTemplates(name string) (html *template.Template, text *template.Template) {
	{ // HTML template
		data, err := templates.ReadFile(fmt.Sprintf("email_templates/%s.html", name))
		if err != nil {
			panic(fmt.Sprintf("failed to load embedded email template %s: %+v", name, err))
		}

		html = template.New(name)
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

		text = template.New(name)
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

func (e *emailCommunicationBase) sendMessage(ctx context.Context, payload messagePayload) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	builder := bytes.NewBuffer(nil)
	builder.WriteString("To: ")
	builder.WriteString(payload.To)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("From: ")
	builder.WriteString(payload.From)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("Subject: ")
	builder.WriteString(payload.Subject)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("Date: ")
	builder.WriteString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
	builder.WriteRune(';')
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("Reply-To: ")
	// TODO Make this a configurable thing.
	builder.WriteString("monetr Support <support@monetr.app>")
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("MIME-version: 1.0;")
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	if payload.TextContent == "" || payload.HTMLContent == "" {
		panic("text or html content missing, are your templates generated properly?")
	}

	// Since the boundaries are named and I really don't know how SMTP works, I don't want someone to be able to use the
	// actual boundary names in the parts of the template that we are filling in from user input. So I'm going to generate
	// a boundary name every time we send an email. This way there isn't a way that someone can guess the boundary names
	// which could result in some weird injection between the plain text part of the email and the HTML part of the email.
	htmlBoundry, textBoundry := uuid.NewString(), uuid.NewString()
	builder.WriteString(`Content-Type: multipart/mixed; boundary=`)
	builder.WriteRune('"')
	builder.WriteString(htmlBoundry)
	builder.WriteRune('"')
	builder.WriteRune(';')
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString("--") // Boundary #1
	builder.WriteString(htmlBoundry)
	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString(`Content-Type: multipart/alternative; boundary=`)
	builder.WriteRune('"')
	builder.WriteString(textBoundry)
	builder.WriteRune('"')
	builder.WriteRune(';')
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString("--") // Boundary #2
	builder.WriteString(textBoundry)
	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString(`Content-Type: text/plain;`)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteRune('\r') // Plain text content
	builder.WriteRune('\n')
	builder.WriteString(payload.TextContent)
	builder.WriteRune('\r') // End of text content, start the next boundary
	builder.WriteRune('\n')

	builder.WriteRune('\r') // Start of the html content
	builder.WriteRune('\n')
	builder.WriteString("--") // Boundary #2
	builder.WriteString(textBoundry)
	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString(`Content-Type: text/html; charset="UTF-8";`)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteRune('\r') // HTML content
	builder.WriteRune('\n')
	builder.WriteString(payload.HTMLContent)
	builder.WriteRune('\r') // End of HTML content, start the next boundary
	builder.WriteRune('\n')

	builder.WriteRune('\r') // Ending of inner boundary
	builder.WriteRune('\n')
	builder.WriteString("--")
	builder.WriteString(textBoundry)
	builder.WriteString("--")
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteRune('\r') // Ending of outer boundary
	builder.WriteRune('\n')
	builder.WriteString("--")
	builder.WriteString(htmlBoundry)
	builder.WriteString("--")

	deadline, ok := span.Context().Deadline()
	if !ok {
		// If we don't have a deadline in our context then just set a 5 second deadline.
		deadline = time.Now().Add(5 * time.Second)
	}

	// TODO Could probably move this into a connection pool of some sort. We won't be sending a ton of emails right away
	// but I don't really think there is a need to create and destroy a connection every time? Why not just keep at least
	// one lying around.
	address := fmt.Sprintf("%s:%d", e.config.Email.SMTP.Host, e.config.Email.SMTP.Port)
	connection, err := net.DialTimeout("tcp", address, deadline.Sub(time.Now()))
	if err != nil {
		return errors.Wrap(err, "failed to dial smtp server")
	}

	// Set the deadline for this connection
	if err = connection.SetDeadline(deadline); err != nil {
		return errors.Wrap(err, "failed to set deadling for smtp connection")
	}

	c, err := smtp.NewClient(connection, e.config.Email.SMTP.Host)
	if err != nil {
		return errors.Wrap(err, "failed to create stmp client")
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		e.log.
			WithField("smtpServer", e.config.Email.SMTP.Host).
			WithContext(span.Context()).
			Trace("negotiated TLS connection with SMTP server")
		tlsConfig := &tls.Config{ServerName: e.config.Email.SMTP.Host}
		if err = c.StartTLS(tlsConfig); err != nil {
			return errors.Wrap(err, "failed to negotiate TLS connection for smtp")
		}
	}
	if err = c.Auth(PlainAuth(
		e.config.Email.SMTP.Identity,
		e.config.Email.SMTP.Username,
		e.config.Email.SMTP.Password,
		e.config.Email.SMTP.Host,
	)); err != nil {
		return errors.Wrap(err, "failed to authenticate smtp connection")
	}
	if err = c.Mail(payload.From); err != nil {
		return errors.Wrap(err, "failed to define from address for stmp message")
	}
	if err = c.Rcpt(payload.To); err != nil {
		return errors.Wrap(err, "failed to define to address for smtp message")
	}
	w, err := c.Data()
	if err != nil {
		return errors.Wrap(err, "failed to create message writer for smtp message")
	}
	_, err = w.Write(builder.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to flush message to writer for smtp")
	}
	if err = w.Close(); err != nil {
		return errors.Wrap(err, "failed to close message writer for smtp")
	}

	return c.Quit()
}
