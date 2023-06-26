package mock_stripe

import (
	"fmt"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func (m *MockStripeHelper) MockNewCheckoutSession(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", RegexPath(t, `/v1/checkout/sessions\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			body, err := ioutil.ReadAll(request.Body)
			require.NoError(t, err, "failed to read request body")

			form, err := url.ParseQuery(string(body))
			require.NoError(t, err, "failed to parse body")

			if _, ok := m.customers[form.Get("customer")]; !ok {
				panic("customer not found")
			}

			stripeForm, err := ParseStripeForm(form)
			if err != nil {
				panic("stripe form must be valid")
			}

			var lineItems []*stripe.LineItem
			if lineItemsForm, ok := stripeForm["line_items"].(map[string]interface{}); ok {
				lineItems = make([]*stripe.LineItem, len(lineItemsForm))
				for indexStr, tupleRaw := range lineItemsForm {
					var index int64
					if index, err = strconv.ParseInt(indexStr, 10, 64); err != nil {
						// We need to be able to parse these keys as indexes to an array.
						panic(err)
					}

					item := lineItems[index]
					if item == nil {
						lineItems[index] = &stripe.LineItem{
							Object: "price",
						}
						item = lineItems[index]
					}

					tuple, ok := tupleRaw.(map[string]interface{})
					require.True(t, ok, "must be able to convert tupleRaw into a map")

					// If a price ID is provided, then include that on the line item object.
					if price, ok := tuple["price"].(string); ok {
						item.Price = &stripe.Price{
							ID: price,
						}
					}

					// If a quantity is included then include that.
					if qtyStr, ok := tuple["quantity"].(string); ok {
						qty, err := strconv.ParseInt(qtyStr, 10, 64)
						require.NoError(t, err, "must be able to parse quantity string")
						item.Quantity = qty
					}
				}
			}

			checkoutSession := &stripe.CheckoutSession{
				AllowPromotionCodes:      stripeForm["allow_promotion_codes"] == "true",
				AmountSubtotal:           0,
				AmountTotal:              0,
				AutomaticTax:             nil,
				BillingAddressCollection: "",
				CancelURL:                form.Get("cancel_url"),
				ClientReferenceID:        "",
				Currency:                 "",
				Customer: &stripe.Customer{
					ID: form.Get("customer"),
				},
				CustomerDetails: nil,
				CustomerEmail:   "",
				Deleted:         false,
				ID:              "",
				LineItems: &stripe.LineItemList{
					Data: lineItems,
				},
				Livemode:                  false,
				Locale:                    "",
				Metadata:                  nil,
				Mode:                      "",
				Object:                    "",
				PaymentIntent:             nil,
				PaymentMethodOptions:      nil,
				PaymentMethodTypes:        nil,
				PaymentStatus:             "",
				SetupIntent:               nil,
				Shipping:                  nil,
				ShippingAddressCollection: nil,
				SubmitType:                "",
				Subscription:              nil,
				SuccessURL:                "",
				TaxIDCollection:           nil,
				TotalDetails:              nil,
				URL:                       "",
			}

			m.CreateCheckoutSession(t, checkoutSession)

			return checkoutSession, http.StatusOK
		},
		StripeHeaders,
	)
}

func (m *MockStripeHelper) MockGetCheckoutSession(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/checkout/sessions/.*\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			checkoutSessionId := strings.TrimSpace(strings.TrimPrefix(request.URL.String(), Path(t, "/v1/checkout/sessions/")))

			if checkoutSessionId == "" {
				return map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Unrecognized request URL (GET: /v1/checkout/sessions/). If you are trying to list objects, remove the trailing slash. If you are trying to retrieve an object, make sure you passed a valid (non-empty) identifier in your code. Please see https://stripe.com/docs or we can help at https://support.stripe.com/.",
						"type":    "invalid_request_error",
					},
				}, http.StatusNotFound
			}

			checkoutSession, ok := m.checkoutSessions[checkoutSessionId]
			if !ok {
				return map[string]interface{}{
					"error": map[string]interface{}{
						"message": fmt.Sprintf("Invalid checkout.session id: %s", checkoutSessionId),
						"type":    "invalid_request_error",
					},
				}, http.StatusNotFound
			}

			return checkoutSession, http.StatusOK
		},
		StripeHeaders,
	)
}

func (m *MockStripeHelper) CreateCheckoutSession(t *testing.T, checkoutSession *stripe.CheckoutSession) {
	for { // Make sure the checkout session ID is unique
		checkoutSession.ID = FakeStripeCheckoutSessionId(t)
		if _, ok := m.checkoutSessions[checkoutSession.ID]; !ok {
			break
		}
	}

	m.checkoutSessions[checkoutSession.ID] = *checkoutSession
}

func (m *MockStripeHelper) CompleteCheckoutSession(t *testing.T, checkoutSessionId string) {
	checkoutSession, ok := m.checkoutSessions[checkoutSessionId]
	assert.Truef(t, ok, "checkout session with ID (%s) not in mock helper", checkoutSessionId)

	customer := m.customers[checkoutSession.Customer.ID]

	subscription := &stripe.Subscription{
		ApplicationFeePercent: 0,
		AutomaticTax:          nil,
		BillingCycleAnchor:    0,
		BillingThresholds:     nil,
		CancelAt:              0,
		CancelAtPeriodEnd:     false,
		CanceledAt:            0,
		CollectionMethod:      "",
		Created:               time.Now().Unix(),
		CurrentPeriodEnd:      time.Now().Add(24 * time.Hour).Unix(),
		CurrentPeriodStart:    time.Now().Unix(),
		Customer: &stripe.Customer{
			ID: customer.ID,
		},
		DefaultSource:                 nil,
		DefaultTaxRates:               nil,
		Discount:                      nil,
		EndedAt:                       0,
		Items:                         nil,
		LatestInvoice:                 nil,
		Livemode:                      false,
		Metadata:                      nil,
		NextPendingInvoiceItemInvoice: 0,
		Object:                        "subscription",
		OnBehalfOf:                    nil,
		PauseCollection:               stripe.SubscriptionPauseCollection{},
		PaymentSettings:               nil,
		PendingInvoiceItemInterval:    stripe.SubscriptionPendingInvoiceItemInterval{},
		PendingSetupIntent:            nil,
		PendingUpdate:                 nil,
		Plan:                          nil,
		Quantity:                      0,
		Schedule:                      nil,
		StartDate:                     0,
		Status:                        stripe.SubscriptionStatusActive,
		TransferData:                  nil,
		TrialEnd:                      0,
		TrialStart:                    0,
	}

	m.CreateSubscription(t, subscription)

	checkoutSession.Subscription = &stripe.Subscription{
		ID: subscription.ID,
	}
	checkoutSession.PaymentStatus = stripe.CheckoutSessionPaymentStatusPaid
	m.checkoutSessions[checkoutSessionId] = checkoutSession
}
