package models

import (
	"time"

	"github.com/monetr/monetr/pkg/feature"
	"github.com/pkg/errors"
)

type Account struct {
	tableName string `pg:"accounts"`

	AccountId                    uint64     `json:"accountId" bun:"account_id,notnull,pk"`
	Timezone                     string     `json:"timezone" bun:"timezone,notnull,default:'UTC'"`
	StripeCustomerId             *string    `json:"-" bun:"stripe_customer_id"`
	StripeSubscriptionId         *string    `json:"-" bun:"stripe_subscription_id"`
	StripeWebhookLatestTimestamp *time.Time `json:"-" bun:"stripe_webhook_latest_timestamp"`
	SubscriptionActiveUntil      *time.Time `json:"subscriptionActiveUntil" bun:"subscription_active_until"`
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
	return a.SubscriptionActiveUntil != nil && a.SubscriptionActiveUntil.After(time.Now())
}
