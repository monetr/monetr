package mock_stripe

import (
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func MockStripeCreateCustomerSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", "/v1/customers",
		func(t *testing.T, request *http.Request) (interface{}, int) {

			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			return stripe.Customer{
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
			}, http.StatusOK
		},
	)
}
