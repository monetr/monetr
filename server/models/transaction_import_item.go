package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

var (
	_ pg.BeforeInsertHook = (*TransactionImportItem)(nil)
	_ Identifiable        = TransactionImportItem{}
)

type TransactionImportItem struct {
	tableName string `pg:"transaction_import_items"`

	TransactionImportItemId ID[TransactionImportItem] `json:"transactionImportItemId" pg:"transaction_import_item_id,notnull,pk"`
	AccountId               ID[Account]               `json:"-" pg:"account_id,notnull,pk"`
	Account                 *Account                  `json:"-" pg:"rel:has-one"`
	TransactionImportId     ID[TransactionImport]     `json:"transactionImportId" pg:"transaction_import_id,notnull,pk"`
	TransactionImport       *TransactionImport        `json:"-" pg:"rel:has-one"`
	// BankAccountId is the ID of the account that this part of the import maps
	// to. If the bank account ID is nil then this is an account that would be
	// created.
	// Unique per transaction import.
	BankAccountId *ID[BankAccount] `json:"bankAccountId" pg:"bank_account_id"`
	BankAccount   *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one"`
	// UploadIdentifier is the hashed value of the account Id field from the camt
	// file. This is typically an account number which is why it is hashed. This
	// must also be unique per import.
	UploadIdentifier string `json:"uploadIdentifier" pg:"upload_identifier,notnull"`
	// UploadIdentifierType is taken from the camt `Acct/Id/...` field. It
	// designates which element was used to uniquely identify the account.
	UploadIdentifierType string `json:"uploadIdentifierType" pg:"upload_identifier_type,notnull"`
	// Name is the account name from the file and is taken from the camt field
	// `Acct/Nm`.
	Name string `json:"name" pg:"name,notnull"`
	// Currency is the currency code taken from the account in the file. It may
	// not be a currency supported by monetr.
	Currency string `json:"currency" pg:"currency,notnull"`
	// TODO false
	Include   bool      `json:"include" pg:"include,notnull"`
	CreatedAt time.Time `json:"createdAt" pg:"created_at,notnull"`
}

func (TransactionImportItem) IdentityPrefix() string {
	return "txit"
}

// BeforeInsert implements orm.BeforeInsertHook.
func (o *TransactionImportItem) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionImportItemId.IsZero() {
		o.TransactionImportItemId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
