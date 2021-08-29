package mock_stripe

import (
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"testing"
)

type MockStripeHelper struct {
	customers        map[string]stripe.Customer
	checkoutSessions map[string]stripe.CheckoutSession
	subscriptions    map[string]stripe.Subscription
}

func NewMockStripeHelper(t *testing.T) *MockStripeHelper {
	return &MockStripeHelper{
		customers:        map[string]stripe.Customer{},
		checkoutSessions: map[string]stripe.CheckoutSession{},
		subscriptions:    map[string]stripe.Subscription{},
	}
}

func (m *MockStripeHelper) AssertNCustomersCreated(t *testing.T, n int) {
	assert.Len(t, m.customers, n, "should have X customers created")
}

func (m *MockStripeHelper) AssertNCheckSessionsCreated(t *testing.T, n int) {
	assert.Len(t, m.checkoutSessions, n, "should have X checkout sessions created")
}

func (m *MockStripeHelper) AssertNSubscriptionsCreated(t *testing.T, n int) {
	assert.Len(t, m.subscriptions, n, "should have X subscriptions created")
}


func StripeHeaders(t *testing.T, request *http.Request, response interface{}, status int) map[string][]string {
	return map[string][]string{
		"Request-Id": {
			FakeStripeRequestId(t),
		},
		"X-Stripe-C-Cost": {
			"0",
		},
		"Stripe-Version": {
			stripe.APIVersion,
		},
	}
}
