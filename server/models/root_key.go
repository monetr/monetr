package models

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/uptrace/bun"
)

const rootKeyFingerprintConstant = "monetr root key"

type RootKeyKind string

const (
	RootKeyKindSecretBox RootKeyKind = "secretbox"
	RootKeyKindAWS       RootKeyKind = "aws"
	RootKeyKindOpenBao   RootKeyKind = "openbao"
)

// RootKey stores in the database the encrypted root key version. When monetr
// starts it attempts to decrypt the highest ID root key in the database using
// the configured wrapper. If the root key is not configured with that wrapper
// (only possible with KMS) then monetr fails to start and provides an error.
// Root keys are always encrypted with the secretbox key even if the secretbox
// key is not used for startup.
type RootKey struct {
	bun.BaseModel `bun:"table:root_keys,alias:root_key"`

	RootKeyId   ID[RootKey]      `json:"-" bun:"root_key_id,pk,notnull"`
	Wrappers    []RootKeyWrapper `json:"-" bun:"wrappers,notnull,type:jsonb"`
	Fingerprint string           `json:"-" bun:"fingerprint,notnull"`
	CreatedAt   time.Time        `json:"-" bun:"created_at,notnull,nullzero"`
}

type RootKeyWrapper struct {
	// Kind just tells us what implementation to use for this wrapper.
	Kind RootKeyKind `json:"kind"`
	// Reference is something like the KMS key ARN or the transit key ID, for
	// secretbox i dont know what this will be yet.
	Reference string `json:"reference"`
	// Version is provided by the wrapper type and its the version of the wrapper
	// used. For example, if its secretbox this might be the version of the
	// secretbox key. Implementation specific and not necessarily the "highest
	// value wins".
	Version *string `json:"version"`
	// Data is the encrypted root key hex encoded.
	Data string `json:"data"`
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

func RootKeyFingerprint(rootKey []byte) string {
	mac := hmac.New(sha256.New, rootKey[:])
	mac.Write([]byte(rootKeyFingerprintConstant))
	return hex.EncodeToString(mac.Sum(nil))
}
