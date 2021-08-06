package secrets

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ PlaidSecretsProvider = &postgresPlaidSecretProvider{}
)

type postgresPlaidSecretProvider struct {
	log *logrus.Entry
	db  pg.DBI
}

func NewPostgresPlaidSecretsProvider(log *logrus.Entry, db pg.DBI) PlaidSecretsProvider {
	return &postgresPlaidSecretProvider{
		log: log,
		db:  db,
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
		span.Status = sentry.SpanStatusInternalError
		return accessToken, errors.Wrap(err, "failed to retrieve access token for plaid link")
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
