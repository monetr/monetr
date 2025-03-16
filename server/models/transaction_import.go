package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
)

var (
	_ pg.BeforeInsertHook = (*TransactionImport)(nil)
	_ Identifiable        = TransactionImport{}
	_ Uploadable          = TransactionImport{}
)

type TransactionImportAffectedBankAccounts struct {
	// UploadIdentifier is the hashed value of the account Id field from the camt
	// file. This is typically an account number which is why it is hashed.
	UploadIdentifier string `json:"uploadIdentifier"`
	// UploadIdentifierType is taken from the camt `Acct/Id/...` field. It
	// designates which element was used to uniquely identify the account.
	UploadIdentifierType string `json:"uploadIdentifierType"`
	// Name is the account name from the file and is taken from the camt field
	// `Acct/Nm`.
	Name string `json:"name"`
	// Currency is the currency code taken from the account in the file. It may
	// not be a currency supported by monetr.
	Currency string `json:"currency"`
	// BankAccountId is the ID of the account that this part of the import maps
	// to. If the bank account ID is nil then this is an account that would be
	// created.
	BankAccountId *ID[BankAccount] `json:"bankAccountId"`
}

// TransactionImport is different from TraansctionUpload, uploads are specific
// to a single bank account. Where as imports are a link level import. Imports
// rely on identifying bank account data should go into, and may import data
// into multiple bank accounts within the same link.
type TransactionImport struct {
	tableName string `pg:"transaction_imports"`

	TransactionImportId  ID[TransactionImport]                   `json:"transactionImportId" pg:"transaction_import_id,notnull,pk"`
	AccountId            ID[Account]                             `json:"-" pg:"account_id,notnull,pk"`
	Account              *Account                                `json:"-" pg:"rel:has-one"`
	LinkId               ID[Link]                                `json:"linkId" pg:"link_id,notnull"`
	Link                 *Link                                   `json:"-,omitempty" pg:"rel:has-one"`
	FileId               ID[File]                                `json:"fileId" pg:"file_id,notnull"`
	File                 *File                                   `json:"file,omitempty" pg:"rel:has-one"`
	AffectedBankAccounts []TransactionImportAffectedBankAccounts `json:"affectedBankAccounts" pg:"affected_bank_accounts"`
	CreatedAt            time.Time                               `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy            ID[User]                                `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser        *User                                   `json:"-" pg:"rel:has-one,fk:created_by"`
}

func (TransactionImport) IdentityPrefix() string {
	return "txim"
}

func (TransactionImport) FileKind() string {
	return "transactions/imports"
}

// TransactionImport files expire after 1 hour be default.
func (TransactionImport) FileExpiration(clock clock.Clock) *time.Time {
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

