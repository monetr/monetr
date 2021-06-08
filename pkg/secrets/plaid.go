package secrets

import "context"

type PlaidSecretsProvider interface {
	UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64, accessToken string) error
	GetAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64) (accessToken string, err error)

	Close() error
}

type plaidLinkWithToken struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkID uint64 `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	ItemId      string `json:"-" pg:"item_id,unique,notnull"`
	AccessToken string `json:"-" pg:"access_token,notnull"`
}
