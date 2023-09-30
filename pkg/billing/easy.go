package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
)

func buildAccountCacheKey(accountId uint64) string {
	return fmt.Sprintf("accounts:%d", accountId)
}

// AccountRepository is used by the pay wall and billing interfaces to retrieve data for an account.
type AccountRepository interface {
	GetAccount(ctx context.Context, accountId uint64) (*models.Account, error)
	GetAccountByCustomerId(ctx context.Context, stripeCustomerId string) (*models.Account, error)
	UpdateAccount(ctx context.Context, account *models.Account) error
}

// BasicBilling is used by the Stripe webhooks to maintain a subscription's status within our application. As the status
// of subscription's change or update these functions can be used to keep the status up to date within monetr.
type BasicBilling interface {
	// UpdateSubscription will set the Stripe customer Id and subscription Id on the account object. It will also set
	// the active until date for the account. The date can be nil. If the date is nil or in the past, the subscription
	// is considered cancelled. A timestamp should also be provided. The timestamp is used to fix race conditions in the
	// webhooks received from Stripe. If the provided timestamp is less than the timestamp of the last change applied
	// to the account, then the change is not applied. The change in subscription is only applied when the provided
	// timestamp is after the timestamp of the previously applied change. This helps solve a problem where sometimes a
	// webhook for a subscription being created (which would have an incomplete status) can be delivered after a update
	// webhook for the same subscription (which would indicate an active status) causing the subscription to incorrectly
	// show as inactive.
	UpdateSubscription(
		ctx context.Context,
		customerId, subscriptionId string,
		status stripe.SubscriptionStatus,
		activeUntil *time.Time,
		timestamp time.Time,
	) error
	// UpdateCustomerSubscription does the same thing that UpdateSubscription does, but does not require that the
	// stripe customerId match any customerId stored. Instead, it will take the provided account and update the customer
	// ID and store it on the account with the new subscription data.
	UpdateCustomerSubscription(
		ctx context.Context,
		account *models.Account,
		customerId, subscriptionId string,
		status stripe.SubscriptionStatus,
		activeUntil *time.Time,
		timestamp time.Time,
	) error
}

var (
	_ AccountRepository = &postgresAccountRepository{}
)

type postgresAccountRepository struct {
	log   *logrus.Entry
	cache cache.Cache
	db    pg.DBI
}

func NewAccountRepository(log *logrus.Entry, cacheClient cache.Cache, db pg.DBI) AccountRepository {
	return &postgresAccountRepository{
		log:   log,
		cache: cacheClient,
		db:    db,
	}
}

func (p *postgresAccountRepository) GetAccount(ctx context.Context, accountId uint64) (*models.Account, error) {
	span := sentry.StartSpan(ctx, "Billing - GetAccount")
	defer span.Finish()

	span.Status = sentry.SpanStatusOK

	log := p.log.WithContext(span.Context()).WithField("accountId", accountId)

	var account models.Account
	if err := p.cache.GetEz(span.Context(), buildAccountCacheKey(accountId), &account); err != nil {
		log.WithError(err).Errorf("failed to retrieve account data from cache")
	}

	if account.AccountId > 0 {
		log.Debugf("returning account from cache")
		return &account, nil
	}

	if err := p.db.ModelContext(span.Context(), &account).
		Where(`"account"."account_id" = ?`, accountId).
		Limit(1).
		Select(&account); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account by Id")
	}

	if err := p.cache.SetEzTTL(span.Context(), buildAccountCacheKey(accountId), account, 30*time.Minute); err != nil {
		log.WithError(err).Warn("failed to store account in cache")
	}

	return &account, nil
}

func (p *postgresAccountRepository) GetAccountByCustomerId(ctx context.Context, stripeCustomerId string) (*models.Account, error) {
	span := sentry.StartSpan(ctx, "Billing - GetAccountByCustomerId")
	defer span.Finish()

	var account models.Account
	if err := p.db.ModelContext(span.Context(), &account).
		Where(`"account"."stripe_customer_id" = ?`, stripeCustomerId).
		Limit(1).
		Select(&account); err != nil {

		span.Status = sentry.SpanStatusInternalError
		if span.Data == nil {
			span.Data = map[string]interface{}{}
		}

		span.Data["stripeCustomerId"] = stripeCustomerId

		return nil, errors.Wrap(err, "failed to retrieve account by customer Id")
	}

	return &account, nil
}

func (p *postgresAccountRepository) UpdateAccount(ctx context.Context, account *models.Account) error {
	span := sentry.StartSpan(ctx, "Billing - UpdateAccount")
	defer span.Finish()

	log := p.log.WithContext(span.Context()).WithField("accountId", account.AccountId)

	log.Trace("updating account")

	_, err := p.db.ModelContext(span.Context(), account).
		Where(`"account"."account_id" = ?`, account.AccountId).
		Update(account)
	if err != nil {
		log.WithError(err).Errorf("failed to update account")
		return errors.Wrap(err, "failed to update account")
	}

	if err = p.cache.SetEzTTL(span.Context(), buildAccountCacheKey(account.AccountId), account, 30*time.Minute); err != nil {
		log.WithError(err).Warn("failed to store account in cache")
	}

	return nil
}

