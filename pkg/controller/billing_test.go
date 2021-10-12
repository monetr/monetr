package controller_test

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/mock_stripe"
	"github.com/monetr/monetr/pkg/swag"
	"net/http"
	"testing"
)

func TestGetAfterCheckout(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		stripeMock := mock_stripe.NewMockStripeHelper(t)

		stripeMock.MockStripeCreateCustomerSuccess(t)
		stripeMock.MockNewCheckoutSession(t)
		stripeMock.MockGetCheckoutSession(t)
		stripeMock.MockGetSubscription(t)

		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.BillingEnabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.InitialPlan = &config.Plan{
			FreeTrialDays: 0,
			Visible:       true,
			StripePriceId: mock_stripe.FakeStripePriceId(t),
			Default:       true,
		}

		e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)

		// Make sure that our customer has been created.
		stripeMock.AssertNCustomersCreated(t, 1)

		var checkoutSessionId string
		{ // Create a checkout session
			result := e.POST("/api/billing/create_checkout").
				WithHeader("M-Token", token).
				WithJSON(swag.CreateCheckoutSessionRequest{
					PriceId:    nil,
					CancelPath: nil,
				}).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.sessionId").String().NotEmpty()
			checkoutSessionId = result.JSON().Path("$.sessionId").String().Raw()
		}

		// Mark the checkout session as complete.
		stripeMock.CompleteCheckoutSession(t, checkoutSessionId)

		{ // Then do the callback from the frontend to complete the checkout session for our application.
			result := e.GET(fmt.Sprintf("/api/billing/checkout/%s", checkoutSessionId)).
				WithHeader("M-Token", token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().True()
			result.JSON().Path("$.nextUrl").String().Equal("/")
		}
	})
}
