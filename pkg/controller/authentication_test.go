package controller_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/monetr/monetr/pkg/internal/mock_stripe"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/swag"
	"github.com/monetr/monetr/pkg/verification"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
	})

	t.Run("cannot login without TOTP when enabled", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		// Then configure the login fixture with TOTP.
		fixtures.GivenIHaveTOTPForLogin(t, user.Login)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    user.Login.Email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusPreconditionRequired)
		response.JSON().Path("$.error").String().Equal("login requires MFA")
		response.JSON().Path("$.code").String().Equal("MFA_REQUIRED")
		response.Cookies().Empty()
	})

	t.Run("can login when TOTP is provided", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		// Then configure the login fixture with TOTP.
		loginTotp := fixtures.GivenIHaveTOTPForLogin(t, user.Login)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    user.Login.Email,
				Password: password,
				TOTP:     loginTotp.Now(),
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
	})

	t.Run("can provide TOTP when it is not enabled", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    user.Login.Email,
				Password: password,
				TOTP:     "123456",
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
	})

	t.Run("bad cookie name", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Server.Cookies.Name = ""
		e := NewTestApplicationWithConfig(t, conf)

		// We need to provision the login directly, because the token should fail otherwise.
		login, password := fixtures.GivenIHaveLogin(t)
		fixtures.GivenIHaveAnAccount(t, login)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    login.Email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().Equal("An internal error occurred.")
		response.Cookies().Empty()
	})

	t.Run("no users", func(t *testing.T) {
		e := NewTestApplication(t)
		// Creating the login fixture directly prevents it from also creating a user and an account.
		login, password := fixtures.GivenIHaveLogin(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    login.Email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().Equal("user has no accounts")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("multiple users", func(t *testing.T) {
		e := NewTestApplication(t)
		// Creating the login fixture directly prevents it from also creating a user and an account.
		login, password := fixtures.GivenIHaveLogin(t)

		user1 := fixtures.GivenIHaveAnAccount(t, login)
		assert.Equal(t, login.LoginId, user1.LoginId, "user should have the given login")
		user2 := fixtures.GivenIHaveAnAccount(t, login)
		assert.Equal(t, login.LoginId, user2.LoginId, "user should have the given login")

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    login.Email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.users").Array().Length().Equal(2) // Should have 2 accounts.
		response.JSON().Path("$.users..accountId").Array().Contains(user1.AccountId)
		response.JSON().Path("$.users..accountId").Array().Contains(user2.AccountId)
	})

	t.Run("invalid email", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "notan.email",
				Password: "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: email address provided is not valid")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("invalid email weird parser", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "Barry Gibbs <bg@example.com>",
				Password: "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: email address provided is not valid")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("password to short", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    "example@example.com",
				Password: "short",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("login is not valid: password must be at least 8 characters")
		response.JSON().Object().NotContainsKey("token")
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

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.nextUrl").String().Equal("/account/subscribe")

		stripeMock.AssertNCustomersCreated(t, 1)
	})

	t.Run("bad password", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    gofakeit.Email(),
				Password: "badPassword",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").Equal("invalid email and password")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("valid captcha", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(t,
			"POST", "https://www.google.com/recaptcha/api/siteverify",
			func(t *testing.T, request *http.Request) (interface{}, int) {
				return map[string]interface{}{
					"success":      true,
					"challenge_ts": time.Now(),
					"hostname":     "monetr.mini",
					"score":        1.0,
				}, http.StatusOK
			},
			nil,
		)

		config := NewTestApplicationConfig(t)
		config.ReCAPTCHA.Enabled = true
		config.ReCAPTCHA.VerifyLogin = true
		config.ReCAPTCHA.PublicKey = gofakeit.UUID()
		config.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		e := NewTestApplicationWithConfig(t, config)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
				Captcha:  myownsanity.StringP("Believe it or not, I am a valid captcha"),
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
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

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("valid ReCAPTCHA is required: captcha is not valid")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithBytes([]byte("{bad json}")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("malformed json: invalid character 'b' looking for beginning of object key string")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("with unverified email", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Email.Enabled = true
		config.Email.Verification.Enabled = true
		config.Email.Verification.TokenLifetime = 5 * time.Second
		config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????/")
		config.Email.Domain = "monetr.mini"
		e := NewTestApplicationWithConfig(t, config)

		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(swag.LoginRequest{
				Email:    email,
				Password: password,
			}).
			Expect()

		response.Status(http.StatusPreconditionRequired)
		response.JSON().Path("$.error").String().Equal("email address is not verified")
		response.JSON().Path("$.code").String().Equal("EMAIL_NOT_VERIFIED")
		response.JSON().Object().NotContainsKey("token")
	})
}

func TestLogout(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		email, password := GivenIHaveLogin(t, e)

		var token string
		{ // Login to monetr and retrieve our token cookie.
			response := e.POST("/api/authentication/login").
				WithJSON(swag.LoginRequest{
					Email:    email,
					Password: password,
				}).
				Expect()

			response.Status(http.StatusOK)
			token = AssertSetTokenCookie(t, response)
		}

		{ // Then logout and make sure it removes the cookie.
			response := e.GET("/api/authentication/logout").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Headers().ContainsKey("Set-Cookie")
			cookies := response.Raw().Cookies()
			assert.Len(t, cookies, 1, "must contain only one cookie")
			cookie := cookies[0]
			assert.Empty(t, cookie.Value, "value should be blank to unset the cookie")
		}
	})

	t.Run("no cookie", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.GET("/api/authentication/logout").
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").Equal("authentication required")
	})
}

