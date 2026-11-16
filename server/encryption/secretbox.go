package encryption

import (
	"bytes"
	"context"
	"crypto/hkdf"
	"crypto/sha256"
	"io/fs"
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
	return openSealed(s.wrapKey, w.Data, aad)
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

func loadOrCreateSecretboxKey(
	log *slog.Logger,
	config config.SecretBox,
) (SecretboxKey, error) {
	switch {
	case config.Secret != nil && config.SecretFile != nil:
		return SecretboxKey{}, errors.New("set only one of encryption.secretBox.secret or encryption.secretBox.secretFile")
	case config.Secret != nil:
		return ParseSecretboxKey(strings.TrimSpace(*config.Secret))
	case config.SecretFile != nil:
		return readOrCreateSecretboxKeyFile(
			log,
			*config.SecretFile,
			config.CreateIfNotExists,
		)
	default:
		return SecretboxKey{}, errors.New("encryption.kind is secretbox but neither secret nor secretFile is configured")
	}
}

// readOrCreateSecretboxKeyFile reads the secret key from path. If create is true
// and the file doesn't exist, a new secret key is generated and written there
// first.
func readOrCreateSecretboxKeyFile(
	log *slog.Logger,
	path string,
	create bool,
) (SecretboxKey, error) {
	info, err := os.Stat(path)
	if create && errors.Is(err, fs.ErrNotExist) {
		key, err := createSecretboxKeyFile(log, path)
		if !errors.Is(err, fs.ErrExist) {
			return key, err
		}
		// Another process created the file between the stat and the create, so
		// read theirs instead.
		info, err = os.Stat(path)
	}
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

// createSecretboxKeyFile generates a new secret key and writes it to path. It
// never overwrites an existing file; if path already exists it returns an error
// that matches fs.ErrExist.
func createSecretboxKeyFile(log *slog.Logger, path string) (SecretboxKey, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return SecretboxKey{}, err
	} else if err != nil {
		return SecretboxKey{}, errors.Wrap(err, "failed to create secret key file")
	}

	key := GenerateSecretboxKey()
	_, err = file.WriteString(key.Encode() + "\n")
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		// Don't leave a partial file behind, the next startup would fail to parse
		// it instead of generating a new one.
		os.Remove(path)
		return SecretboxKey{}, errors.Wrap(err, "failed to write secret key file")
	}

	log.Info("generated a new secret key", "path", path)
	return key, nil
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
