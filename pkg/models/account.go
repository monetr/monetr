package models

type Account struct {
	tableName string `sql:"accounts"`

	AccountId     uint64 `json:"accountId" sql:"account_id,notnull,pk,type:'bigserial'"`
	BillingUserId uint64 `json:"billingUserId" sql:"billing_user_id,notnull,on_delete:CASCADE"`
	Timezone      string `json:"timezone" sql:"timezone,notnull,default:'UTC'"`
}
