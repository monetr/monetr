package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/mock_stripe"
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
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = -1
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}

		_, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)

		// Customers are no longer created on registration, instead created at checkout.
		stripeMock.AssertNCustomersCreated(t, 0)

		var checkoutSessionId string
		{ // Create a checkout session
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"priceId":    nil,
					"cancelPath": nil,
				}).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.url").String().NotEmpty()
			result.JSON().Path("$.sessionId").String().NotEmpty()
			checkoutSessionId = result.JSON().Path("$.sessionId").String().Raw()
		}
		stripeMock.AssertNCustomersCreated(t, 1)

		// Mark the checkout session as complete.
		stripeMock.CompleteCheckoutSession(t, checkoutSessionId)

		{ // Then do the callback from the frontend to complete the checkout session for our application.
			result := e.GET("/api/billing/checkout/{checkoutSessionId}").
				WithPath("checkoutSessionId", checkoutSessionId).
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.nextUrl").String().IsEqual("/")
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
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = -1
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}

		_, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)

		stripeMock.AssertNCustomersCreated(t, 0)

		{ // Make sure that initially the customer's subscription is not present or active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().IsFalse()
			result.JSON().Path("$.hasSubscription").Boolean().IsFalse()
		}

		var checkoutSessionId string
		{ // Create a checkout session
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"priceId":    nil,
					"cancelPath": nil,
				}).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.url").String().NotEmpty()
			result.JSON().Path("$.sessionId").String().NotEmpty()
			checkoutSessionId = result.JSON().Path("$.sessionId").String().Raw()
		}

		stripeMock.AssertNCustomersCreated(t, 1)

		// Mark the checkout session as complete.
		stripeMock.CompleteCheckoutSession(t, checkoutSessionId)

		{ // Then do the callback from the frontend to complete the checkout session for our application.
			result := e.GET("/api/billing/checkout/{checkoutSessionId}").
				WithPath("checkoutSessionId", checkoutSessionId).
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.nextUrl").String().IsEqual("/")
		}

		{ // Then once it's all said and done, make sure the customer's subscription shows as present and active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.hasSubscription").Boolean().IsTrue()
		}

		// Make sure that if the customer attempts to create a checkout session when they already have a subscription
		// that it will fail.
		{
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"priceId":    nil,
					"cancelPath": nil,
				}).
				Expect()

			result.Status(http.StatusBadRequest)
			result.JSON().Path("$.error").String().IsEqual("There is already an active subscription for your account")
		}
	})

	t.Run("allowed to create checkout while trialing", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		stripeMock := mock_stripe.NewMockStripeHelper(t)

		stripeMock.MockStripeCreateCustomerSuccess(t)
		stripeMock.MockNewCheckoutSession(t)
		stripeMock.MockGetCheckoutSession(t)
		stripeMock.MockGetSubscription(t)

		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = 1
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}

		_, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)

		stripeMock.AssertNCustomersCreated(t, 0)

		{ // Make sure that initially the customer's subscription is not present or active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			// When we initially start we are still in a trial status.
			result.JSON().Path("$.isTrialing").Boolean().IsTrue()
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.hasSubscription").Boolean().IsFalse()
		}

		var checkoutSessionId string
		{ // Create a checkout session
			result := e.POST("/api/billing/create_checkout").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"cancelPath": nil,
				}).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.url").String().NotEmpty()
			result.JSON().Path("$.sessionId").String().NotEmpty()
			checkoutSessionId = result.JSON().Path("$.sessionId").String().Raw()
		}

		stripeMock.AssertNCustomersCreated(t, 1)

		// Mark the checkout session as complete.
		stripeMock.CompleteCheckoutSession(t, checkoutSessionId)

		{ // Then do the callback from the frontend to complete the checkout session for our application.
			result := e.GET("/api/billing/checkout/{checkoutSessionId}").
				WithPath("checkoutSessionId", checkoutSessionId).
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.nextUrl").String().IsEqual("/")
		}

		{ // Then once it's all said and done, make sure the customer's subscription shows as present and active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			// Now that we are subscribed we should not be trialing.
			result.JSON().Path("$.isTrialing").Boolean().IsFalse()
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.hasSubscription").Boolean().IsTrue()
		}
	})
}

func TestGetBillingPortal(t *testing.T) {
	t.Run("cannot create portal without subscription", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		stripeMock := mock_stripe.NewMockStripeHelper(t)

		stripeMock.MockStripeCreateCustomerSuccess(t)
		stripeMock.MockNewCheckoutSession(t)
		stripeMock.MockGetCheckoutSession(t)
		stripeMock.MockGetSubscription(t)

		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = 30
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}

		_, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)

		stripeMock.AssertNCustomersCreated(t, 0)

		{ // Make sure that initially the customer's subscription is not present or active.
			result := e.GET("/api/users/me").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusOK)
			// When we initially start we are still in a trial status.
			result.JSON().Path("$.isTrialing").Boolean().IsTrue()
			result.JSON().Path("$.isActive").Boolean().IsTrue()
			result.JSON().Path("$.hasSubscription").Boolean().IsFalse()
		}

		{ // Create a billing portal session
			result := e.GET("/api/billing/portal").
				WithCookie(TestCookieName, token).
				Expect()

			result.Status(http.StatusBadRequest)
			result.JSON().Path("$.error").String().IsEqual("Account does not have a subscription")
		}
	})
}
