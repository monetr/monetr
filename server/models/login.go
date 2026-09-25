package models

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/xlzd/gotp"
)

var (
	ErrTOTPNotValid = errors.New("provided TOTP code is not valid")
)

type Login struct {
	bun.BaseModel `bun:"table:logins,alias:login"`

	LoginId           ID[Login]  `json:"loginId" bun:"login_id,notnull,pk"`
	Email             string     `json:"email" bun:"email,notnull,unique,nullzero"`
	FirstName         string     `json:"firstName" bun:"first_name,notnull,nullzero"`
	LastName          string     `json:"lastName" bun:"last_name,nullzero"`
	PasswordResetAt   *time.Time `json:"passwordResetAt" bun:"password_reset_at"`
	IsEnabled         bool       `json:"-" bun:"is_enabled,notnull"`
	IsEmailVerified   bool       `json:"isEmailVerified" bun:"is_email_verified,notnull"`
	EmailVerifiedAt   *time.Time `json:"emailVerifiedAt" bun:"email_verified_at"`
	TOTP              string     `json:"-" bun:"totp,nullzero"`
	TOTPRecoveryCodes []string   `json:"-" bun:"totp_recovery_codes,array,nullzero"`
	TOTPEnabledAt     *time.Time `json:"totpEnabledAt" bun:"totp_enabled_at"`

	Users []User `json:"-" bun:"rel:has-many,join:login_id=login_id"`
}

func (Login) IdentityPrefix() string {
	return "lgn"
}

// VerifyTOTP will validate that the provided TOTP string is correct for this
// login. It will return ErrTOTPNotValid if the provided input is not valid, or
// if TOTP is not configured for the login.
func (l Login) VerifyTOTP(input string, now time.Time) error {
	// If the login does not have TOTP configured, do not return a special error.
	// To the client it should appear as if the TOTP provided is not valid. I
	// don't know if this really makes a difference at all, but it seems like the
	// intuitive thing to do.
	if l.TOTP == "" {
		return errors.WithStack(ErrTOTPNotValid)
	}

	loginTotp := gotp.NewDefaultTOTP(l.TOTP)

	// Allow a margin of error of 5 seconds relative to the server time.
	allowedError := 5 * time.Second
	// This probably only needs two, just the negative and positive allowed error,
	// but it's not that expensive to just have all 3 to be clear what is
	// happening.
	allowedTimestamps := []int64{
		now.Unix(),
		now.Add(-allowedError).Unix(),
		now.Add(allowedError).Unix(),
	}
	// Test the valid timestamps against the provided code, if the provided code
	// is valid for ANY of the timestamps then consider it a success.
	for _, timestamp := range allowedTimestamps {
		if loginTotp.Verify(input, timestamp) {
			return nil
		}
	}

	// Otherwise return an error indicating that the TOTP code is invalid.
	return errors.WithStack(ErrTOTPNotValid)
}

func (l Login) Name() string {
	return strings.TrimSpace(l.FirstName + " " + l.LastName)
}

type LoginWithHash struct {
	bun.BaseModel `bun:"table:logins,alias:login_with_hash"`

	Login
	Crypt []byte `json:"-" bun:"crypt,nullzero"`
}

func (l Login) GetEmailIsVerified() bool {
	return l.IsEmailVerified && l.EmailVerifiedAt != nil
}

var (
	_ bun.BeforeAppendModelHook = (*LoginWithHash)(nil)
)

func (o *LoginWithHash) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.LoginId.IsZero() {
			o.LoginId = NewID[Login]()
		}
	}

	return nil
}
