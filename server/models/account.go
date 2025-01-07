package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81"
)

var (
	_ pg.BeforeInsertHook = (*Account)(nil)
	_ Identifiable        = Account{}
)

type Account struct {
	tableName string `pg:"accounts"`

	AccountId                     ID[Account]                `json:"accountId" pg:"account_id,notnull,pk"`
	Timezone                      string                     `json:"timezone" pg:"timezone,notnull,default:'UTC'"`
	Locale                        string                     `json:"locale" pg:"locale,notnull"`
	StripeCustomerId              *string                    `json:"-" pg:"stripe_customer_id"`
	StripeSubscriptionId          *string                    `json:"-" pg:"stripe_subscription_id"`
	StripeWebhookLatestTimestamp  *time.Time                 `json:"-" pg:"stripe_webhook_latest_timestamp"`
	SubscriptionActiveUntil       *time.Time                 `json:"subscriptionActiveUntil" pg:"subscription_active_until"`
	SubscriptionStatus            *stripe.SubscriptionStatus `json:"subscriptionStatus" pg:"subscription_status"`
	TrialEndsAt                   *time.Time                 `json:"trialEndsAt" pg:"trial_ends_at"`
	TrialExpiryNotificationSentAt *time.Time                 `json:"-" pg:"trial_expiry_notification_sent_at"`
	CreatedAt                     time.Time                  `json:"createdAt" pg:"created_at,notnull"`
}

func (Account) IdentityPrefix() string {
	return "acct"
}

func (o *Account) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.AccountId.IsZero() {
		o.AccountId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}

func (a *Account) GetTimezone() (*time.Location, error) {
	location, err := time.LoadLocation(a.Timezone)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse account timezone as location")
	}

	return location, nil
}

// IsSubscriptionActive will return true if the SubscriptionActiveUntil date is
// not nill and is in the future. Even if the StripeSubscriptionId or
// StripeCustomerId is nil.
func (a *Account) IsSubscriptionActive(now time.Time) bool {
	activeUntil := myownsanity.MaxNonNilTime(
		a.SubscriptionActiveUntil,
		a.TrialEndsAt,
	)

	consideredActive := activeUntil != nil && activeUntil.After(now)

	// If for some reason the account does not have a subscription status, then
	// only consider the timestamps.
	if a.SubscriptionStatus == nil {
		return consideredActive
	}

	// If they do have a subscription status then only consider these two statuses
	// as active. All other statuses should be considered an inactive
	// subscription.
	switch *a.SubscriptionStatus {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return consideredActive
	default:
		return false
	}
}

// HasSubscription is used to determine if a "active" subscription has already
// been established for the account. This is active in the sense that the
// subscription is an accurate representation of their payment status for the
// application. The subscription could be past due, which would put the
// application in a "not usable" state; but the subscription would still be
// "active" because we would not want to create a new subscription.
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
		// When the subscription is one of these statuses, then the current
		// subscription object should be used in stripe and a new object should not
		// be created.
		return a.StripeSubscriptionId != nil
	case
		stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusTrialing:
		// When the customer's subscription is canceled, it will not be re-used. A
		// new subscription should be created.
		return false
	default:
		return false
	}
}

func (a *Account) IsTrialing(now time.Time) bool {
	// We are in a trial state when we do not have a subscription and the trial
	// ends at date is in the future.
	return !a.HasSubscription() &&
		a.TrialEndsAt != nil &&
		a.TrialEndsAt.After(now)
}
