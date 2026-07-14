package models_test

import (
	"crypto/ed25519"
	"encoding/base32"
	"strings"
	"testing"
	"time"

	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApiKey(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		key, secret, err := NewApiKey()
		require.NoError(t, err, "must be able to generate a new API key")
		require.NotNil(t, key, "generated key must not be nil")

		assert.Len(t, key.PublicKey, ed25519.PublicKeySize, "the stored key should be an ed25519 public key")
		assert.True(t, strings.HasPrefix(secret, ApiKeySecretPrefix), "the secret must have the secret prefix")

		// The secret is the base32 encoded private key seed, and it should be lower
		// case so that it is easier to read and copy.
		encoded := strings.TrimPrefix(secret, ApiKeySecretPrefix)
		assert.Equal(t, strings.ToLower(encoded), encoded, "the encoded secret should be lower case")
		assert.NotContains(t, secret, "=", "the secret should not contain base32 padding characters")

		// A freshly generated key does not have an Id or timestamps yet, those are
		// assigned by the BeforeInsert hook when it is persisted.
		assert.True(t, key.ApiKeyId.IsZero(), "a new key should not have an Id yet")
	})

	t.Run("keys are unique", func(t *testing.T) {
		keyA, secretA, err := NewApiKey()
		require.NoError(t, err, "must be able to generate the first API key")
		keyB, secretB, err := NewApiKey()
		require.NoError(t, err, "must be able to generate the second API key")

		assert.NotEqual(t, secretA, secretB, "two generated secrets must not match")
		assert.NotEqual(t, keyA.PublicKey, keyB.PublicKey, "two generated public keys must not match")
	})
}

func TestApiKey_Verify(t *testing.T) {
	// newKey generates a key and assigns it an Id, mimicking what the database
	// would do on insert. It takes the subtest's own *testing.T so that a failed
	// require targets the correct test goroutine.
	newKey := func(t *testing.T) (*ApiKey, string) {
		key, secret, err := NewApiKey()
		require.NoError(t, err, "must be able to generate a new API key")
		key.ApiKeyId = NewID[ApiKey]()
		return key, secret
	}

	t.Run("happy path", func(t *testing.T) {
		key, secret := newKey(t)
		assert.True(t, key.Verify(key.ApiKeyId, secret), "the issued secret must verify against its own key")
	})

	t.Run("mismatched key Id", func(t *testing.T) {
		key, secret := newKey(t)
		assert.False(t, key.Verify(NewID[ApiKey](), secret), "a secret must not verify against the wrong key Id")
	})

	t.Run("secret without the prefix", func(t *testing.T) {
		key, secret := newKey(t)
		stripped := strings.TrimPrefix(secret, ApiKeySecretPrefix)
		assert.False(t, key.Verify(key.ApiKeyId, stripped), "a secret missing the prefix must not verify")
	})

	t.Run("secret from a different key", func(t *testing.T) {
		key, _ := newKey(t)
		_, otherSecret := newKey(t)
		assert.False(t, key.Verify(key.ApiKeyId, otherSecret), "another key's secret must not verify")
	})

	t.Run("empty secret", func(t *testing.T) {
		key, _ := newKey(t)
		assert.False(t, key.Verify(key.ApiKeyId, ""), "an empty secret must not verify")
	})

	t.Run("only the prefix", func(t *testing.T) {
		key, _ := newKey(t)
		assert.False(t, key.Verify(key.ApiKeyId, ApiKeySecretPrefix), "a secret that is only the prefix must not verify")
	})

	t.Run("secret is not valid base32", func(t *testing.T) {
		key, _ := newKey(t)
		secret := ApiKeySecretPrefix + "this is not base32!"
		assert.False(t, key.Verify(key.ApiKeyId, secret), "a secret that is not valid base32 must not verify")
	})

	t.Run("seed is the wrong length", func(t *testing.T) {
		key, _ := newKey(t)
		// A perfectly valid (unpadded) base32 string, but it decodes to 16 bytes
		// instead of the 32 byte ed25519 seed size.
		shortSeed := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(make([]byte, 16))
		secret := ApiKeySecretPrefix + strings.ToLower(shortSeed)
		assert.False(t, key.Verify(key.ApiKeyId, secret), "a seed that is not the ed25519 seed size must not verify")
	})
}

func TestApiKey_BeforeInsert(t *testing.T) {
	t.Run("assigns an id and timestamps", func(t *testing.T) {
		key := &ApiKey{}
		require.True(t, key.ApiKeyId.IsZero(), "the key should not have an Id before insert")

		_, err := key.BeforeInsert(t.Context())
		require.NoError(t, err, "before insert should not return an error")

		assert.False(t, key.ApiKeyId.IsZero(), "an Id should have been assigned")
		assert.True(t, strings.HasPrefix(key.ApiKeyId.String(), string(ApiKeyIDKind)), "the assigned Id should have the api key prefix")
		assert.False(t, key.CreatedAt.IsZero(), "created at should have been set")
		assert.False(t, key.UpdatedAt.IsZero(), "updated at should have been set")
	})

	t.Run("preserves an existing id and created at", func(t *testing.T) {
		existingId := NewID[ApiKey]()
		existingCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		key := &ApiKey{
			ApiKeyId:  existingId,
			CreatedAt: existingCreatedAt,
		}

		_, err := key.BeforeInsert(t.Context())
		require.NoError(t, err, "before insert should not return an error")

		assert.Equal(t, existingId, key.ApiKeyId, "an already assigned Id should be preserved")
		assert.Equal(t, existingCreatedAt, key.CreatedAt, "an already set created at should be preserved")
		assert.False(t, key.UpdatedAt.IsZero(), "updated at should still be set")
	})
}

func TestApiKey_BeforeUpdate(t *testing.T) {
	t.Run("sets updated at", func(t *testing.T) {
		key := &ApiKey{}
		require.True(t, key.UpdatedAt.IsZero(), "updated at should be zero before update")

		_, err := key.BeforeUpdate(t.Context())
		require.NoError(t, err, "before update should not return an error")

		assert.False(t, key.UpdatedAt.IsZero(), "updated at should have been set")
	})
}

func TestApiKey_IdentityPrefix(t *testing.T) {
	t.Run("matches the id kind", func(t *testing.T) {
		assert.Equal(t, "key", ApiKey{}.IdentityPrefix(), "the api key identity prefix should be 'key'")
		assert.Equal(t, string(ApiKeyIDKind), ApiKey{}.IdentityPrefix(), "the identity prefix should match the api key id kind")
	})
}
