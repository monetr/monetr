package models

import (
	"time"

	"github.com/monetr/monetr/pkg/feature"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
)

type Account struct {
	tableName string `pg:"accounts"`

	AccountId                    uint64                     `json:"accountId" pg:"account_id,notnull,pk,type:'bigserial'"`
	Timezone                     string                     `json:"timezone" pg:"timezone,notnull,default:'UTC'"`
	StripeCustomerId             *string                    `json:"-" pg:"stripe_customer_id"`
	StripeSubscriptionId         *string                    `json:"-" pg:"stripe_subscription_id"`
	StripeWebhookLatestTimestamp *time.Time                 `json:"-" pg:"stripe_webhook_latest_timestamp"`
	SubscriptionActiveUntil      *time.Time                 `json:"subscriptionActiveUntil" pg:"subscription_active_until"`
	SubscriptionStatus           *stripe.SubscriptionStatus `json:"subscriptionStatus" pg:"subscription_status"`
}

func (a *Account) GetTimezone() (*time.Location, error) {
	location, err := time.LoadLocation(a.Timezone)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse account timezone as location")
	}

	return location, nil
}

func (a *Account) HasFeature(feature feature.Feature) bool {
	// TODO Implement feature system with accounts.
	return true
}

// IsSubscriptionActive will return true if the SubscriptionActiveUntil date is not nill and is in the future. Even if
// the StripeSubscriptionId or StripeCustomerId is nil.
func (a *Account) IsSubscriptionActive() bool {
	endsInTheFuture := a.SubscriptionActiveUntil != nil && a.SubscriptionActiveUntil.After(time.Now())
	if a.SubscriptionStatus == nil {
		return endsInTheFuture
	}

	switch *a.SubscriptionStatus {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return endsInTheFuture
	default:
		return false
	}
}

// HasSubscription is used to determine if a "active" subscription has already been established for the account. This is
// active in the sense that the subscription is an accurate representation of their payment status for the application.
// The subscription could be past due, which would put the application in a "not usable" state; but the subscription
// would still be "active" because we would not want to create a new subscription.
func (a *Account) HasSubscription() bool {
	if a.SubscriptionStatus == nil {
		return a.SubscriptionActiveUntil != nil &&
			a.SubscriptionActiveUntil.After(time.Now()) &&
			a.StripeSubscriptionId != nil
	}

	switch *a.SubscriptionStatus {
	case stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing,
		stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete,
		stripe.SubscriptionStatusUnpaid:
		// When the subscription is one of these statuses, then the current subscription object should be used in stripe
		// and a new object should not be created.
		return a.StripeSubscriptionId != nil
	case stripe.SubscriptionStatusCanceled:
		// When the customer's subscription is canceled, it will not be re-used. A new subscription should be created.
		return false
	default:
		return false
	}
}
