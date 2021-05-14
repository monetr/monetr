package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
)

type BillingRepository interface {
}

var (
	_ BillingRepository = &billingRepositoryBase{}
)

type billingRepositoryBase struct {
	db pg.DBI
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
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve an active subscription for the current account")
	}

	return &result, nil
}
