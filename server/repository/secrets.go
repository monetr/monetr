package repository

import (
	"context"
	"encoding/hex"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Secret struct {
	SecretId uint64            `json:"-"`
	Kind     models.SecretKind `json:"-"`
	Secret   string            `json:"-"`
}

type SecretsRepository interface {
	Store(ctx context.Context, secret *Secret) error
	Read(ctx context.Context, secretId uint64) (*Secret, error)
	Delete(ctx context.Context, secretId uint64) error
}

type baseSecretsRepository struct {
	accountId uint64
	clock     clock.Clock
	log       *logrus.Entry
	db        pg.DBI
	kms       secrets.KeyManagement
}

func (b *baseSecretsRepository) AccountId() uint64 {
	return b.accountId
}

func NewSecretsRepository(
	log *logrus.Entry,
	clock clock.Clock,
	db pg.DBI,
	kms secrets.KeyManagement,
	accountId uint64,
) SecretsRepository {
	return &baseSecretsRepository{
		accountId: accountId,
		clock:     clock,
		db:        db,
		kms:       kms,
		log:       log,
	}
}

func (b *baseSecretsRepository) Store(ctx context.Context, secret *Secret) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context())

	var item models.Secret
	if secret.SecretId != 0 {
		log = log.WithField("secretId", secret.SecretId)
		err := b.db.ModelContext(span.Context(), &item).
			Where(`"secret"."account_id" = ?`, b.AccountId()).
			Where(`"secret"."secret_id" = ?`, secret.SecretId).
			Limit(1).
			For(`UPDATE`).
			Select(&item)
		if err != nil {
			log.WithError(err).Error("failed to read an existing secret for update")
			return errors.Wrap(err, "failed to retrieve secretfor update")
		}
		log.Trace("found an existing secret to update")
		item.UpdatedAt = b.clock.Now().UTC()
	} else {
		log.Trace("secret does not exist, a new secret will be stored")
		item = models.Secret{
			AccountId: b.AccountId(),
			Kind:      secret.Kind,
			UpdatedAt: b.clock.Now().UTC(),
			CreatedAt: b.clock.Now().UTC(),
		}
	}

	keyId, version, encrypted, err := b.kms.Encrypt(
		span.Context(),
		[]byte(secret.Secret),
	)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to encrypt access token")
	}
	item.AccountId = b.AccountId()
	item.KeyID = keyId
	item.Version = version
	item.Secret = hex.EncodeToString(encrypted)

	query := b.db.ModelContext(span.Context(), &item)
	if item.SecretId == 0 {
		_, err = query.Insert(&item)
	} else {
		_, err = query.WherePK().Update(&item)
	}
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to store secret")
	}

	log = log.WithField("secretId", item.SecretId)
	log.Trace("successfully stored secret")

	span.Status = sentry.SpanStatusOK

	secret.SecretId = item.SecretId
	return nil
}

func (b *baseSecretsRepository) Read(
	ctx context.Context,
	secretId uint64,
) (*Secret, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item models.Secret
	err := b.db.ModelContext(span.Context(), &item).
		Where(`"secret"."account_id" = ?`, b.AccountId()).
		Where(`"secret"."secret_id" = ?`, secretId).
		Limit(1).
		Select(&item)
	if err != nil {
		// TODO Add proper returning of the ErrNotFound here.
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve secret")
	}

	decoded, err := hex.DecodeString(item.Secret)
	if err != nil {
		span.Status = sentry.SpanStatusDataLoss
		return nil, errors.Wrap(err, "failed to hex decode encrypted secret")
	}
	decrypted, err := b.kms.Decrypt(span.Context(), item.KeyID, item.Version, decoded)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to decrypt secret")
	}

	span.Status = sentry.SpanStatusOK

	return &Secret{
		SecretId: secretId,
		Kind:     item.Kind,
		Secret:   string(decrypted),
	}, nil
}

func (b *baseSecretsRepository) Delete(
	ctx context.Context,
	secretId uint64,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := b.db.ModelContext(span.Context(), new(models.Secret)).
		Where(`"secret"."account_id" = ?`, b.AccountId()).
		Where(`"secret"."secret_id" = ?`, secretId).
		Delete()
	return errors.Wrap(err, "failed to delete secret")
}
