package models

import "time"

type SecretKind string

const (
	PlaidSecretKind  SecretKind = "plaid"
	TellerSecretKind SecretKind = "teller"
)

type Secret struct {
	tableName string `pg:"secrets"`

	SecretId  uint64     `json:"-" pg:"secret_id,pk,notnull,type:'bigserial'"`
	AccountId uint64     `json:"-" pg:"account_id,notnull,pk"`
	Account   *Account   `json:"-" pg:"rel:has-one"`
	Kind      SecretKind `json:"-" pg:"kind,notnull"`
	KeyID     *string    `json:"-" pg:"key_id"`
	Version   *string    `json:"-" pg:"version"`
	Secret    string     `json:"-" pg:"secret,notnull"`
	UpdatedAt time.Time  `json:"-" pg:"updated_at,notnull"`
	CreatedAt time.Time  `json:"-" pg:"created_at,notnull"`
}
