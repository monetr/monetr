package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

// TransactionRule is a starting point for the automation that I want to
// introduce to similar transactions. When you have a group of transactions,
// like say your mortage. You should be able to say "oh yeah just rename that to
// be `Mortgage` every time you see it", or "spend that from the mortgage
// budget". Thats what transaction rules aim to achieve.
type TransactionRule struct {
	tableName string `pg:"transaction_rules"`

	TransactionRuleId    ID[TransactionRule]    `json:"transactionRuleId" pg:"transaction_rule_id,notnull,pk"`
	AccountId            ID[Account]            `json:"-" pg:"account_id,notnull,pk"`
	Account              *Account               `json:"-" pg:"rel:has-one"`
	BankAccountId        ID[BankAccount]        `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	BankAccount          *BankAccount           `json:"-" pg:"rel:has-one"`
	TransactionClusterId ID[TransactionCluster] `json:"transactionClusterId" pg:"transaction_cluster_id,notnull"`
	TransactionCluster   *TransactionCluster    `json:"-" pg:"rel:has-one"`
	// Actions, if a field here is not nil then its action will be applied to
	// transactions.
	// Name indicates that transactions added to this cluster should be renamed.
	// This only effects transactions who's created at date is greater than the
	// created at of this rule.
	Name *string
	// SpendingId indicates that transactions added to this cluster should be
	// assigned to this spending. This only applies to debit transactions and will
	// only be applied to transactions whos created at is greater than the created
	// at of the transaction rule, as well as if the transaction does not already
	// have a spending object associated with it.
	SpendingId *ID[Spending] `json:"spendingId" pg:"spending_id"`
	Spending   *Spending     `json:"spending,omitempty" pg:"rel:has-one"`

	CreatedAt time.Time `json:"createdAt" pg:"created_at,notnull,default:now()"`
	UpdatedAt time.Time `json:"updatedAt" pg:"updated_at,notnull,default:now()"`
}

func (TransactionRule) IdentityPrefix() string {
	return "trl"
}

var (
	_ pg.BeforeInsertHook = (*TransactionRule)(nil)
)

func (o *TransactionRule) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionRuleId.IsZero() {
		o.TransactionRuleId = NewID[TransactionRule]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
