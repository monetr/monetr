package communication

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
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
	span := crumbs.StartFnTrace(ctx)
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

	builder.WriteString("Date: ")
	builder.WriteString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
	builder.WriteRune('\r')
	builder.WriteRune('\n')

	builder.WriteString("Reply-To: ")
	// TODO Make this a configurable thing.
	builder.WriteString("monetr Support <support@monetr.app>")
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

	deadline, ok := span.Context().Deadline()
	if !ok {
		// If we don't have a deadline in our context then just set a 5 second deadline.
		deadline = time.Now().Add(5 * time.Second)
	}

	// TODO Could probably move this into a connection pool of some sort. We won't be sending a ton of emails right away
	// but I don't really think there is a need to create and destroy a connection every time? Why not just keep at least
	// one lying around.
	address := fmt.Sprintf("%s:%d", s.configuration.Host, s.configuration.Port)
	connection, err := net.DialTimeout("tcp", address, deadline.Sub(time.Now()))
	if err != nil {
		return errors.Wrap(err, "failed to dial smtp server")
	}

	// Set the deadline for this connection
	if err = connection.SetDeadline(deadline); err != nil {
		return errors.Wrap(err, "failed to set deadling for smtp connection")
	}

	c, err := smtp.NewClient(connection, s.configuration.Host)
	if err != nil {
		return errors.Wrap(err, "failed to create stmp client")
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: s.configuration.Host}
		if err = c.StartTLS(config); err != nil {
			return errors.Wrap(err, "failed to negotiate TLS connection for smtp")
		}
	}
	if err = c.Auth(s.auth); err != nil {
		return errors.Wrap(err, "failed to authenticate smtp connection")
	}
	if err = c.Mail(request.From); err != nil {
		return errors.Wrap(err, "failed to define from address for stmp message")
	}
	if err = c.Rcpt(request.To); err != nil {
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
