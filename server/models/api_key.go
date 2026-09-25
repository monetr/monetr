package models

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// apiKeySecretEncoding is just base32 without padding because i think trailing
// equal signs looks ugly
var apiKeySecretEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type ApiKey struct {
	bun.BaseModel `bun:"table:api_keys,alias:api_key"`

	ApiKeyId      ID[ApiKey]  `json:"apiKeyId" bun:"api_key_id,notnull,pk"`
	AccountId     ID[Account] `json:"-" bun:"account_id,notnull,nullzero"`
	Account       *Account    `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	Name          string      `json:"name" bun:"name,notnull,nullzero"`
	PublicKey     []byte      `json:"-" bun:"public_key,notnull,nullzero"`
	CreatedAt     time.Time   `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy     ID[User]    `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser *User       `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	UpdatedAt     time.Time   `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt     *time.Time  `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (ApiKey) IdentityPrefix() string {
	return "key"
}

var (
	_ bun.BeforeAppendModelHook = (*ApiKey)(nil)
)

func (o *ApiKey) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.ApiKeyId.IsZero() {
			o.ApiKeyId = NewID[ApiKey]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
		o.UpdatedAt = now
	case *bun.UpdateQuery:
		o.UpdatedAt = time.Now()
	}

	return nil
}

// Verify will take they username (keyId) and secret provided by the client and
// validate it against this [ApiKey]. It simply returns true or false indicating
// whether or not the credentials are valid for this specific record.
func (o *ApiKey) Verify(keyId ID[ApiKey], secret string) bool {
	if o.ApiKeyId != keyId {
		return false
	}

	// I miss clojure :(
	encoded := strings.ToUpper(secret)
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
	}, strings.ToLower(secret), nil
}
