package mock_stripe

import (
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func (m *MockStripeHelper) MockStripeCreateCustomerSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", "/v1/customers",
		func(t *testing.T, request *http.Request) (interface{}, int) {
			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			customer := stripe.Customer{
				Balance:             0,
				Created:             time.Now().Unix(),
				Currency:            "USD",
				DefaultSource:       nil,
				Deleted:             false,
				Delinquent:          false,
				Description:         "",
				Discount:            nil,
				Email:               form.Get("email"),
				ID:                  FakeStripeCustomerId(t),
				InvoicePrefix:       "",
				InvoiceSettings:     nil,
				Livemode:            false,
				Metadata:            nil,
				Name:                form.Get("name"),
				NextInvoiceSequence: 0,
				Object:              "customer",
				Phone:               "",
				PreferredLocales:    nil,
				Shipping:            nil,
				Sources:             nil,
				Subscriptions:       nil,
				Tax:                 nil,
				TaxExempt:           "",
				TaxIDs:              nil,
			}
			m.CreateCustomer(t, &customer)

			return customer, http.StatusOK
		},
		StripeHeaders,
	)
}

func (m *MockStripeHelper) CreateCustomer(t *testing.T, customer *stripe.Customer) {
	for { // Make sure the customer ID is unique
		customer.ID = FakeStripeCustomerId(t)
		if _, ok := m.customers[customer.ID]; !ok {
			break
		}
	}

	m.customers[customer.ID] = *customer
}
