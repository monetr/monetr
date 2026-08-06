package models

import (
	"github.com/uptrace/bun"
)

type PlaidToken struct {
	bun.BaseModel `bun:"table:plaid_tokens,alias:plaid_token"`

	ItemId      string   `bun:"item_id,notnull,nullzero"`
	AccountId   uint64   `bun:"account_id,notnull,nullzero"`
	Account     *Account `bun:"rel:belongs-to,join:account_id=account_id"`
	KeyID       *string  `bun:"key_id"`
	Version     *string  `bun:"version"`
	AccessToken string   `bun:"access_token,notnull,nullzero"`
}
