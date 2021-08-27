package mock_stripe

import (
	"fmt"
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strings"
	"testing"
	"time"
)

const stripeHost = "https://api.stripe.com"

func RegexPath(t *testing.T, relative string) string {

	require.NotEmpty(t, relative, "relative path cannot be empty")
	return fmt.Sprintf("=~^%s", Path(t, relative))
}

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative)
	return fmt.Sprintf("%s%s", stripeHost, relative)
}

func MockStripeGetPriceSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/prices/.+\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			basePath := Path(t, "/v1/prices/")
			priceId := strings.SplitAfter(request.URL.String(), basePath)[1]
			require.NotEmpty(t, priceId, "priceId from URL cannot be empty")

			return stripe.Price{
				Active:        true,
				BillingScheme: stripe.PriceBillingSchemePerUnit,
				Created:       time.Now().Unix(),
				Currency:      "USD",
				Deleted:       false,
				ID:            priceId,
				Livemode:      false,
				LookupKey:     "",
				Metadata:      nil,
				Nickname:      "Some Price",
				Object:        "price",
				Product: &stripe.Product{
					ID: FakeStripeProductId(t),
				},
				Recurring:         nil,
				Tiers:             nil,
				TiersMode:         "",
				TransformQuantity: nil,
				Type:              "recurring",
				UnitAmount:        199,
				UnitAmountDecimal: 1.99,
			}, http.StatusOK
		},
		StripeHeaders,
	)
}

func MockStripeGetPriceNotFound(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/prices/.+\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			basePath := Path(t, "/v1/prices/")
			priceId := strings.SplitAfter(request.URL.String(), basePath)[1]
			require.NotEmpty(t, priceId, "priceId from URL cannot be empty")

			return NewResourceMissingError(t, priceId, "price"), http.StatusNotFound
		},
		StripeHeaders,
	)
}
