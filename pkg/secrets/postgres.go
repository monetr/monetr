package secrets

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
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

func (p *postgresPlaidSecretProvider) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64, accessToken string) error {
	span := sentry.StartSpan(ctx, "UpdateAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()
	panic("implement me")
}

func (p *postgresPlaidSecretProvider) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64) (accessToken string, err error) {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [POSTGRES]")
	defer span.Finish()

	var result plaidLinkWithToken
	err = p.db.ModelContext(span.Context(), &result).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"link"."plaid_link_id" = "plaid_link"."plaid_link_id"`).
		Where(`"link"."account_id" = ?`, accountId).
		Where(`"plaid_link"."plaid_link_id" = ?`, plaidLinkId).
		Limit(1).
		Select(&result)
	if err != nil {
		return accessToken, errors.Wrap(err, "failed to retrieve access token for plaid link")
	}

	return result.AccessToken, nil
}

func (p *postgresPlaidSecretProvider) Close() error {
	panic("implement me")
}
