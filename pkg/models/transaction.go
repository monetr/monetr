package models

import (
	"time"
)

type Transaction struct {
	tableName string `sql:"transactions"`

	TransactionId        uint64       `json:"transactionId" sql:"transaction_id,notnull,pk,type:'bigserial'"`
	AccountId            uint64       `json:"-" sql:"account_id,notnull,pk,on_delete:CASCADE"`
	Account              *Account     `json:"-" sql:"rel:has-one"`
	BankAccountId        uint64       `json:"bankAccountId" sql:"bank_account_id,notnull,pk,on_delete:CASCADE"`
	BankAccount          *BankAccount `json:"bankAccount,omitempty" sql:"rel:has-one"`
	PlaidTransactionId   string       `json:"-" sql:"plaid_transaction_id,notnull"`
	Amount               int64        `json:"amount" sql:"amount,notnull,use_zero"`
	ExpenseId            *uint64      `json:"expenseId" sql:"expense_id,null,on_delete:SET NULL"`
	Expense              *Expense     `json:"expense,omitempty" sql:"rel:has-one"`
	Categories           []string     `json:"categories" sql:"categories,null"`
	OriginalCategories   []string     `json:"originalCategories" sql:"original_categories,null"`
	Date                 time.Time    `json:"date" sql:"date,notnull,type:'date'"`
	AuthorizedDate       *time.Time   `json:"authorizedDate" sql:"authorized_date,null,type:'date'"`
	Name                 string       `json:"name,omitempty" sql:"name,null"`
	OriginalName         string       `json:"originalName" sql:"original_name,notnull"`
	MerchantName         string       `json:"merchantName,omitempty" sql:"merchant_name,null"`
	OriginalMerchantName string       `json:"originalMerchantName" sql:"original_merchant_name,null"`
	IsPending            bool         `json:"isPending" sql:"is_pending,notnull,use_zero"`
	CreatedAt            time.Time    `json:"createdAt" sql:"created_at,notnull,default:now()"`
}
