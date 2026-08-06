package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// TransactionRule is a starting point for the automation that I want to
// introduce to similar transactions. When you have a group of transactions,
// like say your mortage. You should be able to say "oh yeah just rename that to
// be `Mortgage` every time you see it", or "spend that from the mortgage
// budget". Thats what transaction rules aim to achieve.
type TransactionRule struct {
	bun.BaseModel `bun:"table:transaction_rules,alias:transaction_rule"`

	TransactionRuleId    ID[TransactionRule]    `json:"transactionRuleId" bun:"transaction_rule_id,notnull,pk"`
	AccountId            ID[Account]            `json:"-" bun:"account_id,notnull,pk"`
	Account              *Account               `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId        ID[BankAccount]        `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	BankAccount          *BankAccount           `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	TransactionClusterId ID[TransactionCluster] `json:"transactionClusterId" bun:"transaction_cluster_id,notnull,nullzero"`
	TransactionCluster   *TransactionCluster    `json:"-" bun:"rel:belongs-to,join:transaction_cluster_id=transaction_cluster_id,join:account_id=account_id"`
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
	SpendingId *ID[Spending] `json:"spendingId" bun:"spending_id"`
	Spending   *Spending     `json:"spending,omitempty" bun:"rel:belongs-to,join:spending_id=spending_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`

	CreatedAt time.Time `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:"updated_at,notnull,default:now(),nullzero"`
}

func (TransactionRule) IdentityPrefix() string {
	return "trl"
}

var (
	_ bun.BeforeAppendModelHook = (*TransactionRule)(nil)
)

func (o *TransactionRule) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionRuleId.IsZero() {
			o.TransactionRuleId = NewID[TransactionRule]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
