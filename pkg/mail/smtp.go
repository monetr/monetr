package mail

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SendEmailRequest struct {
	From    string
	To      string
	Subject string
	Content string
	IsHTML  bool
}

type Communication interface {
	Send(ctx context.Context, request SendEmailRequest) error
}

var (
	_ Communication = &smtpCommunication{}
)

type smtpCommunication struct {
	log           *logrus.Entry
	configuration config.SMTPClient
	auth          smtp.Auth
}

func NewSMTPCommunication(log *logrus.Entry, configuration config.SMTPClient) Communication {
	auth := PlainAuth(
		configuration.Identity,
		configuration.Username,
		configuration.Password,
		configuration.Host,
	)

	return &smtpCommunication{
		log:           log,
		configuration: configuration,
		auth:          auth,
	}
}

func (s *smtpCommunication) Send(ctx context.Context, request SendEmailRequest) error {
	span := sentry.StartSpan(ctx, "SMTP - Send")
	defer span.Finish()

	builder := bytes.NewBuffer(nil)
	builder.WriteString("To: ")
	builder.WriteString(request.To)
	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString("From: ")
	builder.WriteString(request.From)
	builder.WriteRune('\r')
	builder.WriteRune('\n')
	builder.WriteString("Subject: ")
	builder.WriteString(request.Subject)
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	if request.IsHTML {
		builder.WriteString("MIME-version: 1.0;")
		builder.WriteRune('\r')
		builder.WriteRune('\n')
		builder.WriteString(`Content-Type: text/html; charset="UTF-8";`)
		builder.WriteRune('\r')
		builder.WriteRune('\n')
	}

	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString(request.Content)

	if err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.configuration.Host, s.configuration.Port),
		s.auth,
		request.From,
		[]string{request.To},
		builder.Bytes(),
	); err != nil {
		s.log.WithError(err).Errorf("failed to send email via SMTP")
		return errors.Wrap(err, "failed to send email via SMTP")
	}

	return nil
}
