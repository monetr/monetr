package models

import (
	"context"
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base32"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

const (
	ApiKeySecretPrefix = "monetr_secret_"
)

type ApiKey struct {
	tableName string `pg:"api_keys"`

	ApiKeyId      ID[ApiKey]  `json:"apiKeyId" pg:"api_key_id,notnull,pk"`
	AccountId     ID[Account] `json:"-" pg:"account_id,notnull"`
	Account       *Account    `json:"-" pg:"rel:has-one"`
	PublicKey     []byte      `json:"-" pg:"public_key,notnull"`
	CreatedAt     time.Time   `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy     ID[User]    `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser *User       `json:"-" pg:"rel:has-one,fk:created_by"`
	UpdatedAt     time.Time   `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt     *time.Time  `json:"deletedAt" pg:"deleted_at"`
}

func (ApiKey) IdentityPrefix() string {
	return "key"
}

var (
	_ pg.BeforeInsertHook = (*ApiKey)(nil)
)

func (o *ApiKey) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.ApiKeyId.IsZero() {
		o.ApiKeyId = NewID[ApiKey]()
	}

	return ctx, nil
}

// Verify will take they username (keyId) and secret provided by the client and
// validate it against this [ApiKey]. It simply returns true or false indicating
// whether or not the credentials are valid for this specific record.
func (o *ApiKey) Verify(keyId ID[ApiKey], secret string) bool {
	if o.ApiKeyId != keyId {
		return false
	}

	seed, err := base32.StdEncoding.DecodeString(
		strings.TrimPrefix(secret, ApiKeySecretPrefix),
	)
	if err != nil || len(seed) != ed25519.SeedSize {
		return false
	}

	derived := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	return subtle.ConstantTimeCompare(derived, o.PublicKey) == 1
}
