package models

type Account struct {
	tableName string `pg:"accounts"`

	AccountId uint64 `json:"accountId" pg:"account_id,notnull,pk,type:'bigserial'"`
	Timezone  string `json:"timezone" pg:"timezone,notnull,default:'UTC'"`
}
