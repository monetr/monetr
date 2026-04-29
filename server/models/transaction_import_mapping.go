package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/datasources/table"
)

var (
	_ pg.BeforeInsertHook = (*TransactionImportMapping)(nil)
	_ Identifiable        = TransactionImportMapping{}
)

type TransactionImportMapping struct {
	tableName string `pg:"transaction_import_mappings"`

	TransactionImportMappingId ID[TransactionImportMapping] `json:"transactionImportMappingId" pg:"transaction_import_mapping_id,notnull,pk"`
	AccountId                  ID[Account]                  `json:"-" pg:"account_id,notnull,pk"`
	Account                    *Account                     `json:"-" pg:"rel:has-one"`
	Signature                  string                       `json:"signature" pg:"signature,notnull"`
	Mapping                    table.Mapping                `json:"mapping" pg:"mapping,notnull,jsonb"`
	CreatedAt                  time.Time                    `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt                  time.Time                    `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedBy                  ID[User]                     `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser              *User                        `json:"-" pg:"rel:has-one,fk:created_by"`
}

func (TransactionImportMapping) IdentityPrefix() string {
	return "txix"
}

func (o *TransactionImportMapping) BeforeInsert(ctx context.Context) (context.Context, error) {
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

	return ctx, nil
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
