package billing

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
)

type Manager interface {
	CreateCustomer(ctx context.Context, name, email string) (*stripe.Customer, error)
	CreateSubscription(ctx context.Context, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	GetProduct(ctx context.Context, productId string) (*stripe.Product, error)
	GetPrice(ctx context.Context, priceId string) (*stripe.Price, error)
	GetSubscription(ctx context.Context, subscriptionId string) (*stripe.Subscription, error)
}

var (
	_ Manager = &managerBase{}
)

type managerBase struct {
	configuration config.Stripe
	client        *stripe_client.API
}

func (m *managerBase) CreateCustomer(ctx context.Context, name, email string) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "Stripe - CreateCustomer")
	defer span.Finish()

	customer, err := m.client.Customers.New(&stripe.CustomerParams{
		Email: &email,
		Name:  &name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stripe customer")
	}

	return customer, nil
}

func (m *managerBase) CreateSubscription(ctx context.Context, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	span := sentry.StartSpan(ctx, "Stripe - CreateSubscription")
	defer span.Finish()

	if params.Customer == nil {
		return nil, errors.Errorf("must specify a customer when creating a subscription")
	}

	// If a payment behavior has not been specified we want to default to "default_incomplete" for our application.
	if params.PaymentBehavior == nil {
		params.PaymentBehavior = stripe.String("default_incomplete")
	}

	// This expand is to help with looking at the payment intent for the new subscription.
	params.AddExpand("latest_invoice.payment_intent")

	subscription, err := m.client.Subscriptions.New(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stripe customer")
	}

	return subscription, nil
}

func (m *managerBase) GetProduct(ctx context.Context, productId string) (*stripe.Product, error) {
	panic("implement me")
}

func (m *managerBase) GetPrice(ctx context.Context, priceId string) (*stripe.Price, error) {
	panic("implement me")
}

func (m *managerBase) GetPrices(ctx context.Context) ([]*stripe.Price, error) {
	span := sentry.StartSpan(ctx, "Stripe - GetPrices")
	defer span.Finish()

	result := m.client.Prices.List(&stripe.PriceListParams{
		ListParams: stripe.ListParams{},
		Active:     stripe.Bool(true),
		Currency:   stripe.String("USD"),
		LookupKeys: m.configuration.Prices,
		Type:       stripe.String("recurring"),
	})

	if err := result.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve stripe prices")
	}

	return result.PriceList().Data, nil
}

func (m *managerBase) GetSubscription(ctx context.Context, subscriptionId string) (*stripe.Subscription, error) {
	panic("implement me")
}
