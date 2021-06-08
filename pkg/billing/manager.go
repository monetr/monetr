package billing

import (
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/stripe/stripe-go/v72"
)

type Billing interface {
	CreateCustomer(login models.Login, user models.User) (*stripe.Customer, error)
	GetActiveSubscription(accountId uint64) (*models.Subscription, error)
	CreateSubscription(uint64) error
	CancelSubscription(accountId uint64)
}
