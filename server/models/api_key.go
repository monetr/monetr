package models

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

const (
	ApiKeySecretPrefix = "monetr_secret_"
)

// apiKeySecretEncoding is just base32 without padding because i think trailing
// equal signs looks ugly
var apiKeySecretEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type ApiKey struct {
	tableName string `pg:"api_keys"`

	ApiKeyId      ID[ApiKey]  `json:"apiKeyId" pg:"api_key_id,notnull,pk"`
	AccountId     ID[Account] `json:"-" pg:"account_id,notnull"`
	Account       *Account    `json:"-" pg:"rel:has-one"`
	Name          string      `json:"name" pg:"name,notnull"`
	PublicKey     []byte      `json:"-" pg:"public_key,notnull"`
	CreatedAt     time.Time   `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy     ID[User]    `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser *User       `json:"-" pg:"rel:has-one,fk:created_by"`
	UpdatedAt     time.Time   `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt     *time.Time  `json:"deletedAt,omitempty" pg:"deleted_at"`
}

func (ApiKey) IdentityPrefix() string {
	return "key"
}

var (
	_ pg.BeforeInsertHook = (*ApiKey)(nil)
	_ pg.BeforeUpdateHook = (*ApiKey)(nil)
)

func (o *ApiKey) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.ApiKeyId.IsZero() {
		o.ApiKeyId = NewID[ApiKey]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	o.UpdatedAt = now

	return ctx, nil
}

func (o *ApiKey) BeforeUpdate(ctx context.Context) (context.Context, error) {
	o.UpdatedAt = time.Now()
	return ctx, nil
}

// Verify will take they username (keyId) and secret provided by the client and
// validate it against this [ApiKey]. It simply returns true or false indicating
// whether or not the credentials are valid for this specific record.
func (o *ApiKey) Verify(keyId ID[ApiKey], secret string) bool {
	if o.ApiKeyId != keyId || !strings.HasPrefix(secret, ApiKeySecretPrefix) {
		return false
	}

	// I miss clojure :(
	encoded := strings.ToUpper(strings.TrimPrefix(secret, ApiKeySecretPrefix))
	seed, err := apiKeySecretEncoding.DecodeString(encoded)
	if err != nil || len(seed) != ed25519.SeedSize {
		return false
	}

	// A 32 byte seed only needs 256 of the 260 bits that 52 base32 characters
	// carry, the decoder throws away the 4 leftover bits in the final character.
	// That means 16 different secrets decode to the same seed. Re-encoding and
	// comparing forces the caller to present the one canonical spelling, so a
	// hash of the secret (leak detection, secret scanning) is stable.
	if apiKeySecretEncoding.EncodeToString(seed) != encoded {
		return false
	}

	derived := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	return subtle.ConstantTimeCompare(derived, o.PublicKey) == 1
}

func NewApiKey() (*ApiKey, string, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to generate api key")
	}
	secret := apiKeySecretEncoding.EncodeToString(private.Seed())

	return &ApiKey{
		PublicKey: public,
	}, ApiKeySecretPrefix + strings.ToLower(secret), nil
}
