package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/cache"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

// BasicPayWall is used by the API middleware and other operations to restrict access to some features or functionality
// that requires an active subscription.
type BasicPayWall interface {
	GetSubscriptionIsActive(ctx context.Context, accountId uint64) (bool, error)
}

// BasicBilling is used by the Stripe webhooks to maintain a subscription's status within our application. As the status
// of subscription's change or update these functions can be used to keep the status up to date within monetr.
type BasicBilling interface {
	UpdateSubscription(ctx context.Context, customerId, subscriptionId string, activeUntil *time.Time) error
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

// GetSubscriptionIsActive will retrieve the account data from the AccountRepository interface. This means it is
// possible for it to return a stale response within a few seconds. But in general it should be acceptable. When an
// account is updated -> its cache is invalidated. There is likely a very small window where an invalid state could be
// evaluated but it should be fine.
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
				message = "Subscription is active"
			} else if err == nil {
				message = "Subscription is not active"
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
		return false, errors.Wrap(err, "cannot determine if account subscription is active")
	}

	return account.IsSubscriptionActive(), nil
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

func (b *baseBasicBilling) UpdateSubscription(ctx context.Context, customerId, subscriptionId string, activeUntil *time.Time) error {
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

	log = log.WithField("accountId", account.AccountId)

	currentlyActive := account.IsSubscriptionActive()

	account.StripeSubscriptionId = &subscriptionId
	account.SubscriptionActiveUntil = activeUntil

	// Check to see if the subscription status of the account has changed with this update to be.
	if account.IsSubscriptionActive() != currentlyActive {
		// If it has check to see if it was previously active.
		updatedChannelName := fmt.Sprintf("account:%d:subscription:updated", account.AccountId)
		activatedChannelName := fmt.Sprintf("account:%d:subscription:activated", account.AccountId)
		canceledChannelName := fmt.Sprintf("account:%d:subscription:canceled", account.AccountId)

		if err = b.notify.Notify(span.Context(), updatedChannelName, "0"); err != nil {
			log.WithError(err).WithField("channel", updatedChannelName).Warn("failed to send updated notification")
		}

		if currentlyActive {
			log.Info("account subscription is no longer active")
			if err = b.notify.Notify(span.Context(), canceledChannelName, "0"); err != nil {
				log.WithError(err).Warn("failed to send updated notification")
			}
		} else {
			log.Info("account subscription is now active")
			if err = b.notify.Notify(span.Context(), activatedChannelName, "0"); err != nil {
				log.WithError(err).Warn("failed to send updated notification")
			}
		}
	}

	if err = b.repo.UpdateAccount(span.Context(), account); err != nil {
		log.WithError(err).Errorf("failed to update account subscription status")
		return errors.Wrap(err, "failed to update account subscription status")
	}

	return nil
}
