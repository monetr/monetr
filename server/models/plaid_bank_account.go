package models

import "time"

type PlaidBankAccount struct {
	tableName string `pg:"plaid_bank_accounts"`

	PlaidBankAccountId uint64     `json:"-" pg:"plaid_bank_account_id,notnull,pk,type:'bigserial'"`
	AccountId          uint64     `json:"-" pg:"account_id,notnull,pk"`
	Account            *Account   `json:"-" pg:"rel:has-one"`
	PlaidLinkId        uint64     `json:"-" pg:"plaid_link_id,type:'bigint'"`
	PlaidLink          *PlaidLink `json:"-" pg:"rel:has-one"`
	PlaidId            string     `json:"-" pg:"plaid_id,notnull"`
	Name               string     `json:"name" pg:"name,notnull"`
	OfficialName       string     `json:"officialName" pg:"official_name"`
	Mask               string     `json:"mask" pg:"mask"`
	AvailableBalance   int64      `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance     int64      `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	LimitBalance       *int64     `json:"limitBalance" pg:"limit_balance,use_zero"`
	CreatedAt          time.Time  `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId    uint64     `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser      *User      `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
