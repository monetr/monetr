package secrets

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

var (
	_ PlaidSecretsProvider = &databasePlaidSecretProvider{}
)

type databasePlaidSecretProvider struct {
	log *logrus.Entry
	db  bun.IDB
}

func NewDatabasePlaidSecretsProvider(log *logrus.Entry, db bun.IDB) PlaidSecretsProvider {
	return &databasePlaidSecretProvider{
		log: log,
		db:  db,
	}
}

func (p *databasePlaidSecretProvider) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId, accessToken string) error {
	span := sentry.StartSpan(ctx, "UpdateAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()

	token := models.PlaidToken{
		ItemId:      plaidItemId,
		AccountId:   accountId,
		AccessToken: accessToken,
	}
	_, err := p.db.NewInsert().
		Model(&token).
		// TODO Make sure we test this.
		On(`CONFLICT (item_id, account_id) DO UPDATE`).
		Exec(span.Context(), &token)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update access token")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (p *databasePlaidSecretProvider) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, err error) {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()

	var result models.PlaidToken
	err = p.db.NewSelect().
		Model(&result).
		Where(`plaid_token.account_id = ?`, accountId).
		Where(`plaid_token.item_id = ?`, plaidItemId).
		Limit(1).
		Scan(span.Context(), &result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return accessToken, errors.Wrap(err, "failed to retrieve access token for plaid link")
	}

	span.Status = sentry.SpanStatusOK

	return result.AccessToken, nil
}

func (p *databasePlaidSecretProvider) RemoveAccessTokenForPlaidLink(ctx context.Context, accountId uint64, plaidItemId string) error {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"itemId": plaidItemId,
	}

	_, err := p.db.NewDelete().
		Model(&models.PlaidToken{}).
		Where(`plaid_token.account_id = ?`, accountId).
		Where(`plaid_token.item_id = ?`, plaidItemId).
		Exec(span.Context())
	// TODO Add affected rows count assertion, should be 1.
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete plaid access token")
	}

	return nil
}

func (p *databasePlaidSecretProvider) Close() error {
	return nil
}
