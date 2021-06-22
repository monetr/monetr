package secrets

import "context"

type PlaidSecretsProvider interface {
	UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId, accessToken string) error
	GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, err error)
	Close() error
}
