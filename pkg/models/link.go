package models

type Link struct {
	tableName string `sql:"links"`

	LinkId      uint64   `json:"linkId" sql:"link_id,notnull,pk,type:'bigserial'"`
	AccountId   uint64   `json:"-" sql:"account_id,notnull,pk,on_delete:CASCADE"`
	Account     *Account `json:"-" sql:"rel:has-one"`
	PlaidItemId string   `json:"-" sql:"plaid_item_id,notnull"`
	// TODO (elliotcourant) Allow this access token to be stored elsewhere, such
	//  as vault.
	PlaidAccessToken string `json:"-" sql:"plaid_access_token,notnull"`
	WebhookUrl       string `json:"-" sql:"webhook_url,null"`
}
