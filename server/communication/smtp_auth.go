//go:build !local

package communication

import "github.com/wneessen/go-mail/smtp"

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return smtp.PlainAuth(identity, username, password, host)
}
