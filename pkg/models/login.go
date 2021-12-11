package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Login struct {
	tableName string `pg:"logins"`

	LoginId         uint64       `json:"loginId" bun:"login_id,notnull,pk"`
	Email           string       `json:"email" bun:"email,notnull"`
	FirstName       string       `json:"firstName" bun:"first_name,notnull"`
	LastName        string       `json:"lastName" bun:"last_name"`
	PasswordResetAt *time.Time   `json:"passwordResetAt" bun:"password_reset_at"`
	PhoneNumber     *PhoneNumber `json:"-" bun:"phone_number,type:'text'"`
	IsEnabled       bool         `json:"-" bun:"is_enabled,notnull"`
	IsEmailVerified bool         `json:"isEmailVerified" bun:"is_email_verified,notnull"`
	EmailVerifiedAt *time.Time   `json:"emailVerifiedAt" bun:"email_verified_at"`
	IsPhoneVerified bool         `json:"isPhoneVerified" bun:"is_phone_verified,notnull"`

	Users []User `json:"-" bun:"rel:has-many,join:login_id=login_id"`
}

type LoginWithHash struct {
	bun.BaseModel `bun:"logins"`

	Login
	PasswordHash string `json:"-" bun:"password_hash,notnull"`
}

func (l Login) GetEmailIsVerified() bool {
	return l.IsEmailVerified && l.EmailVerifiedAt != nil
}
