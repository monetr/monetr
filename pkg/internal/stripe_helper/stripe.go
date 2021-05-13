package stripe_helper

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
	"net/http"
	"time"
)

type Stripe interface {
	GetPricesById(ctx context.Context, stripePriceIds []string) ([]stripe.Price, error)
	GetProductsById(ctx context.Context, stripeProductIds []string) ([]stripe.Product, error)
}

var (
	_ Stripe = &stripeBase{}
)

type stripeBase struct {
	log    *logrus.Entry
	client *stripe_client.API
}

func NewStripeHelper(log *logrus.Entry, apiKey string) Stripe {
	return &stripeBase{
		log: log,
		client: stripe_client.New(apiKey, stripe.NewBackends(&http.Client{
			Timeout: time.Second * 30,
		})),
	}
}

func (s *stripeBase) GetPricesById(ctx context.Context, stripePriceIds []string) ([]stripe.Price, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetPricesById")
	defer span.Finish()

	prices := make([]stripe.Price, len(stripePriceIds))
	for i, stripePriceId := range stripePriceIds {
		price, err := s.GetPriceById(span.Context(), stripePriceId)
		if err != nil {
			return nil, err
		}

		prices[i] = *price
	}

	return prices, nil
}

func (s *stripeBase) GetPriceById(ctx context.Context, id string) (*stripe.Price, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetPriceById")
	defer span.Finish()

	log := s.log.WithField("stripePriceId", id)

	result, err := s.client.Prices.Get(id, &stripe.PriceParams{})
	if err != nil {
		log.WithError(err).Error("failed to retrieve stripe price")
		return nil, errors.Wrap(err, "failed to retrieve stripe price")
	}

	return result, nil
}

func (s *stripeBase) GetProductsById(ctx context.Context, stripeProductIds []string) ([]stripe.Product, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetProductsById")
	defer span.Finish()

	productIds := make([]*string, len(stripeProductIds))
	for i, id := range stripeProductIds { productIds[i] = &id }

	productIterator := s.client.Products.List(&stripe.ProductListParams{
		IDs: productIds,
	})

	products := make([]stripe.Product, 0)
	for {
		if err := productIterator.Err(); err != nil {
			return nil, errors.Wrap(err, "failed to retrieve stripe products")
		}

		if !productIterator.Next() {
			break
		}

		if product := productIterator.Product(); product != nil {
			products = append(products, *product)
		}
	}

	return products, nil
}
