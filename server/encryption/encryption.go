package encryption

import (
	"crypto/rand"
	"encoding/base32"
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
