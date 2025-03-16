package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
)

var (
	_ pg.BeforeInsertHook = (*TransactionAttachment)(nil)
	_ Identifiable        = TransactionAttachment{}
	_ Uploadable          = TransactionAttachment{}
)

type TransactionAttachment struct {
	tableName string `pg:"transaction_attachments"`

	TransactionAttachmentId ID[TransactionAttachment] `json:"transactionAttachmentId" pg:"transaction_attachment_id,notnull,pk"`
	AccountId               ID[Account]               `json:"-" pg:"account_id,notnull,pk"`
	Account                 *Account                  `json:"-" pg:"rel:has-one"`
	BankAccountId           ID[BankAccount]           `json:"bankAccountId" pg:"bank_account_id,notnull"`
	BankAccount             *BankAccount              `json:"-" pg:"rel:has-one"`
	TransactionId           ID[Transaction]           `json:"transactionId" pg:"transaction_id,notnull"`
	Transaction             *Transaction              `json:"-" pg:"rel:has-one"`
	FileId                  ID[File]                  `json:"fileId" pg:"file_id,notnull"`
	File                    *File                     `json:"file,omitempty" pg:"rel:has-one"`
	CreatedAt               time.Time                 `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy               ID[User]                  `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser           *User                     `json:"-" pg:"rel:has-one,fk:created_by"`
}

func (TransactionAttachment) FileKind() string {
	return "transactions/atttachments"
}

// TransactionAttachment files do not expire.
func (TransactionAttachment) FileExpiration(clock clock.Clock) *time.Time {
	expiration := clock.Now().Add(1 * time.Hour)
	return &expiration
}

func (TransactionAttachment) IdentityPrefix() string {
	return "txat"
}

// BeforeInsert implements orm.BeforeInsertHook.
func (o *TransactionAttachment) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionAttachmentId.IsZero() {
		o.TransactionAttachmentId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
