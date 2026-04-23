package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

// Key represents an individual user's API key, keys are user specific. They are
// not login level or account level. Two user's in the same account will have
// different API keys. Credentials for the key are generated when the key is
// created. The unencrypted credential is given to the user once, the signature
// used to verify the credentials is then stored on the key row. The secret
// is 64 bytes returned as a base32 string, the verifier is the SHA512 hash of
// the 64 bytes.
type Key struct {
	tableName string `pg:"keys"`

	KeyId     ID[Key]     `json:"keyId" pg:"key_id,notnull,pk"`
	UserId    ID[User]    `json:"-" pg:"user_id,notnull,pk"`
	User      *User       `json:"-" pg:"rel:has-one"`
	AccountId ID[Account] `json:"-" pg:"account_id,notnull,pk"`
	Account   *Account    `json:"-" pg:"rel:has-one"`
	Verifier  []byte      `json:"-" pg:"verifier,notnull"`
	CreatedAt time.Time   `json:"createdAt" pg:"created_at,notnull"`
	DeletedAt *time.Time  `json:"deletedAt" pg:"deleted_at"`
}

func (Key) IdentityPrefix() string {
	return "key"
}

var (
	_ pg.BeforeInsertHook = (*Key)(nil)
)

// BeforeInsert implements [orm.BeforeInsertHook].
func (o *Key) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.KeyId.IsZero() {
		o.KeyId = NewID[Key]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
