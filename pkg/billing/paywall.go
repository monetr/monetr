package billing

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// BasicPayWall is used by the API middleware and other operations to restrict access to some features or functionality
// that requires an active subscription.
type BasicPayWall interface {
	// GetHasSubscription should return whether or not there is a non-canceled subscription object associated with an
	// account. It does not indicate whether or not this subscription object is in a state that the customer should be
	// allowed to use their account, only whether or not the subscription object already exists in such a state that a
	// new subscription should not be created.
	GetHasSubscription(ctx context.Context, accountId uint64) (bool, error)
	// GetSubscriptionIsActive should return whether or not the customer's subscription (or lack thereof) is in a state
	// where the customer should have access to their account and data. If they lack a subscription entirely, or the
	// subscription has been canceled or past due; then the customer should not be permitted to access their
	// application.
	GetSubscriptionIsActive(ctx context.Context, accountId uint64) (bool, error)
}

var (
	_ BasicPayWall = &baseBasicPaywall{}
)

type baseBasicPaywall struct {
	log      *logrus.Entry
	accounts AccountRepository
}

func NewBasicPaywall(log *logrus.Entry, repo AccountRepository) BasicPayWall {
	return &baseBasicPaywall{
		log:      log,
		accounts: repo,
	}
}

func (b *baseBasicPaywall) GetHasSubscription(ctx context.Context, accountId uint64) (bool, error) {
	span := sentry.StartSpan(ctx, "Billing - GetHasSubscription")
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithField("accountId", accountId)
	log.Trace("checking whether or not subscription is present")

	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "could not determine whether subscription was present")
	}

	return account.HasSubscription(), nil
}

// GetSubscriptionIsActive will retrieve the account data from the AccountRepository interface. This means it is
// possible for it to return a stale response within a few seconds. But in general it should be acceptable. When an
// account is updated -> its cache is invalidated. There is likely a very small window where an invalid state could be
// evaluated, but it should be fine.
func (b *baseBasicPaywall) GetSubscriptionIsActive(ctx context.Context, accountId uint64) (active bool, err error) {
	span := sentry.StartSpan(ctx, "Billing - GetSubscriptionIsActive")
	defer span.Finish()

	defer func() {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			level := sentry.LevelDebug
			crumbType := "debug"
			if err != nil {
				crumbType = "error"
				level = sentry.LevelError
			}

			var message string
			if active {
				message = "Subscription is active."
			} else if err == nil {
				message = "Subscription is not active, the current endpoint may require an active subscription."
			} else {
				message = "There was a problem verifying whether or not the subscription was active"
			}

			hub.AddBreadcrumb(&sentry.Breadcrumb{
				Type:      crumbType,
				Category:  "subscription",
				Message:   message,
				Level:     level,
				Timestamp: time.Now(),
			}, nil)
		}
	}()

	log := b.log.WithContext(span.Context()).WithField("accountId", accountId)

	log.Trace("checking if account subscription is active")

	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "cannot determine if account subscription is active")
	}

	span.Status = sentry.SpanStatusOK

	return account.IsSubscriptionActive(), nil
}
