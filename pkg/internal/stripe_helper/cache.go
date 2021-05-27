package stripe_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	"time"
)

type StripeCache interface {
	GetPriceById(ctx context.Context, id string) (*stripe.Price, bool)
	CachePrice(ctx context.Context, price stripe.Price) bool
	Close() error
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
	cache redis.Conn
}

func (r *redisStripeCache) GetPriceById(ctx context.Context, id string) (*stripe.Price, bool) {
	span := sentry.StartSpan(ctx, "Cache - GetPriceById")
	defer span.Finish()

	log := r.log.WithField("stripePriceId", id)

	log.Trace("checking redis cache for Stripe price")
	result, err := r.cache.Do("GET", r.cacheKey(id))
	if err != nil {
		log.WithError(err).Warn("failed to retrieve Stripe price from cache")
		return nil, false
	}

	var data []byte
	switch actual := result.(type) {
	case []byte:
		data = actual
	case string:
		data = []byte(actual)
	case *string:
		data = []byte(*actual)
	default:
		log.Warnf("invalid type returned from redis cache: %T", actual)
		return nil, false
	}

	var price stripe.Price
	if err = json.Unmarshal(data, &price); err != nil {
		log.WithError(err).Warnf("failed to unmarshal Stripe price from redis cache")
		return nil, false
	}

	return &price, true
}

func (r *redisStripeCache) CachePrice(ctx context.Context, price stripe.Price) bool {
	span := sentry.StartSpan(ctx, "Cache - GetPriceById")
	defer span.Finish()

	log := r.log.WithField("stripePriceId", price.ID)

	log.Trace("storing Stripe price in redis cache")

	data, err := json.Marshal(price)
	if err != nil {
		log.WithError(err).Warn("failed to marshal Stripe price for redis cache")
		return false
	}

	if err = r.cache.Send("SET", r.cacheKey(price.ID), data, "EXAT", time.Now().Add(1*time.Hour)); err != nil {
		log.WithError(err).Warn("failed to store Stripe price in redis cache")
		return false
	}

	return true
}

func (r *redisStripeCache) Close() error {
	return r.cache.Close()
}

func (r *redisStripeCache) cacheKey(id string) string {
	return fmt.Sprintf("stripe:prices:%s", id)
}
