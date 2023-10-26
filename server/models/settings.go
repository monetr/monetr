package models

type Settings struct {
	tableName string `pg:"settings"`

	AccountId      uint64   `json:"-" pg:"account_id,notnull,pk"`
	Account        *Account `json:"-" pg:"rel:has-one"`
	MaxSafeToSpend struct {
		Enabled bool  `json:"enabled"`
		Maximum int64 `json:"maximum"`
	} `json:"maxSafeToSpend" pg:"max_safe_to_spend,type:'jsonb'"`
}
