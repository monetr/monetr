//go:build !vault
// +build !vault

// When we are not using vault we will be storing the access token on the plaid link record itself.

package models

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkID     uint64   `json:"-" bun:"plaid_link_id,notnull,pk,type:'bigserial'"`
	ItemId          string   `json:"-" bun:"item_id,unique,notnull"`
	Products        []string `json:"-" bun:"products,array"`
	WebhookUrl      string   `json:"-" bun:"webhook_url"`
	InstitutionId   string   `json:"-" bun:"institution_id"`
	InstitutionName string   `json:"-" bun:"institution_name"`
}
