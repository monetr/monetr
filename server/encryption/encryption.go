package encryption

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"slices"

	"github.com/monetr/monetr/server/crumbs"
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

// unwrapRootKey is also what the keyring calls lazily when a KEK points at a
// root key that isn't loaded yet.
func unwrapRootKey(
	ctx context.Context,
	db bun.IDB,
	unwrapper rootKeyUnwrapper,
	rootKey *models.RootKey,
) ([]byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	i := slices.IndexFunc(rootKey.Wrappers, unwrapper.Matches)
	if i < 0 {
		return nil, errors.Errorf(
			"root key %s is not wrapped by the configured %s key, configure a key that wraps it or re-wrap the root key",
			rootKey.RootKeyId, unwrapper.Kind(),
		)
	}

	aad, err := models.EncryptionAAD(db, rootKey)
	if err != nil {
		return nil, err
	}

	raw, err := unwrapper.Unwrap(ctx, rootKey.Wrappers[i], aad)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unwrap root key %s", rootKey.RootKeyId)
	}

	if models.RootKeyFingerprint(raw) != rootKey.Fingerprint {
		clear(raw)
		return nil, errors.Errorf(
			"root key %s does not match its fingerprint",
			rootKey.RootKeyId,
		)
	}
	return raw, nil
}

// openSealed opens base32(nonce || ciphertext || tag) with key. It's the format
// for secretbox root key wrappers and key_encryption_keys.wrapped.
func openSealed(key []byte, sealed string, aad []byte) ([]byte, error) {
	raw, err := models.EncryptionEncoding.DecodeString(sealed)
	if err != nil {
		return nil, errors.Wrap(err, "sealed key is not valid base32")
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
