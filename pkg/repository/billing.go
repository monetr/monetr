package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
)

func (r *repositoryBase) GetProducts(ctx context.Context) ([]models.Product, error) {
	span := sentry.StartSpan(ctx, "GetProducts")
	defer span.Finish()

	result := make([]models.Product, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("Prices").
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve products")
	}

	return result, nil
}

// GetActiveSubscription will return an account's active subscription (if there is one, if not an error is returned).
// The subscription object returned will have the Items, Items.Price and Items.Price.Product relations populated.
func (r *repositoryBase) GetActiveSubscription(ctx context.Context) (*models.Subscription, error) {
	span := sentry.StartSpan(ctx, "GetActiveSubscription")
	defer span.Finish()

	var result models.Subscription
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("Items").
		Relation("Items.Price").
		Relation("Items.Price.Product").
		Where(`"subscription"."account_id" = ?`, r.AccountId()).
		Where(`"subscription"."status" = ?`, stripe.SubscriptionStatusActive).
		Limit(1).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve an active subscription for the current account")
	}

	return &result, nil
}
