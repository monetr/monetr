package encryption

import (
	"context"
	"log/slog"
	"sync"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

var (
	_ models.DataKeyWrapper = &Keyring{}
)

type Keyring struct {
	lock     sync.RWMutex
	wrapper  rootKeyWrapper
	rootKeys map[models.ID[models.RootKey]][]byte
	keks     map[models.ID[models.KeyEncryptionKey]][]byte

	db   bun.IDB
	repo repository.EncryptionRepository
}

func NewKeyring(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	config config.Configuration,
	db bun.IDB,
) (*Keyring, error) {
	wrapper, err := newRootKeyWrapper(ctx, log, config.Encryption)
	if err != nil {
		return nil, err
	}

	return &Keyring{
		wrapper:  wrapper,
		rootKeys: map[models.ID[models.RootKey]][]byte{},
		keks:     map[models.ID[models.KeyEncryptionKey]][]byte{},
		db:       db,
		repo:     repository.NewEncryptionRepository(log, clock, db),
	}, nil
}

func (k *Keyring) cacheRootKey(id models.ID[models.RootKey], raw []byte) []byte {
	k.lock.Lock()
	defer k.lock.Unlock()
	if existing, ok := k.rootKeys[id]; ok {
		clear(raw)
		return existing
	}
	k.rootKeys[id] = raw
	return raw
}

func (k *Keyring) cacheKEK(id models.ID[models.KeyEncryptionKey], raw []byte) []byte {
	k.lock.Lock()
	defer k.lock.Unlock()
	if existing, ok := k.keks[id]; ok {
		clear(raw)
		return existing
	}
	k.keks[id] = raw
	return raw
}

func (k *Keyring) rootKey(
	ctx context.Context,
	id models.ID[models.RootKey],
) ([]byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	k.lock.RLock()
	raw, ok := k.rootKeys[id]
	k.lock.RUnlock()
	if ok {
		return raw, nil
	}

	row, err := k.repo.GetRootKeyById(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load root key %s", id)
	}

	raw, err = unwrapRootKey(span.Context(), k.db, k.wrapper, row)
	if err != nil {
		return nil, err
	}
	return k.cacheRootKey(id, raw), nil
}

func (k *Keyring) loadKEK(ctx context.Context, row *models.KeyEncryptionKey) ([]byte, error) {
	// Whatever root key the row points at right now. After a root key rotation
	// that's the new one, and it gets unwrapped lazily here.
	rootKey, err := k.rootKey(ctx, row.RootKeyId)
	if err != nil {
		return nil, err
	}

	aad, err := models.EncryptionAAD(row)
	if err != nil {
		return nil, err
	}

	raw, err := unseal(rootKey, row.Wrapped, aad)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unwrap KEK %s", row.KeyEncryptionKeyId)
	}
	return k.cacheKEK(row.KeyEncryptionKeyId, raw), nil
}

func (k *Keyring) activeKEK(
	ctx context.Context,
) (models.ID[models.KeyEncryptionKey], []byte, error) {
	row, err := k.repo.GetKeyEncryptionKey(ctx) // highest ID
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to load active KEK")
	}

	k.lock.RLock()
	raw, ok := k.keks[row.KeyEncryptionKeyId]
	k.lock.RUnlock()
	if ok {
		return row.KeyEncryptionKeyId, raw, nil
	}

	raw, err = k.loadKEK(ctx, row)
	return row.KeyEncryptionKeyId, raw, err
}

// kek returns the raw KEK for an ID stored on an encrypted row.
func (k *Keyring) kek(ctx context.Context, id models.ID[models.KeyEncryptionKey]) ([]byte, error) {
	k.lock.RLock()
	raw, ok := k.keks[id]
	k.lock.RUnlock()
	if ok {
		return raw, nil
	}

	row, err := k.repo.GetKeyEncryptionKeyById(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load KEK %s", id)
	}
	return k.loadKEK(ctx, row)
}

// currentSecretboxKey decrypts the newest secretbox key through the keyring.
// Any configured key that can open the chain works here, it does not have to be
// the current secretbox key.
func (k *Keyring) currentSecretboxKey(
	ctx context.Context,
) (*models.SecretboxKey, string, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	row, err := k.repo.GetSecretboxKey(span.Context())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to load the current secretbox key")
	}

	aad, err := models.EncryptionAAD(row)
	if err != nil {
		return nil, "", err
	}

	encoded, err := row.Decrypt(span.Context(), k, aad)
	if err != nil {
		return nil, "", errors.Wrapf(err,
			"failed to decrypt the current secretbox key %s with the configured %s key",
			row.SecretboxKeyId, k.wrapper.Kind(),
		)
	}

	return row, encoded, nil
}

// verify decrypts the current secretbox key. That goes through the whole chain:
// the configured wrapper opens the root key, the root key opens the KEK, and
// the KEK opens the DEK. So a wrong secretbox key or a KMS that can't unwrap
// fails here. With secretbox it also checks the decrypted key is the
// configured one, since an old secretbox key can still wrap the root key and
// would decrypt fine.
func (k *Keyring) verify(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	row, encoded, err := k.currentSecretboxKey(span.Context())
	if err != nil {
		return err
	}

	if secretbox, ok := k.wrapper.(*secretboxWrapper); ok {
		current, err := ParseSecretboxKey(encoded)
		if err != nil {
			return errors.Wrap(err, "failed to parse the current secretbox key")
		}
		defer clear(current[:])

		if SecretboxKeyFingerprint(current) != secretbox.fingerprint {
			return errors.Errorf(
				"the configured secretbox key is not the current one (%s), configure the current secretbox key",
				row.SecretboxKeyId,
			)
		}
	}

	return nil
}

// UnwrapDataKey implements [models.DataKeyWrapper]. Reads can use any cached
// KEK: a KEK's raw value never changes for its ID, so a cached one is never
// stale. Only a KEK this keyring hasn't seen yet is loaded from the database.
func (k *Keyring) UnwrapDataKey(
	ctx context.Context,
	kekId models.ID[models.KeyEncryptionKey],
	wrapped []byte,
) ([]byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	kek, err := k.kek(span.Context(), kekId)
	if err != nil {
		return nil, err
	}

	return unsealDataKey(kek, kekId, wrapped)
}

// WrapDataKey implements [models.DataKeyWrapper]. Writes always read the
// highest KEK ID from the database, so a KEK rotated in by another process is
// picked up on the next write without a restart. Only the raw KEK for that ID
// comes from the cache.
func (k *Keyring) WrapDataKey(
	ctx context.Context,
	dek []byte,
) (models.ID[models.KeyEncryptionKey], []byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	kekId, kek, err := k.activeKEK(span.Context())
	if err != nil {
		return "", nil, err
	}

	wrapped, err := sealDataKey(kek, kekId, dek)
	if err != nil {
		return "", nil, err
	}
	return kekId, wrapped, nil
}
