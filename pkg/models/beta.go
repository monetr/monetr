package models

import "time"

type Beta struct {
	tableName string `pg:"betas"`

	BetaID       uint64    `json:"betaId" pg:"beta_id,notnull,pk,type:'bigserial'"`
	CodeHash     string    `json:"-" pg:"code_hash,notnull,unique"`
	UsedByUserId *uint64   `json:"usedByUserId" pg:"used_by_user_id,on_delete:CASCADE"`
	UsedByUser   *User     `json:"-" pg:"rel:has-one,fk:used_by_user_id"`
	ExpiresAt    time.Time `json:"expiresAt" pg:"expires_at,notnull"`
}
