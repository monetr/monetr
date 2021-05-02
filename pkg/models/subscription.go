package models

import "github.com/stripe/stripe-go/v72"

type Subscription struct {
	tableName string `sql:"subscriptions"`

	SubscriptionId       uint64                    `json:"subscriptionId" sql:"subscription_id,pk,type:'bigserial'"`
	AccountId            uint64                    `json:"-" sql:"account_id,notnull,on_delete:restrict"` // Think about deletes later.
	Account              *Account                  `json:"-" sql:"rel:has-one"`
	OwnedByUserId        uint64                    `json:"ownedByUserId" sql:"owned_by_user_id,notnull,on_delete:restrict"`
	OwnedByUser          *User                     `json:"ownedByUser,omitempty" sql:"rel:has-one"`
	StripeSubscriptionId string                    `json:"-" sql:"stripe_subscription_id,notnull"`
	StripeCustomerId     string                    `json:"-" sql:"stripe_customer_id,notnull"`
	PriceId              uint64                    `json:"priceId" sql:"price_id,notnull,on_delete:restrict"`
	Price                *Price                    `json:"price,omitempty" sql:"rel:has-one"`
	Status               stripe.SubscriptionStatus `json:"status" sql:"status,notnull"`
}
