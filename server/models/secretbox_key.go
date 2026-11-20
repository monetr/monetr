package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type SecretboxKey struct {
	bun.BaseModel `bun:"table:secretbox_keys,alias:secretbox_key"`

	SecretboxKeyId ID[SecretboxKey] `json:"-" bun:"secretbox_key_id,pk,notnull"`
	Fingerprint    string           `json:"-" bun:"fingerprint,notnull"`
	CreatedAt      time.Time        `json:"-" bun:"created_at,notnull,nullzero"`

	// SecretboxKey is stored encrypted on the row
	Encrypted[string]
}

func (SecretboxKey) IdentityPrefix() string {
	return "sbk"
}

var (
	_ bun.BeforeAppendModelHook = (*Secret)(nil)
)

func (o *SecretboxKey) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.SecretboxKeyId.IsZero() {
			o.SecretboxKeyId = NewID[SecretboxKey]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
