package secrets

import (
	"context"
	"encoding/hex"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ PlaidSecretsProvider = &postgresPlaidSecretProvider{}
)

type postgresPlaidSecretProvider struct {
	log *logrus.Entry
	db  pg.DBI
	kms KeyManagement
}

func NewPostgresPlaidSecretsProvider(log *logrus.Entry, db pg.DBI, kms KeyManagement) PlaidSecretsProvider {
	return &postgresPlaidSecretProvider{
		log: log,
		db:  db,
		kms: kms,
	}
}

func (p *postgresPlaidSecretProvider) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId, accessToken string) error {
	span := sentry.StartSpan(ctx, "UpdateAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()

	token := models.PlaidToken{
		ItemId:      plaidItemId,
		AccountId:   accountId,
		AccessToken: accessToken,
	}

	if p.kms != nil {
		span.Data = map[string]interface{}{
			"kms": true,
		}
		keyId, version, encrypted, err := p.kms.Encrypt(span.Context(), []byte(accessToken))
		if err != nil {
			span.Status = sentry.SpanStatusInternalError
			return errors.Wrap(err, "failed to encrypt access token")
		}

		token.KeyID = &keyId
		if version != "" {
			token.Version = &version
		}
		token.AccessToken = hex.EncodeToString(encrypted)
	} else {
		span.Data = map[string]interface{}{
			"kms": false,
		}
	}

	_, err := p.db.ModelContext(span.Context(), &token).
		OnConflict(`(item_id, account_id) DO UPDATE`).
		Insert(&token)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update access token")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (p *postgresPlaidSecretProvider) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, err error) {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()

	var result models.PlaidToken
	err = p.db.ModelContext(span.Context(), &result).
		Where(`"plaid_token"."account_id" = ?`, accountId).
		Where(`"plaid_token"."item_id" = ?`, plaidItemId).
		Limit(1).
		Select(&result)
	if err != nil {
		// TODO Add proper returning of the ErrNotFound here.
		span.Status = sentry.SpanStatusInternalError
		return accessToken, errors.Wrap(err, "failed to retrieve access token for plaid link")
	}

	if p.kms != nil && result.KeyID != nil {
		span.Data = map[string]interface{}{
			"kms": true,
		}
		version := ""
		if result.Version != nil {
			version = *result.Version
		}
		decoded, err := hex.DecodeString(result.AccessToken)
		if err != nil {
			span.Status = sentry.SpanStatusDataLoss
			return accessToken, errors.Wrap(err, "failed to hex decode encrypted access token")
		}
		decrypted, err := p.kms.Decrypt(span.Context(), *result.KeyID, version, decoded)
		if err != nil {
			span.Status = sentry.SpanStatusInternalError
			return accessToken, errors.Wrap(err, "failed to encrypt access token")
		}

		span.Status = sentry.SpanStatusOK

		return string(decrypted), nil
	} else if p.kms != nil {
		crumbs.Debug(span.Context(), "Not decrypting using KMS because access token was not encrypted", nil)
		span.Data = map[string]interface{}{
			"kms": false,
		}
	} else {
		span.Data = map[string]interface{}{
			"kms": false,
		}
	}

	span.Status = sentry.SpanStatusOK

	return result.AccessToken, nil
}

func (p *postgresPlaidSecretProvider) RemoveAccessTokenForPlaidLink(ctx context.Context, accountId uint64, plaidItemId string) error {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"itemId": plaidItemId,
	}

	_, err := p.db.ModelContext(span.Context(), &models.PlaidToken{}).
		Where(`"plaid_token"."account_id" = ?`, accountId).
		Where(`"plaid_token"."item_id" = ?`, plaidItemId).
		Limit(1).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete plaid access token")
	}

	return nil
}

func (p *postgresPlaidSecretProvider) Close() error {
	return nil
}
