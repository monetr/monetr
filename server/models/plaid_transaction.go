package models

import "time"

type PlaidTransaction struct {
	tableName string `pg:"plaid_transactions"`

	PlaidTransactionId uint64            `json:"plaidTransactionId" pg:"plaid_transaction_id,notnull,pk,type:'bigserial'"`
	AccountId          uint64            `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account            *Account          `json:"-" pg:"rel:has-one"`
	PlaidBankAccountId uint64            `json:"plaidBankAccountId" pg:"plaid_bank_account_id,notnull,type:'bigint',unique:per_bank_account"`
	PlaidBankAccount   *PlaidBankAccount `json:"-" pg:"rel:has-one"`
	PlaidId            string            `json:"-" pg:"plaid_id,notnull,unique:per_bank_account"`
	PendingPlaidId     *string           `json:"-" pg:"pending_plaid_id"`
	Categories         []string          `json:"categories" pg:"categories,type:'text[]'"`
	Date               time.Time         `json:"date" pg:"date,notnull"`
	AuthorizedDate     *time.Time        `json:"authorizedDate" pg:"authorized_date"`
	Name               string            `json:"name,omitempty" pg:"name,notnull"`
	MerchantName       string            `json:"merchantName,omitempty" pg:"merchant_name"`
	Amount             int64             `json:"amount" pg:"amount,notnull,use_zero"`
	Currency           string            `json:"currency" pg:"currency,notnull"`
	IsPending          bool              `json:"isPending" pg:"is_pending,notnull,use_zero"`
	CreatedAt          time.Time         `json:"createdAt" pg:"created_at,notnull,default:now()"`
	DeletedAt          *time.Time        `json:"deletedAt" pg:"deleted_at"`
}
