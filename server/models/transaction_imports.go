package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/uptrace/bun"
)

// TransactionImportStatus covers all of the different states of a transaction
// import. When it is initially created it will be in a `mapping` status.
// 1. `mapping`
// 2. `pending-preview` (user)
// 3. `preview`
// 4. `pending-processing` (user)
// 5. `processing`
// 6. `complete` | `failed`
// 0. `expired`
// Imports move to expired if they remain in a mapping, preview, or processing
// status longer than the file is available in storage.
// Imports are moved to pending preview or pending processing by the user when
// they want to proceed with the import after mapping and confirmation.
type TransactionImportStatus string

const (
	TransactionImportStatusMapping           TransactionImportStatus = "mapping"
	TransactionImportStatusPendingPreview    TransactionImportStatus = "pending-preview"
	TransactionImportStatusPreview           TransactionImportStatus = "preview"
	TransactionImportStatusPendingProcessing TransactionImportStatus = "pending-processing"
	TransactionImportStatusProcessing        TransactionImportStatus = "processing"
	TransactionImportStatusFailed            TransactionImportStatus = "failed"
	TransactionImportStatusComplete          TransactionImportStatus = "complete"
	TransactionImportStatusExpired           TransactionImportStatus = "expired"
)

type TransactionImport struct {
	bun.BaseModel `bun:"table:transaction_imports,alias:transaction_import"`

	TransactionImportId        ID[TransactionImport]         `json:"transactionImportId" bun:"transaction_import_id,notnull,pk"`
	AccountId                  ID[Account]                   `json:"-" bun:"account_id,notnull,pk"`
	Account                    *Account                      `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId              ID[BankAccount]               `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	BankAccount                *BankAccount                  `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	FileId                     ID[File]                      `json:"fileId" bun:"file_id,notnull,nullzero"`
	File                       *File                         `json:"file,omitempty" bun:"rel:belongs-to,join:file_id=file_id,join:account_id=account_id"`
	TransactionImportMappingId *ID[TransactionImportMapping] `json:"transactionImportMappingId" bun:"transaction_import_mapping_id"`
	TransactionImportMapping   *TransactionImportMapping     `json:"transactionImportMapping,omitempty" bun:"rel:belongs-to,join:transaction_import_mapping_id=transaction_import_mapping_id,join:account_id=account_id"`
	Headers                    []string                      `json:"headers" bun:"headers,notnull,array,nullzero"`
	Delimeter                  string                        `json:"delimeter" bun:"delimeter,notnull,nullzero"`
	Status                     TransactionImportStatus       `json:"status" bun:"status,notnull,nullzero"`
	CreatedAt                  time.Time                     `json:"createdAt" bun:"created_at,notnull,nullzero"`
	UpdatedAt                  time.Time                     `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	CompletedAt                *time.Time                    `json:"completedAt" bun:"completed_at"`
	CreatedBy                  ID[User]                      `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser              *User                         `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
}

func (TransactionImport) FileKind() string {
	return "transactions/imports"
}

// FileExpiration will return when a TransactionImport's uploaded file should
// expire, this is one hour by default.
func (TransactionImport) FileExpiration(clock clock.Clock) *time.Time {
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

func (TransactionImport) IdentityPrefix() string {
	return "txim"
}

func (o *TransactionImport) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionImportId.IsZero() {
			o.TransactionImportId = NewID[TransactionImport]()
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
