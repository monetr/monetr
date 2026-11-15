package encryption

import (
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"slices"

	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	SecretKeyPrefix = "monetr_sk_"
	SecretKeySize   = 32
)

var secretKeyEncoding = base32.
	NewEncoding("abcdefghijklmnopqrstuvwxyz234567").
	WithPadding(base32.NoPadding)

// SecretKey is the user-held key that wraps the root keys. It is only ever used
// as HKDF input, never directly as an encryption key.
type SecretKey [SecretKeySize]byte

// GenerateSecretKey returns a new random secret key. Use Encode to get the
// value that's handed to the user.
func GenerateSecretKey() SecretKey {
	var key SecretKey
	// As of Go 1.24, crypto/rand.Read never returns an error; it crashes the
	// program if the OS random source fails, so there's nothing to handle.
	rand.Read(key[:])
	return key
}

// Encode returns the form given to the user and accepted in config: monetr_sk_
// followed by 52 characters of unpadded lowercase base32.
func (k SecretKey) Encode() string {
	return SecretKeyPrefix + secretKeyEncoding.EncodeToString(k[:])
}

// String is deliberately redacted so a secret key passed to a logger or
// fmt.Println by mistake doesn't print the key.
func (k SecretKey) String() string {
	return SecretKeyPrefix + "[redacted]"
}

func unwrapRootKey(
	db bun.IDB,
	secret SecretKey,
	secretboxKeyId models.ID[models.SecretboxKey],
	row *models.RootKey,
) ([]byte, error) {
	// Root keys are encrypted potentially with multiple wrappers, this way monetr
	// can start using the secretbox key or with a kms
	i := slices.IndexFunc(
		row.Wrappers,
		func(w models.RootKeyWrapper) bool {
			return w.Kind == models.RootKeyKindSecretBox &&
				w.Reference == secretboxKeyId.String()
		},
	)
	if i < 0 {
		return nil, errors.Errorf(
			"root key %s is not wrapped by this secret key",
			row.RootKeyId,
		)
	}

	wrapKey, err := hkdf.Key(
		sha256.New,
		secret[:],
		nil,
		"monetr secret key wrap",
		chacha20poly1305.KeySize,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive wrapping key")
	}
	defer clear(wrapKey)

	aad, err := models.EncryptionAAD(db, row)
	if err != nil {
		return nil, err
	}

	rootKey, err := openSealed(wrapKey, row.Wrappers[i].Data, aad)
	if err != nil {
		return nil, err
	}

	if !hmac.Equal([]byte(models.RootKeyFingerprint(rootKey)), []byte(row.Fingerprint)) {
		clear(rootKey)
		return nil, errors.Errorf("root key %s does not match its fingerprint", row.RootKeyId)
	}
	return rootKey, nil
}

// openSealed opens hex(nonce || ciphertext || tag) with key. It's the format
// for secretbox root key wrappers and key_encryption_keys.wrapped.
func openSealed(key []byte, sealed string, aad []byte) ([]byte, error) {
	raw, err := hex.DecodeString(sealed)
	if err != nil {
		return nil, errors.Wrap(err, "sealed key is not valid hex")
	}
	if len(raw) < chacha20poly1305.NonceSizeX+chacha20poly1305.Overhead {
		return nil, errors.New("sealed key is truncated")
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key cipher")
	}

	nonce, ciphertext := raw[:chacha20poly1305.NonceSizeX], raw[chacha20poly1305.NonceSizeX:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, errors.New("failed to unwrap key, wrong key or tampered data")
	}
	return plaintext, nil
}
