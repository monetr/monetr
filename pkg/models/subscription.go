package models

import (
	"github.com/stripe/stripe-go/v72"
)

type Subscription struct {
	tableName string `pg:"subscriptions"`

	SubscriptionId       uint64                    `json:"subscriptionId" pg:"subscription_id,pk,type:'bigserial'"`
	AccountId            uint64                    `json:"-" pg:"account_id,notnull,on_delete:restrict"` // Think about deletes later.
	Account              *Account                  `json:"-" pg:"rel:has-one"`
	OwnedByUserId        uint64                    `json:"ownedByUserId" pg:"owned_by_user_id,notnull,on_delete:restrict"`
	OwnedByUser          *User                     `json:"ownedByUser,omitempty" pg:"rel:has-one"`
	StripeSubscriptionId string                    `json:"-" pg:"stripe_subscription_id,notnull"`
	StripeCustomerId     string                    `json:"-" pg:"stripe_customer_id,notnull"`
	PriceId              uint64                    `json:"priceId" pg:"price_id,notnull,on_delete:restrict"`
	Price                *Price                    `json:"price,omitempty" pg:"rel:has-one"`
	Status               stripe.SubscriptionStatus `json:"status" pg:"status,notnull"`
}
