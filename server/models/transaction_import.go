package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
)

// TransactionImportStatus represents the status and "stage" of the transaction
// import workflow. When an import is created it will be created in a pending
// status. The import is then kicked into a background job which will parse the
// file and establish which accounts will be affected or if the file is even
// valid. If the file is not valid it moves to failed immediately. If the file
// is valid then it moves to a confirming status. This status indicates that the
// user must provide some input to continue the file processing. At this stage
// the user must confirm the data, things like the items in the import before
// continuing. Once they have confirmed the import will move into a processing
// state. If the user does not confirm, eventually the import will move to
// expired as the file for the import is automatically removed after 1 hour. If
// any of the import items fail during processing, the entire import is moved to
// a failed status. If all import items succeed then the import is moved to a
// complete status.
type TransactionImportStatus string

const (
	TransactionImportStatusPending    = "pending"
	TransactionImportStatusConfirming = "confirming"
	TransactionImportStatusProcessing = "processing"
	TransactionImportStatusFailed     = "failed"
	TransactionImportStatusComplete   = "complete"
	TransactionImportStatusExpired    = "expired"
)

var (
	_ pg.BeforeInsertHook = (*TransactionImport)(nil)
	_ Identifiable        = TransactionImport{}
	_ Uploadable          = TransactionImport{}
)

// TransactionImport is different from TraansctionUpload, uploads are specific
// to a single bank account. Where as imports are a link level import. Imports
// rely on identifying bank account data should go into, and may import data
// into multiple bank accounts within the same link.
type TransactionImport struct {
	tableName string `pg:"transaction_imports"`

	TransactionImportId ID[TransactionImport]   `json:"transactionImportId" pg:"transaction_import_id,notnull,pk"`
	AccountId           ID[Account]             `json:"-" pg:"account_id,notnull,pk"`
	Account             *Account                `json:"-" pg:"rel:has-one"`
	LinkId              ID[Link]                `json:"linkId" pg:"link_id,notnull"`
	Link                *Link                   `json:"-,omitempty" pg:"rel:has-one"`
	FileId              ID[File]                `json:"fileId" pg:"file_id,notnull"`
	File                *File                   `json:"file,omitempty" pg:"rel:has-one"`
	Status              TransactionImportStatus `json:"status" pg:"status,notnull"`
	Items               []TransactionImportItem `json:"items" pg:"rel:has-many"`
	ExpiresAt           time.Time               `json:"expiresAt" pg:"expires_at,notnull"`
	CreatedAt           time.Time               `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy           ID[User]                `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser       *User                   `json:"-" pg:"rel:has-one,fk:created_by"`
	UpdateAt            time.Time               `json:"updatedAt" pg:"updated_at,notnull"`
}

func (TransactionImport) IdentityPrefix() string {
	return "txim"
}

func (TransactionImport) FileKind() string {
	return "transactions/imports"
}

func (TransactionImport) FileExpiration(clock clock.Clock) *time.Time {
	// TransactionImport files expire after 1 hour be default.
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

// BeforeInsert implements orm.BeforeInsertHook.
func (o *TransactionImport) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionImportId.IsZero() {
		o.TransactionImportId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