func TestRegister(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)

		response.JSON().Path("$.nextUrl").String().Equal("/setup")
		response.JSON().Path("$.isActive").Boolean().True()
		response.JSON().Path("$.user").Object().NotEmpty()
		response.JSON().Path("$.user.login").Object().NotEmpty()
		response.JSON().Path("$.user.account").Object().NotEmpty()
	})

	t.Run("beta code not provided", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Beta.EnableBetaCodes = true
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("beta code required for registration")
	})

	t.Run("invalid beta code", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Beta.EnableBetaCodes = true
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			BetaCode  string `json:"betaCode"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()
		registerRequest.BetaCode = "123456"

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().Equal("could not verify beta code: record does not exist")
	})

	t.Run("bad password", func(t *testing.T) {
		e := NewTestApplication(t)
		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 4)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("invalid registration: password must be at least 8 characters")
	})

	t.Run("bad timezone", func(t *testing.T) {
		e := NewTestApplication(t)
		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Timezone  string `json:"timezone"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 10)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()
		registerRequest.Timezone = "going for broke"

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("failed to parse timezone: unknown time zone going for broke")
	})

	t.Run("valid captcha", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(t,
			"POST", "https://www.google.com/recaptcha/api/siteverify",
			func(t *testing.T, request *http.Request) (interface{}, int) {
				return map[string]interface{}{
					"success":      true,
					"challenge_ts": time.Now(),
					"hostname":     "monetr.mini",
					"score":        1.0,
				}, http.StatusOK
			},
			nil,
		)

		config := NewTestApplicationConfig(t)
		config.ReCAPTCHA.Enabled = true
		config.ReCAPTCHA.VerifyRegister = true
		config.ReCAPTCHA.PublicKey = gofakeit.UUID()
		config.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Captcha   string `json:"captcha"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()
		registerRequest.Captcha = "I am a valid captcha"

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.nextUrl").String().Equal("/setup")
		response.JSON().Path("$.isActive").Boolean().True()
		response.JSON().Path("$.user").Object().NotEmpty()
		response.JSON().Path("$.user.login").Object().NotEmpty()
		response.JSON().Path("$.user.account").Object().NotEmpty()
	})

	t.Run("invalid captcha", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.ReCAPTCHA.Enabled = true
		config.ReCAPTCHA.VerifyRegister = true
		config.ReCAPTCHA.PublicKey = gofakeit.UUID()
		config.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Captcha   string `json:"captcha"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()
		registerRequest.Captcha = "I am not a valid captcha"

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("valid ReCAPTCHA is required: invalid ReCAPTCHA: remote error codes: [invalid-input-secret]")
	})

	t.Run("invalid json", func(t *testing.T) {
		e := NewTestApplication(t)
		response := e.POST(`/api/authentication/register`).
			WithBytes([]byte("I am not a valid json body")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("invalid register JSON: invalid character 'I' looking for beginning of value")
	})

	t.Run("email already exists", func(t *testing.T) {
		e := NewTestApplication(t)
		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.login").Object().NotEmpty()
			response.JSON().Path("$.user.account").Object().NotEmpty()
		}

		{ // Send the same register request again, this time it should result in an error.
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusInternalServerError)
			response.JSON().Path("$.error").Equal("failed to create login: a login with the same email already exists")
		}
	})

	t.Run("with billing", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		stripeMock := mock_stripe.NewMockStripeHelper(t)
		stripeMock.MockStripeCreateCustomerSuccess(t)

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

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.isActive").Boolean().False()
		response.JSON().Path("$.nextUrl").String().Equal("/account/subscribe")
		response.JSON().Path("$.user").Object().NotEmpty()
		response.JSON().Path("$.user.login").Object().NotEmpty()
		response.JSON().Path("$.user.account").Object().NotEmpty()

		stripeMock.AssertNCustomersCreated(t, 1)
	})

	t.Run("requires email verification", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Email.Enabled = true
		config.Email.Verification.Enabled = true
		config.Email.Verification.TokenLifetime = 5 * time.Second
		config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
		config.Email.Domain = "monetr.mini"
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.message").String().Equal("A verification email has been sent to your email address, please verify your email.")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("verification token lifespan is too short", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Email.Enabled = true
		config.Email.Verification.Enabled = true
		config.Email.Verification.TokenLifetime = 1 * time.Millisecond
		config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
		config.Email.Domain = "monetr.mini"
		e := NewTestApplicationWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().Equal("could not generate email verification token: lifetime must be greater than 1 second")
		response.JSON().Object().NotContainsKey("token")
	})
}

