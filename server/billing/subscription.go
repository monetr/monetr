package billing

import (
	"context"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81"
)

// ReconcileSubscription takes an account who _should_ have a subscription
// associated with it but might not. If the account does not have a subscription
// then nothing is done and nil is returned. If the account does have a
// subscription then the subscription is retrieved from stripe and the details
// of the subscription are persisted to the account as represented by stripe.
func (b *baseBilling) ReconcileSubscription(ctx context.Context, accountId ID[Account]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.With("accountId", accountId)

	// Gather the account details from the repo. This data might be cached but
	// should be considered accurate as all writes for subscription data go
	// through this interface.
	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "cannot determine if account subscription is active")
	}

	if account.StripeCustomerId == nil {
		log.Log(span.Context(), logging.LevelTrace, "account does not have a customer, no subscription to reconcile")
		return nil
	}

	if account.StripeSubscriptionId == nil {
		log.Log(span.Context(), logging.LevelTrace, "account does not have a subscription to reconcile")
		return nil
	}

	log = log.With(
		slog.Group("stripe",
			"subscriptionId", account.StripeSubscriptionId,
			"customerId", account.StripeCustomerId,
		),
	)

	subscription, err := b.stripe.GetSubscription(
		span.Context(),
		*account.StripeSubscriptionId,
	)
	if err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve subscription from stripe for recon", "err", err)
		return err
	}

	currentStatus := *account.SubscriptionStatus
	actualStatus := subscription.Status

	currentActiveUntil := *account.SubscriptionActiveUntil
	actualActiveUntil := time.Unix(subscription.CurrentPeriodEnd, 0)

	currentSubscriptionId := *account.StripeSubscriptionId
	actualSubscriptionId := account.StripeSubscriptionId
	if actualStatus == stripe.SubscriptionStatusCanceled {
		actualSubscriptionId = nil
	}

	log.InfoContext(span.Context(), "subscription is being reconciled",
		slog.Group("status", "old", currentStatus, "new", actualStatus),
		slog.Group("activeUntil", "old", currentActiveUntil, "new", actualActiveUntil),
		slog.Group("stripe",
			"customerId", account.StripeCustomerId,
			slog.Group("subscriptionId", "old", currentSubscriptionId, "new", actualSubscriptionId),
		),
	)

	return b.UpdateCustomerSubscription(
		span.Context(),
		account,
		*account.StripeCustomerId,
		currentSubscriptionId,
		subscription.Status,
		&actualActiveUntil,
		b.clock.Now(),
	)
}
