package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"github.com/xlzd/gotp"
)

var (
	ErrTOTPNotValid = errors.New("provided TOTP code is not valid")
)

type Login struct {
	tableName string `pg:"logins"`

	LoginId         ID[Login]  `json:"loginId" pg:"login_id,notnull,pk"`
	Email           string     `json:"email" pg:"email,notnull,unique"`
	FirstName       string     `json:"firstName" pg:"first_name,notnull"`
	LastName        string     `json:"lastName" pg:"last_name"`
	PasswordResetAt *time.Time `json:"passwordResetAt" pg:"password_reset_at"`
	IsEnabled       bool       `json:"-" pg:"is_enabled,notnull,use_zero"`
	IsEmailVerified bool       `json:"isEmailVerified" pg:"is_email_verified,notnull,use_zero"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt" pg:"email_verified_at"`
	TOTP            string     `json:"-" pg:"totp"`
	TOTPEnabledAt   *time.Time `json:"totpEnabledAt" pg:"totp_enabled_at"`

	Users []User `json:"-" pg:"rel:has-many"`
}

func (Login) IdentityPrefix() string {
	return "lgn"
}

// VerifyTOTP will validate that the provided TOTP string is correct for this login. It will return ErrTOTPNotValid if
// the provided input is not valid, or if TOTP is not configured for the login.
func (l Login) VerifyTOTP(input string, now time.Time) error {
	// If the login does not have TOTP configured, do not return a special error. To the client it should appear as if
	// the TOTP provided is not valid. I don't know if this really makes a difference at all, but it seems like the
	// intuitive thing to do.
	if l.TOTP == "" {
		return errors.WithStack(ErrTOTPNotValid)
	}

	loginTotp := gotp.NewDefaultTOTP(l.TOTP)
	if loginTotp.Verify(input, int(now.Unix())) {
		return nil
	}

	return errors.WithStack(ErrTOTPNotValid)
}

type LoginWithHash struct {
	tableName string `pg:"logins"`

	Login
	Crypt []byte `json:"-" pg:"crypt"`
}

func (l Login) GetEmailIsVerified() bool {
	return l.IsEmailVerified && l.EmailVerifiedAt != nil
}

var (
	_ pg.BeforeInsertHook = (*LoginWithHash)(nil)
)

func (o *LoginWithHash) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LoginId.IsZero() {
		o.LoginId = NewID(&o.Login)
	}

	return ctx, nil
}
