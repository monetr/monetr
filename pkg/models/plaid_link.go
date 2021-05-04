//+build !vault
// When we are not using vault we will be storing the access token on the plaid link record itself.

package models

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkID     uint64   `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	ItemId          string   `json:"-" pg:"item_id,unique,notnull"`
	AccessToken     string   `json:"-" pg:"access_token,notnull"`
	Products        []string `json:"-" pg:"products,type:'text[]'"`
	WebhookUrl      string   `json:"-" pg:"webhook_url"`
	InstitutionId   string   `json:"-" pg:"institution_id"`
	InstitutionName string   `json:"-" pg:"institution_name"`

	Link *Link `json:"-" pg:"rel:has-one"`
}
