package models

import (
	"time"

	"github.com/uptrace/bun"
)

type SecretboxKey struct {
	bun.BaseModel `bun:"table:secretbox_keys,alias:secretbox_key"`

	SecretboxKeyId ID[SecretboxKey] `json:"-" bun:"secretbox_key_id,pk,notnull"`
	Fingerprint    string           `json:"-" bun:"fingerprint,notnull"`
	CreatedAt      time.Time        `json:"-" bun:"created_at,notnull,nullzero"`
}

func (SecretboxKey) IdentityPrefix() string {
	return "sbk"
}
