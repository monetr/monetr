package models

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkID     uint64   `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	ItemId          string   `json:"-" pg:"item_id,unique,notnull"`
	Products        []string `json:"-" pg:"products,type:'text[]'"`
	WebhookUrl      string   `json:"-" pg:"webhook_url"`
	InstitutionId   string   `json:"-" pg:"institution_id"`
	InstitutionName string   `json:"-" pg:"institution_name"`
	UsePlaidSync    bool     `json:"-" pg:"use_plaid_sync,notnull"`
}
