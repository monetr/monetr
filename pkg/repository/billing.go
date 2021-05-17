package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
)


type BillingRepository interface {
	GetAccount() (*models.Account, error)
}

var (
	_ BillingRepository = &billingRepositoryBase{}
)

type billingRepositoryBase struct {
	log     *logrus.Entry
	db      pg.DBI
}

func (b *billingRepositoryBase) GetAccount() (*models.Account, error) {
	panic("implement me")
}

func NewBillingRepository(db pg.DBI) BillingRepository {
	return &billingRepositoryBase{
		db: db,
	}
}

// GetActiveSubscription will return an account's active subscription (if there is one, if not an error is returned).
// The subscription object returned will have the Items, Items.Price and Items.Price.Product relations populated.
func (r *repositoryBase) GetActiveSubscription(ctx context.Context) (*models.Subscription, error) {
	span := sentry.StartSpan(ctx, "GetActiveSubscription")
	defer span.Finish()

	var result models.Subscription
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"subscription"."account_id" = ?`, r.AccountId()).
		Where(`"subscription"."status" = ?`, stripe.SubscriptionStatusActive).
		Limit(1).
		Select(&result)
	switch err {
	case pg.ErrNoRows:
		return nil, nil
	case nil:
		break
	default:
		return nil, errors.Wrap(err, "failed to retrieve an active subscription for the current account")
	}

	if result.SubscriptionId == 0 {
		return nil, nil
	}

	return &result, nil
}

func (r *repositoryBase) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {
	span := sentry.StartSpan(ctx, "CreateSubscription")
	defer span.Finish()

	subscription.AccountId = r.AccountId()
	subscription.OwnedByUserId = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), subscription).
		Insert(subscription)

	return errors.Wrap(err, "failed to create subscription")
}