package stripe_helper

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
)

type StripeCache interface {
	GetPriceById(ctx context.Context, id string) (*stripe.Price, bool)
	CachePrice(ctx context.Context, price stripe.Price) bool
}

var (
	_ StripeCache = &noopStripeCache{}
	_ StripeCache = &redisStripeCache{}
)

type noopStripeCache struct{}

func (n *noopStripeCache) GetPriceById(ctx context.Context, id string) (*stripe.Price, bool) {
	return nil, false
}

func (n *noopStripeCache) CachePrice(ctx context.Context, price stripe.Price) bool {
	return false
}

func (n *noopStripeCache) Close() error {
	return nil
}

type redisStripeCache struct {
	log   *logrus.Entry
	cache cache.Cache
}

func NewRedisStripeCache(log *logrus.Entry, cacheClient cache.Cache) StripeCache {
	return &redisStripeCache{
		log:   log,
		cache: cacheClient,
	}
}

func (r *redisStripeCache) GetPriceById(ctx context.Context, id string) (*stripe.Price, bool) {
	span := sentry.StartSpan(ctx, "Cache - GetPriceById")
	defer span.Finish()

	log := r.log.WithContext(span.Context()).WithField("stripePriceId", id)

	log.Trace("checking redis cache for stripe price")
	var result stripe.Price
	if err := r.cache.GetEz(span.Context(), r.cacheKey(id), &result); err != nil {
		log.WithError(err).Warn("failed to retrieve stripe price from cache")
		return nil, false
	}

	if result.ID == "" {
		return nil, false
	}

	log.Trace("cache hit for stripe price")

	return &result, true
}

func (r *redisStripeCache) CachePrice(ctx context.Context, price stripe.Price) bool {
	span := sentry.StartSpan(ctx, "Cache - GetPriceById")
	defer span.Finish()

	log := r.log.WithContext(span.Context()).WithField("stripePriceId", price.ID)

	log.Trace("storing stripe price in redis cache")
	if err := r.cache.SetEzTTL(span.Context(), r.cacheKey(price.ID), price, 1*time.Hour); err != nil {
		log.WithError(err).Warn("failed to store stripe price in cache")
		return false
	}

	return true
}

func (r *redisStripeCache) cacheKey(id string) string {
	return fmt.Sprintf("stripe:prices:%s", id)
}