func TestVerifyEmail(t *testing.T) {
	config := NewTestApplicationConfig(t)
	config.Email.Enabled = true
	config.Email.Verification.Enabled = true
	config.Email.Verification.TokenLifetime = 5 * time.Second
	config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
	config.Email.Domain = "monetr.mini"

	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

		tokenGenerator := verification.NewTokenGenerator(config.Email.Verification.TokenSecret)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		assert.Empty(t, app.Mail.Sent, "no emails should have been sent yet")

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.message").String().Equal("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Make sure that an email was sent with the token.
			assert.Len(t, app.Mail.Sent, 1, "there should have been one email sent now")
			email := app.Mail.Sent[0]
			assert.Equal(t, registerRequest.Email, email.To, "email should have been sent to the provided address")
		}

		{ // Now that we have registered using this email. Try to login without verifying.
			response := e.POST("/api/authentication/login").
				WithJSON(swag.LoginRequest{
					Email:    registerRequest.Email,
					Password: registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().Equal("email address is not verified")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(swag.VerifyRequest{
					Token: verificationToken,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.nextUrl").String().Equal("/login")
			response.JSON().Path("$.message").String().Equal("Your email is now verified. Please login.")
		}

		{ // Now try to login AFTER we have verified the email address.
			response := e.POST("/api/authentication/login").
				WithJSON(swag.LoginRequest{
					Email:    registerRequest.Email,
					Password: registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
		}
	})

	t.Run("bad verification token", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		// Create a token generator with a different secret so it will always generate invalid tokens.
		tokenGenerator := verification.NewTokenGenerator(gofakeit.UUID())

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.message").String().Equal("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(swag.VerifyRequest{
					Token: verificationToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("invalid email verification: could not verify email: invalid token: signature is invalid")
		}

		{ // Make sure that even when the verify endpoint fails, that our login is still not verified.
			response := e.POST("/api/authentication/login").
				WithJSON(swag.LoginRequest{
					Email:    registerRequest.Email,
					Password: registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().Equal("email address is not verified")
		}
	})

	t.Run("expired verification code", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		tokenGenerator := verification.NewTokenGenerator(config.Email.Verification.TokenSecret)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.message").String().Equal("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 1*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			time.Sleep(2 * time.Second) // Make the code expire

			response := e.POST("/api/authentication/verify").
				WithJSON(swag.VerifyRequest{
					Token: verificationToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("invalid email verification: could not verify email: invalid token: token is expired by 1s")
		}

		{ // Make sure that even when the verify endpoint fails, that our login is still not verified.
			response := e.POST("/api/authentication/login").
				WithJSON(swag.LoginRequest{
					Email:    registerRequest.Email,
					Password: registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().Equal("email address is not verified")
		}
	})

	t.Run("blank verification code", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST("/api/authentication/verify").
			WithJSON(swag.VerifyRequest{
				Token: "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("token cannot be blank")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST("/api/authentication/verify").
			WithBytes([]byte("{bad json}")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("malformed JSON: invalid character 'b' looking for beginning of object key string")
	})
}

func TestResendVerificationEmail(t *testing.T) {
	config := NewTestApplicationConfig(t)
	config.Email.Enabled = true
	config.Email.Verification.Enabled = true
	config.Email.Verification.TokenLifetime = 2 * time.Second
	config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
	config.Email.Domain = "monetr.mini"

	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

		var registerRequest struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		registerRequest.Email = testutils.GetUniqueEmail(t)
		registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
		registerRequest.FirstName = gofakeit.FirstName()
		registerRequest.LastName = gofakeit.LastName()

		assert.Empty(t, app.Mail.Sent, "no emails should have been sent yet")

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.message").String().Equal("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Make sure that an email was sent with the token.
			assert.Len(t, app.Mail.Sent, 1, "there should have been one emails sent now")
			email := app.Mail.Sent[0]
			assert.Equal(t, registerRequest.Email, email.To, "email should have been sent to the provided address")
		}

		{ // Now try to resend the verification email.
			response := e.POST("/api/authentication/verify/resend").
				WithJSON(swag.ResendVerificationRequest{
					Email: registerRequest.Email,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty() // This endpoint should not return anything if it succeeds.
		}

		{ // Now make sure that we have actually sent another email.
			assert.Len(t, app.Mail.Sent, 2, "there should have been two emails sent now")
			email := app.Mail.Sent[1]
			assert.Equal(t, registerRequest.Email, email.To, "email should have been sent to the provided address")
		}
	})

	t.Run("non-existent email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

		response := e.POST("/api/authentication/verify/resend").
			WithJSON(swag.ResendVerificationRequest{
				Email: testutils.GetUniqueEmail(t),
			}).
			Expect()

		response.Status(http.StatusOK)
		response.Body().Empty() // Even if the email provided does not exist, don't indicate anything to the client.

		assert.Empty(t, app.Mail.Sent, "no emails should have been sent, address is not associated with a login")
	})

	t.Run("blank email", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST("/api/authentication/verify/resend").
			WithJSON(swag.ResendVerificationRequest{
				Email: "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("email must be provided to resend verification link")
	})
}

func TestSendForgotPassword(t *testing.T) {
	conf := NewTestApplicationConfig(t)
	conf.Email.Enabled = true
	conf.Email.ForgotPassword.Enabled = true
	conf.Email.ForgotPassword.TokenLifetime = 5 * time.Second
	conf.Email.ForgotPassword.TokenSecret = gofakeit.Generate("????????????????????????")
	conf.Email.Domain = "monetr.mini"

	t.Run("sends email for real login", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, conf)

		email, _ := GivenIHaveLogin(t, e)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = email

		// Make sure we are starting with a clean slate.
		assert.Empty(t, app.Mail.Sent, "no emails should have been sent yet")

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		assert.Len(t, app.Mail.Sent, 1, "should have sent a single email to reset password")
	})

	t.Run("success for non-existent email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, conf)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = testutils.GetUniqueEmail(t)

		// Make sure we are starting with a clean slate.
		assert.Empty(t, app.Mail.Sent, "no emails should have been sent yet")

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		// Make sure that even though the request succeeded, no emails were sent since the email address was not real.
		assert.Empty(t, app.Mail.Sent, "no emails should have been sent yet")
	})

	t.Run("with unverified email", func(t *testing.T) {
		verificationConf := conf
		verificationConf.Email.Verification.Enabled = true
		verificationConf.Email.Verification.TokenLifetime = time.Second * 10
		verificationConf.Email.Verification.TokenSecret = gofakeit.UUID()
		app, e := NewTestApplicationExWithConfig(t, verificationConf)

		email, _ := GivenIHaveLogin(t, e)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = email

		assert.Len(t, app.Mail.Sent, 1, "should contain the verification email")
		assert.Equal(t, "Verify Your Email Address", app.Mail.Sent[0].Subject)

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().Equal("You must verify your email before you can send forgot password requests.")
		}

		assert.Len(t, app.Mail.Sent, 1, "should not have sent another email")
	})

	t.Run("with verified email", func(t *testing.T) {
		verificationConf := conf
		verificationConf.Email.Verification.Enabled = true
		verificationConf.Email.Verification.TokenLifetime = time.Second * 10
		verificationConf.Email.Verification.TokenSecret = gofakeit.UUID()
		tokenGenerator := verification.NewTokenGenerator(verificationConf.Email.Verification.TokenSecret)
		app, e := NewTestApplicationExWithConfig(t, verificationConf)

		email, _ := GivenIHaveLogin(t, e)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = email

		assert.Len(t, app.Mail.Sent, 1, "should contain the verification email")
		assert.Equal(t, "Verify Your Email Address", app.Mail.Sent[0].Subject)

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(swag.VerifyRequest{
					Token: verificationToken,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.nextUrl").String().Equal("/login")
			response.JSON().Path("$.message").String().Equal("Your email is now verified. Please login.")
		}

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		assert.Len(t, app.Mail.Sent, 2, "should now have sent 2 emails")
	})

	t.Run("with bad json body", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithBytes([]byte("not json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("malformed JSON: invalid character 'o' in literal null (expecting 'u')")
		}
	})

	t.Run("with blank email", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(map[string]interface{}{
					"email": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("Must provide an email address.")
		}
	})

	t.Run("with blank captcha", func(t *testing.T) {
		captchaConf := conf
		captchaConf.ReCAPTCHA.Enabled = true
		captchaConf.ReCAPTCHA.PublicKey = gofakeit.UUID()
		captchaConf.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		captchaConf.ReCAPTCHA.VerifyForgotPassword = true
		e := NewTestApplicationWithConfig(t, captchaConf)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(map[string]interface{}{
					"email":   testutils.GetUniqueEmail(t),
					"captcha": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("Must provide a valid ReCAPTCHA.")
		}
	})
}

func TestResetPassword(t *testing.T) {
	conf := NewTestApplicationConfig(t)
	conf.Email.Enabled = true
	conf.Email.ForgotPassword.Enabled = true
	conf.Email.ForgotPassword.TokenLifetime = 5 * time.Second
	conf.Email.ForgotPassword.TokenSecret = gofakeit.Generate("????????????????????????")
	conf.Email.Domain = "monetr.mini"

	tokenGenerator := verification.NewTokenGenerator(conf.Email.ForgotPassword.TokenSecret)

	t.Run("happy path", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)
		user, password := fixtures.GivenIHaveABasicAccount(t)

		{ // Make sure we can log in with the current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": password,
				}).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
		}

		// Generate a new password to reset to.
		newPassword := gofakeit.Generate("????????")
		assert.NotEqual(t, password, newPassword, "make sure the new password does not match the old one")

		{ // Reset the password.
			token, err := tokenGenerator.GenerateToken(context.Background(), user.Login.Email, time.Second*5)
			assert.NoError(t, err, "must be able to generate a password reset token")

			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": newPassword,
					"token":    token,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		{ // Try to log in with the old password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": password,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
			response.JSON().Path("$.error").String().Equal("invalid email and password")
			response.Cookies().Empty()
		}

		{ // Try to log in with the new password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": newPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
		}
	})

	t.Run("invalidates multiple tokens after reset", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)
		user, _ := fixtures.GivenIHaveABasicAccount(t)

		firstToken, err := tokenGenerator.GenerateToken(context.Background(), user.Login.Email, time.Second*5)
		assert.NoError(t, err, "must be able to generate a password reset token")

		secondToken, err := tokenGenerator.GenerateToken(context.Background(), user.Login.Email, time.Second*5)
		assert.NoError(t, err, "must be able to generate a password reset token")

		// Make sure the issued_at on the tokens are definitely in the past.
		time.Sleep(1 * time.Second)

		// Generate a new password to reset to.
		newPassword := gofakeit.Generate("????????")

		{ // Reset the password using the first token.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": newPassword,
					"token":    firstToken,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		{ // Try to reset the password with the second token, it should fail.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": "aDifferentPassword",
					"token":    secondToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("Password has already been reset, you must request another password reset link.")
		}
	})

	t.Run("token cannot be used twice", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)
		user, _ := fixtures.GivenIHaveABasicAccount(t)

		token, err := tokenGenerator.GenerateToken(context.Background(), user.Login.Email, time.Second*5)
		assert.NoError(t, err, "must be able to generate a password reset token")

		// Generate a new password to reset to.
		newPassword := gofakeit.Generate("????????")

		{ // Reset the password using the first token.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": newPassword,
					"token":    token,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
		}

		{ // Try to reset the password with the second token, it should fail.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": "aDifferentPassword",
					"token":    token,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("Password has already been reset, you must request another password reset link.")
		}
	})

	t.Run("token must not be expired", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)
		user, _ := fixtures.GivenIHaveABasicAccount(t)

		token, err := tokenGenerator.GenerateToken(context.Background(), user.Login.Email, time.Second*1)
		assert.NoError(t, err, "must be able to generate a password reset token")

		// Wait for the token to expire.
		time.Sleep(2 * time.Second)

		// Generate a new password to reset to.
		newPassword := gofakeit.Generate("????????")

		{ // Try to reset the password using the expired token.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": newPassword,
					"token":    token,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Contains("Failed to validate password reset token: invalid token: token is expired by")
		}
	})

	t.Run("reset password login does not exist", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		email := testutils.GetUniqueEmail(t)
		token, err := tokenGenerator.GenerateToken(context.Background(), email, time.Second*5)
		assert.NoError(t, err, "must be able to generate a password reset token")

		response := e.POST(`/api/authentication/reset`).
			WithJSON(map[string]interface{}{
				"password": "doesn'tEvenMatter",
				"token":    token,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().Equal("Failed to verify login for email address: record does not exist")
	})

	t.Run("invalid json", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		response := e.POST(`/api/authentication/reset`).
			WithBytes([]byte("I am not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("malformed JSON: invalid character 'I' looking for beginning of value")
	})

	t.Run("token is empty", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		response := e.POST(`/api/authentication/reset`).
			WithJSON(map[string]interface{}{
				"password": "doesn'tEvenMatter",
				"token":    "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("Token must be provided to reset password.")
	})

	t.Run("password too short", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		email := testutils.GetUniqueEmail(t)
		token, err := tokenGenerator.GenerateToken(context.Background(), email, time.Second*5)
		assert.NoError(t, err, "must be able to generate a password reset token")

		response := e.POST(`/api/authentication/reset`).
			WithJSON(map[string]interface{}{
				"password": "short",
				"token":    token,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("Password must be at least 8 characters long.")
	})

	t.Run("invalid token", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		response := e.POST(`/api/authentication/reset`).
			WithJSON(map[string]interface{}{
				"password": "thisIsAPasswordForSure",
				"token":    "notAToken",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("Failed to validate password reset token: invalid token: token contains an invalid number of segments")
	})
}
