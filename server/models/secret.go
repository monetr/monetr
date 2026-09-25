package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type SecretKind string

const (
	SecretKindPlaid     SecretKind = "plaid"
	SecretKindLunchFlow SecretKind = "lunchflow"
)

type Secret struct {
	bun.BaseModel `bun:"table:secrets,alias:secret"`

	SecretId  ID[Secret]  `json:"-" bun:"secret_id,pk,notnull"`
	AccountId ID[Account] `json:"-" bun:"account_id,notnull,pk"`
	Account   *Account    `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	Kind      SecretKind  `json:"-" bun:"kind,notnull,nullzero"`
	KeyID     *string     `json:"-" bun:"key_id"`
	Version   *string     `json:"-" bun:"version"`
	Secret    string      `json:"-" bun:"secret,notnull,nullzero"`
	UpdatedAt time.Time   `json:"-" bun:"updated_at,notnull,nullzero"`
	CreatedAt time.Time   `json:"-" bun:"created_at,notnull,nullzero"`
}

func (Secret) IdentityPrefix() string {
	return "scrt"
}

var (
	_ bun.BeforeAppendModelHook = (*Secret)(nil)
)

func (o *Secret) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
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
	}

	return nil
}
