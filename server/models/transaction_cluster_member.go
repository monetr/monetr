package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type TransactionClusterMember struct {
	bun.BaseModel `bun:"table:transaction_cluster_members,alias:transaction_cluster_member"`

	TransactionId        ID[Transaction]        `json:"transactionId" bun:"transaction_id,notnull,pk"`
	Transaction          *Transaction           `json:"-" bun:"rel:belongs-to,join:transaction_id=transaction_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	AccountId            ID[Account]            `json:"-" bun:"account_id,notnull,pk"`
	Account              *Account               `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId        ID[BankAccount]        `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	BankAccount          *BankAccount           `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	TransactionClusterId ID[TransactionCluster] `json:"transactionClusterId" bun:"transaction_cluster_id,notnull"`
	TransactionCluster   *TransactionCluster    `json:"-" bun:"rel:belongs-to,join:transaction_cluster_id=transaction_cluster_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	CreatedAt            time.Time              `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	UpdatedAt            time.Time              `json:"updatedAt" bun:"updated_at,notnull,default:now(),nullzero"`
}

var (
	_ bun.BeforeAppendModelHook = (*TransactionClusterMember)(nil)
)

// BeforeAppendModel implements [bun.BeforeAppendModelHook].
func (o *TransactionClusterMember) BeforeAppendModel(
	ctx context.Context,
	query bun.Query,
) error {
	switch query.(type) {
	case *bun.InsertQuery:
		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}

		o.UpdatedAt = now
	}

	return nil
}
