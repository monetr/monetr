package secrets

import (
	"context"
	"encoding/hex"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ SecretsProvider = &postgresSecretStorage{}
)

type postgresSecretStorage struct {
	clock clock.Clock
	log   *logrus.Entry
	db    pg.DBI
	kms   KeyManagement
}

func NewPostgresSecretsStorage(log *logrus.Entry, db pg.DBI, kms KeyManagement) SecretsProvider {
	myownsanity.ASSERT_NOTNIL(kms, "key management interface must be provided to postgres secret storage")
	return &postgresSecretStorage{
		log: log,
		db:  db,
		kms: kms,
	}
}

func (p *postgresSecretStorage) Store(ctx context.Context, secret *Data) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log.WithField("accountId", secret.AccountId)

	var item models.Secret
	if secret.SecretId != 0 {
		log = log.WithField("secretId", secret.SecretId)
		err := p.db.ModelContext(span.Context(), &item).
			Where(`"secret"."account_id" = ?`, secret.AccountId).
			Where(`"secret"."secret_id" = ?`, secret.SecretId).
			Limit(1).
			For(`UPDATE`).
			Select(&item)
		if err != nil {
			log.WithError(err).Error("failed to read an existing secret for update")
			return errors.Wrap(err, "failed to retrieve secretfor update")
		}
		log.Trace("found an existing secret to update")
		item.UpdatedAt = p.clock.Now().UTC()
	} else {
		log.Trace("secret does not exist, a new secret will be stored")
		item = models.Secret{
			AccountId: secret.AccountId,
			Kind:      secret.Kind,
			UpdatedAt: p.clock.Now().UTC(),
			CreatedAt: p.clock.Now().UTC(),
		}
	}

	keyId, version, encrypted, err := p.kms.Encrypt(
		span.Context(),
		[]byte(secret.Secret),
	)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to encrypt access token")
	}
	item.KeyID = keyId
	item.Version = version
	item.Secret = hex.EncodeToString(encrypted)

	query := p.db.ModelContext(span.Context(), &item)
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

func (p *postgresSecretStorage) Read(
	ctx context.Context,
	secretId, accountId uint64,
) (*Data, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item models.Secret
	err := p.db.ModelContext(span.Context(), &item).
		Where(`"secret"."account_id" = ?`, accountId).
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
	decrypted, err := p.kms.Decrypt(span.Context(), item.KeyID, item.Version, decoded)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to decrypt secret")
	}

	span.Status = sentry.SpanStatusOK

	return &Data{
		SecretId:  secretId,
		AccountId: accountId,
		Kind:      item.Kind,
		Secret:    string(decrypted),
	}, nil
}

func (p *postgresSecretStorage) Delete(
	ctx context.Context,
	secretId, accountId uint64,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := p.db.ModelContext(span.Context(), new(models.Secret)).
		Where(`"secret"."account_id" = ?`, accountId).
		Where(`"secret"."secret_id" = ?`, secretId).
		Delete()
	return errors.Wrap(err, "failed to delete secret")
}
