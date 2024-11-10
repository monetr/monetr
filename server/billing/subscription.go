package billing

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId": accountId,
	})

	// Gather the account details from the repo. This data might be cached but
	// should be considered accurate as all writes for subscription data go
	// through this interface.
	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "cannot determine if account subscription is active")
	}

	if account.StripeCustomerId == nil {
		log.Trace("account does not have a customer, no subscription to reconcile")
		return nil
	}

	if account.StripeSubscriptionId == nil {
		log.Trace("account does not have a subscription to reconcile")
		return nil
	}

	log = log.WithFields(logrus.Fields{
		"stripe": logrus.Fields{
			"subscriptionId": account.StripeSubscriptionId,
			"customerId":     account.StripeCustomerId,
		},
	})

	subscription, err := b.stripe.GetSubscription(
		span.Context(),
		*account.StripeSubscriptionId,
	)
	if err != nil {
		log.WithError(err).Error("failed to retrieve subscription from stripe for recon")
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

	log.WithFields(logrus.Fields{
		"status": map[string]any{
			"old": currentStatus,
			"new": actualStatus,
		},
		"activeUntil": map[string]any{
			"old": currentActiveUntil,
			"new": actualActiveUntil,
		},
		"stripe": logrus.Fields{
			"customerId": account.StripeCustomerId,
			"subscriptionId": map[string]any{
				"old": currentSubscriptionId,
				"new": actualSubscriptionId,
			},
		},
	}).Info("subscription is being reconciled")

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
