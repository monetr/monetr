package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/mock_stripe"
	"github.com/monetr/monetr/pkg/swag"
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
				WithCookie(TestCookieName, token).
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
			result := e.GET("/api/billing/checkout/{checkoutSessionId}").
				WithPath("checkoutSessionId", checkoutSessionId).
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().True()
			result.JSON().Path("$.nextUrl").String().Equal("/")
		}
	})

	t.Run("will show active", func(t *testing.T) {
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

		{ // Make sure that initially the customer's subscription is not present or active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().False()
			result.JSON().Path("$.hasSubscription").Boolean().False()
		}

		var checkoutSessionId string
		{ // Create a checkout session
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
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
			result := e.GET("/api/billing/checkout/{checkoutSessionId}").
				WithPath("checkoutSessionId", checkoutSessionId).
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().True()
			result.JSON().Path("$.nextUrl").String().Equal("/")
		}

		{ // Then once it's all said and done, make sure the customer's subscription shows as present and active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().True()
			result.JSON().Path("$.hasSubscription").Boolean().True()
		}

		// Make sure that if the customer attempts to create a checkout session when they already have a subscription
		// that it will fail.
		{
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
				WithJSON(swag.CreateCheckoutSessionRequest{
					PriceId:    nil,
					CancelPath: nil,
				}).
				Expect()

			result.Status(http.StatusBadRequest)
			result.JSON().Path("$.error").String().Equal("there is already an active subscription for your account")
		}
	})
}
