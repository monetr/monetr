package models

type Beta struct {
	tableName string `pg:"betas"`

	BetaID uint64 `json:"betaId" pg:"beta_id,notnull,pk,type:'bigserial'"`

}