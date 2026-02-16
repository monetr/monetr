package billing

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
)

var (
	// This is added onto the subscription active until timestamp. This way if the
	// payment fails or Stripe doesn't process the payment in time, we can still
	// have some padding. If the payment fails this is enough time for Stripe to
	// retry the payment at least once.
	SubscriptionPaddingDays = 4
)

// Billing is used by the Stripe webhooks to maintain a subscription's status
// within our application. As the status of subscription's change or update
// these functions can be used to keep the status up to date within monetr.
type Billing interface {
	// UpdateSubscription will set the Stripe customer Id and subscription Id on
	// the account object. It will also set the active until date for the account.
	// The date can be nil. If the date is nil or in the past, the subscription is
	// considered cancelled. A timestamp should also be provided. The timestamp is
	// used to fix race conditions in the webhooks received from Stripe. If the
	// provided timestamp is less than the timestamp of the last change applied to
	// the account, then the change is not applied. The change in subscription is
	// only applied when the provided timestamp is after the timestamp of the
	// previously applied change. This helps solve a problem where sometimes a
	// webhook for a subscription being created (which would have an incomplete
	// status) can be delivered after a update webhook for the same subscription
	// (which would indicate an active status) causing the subscription to
	// incorrectly show as inactive.
	UpdateSubscription(
		ctx context.Context,
		customerId, subscriptionId string,
		status stripe.SubscriptionStatus,
		activeUntil *time.Time,
		timestamp time.Time,
	) error

	// UpdateCustomerSubscription does the same thing that UpdateSubscription
	// does, but does not require that the stripe customerId match any customerId
	// stored. Instead, it will take the provided account and update the customer
	// ID and store it on the account with the new subscription data.
	UpdateCustomerSubscription(
		ctx context.Context,
		account *Account,
		customerId, subscriptionId string,
		status stripe.SubscriptionStatus,
		activeUntil *time.Time,
		timestamp time.Time,
	) error

	// HandleStripeWebhook takes a stripe webhook event and properly updates the
	// subscription details for the associated account in monetr.
	HandleStripeWebhook(ctx context.Context, event stripe.Event) error

	// GetHasSubscription should return whether or not there is a non-canceled
	// subscription object associated with an account. It does not indicate
	// whether or not this subscription object is in a state that the customer
	// should be allowed to use their account, only whether or not the
	// subscription object already exists in such a state that a new subscription
	// should not be created.
	GetHasSubscription(ctx context.Context, accountId ID[Account]) (bool, error)
	// GetSubscriptionIsActive should return whether or not the customer's
	// subscription (or lack thereof) is in a state where the customer should have
	// access to their account and data. If they lack a subscription entirely, or
	// the subscription has been canceled or past due; then the customer should
	// not be permitted to access their application.
	GetSubscriptionIsActive(ctx context.Context, accountId ID[Account]) (bool, error)
	GetSubscriptionIsTrialing(ctx context.Context, accountId ID[Account]) (trialing bool, err error)

	// CreateBillingPortal will return a Stripe billing portal URL if the
	// specified account ID has a subscription that is "active" and can be
	// updated. If the subscription has been cancelled or some other terminal
	// status then the portal function will not work. Instead you must create a
	// checkout session. Owner represents the Login which "owns" the account and
	// is the one on the billing information.
	CreateBillingPortal(ctx context.Context, owner Login, accountId ID[Account]) (string, error)

	// CreateCustomer takes an account object and will create a stripe customer
	// for that account if one does not already exist. It will then store this in
	// the database and update any cache. As well as update it on the provided
	// object. If the account provided already has a Stripe customer associated
	// with it then this function will do nothing and return nil.
	CreateCustomer(ctx context.Context, owner Login, account *Account) error

	// CreateCheckout takes a Login object that represents the "billing owner" as
	// well as an account ID and an optional cancel path. If the specified account
	// has a subscription at all then this function will return an error. If the
	// subscription is active it will return an `ErrSubscriptionAlreadyActive` and
	// if the subscription is not active but exists it will return
	// `ErrSubscriptionAlreadyExists`. Both of these errors should be treated as an
	// indication that the user cannot use the checkout to establish billing.
	// Instead a billing portal should be created. If the optional cancel path is
	// specified then that will be used for the non-success return route from the
	// checkout. The default is `/account/subscribe` if one is not provided. The
	// success path of the checkout will always be `/account/subscribe/after` and
	// cannot be overwritten.
	CreateCheckout(ctx context.Context, owner Login, accountId ID[Account], cancelPath *string) (*stripe.CheckoutSession, error)

	// AfterCheckout is called when the user is directed back to the application
	// after completing a checkout in stripe. This function takes the checkout
	// session ID that is returned during the success redirect from stripe. This
	// function then checks to see if the user completed their subscription as part
	// of the checkout and takes their subscription status and stores it on the
	// account. It then returns true or false indicating whether the subscription is
	// now active.
	AfterCheckout(ctx context.Context, accountId ID[Account], checkoutSessionId string) (active bool, _ error)

	// ReconcileSubscription takes an account who _should_ have a subscription
	// associated with it but might not. If the account does not have a subscription
	// then nothing is done and nil is returned. If the account does have a
	// subscription then the subscription is retrieved from stripe and the details
	// of the subscription are persisted to the account as represented by stripe.
	ReconcileSubscription(ctx context.Context, accountId ID[Account]) error
}

