package repository

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SecretData struct {
	SecretId ID[Secret] `json:"-"`
	Kind     SecretKind `json:"-"`
	Value    string     `json:"-"`
}

type SecretsRepository interface {
	Store(ctx context.Context, secret *SecretData) error
	Read(ctx context.Context, secretId ID[Secret]) (*SecretData, error)
	Delete(ctx context.Context, secretId ID[Secret]) error
}

type baseSecretsRepository struct {
	accountId ID[Account]
	clock     clock.Clock
	log       *logrus.Entry
	db        pg.DBI
	kms       secrets.KeyManagement
}

func (b *baseSecretsRepository) AccountId() ID[Account] {
	return b.accountId
}

func NewSecretsRepository(
	log *logrus.Entry,
	clock clock.Clock,
	db pg.DBI,
	kms secrets.KeyManagement,
	accountId ID[Account],
) SecretsRepository {
	return &baseSecretsRepository{
		accountId: accountId,
		clock:     clock,
		db:        db,
		kms:       kms,
		log:       log,
	}
}

// Store will take a secret that may be existing or may not be. If the ID of the
// secret is unset then a new secret will be created and an ID will be
// generated. If the ID is set, then the existing secret in the database will be
// updated.
func (b *baseSecretsRepository) Store(ctx context.Context, secret *SecretData) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context())

	var item Secret
	if !secret.SecretId.IsZero() {
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
		item = Secret{
			AccountId: b.AccountId(),
			Kind:      secret.Kind,
			UpdatedAt: b.clock.Now().UTC(),
			CreatedAt: b.clock.Now().UTC(),
		}
	}

	keyId, version, encrypted, err := b.kms.Encrypt(
		span.Context(),
		secret.Value,
	)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to encrypt access token")
	}
	item.AccountId = b.AccountId()
	item.KeyID = keyId
	item.Version = version
	item.Secret = encrypted

	query := b.db.ModelContext(span.Context(), &item)
	if item.SecretId.IsZero() {
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
	secretId ID[Secret],
) (*SecretData, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item Secret
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

	decrypted, err := b.kms.Decrypt(span.Context(), item.KeyID, item.Version, item.Secret)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to decrypt secret")
	}

	span.Status = sentry.SpanStatusOK

	return &SecretData{
		SecretId: secretId,
		Kind:     item.Kind,
		Value:    string(decrypted),
	}, nil
}

func (b *baseSecretsRepository) Delete(
	ctx context.Context,
	secretId ID[Secret],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := b.db.ModelContext(span.Context(), new(Secret)).
		Where(`"secret"."account_id" = ?`, b.AccountId()).
		Where(`"secret"."secret_id" = ?`, secretId).
		Delete()
	return errors.Wrap(err, "failed to delete secret")
}
