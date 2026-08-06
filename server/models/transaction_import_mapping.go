package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TransactionImportMapping)(nil)
	_ Identifiable              = TransactionImportMapping{}
)

type TransactionImportMapping struct {
	bun.BaseModel `bun:"table:transaction_import_mappings,alias:transaction_import_mapping"`

	TransactionImportMappingId ID[TransactionImportMapping] `json:"transactionImportMappingId" bun:"transaction_import_mapping_id,notnull,pk"`
	AccountId                  ID[Account]                  `json:"-" bun:"account_id,notnull,pk"`
	Account                    *Account                     `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	Signature                  string                       `json:"signature" bun:"signature,notnull,nullzero"`
	Mapping                    table.Mapping                `json:"mapping" bun:"mapping,notnull,type:jsonb"`
	CreatedAt                  time.Time                    `json:"createdAt" bun:"created_at,notnull,nullzero"`
	UpdatedAt                  time.Time                    `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	CreatedBy                  ID[User]                     `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser              *User                        `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
}

func (TransactionImportMapping) IdentityPrefix() string {
	return "txix"
}

func (o *TransactionImportMapping) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionImportMappingId.IsZero() {
			o.TransactionImportMappingId = NewID[TransactionImportMapping]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = now
		}
		if o.Signature == "" {
			o.Signature = signMappingHeaders(o.Mapping.Headers)
		}
	}

	return nil
}

func signMappingHeaders(headers []string) string {
	normalized := make([]string, len(headers))
	for i, h := range headers {
		normalized[i] = strings.ToLower(h)
	}
	sort.Strings(normalized)
	sum := sha256.Sum256([]byte(strings.Join(normalized, ",")))
	return hex.EncodeToString(sum[:])
}
