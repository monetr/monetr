package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type TransactionClusterDebugItem struct {
	Word      string  `json:"word"`
	Sanitized string  `json:"sanitized"`
	Order     int     `json:"order"`
	Value     float32 `json:"value"`
	Rank      float32 `json:"rank"`
}

type TransactionCluster struct {
	tableName string `pg:"transaction_clusters"`

	TransactionClusterId ID[TransactionCluster]        `json:"transactionClusterId" pg:"transaction_cluster_id,notnull,pk"`
	AccountId            ID[Account]                   `json:"-" pg:"account_id,notnull"`
	Account              *Account                      `json:"-" pg:"rel:has-one"`
	BankAccountId        ID[BankAccount]               `json:"bankAccountId" pg:"bank_account_id,notnull"`
	BankAccount          *BankAccount                  `json:"-" pg:"rel:has-one"`
	Signature            string                        `json:"signature" pg:"signature"`
	Name                 string                        `json:"name" pg:"name,notnull"`
	Members              []ID[Transaction]             `json:"members" pg:"members,notnull,type:'varchar(32)[]'"`
	Debug                []TransactionClusterDebugItem `json:"debug" pg:"debug,type:'jsonb'"`
	Merchant             []TransactionClusterDebugItem `json:"merchant" pg:"merchant,type:'jsonb'"`
	CreatedAt            time.Time                     `json:"createdAt" pg:"created_at,notnull,default:now()"`
}

func (TransactionCluster) IdentityPrefix() string {
	return "tcl"
}

var (
	_ pg.BeforeInsertHook = (*TransactionCluster)(nil)
)

func (o *TransactionCluster) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionClusterId.IsZero() {
		o.TransactionClusterId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
