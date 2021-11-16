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

		response.Status(http.StatusForbidden)
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

		response.Status(http.StatusForbidden)
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

		response.Status(http.StatusForbidden)
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
		response.JSON().Path("$.error").String().Equal("valid ReCAPTCHA is required: remote error codes: [invalid-input-secret]")
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

		tokenGenerator := verification.NewJWTEmailVerification(config.Email.Verification.TokenSecret)

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
		tokenGenerator := verification.NewJWTEmailVerification(gofakeit.UUID())

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

		tokenGenerator := verification.NewJWTEmailVerification(config.Email.Verification.TokenSecret)

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
