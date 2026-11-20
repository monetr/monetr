package encryption

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"log/slog"
	"slices"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	SecretboxKeyPrefix = "monetr_sbk_"
	SecretboxKeySize   = 32
)

// encryptionSetupLockId is the transaction scoped advisory lock that serializes
// first time encryption setup across servers starting at the same time.
const encryptionSetupLockId int64 = 20261117

// ErrEncryptionNotInitialized is returned by Start when no keys exist yet and
// setup wasn't requested.
var ErrEncryptionNotInitialized = errors.New("envelope encryption has not been set up")

// ErrEncryptionAlreadyInitialized is returned by Initialize when keys already
// exist.
var ErrEncryptionAlreadyInitialized = errors.New("envelope encryption has already been set up")

var secretKeyEncoding = base32.
	NewEncoding("abcdefghijklmnopqrstuvwxyz234567").
	WithPadding(base32.NoPadding)

// SecretboxKey is the user-held key that wraps the root keys. It is only ever
// used as HKDF input, never directly as an encryption key.
type SecretboxKey [SecretboxKeySize]byte

// GenerateSecretboxKey returns a new random secret key. Use Encode to get the
// value that's handed to the user.
func GenerateSecretboxKey() SecretboxKey {
	var key SecretboxKey
	_, _ = rand.Read(key[:])
	return key
}

// Encode returns the form given to the user and accepted in config: monetr_sbk_
// followed by 52 characters of unpadded lowercase base32.
func (k SecretboxKey) Encode() string {
	return SecretboxKeyPrefix + secretKeyEncoding.EncodeToString(k[:])
}

// String is deliberately redacted so a secret key passed to a logger or
// fmt.Println by mistake doesn't print the key.
func (k SecretboxKey) String() string {
	return SecretboxKeyPrefix + "[redacted]"
}

// unwrapRootKey is also what the keyring calls lazily when a KEK points at a
// root key that isn't loaded yet.
func unwrapRootKey(
	ctx context.Context,
	db bun.IDB,
	wrapper rootKeyWrapper,
	rootKey *models.RootKey,
) ([]byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	i := slices.IndexFunc(rootKey.Wrappers, wrapper.Matches)
	if i < 0 {
		return nil, errors.Errorf(
			"root key %s is not wrapped by the configured %s key, configure a key that wraps it or re-wrap the root key",
			rootKey.RootKeyId, wrapper.Kind(),
		)
	}

	aad, err := models.EncryptionAAD(rootKey)
	if err != nil {
		return nil, err
	}

	raw, err := wrapper.Unwrap(ctx, rootKey.Wrappers[i], aad)
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

// seal encrypts plaintext with key and returns base32(nonce || ciphertext ||
// tag), the format unseal reads. It's used for secretbox root key wrappers and
// key_encryption_keys.wrapped.
func seal(key, plaintext, aad []byte) (string, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return "", errors.Wrap(err, "failed to create key cipher")
	}

	nonce := make(
		[]byte,
		chacha20poly1305.NonceSizeX,
		chacha20poly1305.NonceSizeX+len(plaintext)+aead.Overhead(),
	)
	rand.Read(nonce)
	sealed := aead.Seal(nonce, nonce, plaintext, aad)
	return models.EncryptionEncoding.EncodeToString(sealed), nil
}

