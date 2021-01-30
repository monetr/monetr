package models

type User struct {
	tableName string `pg:"users"`

	UserId           uint64   `json:"userId" pg:"user_id,notnull,pk,type:'bigserial'"`
	LoginId          uint64   `json:"loginId" pg:"login_id,notnull,on_delete:CASCADE"`
	Login            *Login   `json:"-" pg:"rel:has-one"`
	AccountId        uint64   `json:"accountId" pg:"account_id,notnull,on_delete:CASCADE"`
	Account          *Account `json:"account,omitempty" pg:"rel:has-one"`
	StripeCustomerId string   `json:"-" pg:"stripe_customer_id"`
	FirstName        string   `json:"firstName" pg:"first_name,notnull"`
	LastName         string   `json:"lastName" pg:"last_name"`
}
