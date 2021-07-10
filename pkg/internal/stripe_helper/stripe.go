package stripe_helper

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
	"net/http"
	"time"
)

type Stripe interface {
	GetPricesById(ctx context.Context, stripePriceIds []string) ([]stripe.Price, error)
	GetPriceById(ctx context.Context, id string) (*stripe.Price, error)
	GetProductsById(ctx context.Context, stripeProductIds []string) ([]stripe.Product, error)
	CreateCustomer(ctx context.Context, customer stripe.CustomerParams) (*stripe.Customer, error)
	UpdateCustomer(ctx context.Context, id string, customer stripe.CustomerParams) (*stripe.Customer, error)
	GetCustomer(ctx context.Context, id string) (*stripe.Customer, error)
	GetSubscription(ctx context.Context, stripeSubscriptionId string) (*stripe.Subscription, error)
	NewPortalSession(ctx context.Context, params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error)
}

var (
	_ Stripe = &stripeBase{}
)

type stripeBase struct {
	log    *logrus.Entry
	client *stripe_client.API
	cache  StripeCache
}

func NewStripeHelper(log *logrus.Entry, apiKey string) Stripe {
	return &stripeBase{
		log: log,
		client: stripe_client.New(apiKey, stripe.NewBackends(&http.Client{
			Timeout: time.Second * 30,
		})),
		cache: &noopStripeCache{},
	}
}

func NewStripeHelperWithCache(log *logrus.Entry, apiKey string, cacheClient cache.Cache) Stripe {
	return &stripeBase{
		log: log,
		client: stripe_client.New(apiKey, stripe.NewBackends(&http.Client{
			Timeout: time.Second * 30,
		})),
		cache: NewRedisStripeCache(log, cacheClient),
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

	if price, ok := s.cache.GetPriceById(span.Context(), id); ok {
		return price, nil
	}

	result, err := s.client.Prices.Get(id, &stripe.PriceParams{})
	if err != nil {
		log.WithError(err).Error("failed to retrieve stripe price")
		return nil, s.wrapStripeError(span.Context(), err, "failed to retrieve stripe price")
	}

	s.cache.CachePrice(span.Context(), *result)

	return result, nil
}

func (s *stripeBase) GetProductsById(ctx context.Context, stripeProductIds []string) ([]stripe.Product, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetProductsById")
	defer span.Finish()

	productIds := make([]*string, len(stripeProductIds))
	for i := range stripeProductIds {
		productIds[i] = &stripeProductIds[i]
	}

	productIterator := s.client.Products.List(&stripe.ProductListParams{
		IDs: productIds,
	})

	products := make([]stripe.Product, 0)
	for {
		if err := productIterator.Err(); err != nil {
			return nil, s.wrapStripeError(span.Context(), err, "failed to retrieve stripe products")
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

func (s *stripeBase) CreateSubscription(ctx context.Context, subscription stripe.SubscriptionParams) (*stripe.Subscription, error) {
	span := sentry.StartSpan(ctx, "Stripe - CreateSubscription")
	defer span.Finish()

	result, err := s.client.Subscriptions.New(&subscription)
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to create subscription")
	}

	return result, nil
}

func (s *stripeBase) GetSubscription(ctx context.Context, stripeSubscriptionId string) (*stripe.Subscription, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetSubscription")
	defer span.Finish()

	result, err := s.client.Subscriptions.Get(stripeSubscriptionId, &stripe.SubscriptionParams{})
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to create customer")
	}

	return result, nil
}

func (s *stripeBase) CreateCustomer(ctx context.Context, customer stripe.CustomerParams) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - CreateCustomer")
	defer span.Finish()

	result, err := s.client.Customers.New(&customer)
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to create customer")
	}

	return result, nil
}

func (s *stripeBase) UpdateCustomer(ctx context.Context, id string, customer stripe.CustomerParams) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - UpdateCustomer")
	defer span.Finish()

	result, err := s.client.Customers.Update(id, &customer)
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to update customer")
	}

	return result, nil
}

func (s *stripeBase) GetCustomer(ctx context.Context, id string) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - UpdateCustomer")
	defer span.Finish()

	result, err := s.client.Customers.Get(id, &stripe.CustomerParams{})
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to retrieve customer")
	}

	return result, nil
}

func (s *stripeBase) NewPortalSession(ctx context.Context, params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
	span := sentry.StartSpan(ctx, "Stripe - NewPortalSession")
	defer span.Finish()

	result, err := s.client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, s.wrapStripeError(span.Context(), err, "failed to create billing portal session")
	}

	return result, nil
}

const (
	cardDeclined = "card_declined"
)

var (
	ErrCardDeclined = errors.New("card was declined")
)

func (s *stripeBase) wrapStripeError(ctx context.Context, input error, msg string) error {
	log := logrus.WithContext(ctx)

	switch err := input.(type) {
	case nil:
		return nil
	case *stripe.Error:
		log = log.WithField("stripeRequestId", err.RequestID)

		switch err.Code {
		case cardDeclined:
			return errors.Wrap(ErrCardDeclined, msg)
		default:
			return errors.Errorf("%s: %s", msg, err.Msg)
		}
	default:
		return errors.Wrap(err, msg)
	}
}
