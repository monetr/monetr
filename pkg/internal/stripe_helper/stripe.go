package stripe_helper

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/round"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
)

type Stripe interface {
	GetPricesById(ctx context.Context, stripePriceIds []string) ([]stripe.Price, error)
	GetPriceById(ctx context.Context, id string) (*stripe.Price, error)
	GetProductsById(ctx context.Context, stripeProductIds []string) ([]stripe.Product, error)
	CreateCustomer(ctx context.Context, customer stripe.CustomerParams) (*stripe.Customer, error)
	UpdateCustomer(ctx context.Context, id string, customer stripe.CustomerParams) (*stripe.Customer, error)
	GetCustomer(ctx context.Context, id string) (*stripe.Customer, error)
	GetSubscription(ctx context.Context, stripeSubscriptionId string) (*stripe.Subscription, error)
	NewCheckoutSession(ctx context.Context, params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error)
	GetCheckoutSession(ctx context.Context, checkoutSessionId string) (*stripe.CheckoutSession, error)
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
	base := &stripeBase{
		log:    log,
		client: nil,
		cache:  &noopStripeCache{},
	}

	config := &stripe.BackendConfig{
		HTTPClient: &http.Client{
			Transport: round.NewObservabilityRoundTripper(http.DefaultTransport, base.stripeRoundTripper),
			Timeout:   time.Second * 30,
		},
		LeveledLogger: log,
	}

	base.client = stripe_client.New(apiKey, &stripe.Backends{
		API:     stripe.GetBackendWithConfig(stripe.APIBackend, config),
		Connect: stripe.GetBackendWithConfig(stripe.ConnectBackend, config),
		Uploads: stripe.GetBackendWithConfig(stripe.UploadsBackend, config),
	})

	return base
}

func NewStripeHelperWithCache(log *logrus.Entry, apiKey string, cacheClient cache.Cache) Stripe {
	base := &stripeBase{
		log:    log,
		client: nil,
		cache:  NewRedisStripeCache(log, cacheClient),
	}

	config := &stripe.BackendConfig{
		HTTPClient: &http.Client{
			Transport: round.NewObservabilityRoundTripper(http.DefaultTransport, base.stripeRoundTripper),
			Timeout:   time.Second * 30,
		},
		LeveledLogger: log,
	}

	base.client = stripe_client.New(apiKey, &stripe.Backends{
		API:     stripe.GetBackendWithConfig(stripe.APIBackend, config),
		Connect: stripe.GetBackendWithConfig(stripe.ConnectBackend, config),
		Uploads: stripe.GetBackendWithConfig(stripe.UploadsBackend, config),
	})

	return base
}

func (s *stripeBase) stripeRoundTripper(
	ctx context.Context,
	request *http.Request,
	response *http.Response,
	err error,
) {
	var statusCode int
	var requestId string
	if response != nil {
		statusCode = response.StatusCode
		requestId = response.Header.Get("Request-Id")
	}
	// If you get a nil reference panic here during testing, its probably because you forgot to mock a certain endpoint.
	// Check to see if the error is a "no responder found" error.
	crumbs.HTTP(ctx,
		"Stripe API Call",
		"stripe",
		request.URL.String(),
		request.Method,
		statusCode,
		map[string]interface{}{
			"Request-Id": requestId,
		},
	)
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

	log := s.log.WithContext(span.Context()).WithField("stripePriceId", id)

	if price, ok := s.cache.GetPriceById(span.Context(), id); ok {
		return price, nil
	}

	result, err := s.client.Prices.Get(id, &stripe.PriceParams{
		Params: stripe.Params{
			Context: span.Context(),
		},
	})
	if err != nil {
		log.WithError(err).Error("failed to retrieve stripe price")
		return nil, errors.Wrap(err, "failed to retrieve stripe price")
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
		ListParams: stripe.ListParams{
			Context: span.Context(),
		},
		IDs: productIds,
	})

	products := make([]stripe.Product, 0)
	for {
		if err := productIterator.Err(); err != nil {
			return nil, errors.Wrap(err, "failed to iterate over products")
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

func (s *stripeBase) GetSubscription(ctx context.Context, stripeSubscriptionId string) (*stripe.Subscription, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetSubscription")
	defer span.Finish()

	result, err := s.client.Subscriptions.Get(stripeSubscriptionId, &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: span.Context(),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve subscription")
	}

	return result, nil
}

func (s *stripeBase) CreateCustomer(ctx context.Context, customer stripe.CustomerParams) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - CreateCustomer")
	defer span.Finish()

	span.Status = sentry.SpanStatusOK

	customer.Context = span.Context()

	result, err := s.client.Customers.New(&customer)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to create customer")
	}

	return result, err
}

func (s *stripeBase) UpdateCustomer(ctx context.Context, id string, customer stripe.CustomerParams) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - UpdateCustomer")
	defer span.Finish()

	customer.Context = span.Context()

	result, err := s.client.Customers.Update(id, &customer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update customer")
	}

	return result, nil
}

func (s *stripeBase) GetCustomer(ctx context.Context, id string) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - UpdateCustomer")
	defer span.Finish()

	result, err := s.client.Customers.Get(id, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: span.Context(),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve customer")
	}

	return result, nil
}

func (s *stripeBase) NewPortalSession(ctx context.Context, params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
	span := sentry.StartSpan(ctx, "Stripe - NewPortalSession")
	defer span.Finish()

	params.Context = span.Context()

	result, err := s.client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create billing portal session")
	}

	return result, nil
}

func (s *stripeBase) NewCheckoutSession(ctx context.Context, params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	span := sentry.StartSpan(ctx, "Stripe - NewCheckoutSession")
	defer span.Finish()

	span.Status = sentry.SpanStatusOK

	params.Context = span.Context()

	result, err := s.client.CheckoutSessions.New(params)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to create new checkout session")
	}

	if result != nil {
		span.Data = map[string]interface{}{
			"checkoutSessionId": result.ID,
		}
	} else {
		span.Data = map[string]interface{}{
			"checkoutSessionId": nil,
		}
	}

	return result, err
}

func (s *stripeBase) GetCheckoutSession(ctx context.Context, checkoutSessionId string) (*stripe.CheckoutSession, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetCheckoutSession")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"checkoutSessionId": checkoutSessionId,
	}

	span.Status = sentry.SpanStatusOK

	result, err := s.client.CheckoutSessions.Get(checkoutSessionId, &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: span.Context(),
		},
	})
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve stripe checkout session")
	}

	return result, err
}
