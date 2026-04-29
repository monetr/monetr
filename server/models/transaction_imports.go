package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
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
	tableName string `pg:"transaction_imports"`

	TransactionImportId        ID[TransactionImport]         `json:"transactionImportId" pg:"transaction_import_id,notnull,pk"`
	AccountId                  ID[Account]                   `json:"-" pg:"account_id,notnull,pk"`
	Account                    *Account                      `json:"-" pg:"rel:has-one"`
	BankAccountId              ID[BankAccount]               `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	BankAccount                *BankAccount                  `json:"-" pg:"rel:has-one"`
	FileId                     ID[File]                      `json:"fileId" pg:"file_id,notnull"`
	File                       *File                         `json:"file,omitempty" pg:"rel:has-one"`
	TransactionImportMappingId *ID[TransactionImportMapping] `json:"transactionImportMappingId" pg:"transaction_import_mapping_id"`
	TransactionImportMapping   *TransactionImportMapping     `json:"transactionImportMapping,omitempty" pg:"rel:has-one"`
	Headers                    []string                      `json:"headers" pg:"headers,notnull,type:'text[]'"`
	Status                     TransactionImportStatus       `json:"status" pg:"status,notnull"`
	CreatedAt                  time.Time                     `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt                  time.Time                     `json:"updatedAt" pg:"updated_at,notnull"`
	CompletedAt                *time.Time                    `json:"completedAt" pg:"completed_at"`
	CreatedBy                  ID[User]                      `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser              *User                         `json:"-" pg:"rel:has-one,fk:created_by"`
}

func (TransactionImport) FileKind() string {
	return "transactions/imports"
}

// TransactionImport files expire after 1 hour be default.
func (TransactionImport) FileExpiration(clock clock.Clock) *time.Time {
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

func (TransactionImport) IdentityPrefix() string {
	return "txim"
}

func (o *TransactionImport) BeforeInsert(ctx context.Context) (context.Context, error) {
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

	return ctx, nil
}

func (TransactionImport) PatchSchemas() []validation.MapRule {
	return []validation.MapRule{
		// When we do not yet have a mapping then they must specify a new mapping ID
		// and say that we are moving to the pending preview status.
		validation.Map(
			validation.Key(
				"transactionImportMappingId",
				ValidID[TransactionImportMapping]().Error("Transaction import mapping ID must be valid if provided"),
			).Required(validators.Require),
			validation.Key(
				"status",
				validators.In(string(TransactionImportStatusPendingPreview)),
			).Required(validators.Require),
		),
		// Otherwise the user can only progress the import to pending processing.
		validation.Map(
			validation.Key(
				"status",
				validators.In(string(TransactionImportStatusPendingProcessing)),
			).Required(validators.Require),
		),
	}
}
