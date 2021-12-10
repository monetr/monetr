package models

type User struct {
	tableName string `pg:"users"`

	UserId           uint64   `json:"userId" bun:"user_id,notnull,pk,type:'bigserial'"`
	LoginId          uint64   `json:"loginId" bun:"login_id,notnull,on_delete:CASCADE,unique:per_account"`
	Login            *Login   `json:"login,omitempty" bun:"rel:has-one,join:login_id=login_id"`
	AccountId        uint64   `json:"accountId" bun:"account_id,notnull,on_delete:CASCADE,unique:per_account"`
	Account          *Account `json:"account" bun:"rel:has-one,join:account_id=account_id"`
	FirstName        string   `json:"firstName" bun:"first_name,notnull"`
	LastName         string   `json:"lastName" bun:"last_name"`
	StripeCustomerId *string  `json:"-" bun:"stripe_customer_id"`
}