// unseal opens base32(nonce || ciphertext || tag) with key. It's the format
// for secretbox root key wrappers and key_encryption_keys.wrapped.
func unseal(key []byte, sealed string, aad []byte) ([]byte, error) {
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

// Bootstrap controls what Start does when encryption hasn't been set up yet.
type Bootstrap struct {
	// Initialize is the user's opt in to generating keys when none exist.
	Initialize bool
	// DeliverSecret gives the new secretbox key to the user. It runs inside the
	// setup transaction before commit, so if it fails nothing is saved. Required
	// when Initialize is true. The secret is in its encoded monetr_sbk_ form.
	DeliverSecret func(ctx context.Context, secret string) error
}

// Start sets up envelope encryption if it hasn't been and the user asked for
// it, then returns the keyring. If encryption isn't set up and Initialize is
// false it returns ErrEncryptionNotInitialized.
func Start(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	configuration config.Configuration,
	db *bun.DB,
	bootstrap Bootstrap,
) (*Keyring, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	repo := repository.NewEncryptionRepository(log, clock, db)
	initialized, err := isEncryptionInitialized(span.Context(), repo)
	if err != nil {
		return nil, err
	}

	if !initialized {
		if !bootstrap.Initialize {
			// The caller decides what this means: a warning during the
			// transition, a hard failure once encryption is required.
			return nil, ErrEncryptionNotInitialized
		}
		err := Initialize(
			span.Context(),
			log,
			clock,
			configuration.Encryption,
			db,
			bootstrap.DeliverSecret,
		)
		// Already initialized means another server won the race between the check
		// above and the setup lock. That's fine here, the keys exist either way.
		if err != nil && !errors.Is(err, ErrEncryptionAlreadyInitialized) {
			return nil, errors.Wrap(err, "failed to set up envelope encryption")
		}
	}

	// Encryption is set up by this point, so building the wrapper from config
	// is safe even for secretbox.
	keyring, err := NewKeyring(span.Context(), log, clock, configuration, db)
	if err != nil {
		return nil, err
	}

	// Building the keyring doesn't touch any keys, so prove the configured key
	// works now instead of on the first encrypted read or write.
	if err := keyring.verify(span.Context()); err != nil {
		return nil, err
	}

	return keyring, nil
}

// CurrentSecretboxKey decrypts the newest secretbox key with the configured key
// and returns it encoded. Unlike Start it does not require the configured key to
// be the current secretbox key, any key that still wraps the root key works. So
// with a KMS it recovers the secretbox key, and with an old secretbox key it
// returns the current one. Returns ErrEncryptionNotInitialized if encryption
// has not been set up.
func CurrentSecretboxKey(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	configuration config.Configuration,
	db *bun.DB,
) (string, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	repo := repository.NewEncryptionRepository(log, clock, db)
	initialized, err := isEncryptionInitialized(span.Context(), repo)
	if err != nil {
		return "", err
	}
	if !initialized {
		return "", ErrEncryptionNotInitialized
	}

	keyring, err := NewKeyring(span.Context(), log, clock, configuration, db)
	if err != nil {
		return "", err
	}

	_, encoded, err := keyring.currentSecretboxKey(span.Context())
	return encoded, err
}

// isEncryptionInitialized reports whether setup has run. Setup creates every
// key in one transaction, so the secretbox row existing means all of it does.
func isEncryptionInitialized(
	ctx context.Context,
	repo repository.EncryptionRepository,
) (bool, error) {
	_, err := repo.GetSecretboxKey(ctx)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	default:
		return false, errors.Wrap(err, "failed to check whether encryption is set up")
	}
}

// Initialize generates the secretbox key, root key, and first KEK and stores
// them in a single transaction. deliver gets the encoded secretbox key before
// commit, so if it fails nothing is saved. If encryption is already set up it
// returns ErrEncryptionAlreadyInitialized and deliver is never called.
func Initialize(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	configuration config.Encryption,
	db *bun.DB,
	deliver func(ctx context.Context, secret string) error,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if deliver == nil {
		return errors.New("encryption setup requires a way to deliver the secretbox key")
	}

	// With a KMS kind the root key also gets a KMS wrapper. Build that client
	// before the transaction so a misconfigured KMS fails before anything is
	// written. An unknown kind or invalid secretbox config fails here too:
	// otherwise setup would commit keys that startup can never open.
	var kms rootKeyWrapper
	switch models.RootKeyKind(configuration.Kind) {
	case models.RootKeyKindSecretBox:
		// Same check startup makes when loading the key. Catch it here so a config
		// that can never start is rejected before keys are committed.
		if configuration.Secretbox.Secret != nil && configuration.Secretbox.Path != nil {
			return errors.New("set only one of encryption.secretbox.secret or encryption.secretbox.path")
		}
	case models.RootKeyKindAWS, models.RootKeyKindOpenBao:
		var err error
		if kms, err = newRootKeyWrapper(span.Context(), log, configuration); err != nil {
			return err
		}
	case "":
		return errors.New("encryption.kind is not configured")
	default:
		return errors.Errorf("unknown encryption.kind %q, must be secretbox, aws, or openbao", configuration.Kind)
	}

	// Default isolation (READ COMMITTED) on purpose. Each statement sees a fresh
	// snapshot, so a server that waited on the lock sees the rows the first one
	// committed. Under REPEATABLE READ or SERIALIZABLE the snapshot is taken
	// before the lock wait, so the re-check would see an empty table.
	return db.RunInTx(span.Context(), nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			"SELECT pg_advisory_xact_lock(?)",
			encryptionSetupLockId,
		); err != nil {
			return errors.Wrap(err, "failed to acquire encryption setup lock")
		}

		repo := repository.NewEncryptionRepository(log, clock, tx)
		if initialized, err := isEncryptionInitialized(ctx, repo); err != nil {
			return err
		} else if initialized {
			return ErrEncryptionAlreadyInitialized
		}

		// Setup always generates its own secret. Warn here, once it's certain
		// setup will actually run, so a configured secret isn't mistaken for the
		// one being used.
		if configuration.Secretbox.Secret != nil {
			log.WarnContext(ctx,
				"encryption.secretbox.secret is configured but this secret will not be used for init, a new secret will be generated and the configured one must be replaced with it",
			)
		}

		secret := GenerateSecretboxKey()
		secretbox, err := newSecretboxWrapper(secret)
		if err != nil {
			return err
		}

		// Root key, wrapped by the secretbox key and by the KMS if one is
		// configured.
		rawRootKey := make([]byte, chacha20poly1305.KeySize)
		rand.Read(rawRootKey)
		defer clear(rawRootKey)

		rootKey := &models.RootKey{
			// The ID is part of the AAD so it has to be set before wrapping.
			RootKeyId:   models.NewID[models.RootKey](),
			Fingerprint: models.RootKeyFingerprint(rawRootKey),
		}
		rootKeyAAD, err := models.EncryptionAAD(rootKey)
		if err != nil {
			return err
		}
		wrappers := []rootKeyWrapper{secretbox}
		if kms != nil {
			wrappers = append(wrappers, kms)
		}
		for _, wrapper := range wrappers {
			wrapped, err := wrapper.Wrap(ctx, rawRootKey, rootKeyAAD)
			if err != nil {
				return errors.Wrapf(err, "failed to wrap root key with %s", wrapper.Kind())
			}
			rootKey.Wrappers = append(rootKey.Wrappers, wrapped)
		}
		if err := repo.CreateRootKey(ctx, rootKey); err != nil {
			return err
		}

		// First KEK, sealed by the root key.
		rawKEK := make([]byte, chacha20poly1305.KeySize)
		rand.Read(rawKEK)
		defer clear(rawKEK)

		kek := &models.KeyEncryptionKey{
			KeyEncryptionKeyId: models.NewID[models.KeyEncryptionKey](),
			RootKeyId:          rootKey.RootKeyId,
		}
		kekAAD, err := models.EncryptionAAD(kek)
		if err != nil {
			return err
		}
		if kek.Wrapped, err = seal(rawRootKey, rawKEK, kekAAD); err != nil {
			return err
		}
		if err := repo.CreateKeyEncryptionKey(ctx, kek); err != nil {
			return err
		}

		// The secretbox key itself, encrypted through root key -> KEK -> DEK so
		// show-secret-key can recover it later. This row is also what
		// isEncryptionInitialized looks for.
		secretboxKey := &models.SecretboxKey{
			SecretboxKeyId: models.NewID[models.SecretboxKey](),
			Fingerprint:    SecretboxKeyFingerprint(secret),
		}
		secretboxKeyAAD, err := models.EncryptionAAD(secretboxKey)
		if err != nil {
			return err
		}
		if err := secretboxKey.Encrypt(
			ctx,
			setupKEK{id: kek.KeyEncryptionKeyId, raw: rawKEK},
			secretboxKeyAAD,
			secret.Encode(),
		); err != nil {
			return err
		}
		if err := repo.CreateSecretboxKey(ctx, secretboxKey); err != nil {
			return err
		}

		// Last, and still inside the transaction, so a failed delivery rolls
		// everything back.
		if err := deliver(ctx, secret.Encode()); err != nil {
			return errors.Wrap(err, "failed to deliver secretbox key")
		}
		return nil
	})
}

