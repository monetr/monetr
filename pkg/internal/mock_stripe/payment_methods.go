package mock_stripe

import (
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func MockStripeAttachPaymentMethodSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", RegexPath(t, `/v1/payment_methods/.+/attach`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			paymentMethodId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/v1/payment_methods/"), "/attach")
			require.NotEmpty(t, paymentMethodId, "payment method Id must be provided")

			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			customerId := form.Get("customer")
			require.NotEmpty(t, customerId, "customer Id must be provided")

			return &stripe.PaymentMethod{
				APIResource: stripe.APIResource{},
				Card: &stripe.PaymentMethodCard{
					Brand:       "visa",
					Country:     "US",
					Description: "Test credit card",
					ExpMonth:    uint64(time.Now().Month() + 1),
					ExpYear:     uint64(time.Now().Year() + 1),
					Last4:       "1234",
				},
				Customer: &stripe.Customer{
					ID: customerId,
				},
				Created:  time.Now().Add(-1 * time.Hour).Unix(),
				ID:       paymentMethodId,
				Livemode: false,
				Object:   "payment_method",
				Type:     "card",
			}, http.StatusOK
		},
		StripeHeaders,
	)
}

func MockStripeAttachPaymentMethodBadCustomerError(t *testing.T) {
	mockStripeAttachPaymentMethodError(t, NewInternalServerError(t, "payment_method"), http.StatusInternalServerError)
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", RegexPath(t, `/v1/payment_methods/.+/attach`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			paymentMethodId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/v1/payment_methods/"), "/attach")
			require.NotEmpty(t, paymentMethodId, "payment method Id must be provided")

			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			customerId := form.Get("customer")
			require.NotEmpty(t, customerId, "customer Id must be provided")

			return NewResourceMissingError(t, customerId, "customer"), http.StatusBadRequest
		},
		StripeHeaders,
	)
}

func MockStripeAttachPaymentMethodInternalServerError(t *testing.T) {
	mockStripeAttachPaymentMethodError(t, NewInternalServerError(t, "payment_method"), http.StatusInternalServerError)
}

func MockStripeAttachPaymentMethodCardDeclinedError(t *testing.T) {
	mockStripeAttachPaymentMethodError(t, StripeError{
		Error: struct {
			Code    string `json:"code"`
			DocURL  string `json:"doc_url"`
			Message string `json:"message"`
			Param   string `json:"param"`
			Type    string `json:"type"`
		}{
			Code:    "card_declined",
			DocURL:  "https://stripe.com/docs/error-codes/card-declined",
			Message: "Your card was declined.",
			Param:   "",
			Type:    "card_error",
		},
	}, http.StatusPaymentRequired)
}

func mockStripeAttachPaymentMethodError(t *testing.T, stripeError StripeError, statusCode int) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", RegexPath(t, `/v1/payment_methods/.+/attach`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			paymentMethodId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/v1/payment_methods/"), "/attach")
			require.NotEmpty(t, paymentMethodId, "payment method Id must be provided")

			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			customerId := form.Get("customer")
			require.NotEmpty(t, customerId, "customer Id must be provided")

			return stripeError, statusCode
		},
		StripeHeaders,
	)
}

