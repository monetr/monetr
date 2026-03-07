package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type TransactionClusterDebugItem struct {
	Word      string  `json:"word"`
	Sanitized string  `json:"sanitized"`
	Order     float32 `json:"order"`
	Value     float32 `json:"value"`
	Rank      float32 `json:"rank"`
	Count     float32 `json:"count"`
}

type TransactionCluster struct {
	tableName string `pg:"transaction_clusters"`

	TransactionClusterId ID[TransactionCluster]        `json:"transactionClusterId" pg:"transaction_cluster_id,notnull,pk"`
	AccountId            ID[Account]                   `json:"-" pg:"account_id,notnull,pk"`
	Account              *Account                      `json:"-" pg:"rel:has-one"`
	BankAccountId        ID[BankAccount]               `json:"bankAccountId" pg:"bank_account_id,notnull"`
	BankAccount          *BankAccount                  `json:"-" pg:"rel:has-one"`
	Signature            string                        `json:"signature" pg:"signature"`
	Centroid             *ID[Transaction]              `json:"centroid" pg:"centroid"`
	Name                 string                        `json:"name" pg:"name,notnull"`
	OriginalName         string                        `json:"originalName" pg:"original_name,notnull"`
	Members              []ID[Transaction]             `json:"members" pg:"members,notnull,type:'varchar(32)[]'"`
	Debug                []TransactionClusterDebugItem `json:"debug" pg:"debug,type:'jsonb'"`
	Merchant             []TransactionClusterDebugItem `json:"merchant" pg:"merchant,type:'jsonb'"`
	CreatedAt            time.Time                     `json:"createdAt" pg:"created_at,notnull,default:now()"`
	UpdatedAt            time.Time                     `json:"updatedAt" pg:"updated_at,notnull,default:now()"`

	TransactionRules []TransactionRule `json:"rules,omitempty" pg:"rel:has-many"`
}

func (TransactionCluster) IdentityPrefix() string {
	return "tcl"
}

var (
	_ pg.BeforeInsertHook = (*TransactionCluster)(nil)
)

func (o *TransactionCluster) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionClusterId.IsZero() {
		o.TransactionClusterId = NewID[TransactionCluster]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	o.UpdatedAt = now

	return ctx, nil
}
