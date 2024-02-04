package models

type PlaidToken struct {
	tableName string `pg:"plaid_tokens"`

	ItemId      string   `pg:"item_id,notnull"`
	AccountId   uint64   `pg:"account_id,notnull"`
	Account     *Account `pg:"rel:has-one"`
	KeyID       *string  `pg:"key_id"`
	Version     *string  `pg:"version"`
	AccessToken string   `pg:"access_token,notnull"`
}