// SubscriptionIsActive is a helper function that takes in a Stripe subscription
// object and returns true or false based on the state of that object's
// subscription. This is used to handle scenarios where multiple factors could
// lead to a subscription being active or inactive. At the time of writing this
// it will return active if the subscription is in an active state, or is
// trialing.
func SubscriptionIsActive(subscription stripe.Subscription) bool {
	switch subscription.Status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return true
	default:
		return false
	}
}

var (
	_ Billing = &baseBilling{}
)

type baseBilling struct {
	log      *logrus.Entry
	clock    clock.Clock
	config   config.Configuration
	accounts repository.AccountsRepository
	stripe   stripe_helper.Stripe
	notify   pubsub.Publisher
}

func NewBilling(
	log *logrus.Entry,
	clock clock.Clock,
	config config.Configuration,
	repo repository.AccountsRepository,
	stripe stripe_helper.Stripe,
	notifications pubsub.Publisher,
) Billing {
	return &baseBilling{
		log:      log,
		clock:    clock,
		config:   config,
		accounts: repo,
		stripe:   stripe,
		notify:   notifications,
	}
}

func (b *baseBilling) UpdateSubscription(
	ctx context.Context,
	customerId, subscriptionId string,
	status stripe.SubscriptionStatus,
	activeUntil *time.Time,
	timestamp time.Time,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"customerId":     customerId,
		"subscriptionId": subscriptionId,
	})

	log.Trace("retrieving account by customer Id")

	account, err := b.accounts.GetAccountByCustomerId(span.Context(), customerId)
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve account by stripe customer Id")
		return errors.Wrap(err, "failed to retrieve account by stripe customer Id")
	}

	return b.UpdateCustomerSubscription(
		span.Context(),
		account,
		customerId, subscriptionId,
		status,
		activeUntil,
		timestamp,
	)
}

