package models

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/uptrace/bun"
)

type RootKey struct {
	bun.BaseModel `bun:"table:root_keys,alias:root_key"`

	RootKeyId   ID[RootKey] `json:"-" bun:"root_key_id,pk,notnull"`
	Fingerprint []byte      `json:"-" bun:"fingerprint,notnull"`
	CreatedAt   time.Time   `json:"-" bun:"created_at,notnull,nullzero"`
}

func (RootKey) IdentityPrefix() string {
	return "rtk"
}

var (
	_ bun.BeforeAppendModelHook = (*RootKey)(nil)
)

func (o *RootKey) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.RootKeyId.IsZero() {
			o.RootKeyId = NewID[RootKey]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}

func (o *RootKey) CompareFingerprint(rootKey []byte) bool {
	mac := hmac.New(sha256.New, rootKey[:])
	mac.Write([]byte(o.RootKeyId.String()))
	fingerprint := mac.Sum(nil)
	return subtle.ConstantTimeCompare(o.Fingerprint, fingerprint) == 1
}
