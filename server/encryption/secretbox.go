package encryption

import (
	"bytes"
	"context"
	"crypto/hkdf"
	"crypto/sha256"
	"log/slog"
	"os"
	"strings"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

type secretboxUnwrapper struct {
	fingerprint string
	wrapKey     []byte
}

func newSecretboxUnwrapper(secret SecretboxKey) (*secretboxUnwrapper, error) {
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
		fingerprint: SecretboxKeyFingerprint(secret),
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
	return unseal(s.wrapKey, w.Data, aad)
}

// SecretboxKeyFingerprint identifies a secret key without revealing it. Stored
// as secretbox_keys.fingerprint and as the secretbox wrapper's Reference.
func SecretboxKeyFingerprint(secret SecretboxKey) string {
	fp, _ := hkdf.Key(
		sha256.New,
		secret[:],
		nil,
		"monetr secret key fingerprint",
		32,
	)
	return models.EncryptionEncoding.EncodeToString(fp)
}

func loadSecretboxKey(
	log *slog.Logger,
	config config.SecretBox,
) (SecretboxKey, error) {
	switch {
	case config.Secret != nil && config.SecretFile != nil:
		return SecretboxKey{}, errors.New("set only one of encryption.secretBox.secret or encryption.secretBox.secretFile")
	case config.Secret != nil:
		return ParseSecretboxKey(strings.TrimSpace(*config.Secret))
	case config.SecretFile != nil:
		return readSecretboxKeyFile(log, *config.SecretFile)
	default:
		return SecretboxKey{}, errors.New("encryption.kind is secretbox but neither secret nor secretFile is configured")
	}
}

func readSecretboxKeyFile(log *slog.Logger, path string) (SecretboxKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return SecretboxKey{}, errors.Wrap(err, "failed to read secret key file")
	}
	if !info.Mode().IsRegular() {
		return SecretboxKey{}, errors.Errorf("secret key file %s is not a regular file", path)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		log.Warn("secret key file is accessible by group or others, it should be 0600",
			"path", path,
			"mode", perm.String(),
		)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return SecretboxKey{}, errors.Wrap(err, "failed to read secret key file")
	}
	defer clear(data)

	return ParseSecretboxKey(string(bytes.TrimSpace(data)))
}

// ParseSecretboxKey parses the monetr_sbk_ form returned by Encode.
func ParseSecretboxKey(encoded string) (SecretboxKey, error) {
	body, ok := strings.CutPrefix(encoded, SecretboxKeyPrefix)
	if !ok {
		return SecretboxKey{}, errors.Errorf("secret key must start with %s", SecretboxKeyPrefix)
	}

	decoded, err := secretKeyEncoding.DecodeString(body)
	if err != nil {
		return SecretboxKey{}, errors.New("secret key is not valid base32")
	}
	defer clear(decoded)

	if len(decoded) != SecretboxKeySize {
		return SecretboxKey{}, errors.Errorf("secret key must be %d bytes, got %d", SecretboxKeySize, len(decoded))
	}

	var key SecretboxKey
	copy(key[:], decoded)
	return key, nil
}
