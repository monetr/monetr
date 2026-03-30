package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type TransactionClusterMember struct {
	tableName string `pg:"transaction_cluster_members"`

	TransactionId        ID[Transaction]        `json:"transactionId" pg:"transaction_id,notnull,pk"`
	Transaction          *Transaction           `json:"-" pg:"rel:has-one"`
	AccountId            ID[Account]            `json:"-" pg:"account_id,notnull,pk"`
	Account              *Account               `json:"-" pg:"rel:has-one"`
	BankAccountId        ID[BankAccount]        `json:"bankAccountId" pg:"bank_account_id,pk,notnull"`
	BankAccount          *BankAccount           `json:"-" pg:"rel:has-one"`
	TransactionClusterId ID[TransactionCluster] `json:"transactionClusterId" pg:"transaction_cluster_id,notnull"`
	TransactionCluster   *TransactionCluster    `json:"-" pg:"rel:has-one"`
	CreatedAt            time.Time              `json:"createdAt" pg:"created_at,notnull,default:now()"`
	UpdatedAt            time.Time              `json:"updatedAt" pg:"updated_at,notnull,default:now()"`
}

var (
	_ pg.BeforeInsertHook = (*TransactionClusterMember)(nil)
)

// BeforeInsert implements [orm.BeforeInsertHook].
func (o *TransactionClusterMember) BeforeInsert(
	ctx context.Context,
) (context.Context, error) {
	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	o.UpdatedAt = now

	return ctx, nil
}
