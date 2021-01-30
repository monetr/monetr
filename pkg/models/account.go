package models

type Account struct {
	tableName string `pg:"accounts"`

	AccountId     uint64 `json:"accountId" pg:"account_id,notnull,pk,type:'bigserial'"`
	BillingUserId uint64 `json:"billingUserId" pg:"billing_user_id,notnull,on_delete:CASCADE"`
	Timezone      string `json:"timezone" pg:"timezone,notnull,default:'UTC'"`
}
