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

	aad, err := models.EncryptionAAD(k.db, row)
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

// UnwrapDataKey implements [models.DataKeyWrapper].
func (k *Keyring) UnwrapDataKey(
	ctx context.Context,
	kekId models.ID[models.KeyEncryptionKey],
	wrapped []byte,
) ([]byte, error) {
	panic("unimplemented")
}

// WrapDataKey implements [models.DataKeyWrapper].
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
