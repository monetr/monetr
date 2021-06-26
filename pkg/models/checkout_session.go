package models

type CheckoutSession struct {
	tableName string `pg:"checkout_sessions"`

	CheckoutSessionId string   `json:"checkoutSessionId" pg:"checkout_session_id,pk"`
	AccountId         uint64   `json:"-" pg:"account_id,notnull,on_delete:CASCADE"`
	Account           *Account `json:"-" pg:"rel:has-one"`
	IsComplete bool `json:"-"`
}
