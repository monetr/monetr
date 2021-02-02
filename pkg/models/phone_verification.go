package models

import (
	"time"
)

type PhoneVerification struct {
	tableName string `pg:"phone_verifications"`

	PhoneVerificationId uint64      `json:"-" pg:"phone_verification_id,notnull,pk,type:'bigserial'"`
	LoginId             uint64      `json:"-" pg:"login_id,notnull,on_delete:CASCADE,unique:per_login_0,per_login_1"`
	Login               *Login      `json:"-" pg:"rel:has-one"`
	Code                string      `json:"-" pg:"code,notnull,unique:per_login_0"`
	PhoneNumber         PhoneNumber `json:"phoneNumber" pg:"phone_number,notnull,unique:per_login_1,type:'text'"`
	IsVerified          bool        `json:"isVerified" pg:"is_verified,notnull,use_zero"`
	CreatedAt           time.Time   `json:"createdAt" pg:"created_at,notnull,default:now()"`
	ExpiresAt           time.Time   `json:"expiresAt" pg:"expires_at,notnull"`
	VerifiedAt          *time.Time  `json:"verifiedAt" pg:"verified_at"`
}
