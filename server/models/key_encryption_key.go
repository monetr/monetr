package models

import (
	"time"

	"github.com/uptrace/bun"
)

type KeyEncryptionKey struct {
	bun.BaseModel `bun:"table:key_encryption_keys,alias:key_encryption_key"`

	KeyEncryptionKeyId ID[KeyEncryptionKey] `json:"-" bun:"key_encryption_key_id,notnull,pk"`
	RootKeyId          ID[RootKey]          `json:"-" bun:"root_key_id,notnull"`
	RootKey            *RootKey             `json:"-" bun:"rel:belongs-to,join:root_key_id=root_key_id"`
	CreatedAt          time.Time            `json:"-" bun:"created_at,notnull,nullzero"`
}

func (KeyEncryptionKey) IdentityPrefix() string {
	return "kek"
}
