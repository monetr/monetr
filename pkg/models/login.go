package models

import (
	"time"
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

	Users []User `json:"-" pg:"rel:has-many"`
}

type LoginWithHash struct {
	tableName string `pg:"logins"`

	Login
	PasswordHash string `json:"-" pg:"password_hash,notnull"`
}

func (l Login) GetEmailIsVerified() bool {
	return l.IsEmailVerified && l.EmailVerifiedAt != nil
}
