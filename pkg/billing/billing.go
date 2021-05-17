package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type BillingHelper interface {
	GetSubscriptionIsActive(ctx context.Context, userId, accountId uint64) (active bool, _ error)
	GetActiveSubscription(ctx context.Context, userId, accountId uint64) (*models.Subscription, error)
	PurgeCache(ctx context.Context, accountId uint64) error
}

var (
	_ BillingHelper = &billingHelperBase{}
)

type billingHelperBase struct {
	log   *logrus.Entry
	cache *redis.Pool
	db    pg.DBI
}

func NewBillingHelper(log *logrus.Entry, cache *redis.Pool, db pg.DBI) BillingHelper {
	return &billingHelperBase{
		log:   log,
		cache: cache,
		db:    db,
	}
}

func (b *billingHelperBase) GetSubscriptionIsActive(ctx context.Context, userId, accountId uint64) (active bool, _ error) {
	span := sentry.StartSpan(ctx, "Billing - GetSubscriptionIsActive")
	defer span.Finish()

	subscription, err := b.GetActiveSubscription(span.Context(), userId, accountId)
	if err != nil {
		return false, err
	}

	return subscription.IsActive(), nil
}

func (b *billingHelperBase) GetActiveSubscription(ctx context.Context, userId, accountId uint64) (*models.Subscription, error) {
	span := sentry.StartSpan(ctx, "Billing - GetActiveSubscription")
	defer span.Finish()

	if subscription, err := b.checkCacheForSubscription(span.Context(), accountId); err == nil && subscription != nil {
		return subscription, nil
	}

	return b.getSubscriptionFromDatabase(span.Context(), userId, accountId)
}

func (b *billingHelperBase) PurgeCache(ctx context.Context, accountId uint64) error {
	span := sentry.StartSpan(ctx, "Billing - PurgeCache")
	defer span.Finish()

	conn := b.cache.Get()
	defer conn.Close()

	return errors.Wrap(conn.Send("DEL", b.getCacheKey(accountId)), "failed to purge cache")
}

func (b *billingHelperBase) getSubscriptionFromDatabase(ctx context.Context, userId, accountId uint64) (*models.Subscription, error) {
	span := sentry.StartSpan(ctx, "Billing - GetSubscriptionFromDatabase")
	defer span.Finish()

	log := b.log.WithFields(logrus.Fields{
		"userId":    userId,
		"accountId": accountId,
	}).WithContext(ctx)

	repo := repository.NewRepositoryFromSession(userId, accountId, b.db)

	subscription, err := repo.GetActiveSubscription(span.Context())
	if err != nil || subscription == nil {
		return subscription, err
	}

	data, err := json.Marshal(subscription)
	if err != nil {
		log.WithError(err).Warn("failed to marshal subscription for caching, it will not be cached")
		return subscription, nil
	}

	conn := b.cache.Get()
	defer conn.Close()

	// Store the subscription in the cache for an hour.
	if err = conn.Send(
		"SET", b.getCacheKey(accountId), string(data),
		"EXAT", time.Now().Add(1*time.Hour).Unix(),
	); err != nil {
		log.WithError(err).Warn("failed to store subscription in cache")
	}

	return subscription, nil
}

func (b *billingHelperBase) checkCacheForSubscription(ctx context.Context, accountId uint64) (*models.Subscription, error) {
	span := sentry.StartSpan(ctx, "Billing - CheckCacheForSubscription")
	defer span.Finish()

	conn := b.cache.Get()
	defer conn.Close()

	result, err := conn.Do("GET", b.getCacheKey(accountId))
	if err != nil {
		return nil, errors.Wrap(err, "failed to check cache for subscription")
	}

	var data []byte
	switch actual := result.(type) {
	case string:
		data = []byte(actual)
	case *string:
		if actual != nil {
			data = []byte(*actual)
		}
	case []byte:
		data = actual
	}

	if len(data) == 0 {
		return nil, nil
	}

	var subscription models.Subscription
	if err = json.Unmarshal(data, &subscription); err != nil {
		return nil, errors.Wrap(err, "failed to read subscription data from cache")
	}

	return &subscription, nil
}

func (b *billingHelperBase) getCacheKey(accountId uint64) string {
	return fmt.Sprintf("subscription:%d", accountId)
}
