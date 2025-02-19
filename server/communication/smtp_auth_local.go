//go:build local

package communication

import (
	"github.com/wneessen/go-mail"
)

const TLSPolicy = mail.NoTLS
const SMTPAuth = mail.SMTPAuthPlainNoEnc
