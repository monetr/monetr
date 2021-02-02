package models

type Link struct {
	tableName string `pg:"links"`

	LinkId      uint64   `json:"linkId" pg:"link_id,notnull,pk,type:'bigserial'"`
	AccountId   uint64   `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account     *Account `json:"-" pg:"rel:has-one"`
	PlaidItemId string   `json:"-" pg:"plaid_item_id,notnull"`
	// TODO (elliotcourant) Allow this access token to be stored elsewhere, such
	//  as vault.
	PlaidAccessToken      string   `json:"-" pg:"plaid_access_token,notnull"`
	PlaidProducts         []string `json:"-" pg:"plaid_products,notnull,type:'text[]'"`
	WebhookUrl            string   `json:"-" pg:"webhook_url"`
	InstitutionId         string   `json:"institutionId" pg:"institution_id,notnull"`
	InstitutionName       string   `json:"institutionName" pg:"institution_name,notnull"`
	CustomInstitutionName string   `json:"customInstitutionName,omitempty" pg:"custom_institution_name"`
	CreatedByUserId       uint64   `json:"createdByUserId" pg:"created_by_user_id,notnull,on_delete:CASCADE"`
	CreatedByUser         *User    `json:"createdByUser,omitempty" pg:"rel:has-one,fk:created_by_user_id"`
}
