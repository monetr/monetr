//+build vault
// When we are using vault we want to store the access token in vault, thus it is excluded here.

package models

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkID     uint64   `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	ItemId          string   `json:"-" pg:"item_id,notnull"`
	Products        []string `json:"-" pg:"products,type:'text[]'"`
	AccessToken     string   `json:"-" pg:"-"` // Omit this column if we are storing it in vault.
	WebhookUrl      string   `json:"-" pg:"webhook_url"`
	InstitutionId   string   `json:"-" pg:"institution_id"`
	InstitutionName string   `json:"-" pg:"institution_name"`
}
