package repository

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type EncryptionRepository interface {
	CreateKeyEncryptionKey(ctx context.Context, key *models.KeyEncryptionKey) error
	CreateRootKey(ctx context.Context, key *models.RootKey) error
	CreateSecretboxKey(ctx context.Context, key *models.SecretboxKey) error
	GetKeyEncryptionKey(ctx context.Context) (*models.KeyEncryptionKey, error)
	GetKeyEncryptionKeyById(ctx context.Context, id models.ID[models.KeyEncryptionKey]) (*models.KeyEncryptionKey, error)
	GetRootKey(ctx context.Context) (*models.RootKey, error)
	GetRootKeyById(ctx context.Context, id models.ID[models.RootKey]) (*models.RootKey, error)
	GetSecretboxKey(ctx context.Context) (*models.SecretboxKey, error)
}

type encryptionRepositoryBase struct {
	clock clock.Clock
	db    bun.IDB
	log   *slog.Logger
}

func NewEncryptionRepository(
	log *slog.Logger,
	clock clock.Clock,
	db bun.IDB,
) EncryptionRepository {
	return &encryptionRepositoryBase{
		clock: clock,
		db:    db,
		log:   log,
	}
}

// CreateKeyEncryptionKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) CreateKeyEncryptionKey(
	ctx context.Context,
	key *models.KeyEncryptionKey,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := e.db.NewInsert().
		Model(key).
		Returning("*").
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to store key encryption key")
	}

	e.log.DebugContext(span.Context(), "stored key encryption key",
		"keyEncryptionKeyId", key.KeyEncryptionKeyId,
		"rootKeyId", key.RootKeyId,
	)

	span.Status = sentry.SpanStatusOK

	return nil
}

// CreateRootKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) CreateRootKey(
	ctx context.Context,
	key *models.RootKey,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := e.db.NewInsert().
		Model(key).
		Returning("*").
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to store root key")
	}

	e.log.DebugContext(span.Context(), "stored key root key",
		"rootKeyId", key.RootKeyId,
		"fingerprint", key.Fingerprint,
		"wrappers", myownsanity.Map(key.Wrappers, func(item models.RootKeyWrapper) map[string]any {
			return map[string]any{
				"kind":      item.Kind,
				"reference": item.Reference,
				"version":   item.Version,
			}
		}),
	)

	span.Status = sentry.SpanStatusOK

	return nil
}

// CreateSecretboxKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) CreateSecretboxKey(ctx context.Context, key *models.SecretboxKey) error {
	panic("unimplemented")
}

// GetKeyEncryptionKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) GetKeyEncryptionKey(ctx context.Context) (*models.KeyEncryptionKey, error) {
	panic("unimplemented")
}

// GetKeyEncryptionKeyById implements [EncryptionRepository].
func (e *encryptionRepositoryBase) GetKeyEncryptionKeyById(ctx context.Context, id models.ID[models.KeyEncryptionKey]) (*models.KeyEncryptionKey, error) {
	panic("unimplemented")
}

// GetRootKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) GetRootKey(ctx context.Context) (*models.RootKey, error) {
	panic("unimplemented")
}

// GetRootKeyById implements [EncryptionRepository].
func (e *encryptionRepositoryBase) GetRootKeyById(ctx context.Context, id models.ID[models.RootKey]) (*models.RootKey, error) {
	panic("unimplemented")
}

// GetSecretboxKey implements [EncryptionRepository].
func (e *encryptionRepositoryBase) GetSecretboxKey(ctx context.Context) (*models.SecretboxKey, error) {
	panic("unimplemented")
}
