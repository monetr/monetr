package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/uptrace/bun"
)

type TransactionUploadStatus string

const (
	TransactionUploadStatusPending    TransactionUploadStatus = "pending"
	TransactionUploadStatusProcessing TransactionUploadStatus = "processing"
	TransactionUploadStatusFailed     TransactionUploadStatus = "failed"
	TransactionUploadStatusComplete   TransactionUploadStatus = "complete"
)

var (
	_ bun.BeforeAppendModelHook = (*TransactionUpload)(nil)
	_ Identifiable              = TransactionUpload{}
)

type TransactionUpload struct {
	bun.BaseModel `bun:"table:transaction_uploads,alias:transaction_upload"`

	TransactionUploadId ID[TransactionUpload]   `json:"transactionUploadId" bun:"transaction_upload_id,notnull,pk"`
	AccountId           ID[Account]             `json:"-" bun:"account_id,notnull,pk"`
	Account             *Account                `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId       ID[BankAccount]         `json:"bankAccountId" bun:"bank_account_id,notnull,nullzero"`
	BankAccount         *BankAccount            `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	FileId              ID[File]                `json:"fileId" bun:"file_id,notnull,nullzero"`
	File                *File                   `json:"file,omitempty" bun:"rel:belongs-to,join:file_id=file_id,join:account_id=account_id"`
	Status              TransactionUploadStatus `json:"status" bun:"status,notnull,nullzero"`
	Error               *string                 `json:"error,omitempty" bun:"error"`
	CreatedAt           time.Time               `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy           ID[User]                `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser       *User                   `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	ProcessedAt         *time.Time              `json:"processedAt" bun:"processed_at"`
	CompletedAt         *time.Time              `json:"completedAt" bun:"completed_at"`
}

func (TransactionUpload) FileKind() string {
	return "transactions/uploads"
}

// FileExpiration will return when a TransactionUpload's uploaded file should
// expire, this is one hour by default.
func (TransactionUpload) FileExpiration(clock clock.Clock) *time.Time {
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

func (TransactionUpload) IdentityPrefix() string {
	return "txup"
}

func (o *TransactionUpload) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionUploadId.IsZero() {
			o.TransactionUploadId = NewID[TransactionUpload]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
