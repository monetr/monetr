package stripe_helper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/stripe/stripe-go/v81"
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
	log   *slog.Logger
	cache cache.Cache
}

func NewRedisStripeCache(log *slog.Logger, cacheClient cache.Cache) StripeCache {
	return &redisStripeCache{
		log:   log,
		cache: cacheClient,
	}
}

func (r *redisStripeCache) GetPriceById(ctx context.Context, id string) (*stripe.Price, bool) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := r.log.With("stripePriceId", id)

	log.Log(span.Context(), logging.LevelTrace, "checking cache for stripe price")
	var result stripe.Price
	if err := r.cache.GetEz(span.Context(), r.cacheKey(id), &result); err != nil {
		log.WarnContext(span.Context(), "failed to retrieve stripe price from cache", "err", err)
		return nil, false
	}

	if result.ID == "" {
		return nil, false
	}

	log.Log(span.Context(), logging.LevelTrace, "cache hit for stripe price")

	return &result, true
}

func (r *redisStripeCache) CachePrice(ctx context.Context, price stripe.Price) bool {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := r.log.With("stripePriceId", price.ID)

	log.Log(span.Context(), logging.LevelTrace, "storing stripe price in cache")
	if err := r.cache.SetEzTTL(span.Context(), r.cacheKey(price.ID), price, 1*time.Hour); err != nil {
		log.WarnContext(span.Context(), "failed to store stripe price in cache", "err", err)
		return false
	}

	return true
}

func (r *redisStripeCache) cacheKey(id string) string {
	return fmt.Sprintf("stripe:prices:%s", id)
}
