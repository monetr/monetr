package models

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/datasources/table"
)

type TransactionImportPreview struct {
	tableName string `pg:"transaction_import_previews"`

	TransactionImportPreviewId ID[TransactionImportPreview]   `json:"transactionImportPreviewId" pg:"transaction_import_preview_id,notnull,pk"`
	AccountId                  ID[Account]                    `json:"-" pg:"account_id,notnull,pk"`
	Account                    *Account                       `json:"-" pg:"rel:has-one"`
	BankAccountId              ID[BankAccount]                `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	BankAccount                *BankAccount                   `json:"-" pg:"rel:has-one"`
	TransactionImportId        ID[TransactionImport]          `json:"transactionImportId" pg:"transaction_import_id,notnull"`
	TransactionImport          *TransactionImport             `json:"-" pg:"rel:has-one"`
	Rows                       []TransactionImportPreviewItem `json:"rows" pg:"rows,type:'jsonb'"`
	AvailableBalance           int64                          `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance             int64                          `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	CreatedAt                  time.Time                      `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt                  time.Time                      `json:"updatedAt" pg:"updated_at,notnull"`
}

type TransactionImportPreviewItem struct {
	TransactionImportPreviewItemId string            `json:"itemId"`
	Data                           table.Row         `json:"data"`
	ExistingTransactionIds         []ID[Transaction] `json:"existingTransactionIds"`
}

func (TransactionImportPreview) IdentityPrefix() string {
	return "txmp"
}

func (o *TransactionImportPreview) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionImportPreviewId.IsZero() {
		o.TransactionImportPreviewId = NewID[TransactionImportPreview]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = now
	}

	return ctx, nil
}
