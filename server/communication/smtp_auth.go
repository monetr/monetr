//go:build !local

package communication

import (
	"github.com/wneessen/go-mail"
)

const TLSPolicy = mail.TLSMandatory
const SMTPAuth = mail.SMTPAuthAutoDiscover
