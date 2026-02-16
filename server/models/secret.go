package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type SecretKind string

const (
	SecretKindPlaid     SecretKind = "plaid"
	SecretKindLunchFlow SecretKind = "lunchflow"
)

type Secret struct {
	tableName string `pg:"secrets"`

	SecretId  ID[Secret]  `json:"-" pg:"secret_id,pk,notnull"`
	AccountId ID[Account] `json:"-" pg:"account_id,notnull,pk"`
	Account   *Account    `json:"-" pg:"rel:has-one"`
	Kind      SecretKind  `json:"-" pg:"kind,notnull"`
	KeyID     *string     `json:"-" pg:"key_id"`
	Version   *string     `json:"-" pg:"version"`
	Secret    string      `json:"-" pg:"secret,notnull"`
	UpdatedAt time.Time   `json:"-" pg:"updated_at,notnull"`
	CreatedAt time.Time   `json:"-" pg:"created_at,notnull"`
}

func (Secret) IdentityPrefix() string {
	return "scrt"
}

var (
	_ pg.BeforeInsertHook = (*Secret)(nil)
)

func (o *Secret) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.SecretId.IsZero() {
		o.SecretId = NewID[Secret]()
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
