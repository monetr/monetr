package models

import (
	"time"
)

type Transaction struct {
	tableName string `pg:"transactions"`

	TransactionId             uint64       `json:"transactionId" pg:"transaction_id,notnull,pk,type:'bigserial'"`
	AccountId                 uint64       `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                   *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId             uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,pk,on_delete:CASCADE,type:'bigint',unique:per_bank_account"`
	BankAccount               *BankAccount `json:"bankAccount,omitempty" pg:"rel:has-one"`
	PlaidTransactionId        string       `json:"-" pg:"plaid_transaction_id,unique:per_bank_account"`
	PendingPlaidTransactionId *string      `json:"-" pg:"pending_plaid_transaction_id"`
	Amount                    int64        `json:"amount" pg:"amount,notnull,use_zero"`
	SpendingId                *uint64      `json:"spendingId" pg:"spending_id,on_delete:SET NULL"`
	Spending                  *Spending    `json:"spending,omitempty" pg:"rel:has-one"`
	// SpendingAmount is the amount deducted from the expense this transaction was spent from. This is used when a
	// transaction is more than the expense currently has allocated. If the transaction were to be deleted or changed we
	// want to make sure we return the correct amount to the expense.
	SpendingAmount       *int64     `json:"spendingAmount,omitempty" pg:"spending_amount,use_zero"`
	Categories           []string   `json:"categories" pg:"categories,type:'text[]'"`
	OriginalCategories   []string   `json:"originalCategories" pg:"original_categories,type:'text[]'"`
	Date                 time.Time  `json:"date" pg:"date,notnull"`
	AuthorizedDate       *time.Time `json:"authorizedDate" pg:"authorized_date"`
	Name                 string     `json:"name,omitempty" pg:"name"`
	OriginalName         string     `json:"originalName" pg:"original_name,notnull"`
	MerchantName         string     `json:"merchantName,omitempty" pg:"merchant_name"`
	OriginalMerchantName string     `json:"originalMerchantName" pg:"original_merchant_name"`
	IsPending            bool       `json:"isPending" pg:"is_pending,notnull,use_zero"`
	CreatedAt            time.Time  `json:"createdAt" pg:"created_at,notnull,default:now()"`
}
