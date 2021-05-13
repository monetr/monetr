package secrets

import "context"

type PlaidSecretsProvider interface {
	UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64, accessToken string) error
	GetAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64) (accessToken string, err error)

	Close() error
}
