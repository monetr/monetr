package secrets

import "context"

type TellerSecretsProvider interface {
	UpdateAccessTokenForTellerLinkId(ctx context.Context, accountId uint64, tellerLinkId uint64, accessToken string) error
	GetAccessTokenForTellerLinkId(ctx context.Context, accountId uint64, tellerLinkId uint64) (accessToken string, err error)
	RemoveAccessTokenForTellerLinkId(ctx context.Context, accountId uint64, tellerLinkId uint64) error
	Close() error
}