var (
	_ BasicBilling = &baseBasicBilling{}
)

type baseBasicBilling struct {
	log    *logrus.Entry
	repo   AccountRepository
	notify pubsub.Publisher
}

func NewBasicBilling(log *logrus.Entry, repo AccountRepository, notifications pubsub.Publisher) BasicBilling {
	return &baseBasicBilling{
		log:    log,
		repo:   repo,
		notify: notifications,
	}
}

func (b *baseBasicBilling) UpdateSubscription(
	ctx context.Context,
	customerId, subscriptionId string,
	status stripe.SubscriptionStatus,
	activeUntil *time.Time,
	timestamp time.Time,
) error {
	span := sentry.StartSpan(ctx, "Billing - UpdateSubscription")
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"customerId":     customerId,
		"subscriptionId": subscriptionId,
	})

	log.Trace("retrieving account by customer Id")

	account, err := b.repo.GetAccountByCustomerId(span.Context(), customerId)
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

func (b *baseBasicBilling) UpdateCustomerSubscription(
	ctx context.Context,
	account *models.Account,
	customerId, subscriptionId string,
	status stripe.SubscriptionStatus,
	activeUntil *time.Time,
	timestamp time.Time,
) error {
	span := sentry.StartSpan(ctx, "Billing - UpdateCustomerSubscription")
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"customerId":     customerId,
		"subscriptionId": subscriptionId,
		"accountId":      account.AccountId,
	})

	// Set the user for this event, this way webhooks are properly associated with the destination user in our
	// application.
	crumbs.IncludeUserInScope(span.Context(), account.AccountId)

	currentlyActive := account.IsSubscriptionActive()

	// If the timestamp for the last webhook is not nil, and the provided timestamp is not after (<= basically) then
	// perform the update. This is to solve potential race conditions in the order we receive webhooks from Stripe.
	if account.StripeWebhookLatestTimestamp != nil {
		if timestamp.Before(*account.StripeWebhookLatestTimestamp) {
			crumbs.Debug(span.Context(), "Provided timestamp is older than the current subscription timestamp", map[string]interface{}{
				"stored":   *account.StripeWebhookLatestTimestamp,
				"provided": timestamp,
			})
			return nil
		} else if timestamp.Equal(*account.StripeWebhookLatestTimestamp) {
			crumbs.Warn(span.Context(), "Provided timestamp is equal to the current subscription timestamp", "stripe", map[string]interface{}{
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

		crumbs.Debug(span.Context(), "Provided timestamp is after the current subscription timestamp, change will be applied", map[string]interface{}{
			"stored":   *account.StripeWebhookLatestTimestamp,
			"provided": timestamp,
		})
	} else {
		crumbs.Debug(span.Context(), "Current subscription timestamp is nil, webhook will be accepted", map[string]interface{}{
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
	// Add 24 hours to the subscription window. This way Stripe has time to process the subscription payment and update
	// the status for us even if things are running a bit slow. This resolves an issue where the active until date can
	// pass before Stripe has processed the renewal. Causing (usually) around an hour or more of time where monetr
	// believed the subscription to not be active anymore.
	account.SubscriptionActiveUntil = myownsanity.TimeP(activeUntil.Add(24 * time.Hour))
	account.StripeWebhookLatestTimestamp = &timestamp
	account.SubscriptionStatus = &status

	if err := b.repo.UpdateAccount(span.Context(), account); err != nil {
		log.WithError(err).Errorf("failed to update account subscription status")
		return errors.Wrap(err, "failed to update account subscription status")
	}

	// Check to see if the subscription status of the account has changed with this update to be.
	if account.IsSubscriptionActive() != currentlyActive {
		// If it has check to see if it was previously active.
		updatedChannelName := fmt.Sprintf("account:%d:subscription:updated", account.AccountId)
		activatedChannelName := fmt.Sprintf("account:%d:subscription:activated", account.AccountId)
		canceledChannelName := fmt.Sprintf("account:%d:subscription:canceled", account.AccountId)

		if err := b.notify.Notify(span.Context(), updatedChannelName, "0"); err != nil {
			log.WithError(err).WithField("channel", updatedChannelName).Warn("failed to send updated notification")
		}

		if currentlyActive {
			log.Info("account subscription is no longer active")
			if err := b.notify.Notify(span.Context(), canceledChannelName, "0"); err != nil {
				log.WithError(err).Warn("failed to send updated notification")
			}
		} else {
			log.Info("account subscription is now active")
			if err := b.notify.Notify(span.Context(), activatedChannelName, "0"); err != nil {
				log.WithError(err).Warn("failed to send updated notification")
			}
		}
	}

	return nil
}
