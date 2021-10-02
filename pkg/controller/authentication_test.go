package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/rest-api/pkg/internal/mock_stripe"
	"github.com/monetr/rest-api/pkg/swag"
)

func TestLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.token").String().NotEmpty()
	})

	t.Run("invalid email", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "notan.email",
				Password: "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: email address provided is not valid")
	})

	t.Run("invalid email weird parser", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "Barry Gibbs <bg@example.com>",
				Password: "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: email address provided is not valid")
	})

	t.Run("password to short", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "example@example.com",
				Password: "short",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: password must be at least 8 characters")
	})

	t.Run("inactive subscription", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.BillingEnabled = true
		e := NewTestApplicationWithConfig(t, config)

		stripeMock := mock_stripe.NewMockStripeHelper(t)

		stripeMock.MockStripeCreateCustomerSuccess(t)

		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.token").String().NotEmpty()
		response.JSON().Path("$.nextUrl").String().Equal("/account/subscribe/failure")

		stripeMock.AssertNCustomersCreated(t, 1)
	})

	t.Run("bad password", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    gofakeit.Email(),
				Password: "badPassword",
			}).
			Expect()

		response.Status(http.StatusForbidden)
		response.JSON().Path("$.error").Equal("invalid email and password")
	})

	t.Run("bad captcha", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.ReCAPTCHA.Enabled = true
		config.ReCAPTCHA.VerifyLogin = true
		config.ReCAPTCHA.VerifyRegister = false
		config.ReCAPTCHA.PublicKey = gofakeit.UUID()
		config.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		e := NewTestApplicationWithConfig(t, config)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("valid ReCAPTCHA is required: captcha is not valid")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/authentication/login").
			WithBytes([]byte("{bad json}")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("malformed json: invalid character 'b' looking for beginning of object key string")
	})
}
