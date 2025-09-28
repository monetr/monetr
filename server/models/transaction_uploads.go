package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
)

type TransactionUploadStatus string

const (
	TransactionUploadStatusPending    TransactionUploadStatus = "pending"
	TransactionUploadStatusProcessing TransactionUploadStatus = "processing"
	TransactionUploadStatusFailed     TransactionUploadStatus = "failed"
	TransactionUploadStatusComplete   TransactionUploadStatus = "complete"
)

var (
	_ pg.BeforeInsertHook = (*TransactionUpload)(nil)
	_ Identifiable        = TransactionUpload{}
	_ Uploadable          = TransactionUpload{}
)

type TransactionUpload struct {
	tableName string `pg:"transaction_uploads"`

	TransactionUploadId ID[TransactionUpload]   `json:"transactionUploadId" pg:"transaction_upload_id,notnull,pk"`
	AccountId           ID[Account]             `json:"-" pg:"account_id,notnull,pk"`
	Account             *Account                `json:"-" pg:"rel:has-one"`
	BankAccountId       ID[BankAccount]         `json:"bankAccountId" pg:"bank_account_id,notnull"`
	BankAccount         *BankAccount            `json:"-" pg:"rel:has-one"`
	FileId              ID[File]                `json:"fileId" pg:"file_id,notnull"`
	File                *File                   `json:"file,omitempty" pg:"rel:has-one"`
	Status              TransactionUploadStatus `json:"status" pg:"status,notnull"`
	Error               *string                 `json:"error,omitempty" pg:"error"`
	CreatedAt           time.Time               `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy           ID[User]                `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser       *User                   `json:"-" pg:"rel:has-one,fk:created_by"`
	ProcessedAt         *time.Time              `json:"processedAt" pg:"processed_at"`
	CompletedAt         *time.Time              `json:"completedAt" pg:"completed_at"`
}

func (TransactionUpload) FileKind() string {
	return "transactions/uploads"
}

// TransactionUpload files expire after 1 hour be default.
func (TransactionUpload) FileExpiration(clock clock.Clock) *time.Time {
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

func (TransactionUpload) IdentityPrefix() string {
	return "txup"
}

func (o *TransactionUpload) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionUploadId.IsZero() {
		o.TransactionUploadId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
