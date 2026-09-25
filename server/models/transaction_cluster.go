package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
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
	bun.BaseModel `bun:"table:transaction_clusters,alias:transaction_cluster"`

	TransactionClusterId ID[TransactionCluster]        `json:"transactionClusterId" bun:"transaction_cluster_id,notnull,pk"`
	AccountId            ID[Account]                   `json:"-" bun:"account_id,notnull,pk"`
	Account              *Account                      `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId        ID[BankAccount]               `json:"bankAccountId" bun:"bank_account_id,notnull,nullzero"`
	BankAccount          *BankAccount                  `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	Signature            string                        `json:"signature" bun:"signature,nullzero"`
	Centroid             *ID[Transaction]              `json:"centroid" bun:"centroid"`
	Name                 string                        `json:"name" bun:"name,notnull,nullzero"`
	OriginalName         string                        `json:"originalName" bun:"original_name,notnull,nullzero"`
	Members              []ID[Transaction]             `json:"members" bun:"members,notnull,array,nullzero"`
	Debug                []TransactionClusterDebugItem `json:"debug" bun:"debug,type:jsonb,nullzero"`
	Merchant             []TransactionClusterDebugItem `json:"merchant" bun:"merchant,type:jsonb,nullzero"`
	CreatedAt            time.Time                     `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	UpdatedAt            time.Time                     `json:"updatedAt" bun:"updated_at,notnull,default:now(),nullzero"`

	TransactionRules []TransactionRule `json:"rules,omitempty" bun:"rel:has-many,join:transaction_cluster_id=transaction_cluster_id,join:account_id=account_id"`
}

func (TransactionCluster) IdentityPrefix() string {
	return "tcl"
}

var (
	_ bun.BeforeAppendModelHook = (*TransactionCluster)(nil)
)

func (o *TransactionCluster) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionClusterId.IsZero() {
			o.TransactionClusterId = NewID[TransactionCluster]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}

		o.UpdatedAt = now
	}

	return nil
}
