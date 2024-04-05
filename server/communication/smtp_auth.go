//go:build !local

package communication

import (
	"github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/smtp"
)

const TLSPolicy = mail.TLSMandatory

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return smtp.PlainAuth(identity, username, password, host)
}
