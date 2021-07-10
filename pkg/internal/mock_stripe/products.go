package mock_stripe

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/stripe/stripe-go/v72"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"
)

func MockStripeListProductsSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/products(?)*\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			productIds := make([]string, 0)
			for query, value := range request.URL.Query() {
				queryNonIndexed := strings.Split(query, "[")[0]
				switch queryNonIndexed {
				case "ids":
					productIds = append(productIds, value...)
				}
			}

			var products []*stripe.Product
			if len(productIds) > 0 {
				products = make([]*stripe.Product, len(productIds))
			} else {
				products = make([]*stripe.Product, rand.Int31n(10))
			}

			for i := range products {
				var id string
				if len(productIds) > 0 {
					id = productIds[i]
				} else {
					id = FakeStripeProductId(t)
				}

				products[i] = &stripe.Product{
					Active:              true,
					Attributes:          nil,
					Caption:             "Test caption",
					Created:             time.Now().Unix(),
					DeactivateOn:        nil,
					Deleted:             false,
					Description:         "I am a description",
					ID:                  id,
					Images:              nil,
					Livemode:            false,
					Metadata:            nil,
					Name:                fmt.Sprintf("%s %d", gofakeit.Noun(), i+1),
					Object:              "product",
					PackageDimensions:   nil,
					Shippable:           false,
					StatementDescriptor: "",
					Type:                "service",
					UnitLabel:           "Link",
					Updated:             time.Now().Unix(),
					URL:                 "",
				}
			}

			return stripe.ProductList{
				ListMeta: stripe.ListMeta{
					HasMore: false,
					URL:     "/v1/products",
				},
				Data: products,
			}, http.StatusOK
		},
	)
}

func MockStripeListProductsFailure(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/products(?)*\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			return NewInternalServerError(t, "product"), http.StatusInternalServerError
		},
	)
}
