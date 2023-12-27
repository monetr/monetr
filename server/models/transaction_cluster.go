package models

import "time"

type TransactionCluster struct {
	tableName string `pg:"transaction_clusters"`

	TransactionClusterId string       `json:"transactionClusterId" pg:"transaction_cluster_id,notnull,pk"`
	AccountId            uint64       `json:"-" pg:"account_id,notnull,type:'bigint'"`
	Account              *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId        uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,type:'bigint'"`
	BankAccount          *BankAccount `json:"-" pg:"rel:has-one"`
	Name                 string       `json:"name" pg:"name,notnull"`
	Members              []uint64     `json:"members" pg:"members,notnull,type:'bigint[]'"`
	CreatedAt            time.Time    `json:"createdAt" pg:"created_at,notnull,default:now()"`
}
