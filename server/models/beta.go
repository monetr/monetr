package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Beta struct {
	bun.BaseModel `bun:"table:betas,alias:beta"`

	BetaId       ID[Beta]  `json:"betaId" bun:"beta_id,notnull,pk"`
	CodeHash     string    `json:"-" bun:"code_hash,notnull,unique,nullzero"`
	UsedByUserId *uint64   `json:"usedByUserId" bun:"used_by"`
	UsedByUser   *User     `json:"-" bun:"rel:belongs-to,join:used_by=user_id"`
	ExpiresAt    time.Time `json:"expiresAt" bun:"expires_at,notnull,nullzero"`
}

func (Beta) IdentityPrefix() string {
	return "beta"
}

var (
	_ bun.BeforeAppendModelHook = (*Beta)(nil)
)

func (o *Beta) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.BetaId.IsZero() {
			o.BetaId = NewID[Beta]()
		}
	}

	return nil
}
