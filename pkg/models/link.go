package models

type Link struct {
	tableName string `pg:"links"`

	LinkId      uint64   `json:"linkId" pg:"link_id,notnull,pk,type:'bigserial'"`
	AccountId   uint64   `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account     *Account `json:"-" pg:"rel:has-one"`
	PlaidItemId string   `json:"-" pg:"plaid_item_id,notnull"`
	// TODO (elliotcourant) Allow this access token to be stored elsewhere, such
	//  as vault.
	PlaidAccessToken string `json:"-" pg:"plaid_access_token,notnull"`
	WebhookUrl       string `json:"-" pg:"webhook_url"`
}
