package billing

import (
	"context"
	"github.com/monetrapp/rest-api/pkg/config"
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
	panic("not implemented")
}

func (m *managerBase) CreateSubscription(ctx context.Context, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	panic("not implemented")
}

func (m *managerBase) GetProduct(ctx context.Context, productId string) (*stripe.Product, error) {
	panic("implement me")
}

func (m *managerBase) GetPrice(ctx context.Context, priceId string) (*stripe.Price, error) {
	panic("implement me")
}

func (m *managerBase) GetPrices(ctx context.Context) ([]*stripe.Price, error) {
	panic("not implemented")
}

func (m *managerBase) GetSubscription(ctx context.Context, subscriptionId string) (*stripe.Subscription, error) {
	panic("implement me")
}
