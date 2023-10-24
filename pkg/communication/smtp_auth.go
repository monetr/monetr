//go:build !local

package communication

import (
	"net/smtp"
)

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return smtp.PlainAuth(identity, username, password, host)
}
