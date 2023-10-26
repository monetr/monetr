package secrets

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("access token not found")
)

type PlaidSecretsProvider interface {
	UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId, accessToken string) error
	GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, err error)
	RemoveAccessTokenForPlaidLink(ctx context.Context, accountId uint64, plaidItemId string) error
	Close() error
}