var _ models.DataKeyWrapper = setupKEK{}

// setupKEK wraps DEKs with the KEK being created during setup, before any
// keyring exists to do it.
type setupKEK struct {
	id  models.ID[models.KeyEncryptionKey]
	raw []byte
}

func (k setupKEK) WrapDataKey(
	_ context.Context,
	dek []byte,
) (models.ID[models.KeyEncryptionKey], []byte, error) {
	wrapped, err := sealDataKey(k.raw, k.id, dek)
	return k.id, wrapped, err
}

func (k setupKEK) UnwrapDataKey(
	context.Context,
	models.ID[models.KeyEncryptionKey],
	[]byte,
) ([]byte, error) {
	return nil, errors.New("unwrapping is not supported during encryption setup")
}

// sealDataKey wraps a DEK with a KEK. The AAD is the KEK ID, so a wrapped DEK
// can't be relabeled as belonging to a different KEK. Result is nonce ||
// ciphertext.
func sealDataKey(
	kek []byte,
	kekId models.ID[models.KeyEncryptionKey],
	dek []byte,
) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key cipher")
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	rand.Read(nonce)
	return aead.Seal(nonce, nonce, dek, []byte(kekId)), nil
}

// unsealDataKey opens a DEK wrapped by sealDataKey. kekId has to be the ID the
// DEK was wrapped under, since its the AAD.
func unsealDataKey(
	kek []byte,
	kekId models.ID[models.KeyEncryptionKey],
	wrapped []byte,
) ([]byte, error) {
	if len(wrapped) < chacha20poly1305.NonceSizeX+chacha20poly1305.Overhead {
		return nil, errors.New("wrapped data key is truncated")
	}

	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key cipher")
	}

	nonce := wrapped[:chacha20poly1305.NonceSizeX]
	ciphertext := wrapped[chacha20poly1305.NonceSizeX:]

	dek, err := aead.Open(nil, nonce, ciphertext, []byte(kekId))
	if err != nil {
		return nil, errors.Errorf("failed to unwrap data key with KEK %s, wrong key or tampered data", kekId)
	}

	return dek, nil
}
