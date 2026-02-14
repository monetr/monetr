package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type Beta struct {
	tableName string `pg:"betas"`

	BetaId       ID[Beta]  `json:"betaId" pg:"beta_id,notnull,pk"`
	CodeHash     string    `json:"-" pg:"code_hash,notnull,unique"`
	UsedByUserId *uint64   `json:"usedByUserId" pg:"used_by"`
	UsedByUser   *User     `json:"-" pg:"rel:has-one,fk:used_by"`
	ExpiresAt    time.Time `json:"expiresAt" pg:"expires_at,notnull"`
}

func (o Beta) IdentityPrefix() string {
	return "beta"
}

var (
	_ pg.BeforeInsertHook = (*Beta)(nil)
)

func (o *Beta) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.BetaId.IsZero() {
		o.BetaId = NewID[Beta]()
	}

	return ctx, nil
}
