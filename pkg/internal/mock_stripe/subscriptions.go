package mock_stripe

import (
	"fmt"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strings"
	"testing"
)

func (m *MockStripeHelper) MockGetSubscription(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"GET", RegexPath(t, `/v1/subscriptions/.*\z`),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			subscriptionId := strings.TrimSpace(strings.TrimPrefix(request.URL.String(), Path(t, "/v1/subscriptions/")))

			if subscriptionId == "" {
				return map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Unrecognized request URL (GET: /v1/subscriptions/). If you are trying to list objects, remove the trailing slash. If you are trying to retrieve an object, make sure you passed a valid (non-empty) identifier in your code. Please see https://stripe.com/docs or we can help at https://support.stripe.com/.",
						"type":    "invalid_request_error",
					},
				}, http.StatusNotFound
			}

			subscription, ok := m.subscriptions[subscriptionId]
			if !ok {
				return map[string]interface{}{
					"error": map[string]interface{}{
						"message": fmt.Sprintf("Invalid subscription id: %s", subscriptionId),
						"type":    "invalid_request_error",
					},
				}, http.StatusNotFound
			}

			return subscription, http.StatusOK
		},
		StripeHeaders,
	)
}

func (m *MockStripeHelper) CreateSubscription(t *testing.T, subscription *stripe.Subscription) {
	for { // Make sure the subscription ID is unique
		subscription.ID = FakeStripeSubscriptionId(t)
		if _, ok := m.subscriptions[subscription.ID]; !ok {
			break
		}
	}

	m.subscriptions[subscription.ID] = *subscription
}
