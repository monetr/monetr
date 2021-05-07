package stripe_helper

import (
	"context"
	"github.com/stripe/stripe-go/v72"
)

type Stripe interface {
	GetActiveRecurringPrices(ctx context.Context) ([]stripe.Price, error)
}
