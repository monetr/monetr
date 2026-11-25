package models

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TransactionImportPreview)(nil)
	_ Identifiable              = TransactionImportPreview{}
)

type TransactionImportPreview struct {
	bun.BaseModel `bun:"table:transaction_import_previews,alias:transaction_import_preview"`

	TransactionImportPreviewId ID[TransactionImportPreview]   `json:"transactionImportPreviewId" bun:"transaction_import_preview_id,notnull,pk"`
	AccountId                  ID[Account]                    `json:"-" bun:"account_id,notnull,pk"`
	Account                    *Account                       `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId              ID[BankAccount]                `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	BankAccount                *BankAccount                   `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	TransactionImportId        ID[TransactionImport]          `json:"transactionImportId" bun:"transaction_import_id,notnull"`
	TransactionImport          *TransactionImport             `json:"-" bun:"rel:belongs-to,join:transaction_import_id=transaction_import_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	Rows                       []TransactionImportPreviewItem `json:"rows" bun:"rows,notnull,type:jsonb"`
	AvailableBalance           int64                          `json:"availableBalance" bun:"available_balance,notnull"`
	CurrentBalance             int64                          `json:"currentBalance" bun:"current_balance,notnull"`
	CreatedAt                  time.Time                      `json:"createdAt" bun:"created_at,notnull,nullzero"`
	UpdatedAt                  time.Time                      `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
}

type TransactionImportPreviewItem struct {
	TransactionImportPreviewItemId string            `json:"itemId"`
	Data                           table.Row         `json:"data"`
	ExistingTransactionIds         []ID[Transaction] `json:"existingTransactionIds"`
}

func (TransactionImportPreview) IdentityPrefix() string {
	return "txmp"
}

func (o *TransactionImportPreview) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
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
	}

	return nil
}
