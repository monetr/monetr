package models

import (
	"time"

	"github.com/pkg/errors"
	"github.com/xlzd/gotp"
)

var (
	ErrTOTPNotValid = errors.New("provided TOTP code is not valid")
)

type Login struct {
	tableName string `pg:"logins"`

	LoginId         uint64       `json:"loginId" pg:"login_id,notnull,pk,type:'bigserial'"`
	Email           string       `json:"email" pg:"email,notnull,unique"`
	FirstName       string       `json:"firstName" pg:"first_name,notnull"`
	LastName        string       `json:"lastName" pg:"last_name"`
	PasswordResetAt *time.Time   `json:"passwordResetAt" pg:"password_reset_at"`
	PhoneNumber     *PhoneNumber `json:"-" pg:"phone_number,type:'text'"`
	IsEnabled       bool         `json:"-" pg:"is_enabled,notnull,use_zero"`
	IsEmailVerified bool         `json:"isEmailVerified" pg:"is_email_verified,notnull,use_zero"`
	EmailVerifiedAt *time.Time   `json:"emailVerifiedAt" pg:"email_verified_at"`
	IsPhoneVerified bool         `json:"isPhoneVerified" pg:"is_phone_verified,notnull,use_zero"`
	TOTP            string       `json:"-" pg:"totp"`
	TOTPEnabledAt   *time.Time   `json:"totpEnabledAt" pg:"totp_enabled_at"`

	Users []User `json:"-" pg:"rel:has-many"`
}

// VerifyTOTP will validate that the provided TOTP string is correct for this login. It will return ErrTOTPNotValid if
// the provided input is not valid, or if TOTP is not configured for the login.
func (l Login) VerifyTOTP(input string) error {
	// If the login does not have TOTP configured, do not return a special error. To the client it should appear as if
	// the TOTP provided is not valid. I don't know if this really makes a difference at all, but it seems like the
	// intuitive thing to do.
	if l.TOTP == "" {
		return errors.WithStack(ErrTOTPNotValid)
	}

	loginTotp := gotp.NewDefaultTOTP(l.TOTP)
	if loginTotp.Verify(input, int(time.Now().Unix())) {
		return nil
	}

	return errors.WithStack(ErrTOTPNotValid)
}

func (l Login) GetEmailIsVerified() bool {
	return l.IsEmailVerified && l.EmailVerifiedAt != nil
}

type LoginWithHash struct {
	tableName string `pg:"logins"`

	Login
	PasswordHash string `json:"-" pg:"password_hash,notnull"`
}

// LoginWithVerifier gives us access to the fields needed for secure remote password authentication. The notnull tags
// are meant to enforce the ORM and are not representative of the database constraints.
type LoginWithVerifier struct {
	tableName string `pg:"logins"`

	Login
	Verifier []byte `json:"-" pg:"verifier,notnull"`
	Salt     []byte `json:"-" pg:"salt,notnull"`
}
