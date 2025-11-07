package mock_stripe

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/stretchr/testify/require"
)

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
		func(t *testing.T, request *http.Request) (any, int) {
			paymentMethodId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/v1/payment_methods/"), "/attach")
			require.NotEmpty(t, paymentMethodId, "payment method Id must be provided")

			body, err := io.ReadAll(request.Body)
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
