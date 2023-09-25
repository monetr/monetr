package models

import (
	"time"

	"github.com/monetr/monetr/pkg/feature"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
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
	TrialEndsAt                  *time.Time                 `json:"trialEndsAt" pg:"trial_ends_at"`
	CreatedAt                    time.Time                  `json:"createdAt" pg:"created_at,notnull"`
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
	activeUntil := myownsanity.MaxNonNilTime(
		a.SubscriptionActiveUntil,
		a.TrialEndsAt,
	)

	consideredActive := activeUntil != nil && activeUntil.After(time.Now())

	// If for some reason the account does not have a subscription status, then only consider the timestamps.
	if a.SubscriptionStatus == nil {
		return consideredActive
	}

	// If they do have a subscription status then only consider these two statuses as active. All other statuses should be
	// considered an inactive subscription.
	switch *a.SubscriptionStatus {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return consideredActive
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
		return false
	}

	switch *a.SubscriptionStatus {
	case
		stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete,
		stripe.SubscriptionStatusUnpaid:
		// When the subscription is one of these statuses, then the current subscription object should be used in stripe
		// and a new object should not be created.
		return a.StripeSubscriptionId != nil
	case
		stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusTrialing:
		// When the customer's subscription is canceled, it will not be re-used. A new subscription should be created.
		return false
	default:
		return false
	}
}

func (a *Account) IsTrialing() bool {
	return a.TrialEndsAt != nil && a.TrialEndsAt.After(time.Now())
}
