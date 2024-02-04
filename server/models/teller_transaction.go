package models

import "time"

type TellerTransaction struct {
	tableName string `pg:"teller_transactions"`

	TellerTransactionId uint64             `json:"-" pg:"teller_transaction_id,notnull,pk,type:'bigserial'"`
	AccountId           uint64             `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account             *Account           `json:"-" pg:"rel:has-one"`
	TellerBankAccountId uint64             `json:"-" pg:"teller_bank_account_id,type:'bigint',unique:per_account"`
	TellerBankAccount   *TellerBankAccount `json:"-" pg:"rel:has-one"`
	TellerId            string             `json:"-" pg:"teller_id,notnull,unique:per_account"`
	Name                string             `json:"name" pg:"name,notnull"`
	Category            string             `json:"category" pg:"category"`
	Type                string             `json:"type" pg:"type"`
	Date                time.Time          `json:"date" pg:"date,notnull"`
	IsPending           bool               `json:"isPending" pg:"is_pending,notnull,use_zero"`
	Amount              int64              `json:"amount" pg:"amount,notnull,use_zero"`
	RunningBalance      *int64             `json:"runningBalance" pg:"running_balance"`
	CreatedAt           time.Time          `json:"createdAt" pg:"created_at,notnull,default:now()"`
	UpdatedAt           time.Time          `json:"updatedAt" pg:"updated_at,notnull,default:now()"`
	DeletedAt           *time.Time         `json:"deletedAt" pg:"deleted_at"`
}