func (b *baseBilling) UpdateCustomerSubscription(
	ctx context.Context,
	account *Account,
	customerId, subscriptionId string,
	status stripe.SubscriptionStatus,
	activeUntil *time.Time,
	timestamp time.Time,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"customerId":     customerId,
		"subscriptionId": subscriptionId,
		"accountId":      account.AccountId,
	})

	// Set the user for this event, this way webhooks are properly associated with
	// the destination user in our application.
	crumbs.IncludeUserInScope(span.Context(), account.AccountId)

	currentlyActive := account.IsSubscriptionActive(b.clock.Now())

	// If the timestamp for the last webhook is not nil, and the provided
	// timestamp is not after (<= basically) then perform the update. This is to
	// solve potential race conditions in the order we receive webhooks from Stripe.
	if account.StripeWebhookLatestTimestamp != nil {
		if timestamp.Before(*account.StripeWebhookLatestTimestamp) {
			crumbs.Debug(span.Context(), "Provided timestamp is older than the current subscription timestamp", map[string]any{
				"stored":   *account.StripeWebhookLatestTimestamp,
				"provided": timestamp,
			})
			return nil
		} else if timestamp.Equal(*account.StripeWebhookLatestTimestamp) {
			crumbs.Warn(span.Context(), "Provided timestamp is equal to the current subscription timestamp", "stripe", map[string]any{
				"stored":   *account.StripeWebhookLatestTimestamp,
				"provided": timestamp,
			})
			// Set the user for this event, this way webhooks are properly associated with the destination user in our
			// application.
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetTag("potentialBug", "true")
				})
			}

			return nil
		}

		crumbs.Debug(span.Context(), "Provided timestamp is after the current subscription timestamp, change will be applied", map[string]any{
			"stored":   *account.StripeWebhookLatestTimestamp,
			"provided": timestamp,
		})
	} else {
		crumbs.Debug(span.Context(), "Current subscription timestamp is nil, webhook will be accepted", map[string]any{
			"provided": timestamp,
		})
	}

	account.StripeCustomerId = &customerId
	if status == stripe.SubscriptionStatusCanceled {
		// If we are canceling the subscription, then set this to nil.
		account.StripeSubscriptionId = nil
	} else {
		// Otherwise do this. If its adding a value great, otherwise itll update the existing value and overwrite it.
		account.StripeSubscriptionId = &subscriptionId
	}
	// Add padding to the subscription window. This way Stripe has time to process
	// the subscription payment and update the status for us even if things are
	// running a bit slow. This resolves an issue where the active until date can
	// pass before Stripe has processed the renewal. Causing (usually) around an
	// hour or more of time where monetr believed the subscription to not be
	// active anymore.
	account.SubscriptionActiveUntil = myownsanity.Pointer(
		activeUntil.AddDate(0, 0, SubscriptionPaddingDays),
	)
	account.StripeWebhookLatestTimestamp = &timestamp
	account.SubscriptionStatus = &status

	if err := b.accounts.UpdateAccount(span.Context(), account); err != nil {
		log.WithError(err).Errorf("failed to update account subscription status")
		return errors.Wrap(err, "failed to update account subscription status")
	}

	// Check to see if the subscription status of the account has changed with this update to be.
	if account.IsSubscriptionActive(b.clock.Now()) != currentlyActive {
		if currentlyActive {
			log.Info("account subscription is no longer active")
		} else {
			log.Info("account subscription is now active")
		}
	}

	return nil
}

func (b *baseBilling) GetHasSubscription(ctx context.Context, accountId ID[Account]) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
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

// GetSubscriptionIsActive will retrieve the account data from the
// AccountRepository interface. This means it is possible for it to return a
// stale response within a few seconds. But in general it should be acceptable.
// When an account is updated -> its cache is invalidated. There is likely a
// very small window where an invalid state could be evaluated, but it should be
// fine.
func (b *baseBilling) GetSubscriptionIsActive(
	ctx context.Context,
	accountId ID[Account],
) (active bool, err error) {
	span := crumbs.StartFnTrace(ctx)
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
				Timestamp: b.clock.Now(),
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

	return account.IsSubscriptionActive(b.clock.Now()), nil
}

func (b *baseBilling) GetSubscriptionIsTrialing(
	ctx context.Context,
	accountId ID[Account],
) (trialing bool, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithField("accountId", accountId)

	log.Debug("checking if account subscription is trial")

	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "cannot determine if account subscription is trial")
	}

	span.Status = sentry.SpanStatusOK

	return account.IsTrialing(b.clock.Now()), nil
}
