package models

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/pkg/errors"
)

type EmailVerification struct {
	tableName string `pg:"email_verifications"`

	EmailVerificationId uint64     `json:"-" pg:"email_verification_id,notnull,pk,type:'bigserial'"`
	LoginId             uint64     `json:"-" pg:"login_id,notnull,on_delete:CASCADE,unique:per_login"`
	Login               *Login     `json:"-" pg:"rel:has-one"`
	EmailAddress        string     `json:"emailAddress" pg:"email_address,notnull,unique:per_login"`
	IsVerified          bool       `json:"isVerified" pg:"is_verified,notnull,use_zero"`
	CreatedAt           time.Time  `json:"createdAt" pg:"created_at,notnull,default:now()"`
	ExpiresAt           time.Time  `json:"expiresAt" pg:"expires_at,notnull"`
	VerifiedAt          *time.Time `json:"verifiedAt" pg:"verified_at"`
}

func (e EmailVerification) GetVerificationKey() (string, error) {
	// TODO (elliotcourant) This will generate a base64 string of the
	//  verificationId and the loginId, but given that it is possible for the
	//  user to know their loginId, is this secure? Should something be
	//  generated specifically for this rather than a sequence?
	if e.EmailVerificationId == 0 {
		return "", errors.Errorf("cannot create verification key with no id")
	}

	id := make([]byte, 16)
	binary.BigEndian.PutUint64(id, e.EmailVerificationId)
	binary.BigEndian.PutUint64(id[8:], e.LoginId)
	return base64.StdEncoding.EncodeToString(id), nil
}
