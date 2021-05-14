package models

import (
	"github.com/monetrapp/rest-api/pkg/feature"
	"github.com/stripe/stripe-go/v72"
	"time"
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
	StripePriceId        string                    `json:"-" pg:"stripe_price_id,notnull"`
	Features             []feature.Feature         `json:"features" pg:"features,type:'text[]'"`
	Status               stripe.SubscriptionStatus `json:"status" pg:"status,notnull"`
	TrialStart           *time.Time                `json:"trialStart" pg:"trial_start"`
	TrialEnd             *time.Time                `json:"trialEnd" pg:"trial_end"`
}
