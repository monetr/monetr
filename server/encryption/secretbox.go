package encryption

import (
	"bytes"
	"context"
	"crypto/hkdf"
	"crypto/sha256"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

type secretboxWrapper struct {
	fingerprint string
	wrapKey     []byte
}

func newSecretboxWrapper(secret SecretboxKey) (*secretboxWrapper, error) {
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
	return &secretboxWrapper{
		fingerprint: SecretboxKeyFingerprint(secret),
		wrapKey:     wrapKey,
	}, nil
}

func (s *secretboxWrapper) Kind() models.RootKeyKind {
	return models.RootKeyKindSecretBox
}

func (s *secretboxWrapper) Matches(w models.RootKeyWrapper) bool {
	return w.Kind == models.RootKeyKindSecretBox && w.Reference == s.fingerprint
}

func (s *secretboxWrapper) Unwrap(_ context.Context, w models.RootKeyWrapper, aad []byte) ([]byte, error) {
	return unseal(s.wrapKey, w.Data, aad)
}

func (s *secretboxWrapper) Wrap(
	_ context.Context,
	rootKey, aad []byte,
) (models.RootKeyWrapper, error) {
	data, err := seal(s.wrapKey, rootKey, aad)
	if err != nil {
		return models.RootKeyWrapper{}, err
	}
	return models.RootKeyWrapper{
		Kind:      models.RootKeyKindSecretBox,
		Reference: s.fingerprint,
		Data:      data,
	}, nil
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
	config config.Secretbox,
) (SecretboxKey, error) {
	switch {
	case config.Secret != nil && config.Path != nil:
		return SecretboxKey{}, errors.New("set only one of encryption.secretbox.secret or encryption.secretbox.path")
	case config.Secret != nil:
		return ParseSecretboxKey(strings.TrimSpace(*config.Secret))
	case config.Path != nil:
		return readSecretboxKeyFile(log, *config.Path)
	default:
		return SecretboxKey{}, errors.New("encryption.kind is secretbox but neither secret nor path is configured")
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

// WriteSecretboxKeyFile writes contents to a new file at path, readable and
// writable only by the owner (0600). Missing parent directories are created
// owner only (0700); existing ones are left as they are. It never overwrites:
// if anything already exists at path it fails. If the write fails part way the
// file is removed so a retry isn't blocked by a partial key.
func WriteSecretboxKeyFile(path string, contents string) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return errors.Wrap(err, "failed to create secret key file directory")
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return errors.Errorf("secret key file %s already exists, refusing to overwrite it", path)
		}
		return errors.Wrap(err, "failed to create secret key file")
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "failed to close secret key file")
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()

	// The mode passed to OpenFile is filtered by the umask. Set it explicitly so
	// the result is exactly 0600 regardless.
	if err := file.Chmod(0o600); err != nil {
		return errors.Wrap(err, "failed to set secret key file permissions")
	}
	if _, err := file.WriteString(contents); err != nil {
		return errors.Wrap(err, "failed to write secret key file")
	}
	// Make sure the key is on disk before the caller commits anything that
	// depends on it.
	if err := file.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync secret key file")
	}
	return nil
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
