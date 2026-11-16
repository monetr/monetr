package encryption

import (
	"context"
	"crypto/hkdf"
	"crypto/sha256"

	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

type secretboxUnwrapper struct {
	fingerprint string
	wrapKey     []byte
}

func newSecretboxUnwrapper(secret SecretKey) (*secretboxUnwrapper, error) {
	wrapKey, err := hkdf.Key(
		sha256.New,
		secret[:],
		nil,
		"monetr secret key wrap",
		chacha20poly1305.KeySize,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive secret key wrapping key")
	}
	return &secretboxUnwrapper{
		fingerprint: SecretKeyFingerprint(secret),
		wrapKey:     wrapKey,
	}, nil
}

func (s *secretboxUnwrapper) Kind() models.RootKeyKind {
	return models.RootKeyKindSecretBox
}

func (s *secretboxUnwrapper) Matches(w models.RootKeyWrapper) bool {
	return w.Kind == models.RootKeyKindSecretBox && w.Reference == s.fingerprint
}

func (s *secretboxUnwrapper) Unwrap(_ context.Context, w models.RootKeyWrapper, aad []byte) ([]byte, error) {
	return openSealed(s.wrapKey, w.Data, aad)
}

// SecretKeyFingerprint identifies a secret key without revealing it. Stored as
// secretbox_keys.fingerprint and as the secretbox wrapper's Reference.
func SecretKeyFingerprint(secret SecretKey) string {
	fp, _ := hkdf.Key(sha256.New, secret[:], nil, "monetr secret key fingerprint", 32)
	return models.EncryptionEncoding.EncodeToString(fp)
}
