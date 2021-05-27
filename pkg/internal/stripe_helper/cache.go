package stripe_helper

import (
	"context"
	"github.com/stripe/stripe-go/v72"
)

type StripeCache interface {
	GetPriceById(ctx context.Context, id string) (*stripe.Price, bool)
	CachePrice(ctx context.Context, price stripe.Price) bool
	Close() error
}

var (
	_ StripeCache = &noopStripeCache{}
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
