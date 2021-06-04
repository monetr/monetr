package models

import "time"

type PaymentMethod struct {
	tableName string `pg:"payment_method"`

	PaymentMethodId       uint64    `json:"paymentMethodId" pg:"payment_method_id,pk,type:'bigserial'"`
	AccountId             uint64    `json:"-" pg:"account_id,notnull"`
	UserId                uint64    `json:"user_id,notnull,on_delete:RESTRICT"`
	User                  *User     `json:"-" pg:"rel:has-one"`
	StripePaymentMethodId string    `json:"-" pg:"stripe_payment_method_id,notnull"`
	Name                  string    `json:"name,notnull"`
	LastFour              string    `json:"lastFour,notnull"`
	Brand                 string    `json:"brand,notnull"`
	Expires               time.Time `json:"expiresAt"`
	Hash                  string    `json:"-" pg:"hash,notnull"`
	DateCreated           time.Time `json:"createdAt" pg:"created_at,notnull"`
}
