package models

type User struct {
	tableName string `sql:"users"`

	UserId           uint64   `json:"userId" sql:"user_id,notnull,pk,type:'bigserial'"`
	LoginId          uint64   `json:"loginId" sql:"login_id,notnull,on_delete:CASCADE"`
	Login            *Login   `json:"-" sql:"rel:has-one"`
	AccountId        uint64   `json:"accountId" sql:"account_id,notnull,on_delete:CASCADE"`
	Account          *Account `json:"account,omitempty" sql:"rel:has-one"`
	StripeCustomerId string   `json:"-" sql:"stripe_customer_id,null"`
	FirstName        string   `json:"firstName" sql:"first_name,notnull"`
	LastName         string   `json:"lastName" sql:"last_name,null"`
}
