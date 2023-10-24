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
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/verification"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
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
			WithJSON(map[string]interface{}{
				"email":    user.Login.Email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusPreconditionRequired)
		response.JSON().Path("$.error").String().IsEqual("login requires MFA")
		response.JSON().Path("$.code").String().IsEqual("MFA_REQUIRED")
		response.Cookies().IsEmpty()
	})

	t.Run("login true path with TOTP", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		// Then configure the login fixture with TOTP.
		loginTotp := fixtures.GivenIHaveTOTPForLogin(t, user.Login)

		{ // Send the initial request and make sure it responds with the error.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("login requires MFA")
			response.JSON().Path("$.code").String().IsEqual("MFA_REQUIRED")
		}

		{ // Then try to authenticate using the code.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": password,
					"totp":     loginTotp.Now(),
				}).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
		}
	})

	t.Run("can login when TOTP is provided", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		// Then configure the login fixture with TOTP.
		totpCode := fixtures.GivenIHaveTOTPCodeForLogin(t, user.Login)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    user.Login.Email,
				"password": password,
				"totp":     totpCode,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
	})

	t.Run("can provide TOTP when it is not enabled", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    user.Login.Email,
				"password": password,
				"totp":     "123456",
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
			WithJSON(map[string]interface{}{
				"email":    login.Email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("An internal error occurred.")
		response.Cookies().IsEmpty()
	})

	t.Run("no users", func(t *testing.T) {
		e := NewTestApplication(t)
		// Creating the login fixture directly prevents it from also creating a user and an account.
		login, password := fixtures.GivenIHaveLogin(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    login.Email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("user has no accounts")
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
			WithJSON(map[string]interface{}{
				"email":    login.Email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.users").Array().Length().IsEqual(2) // Should have 2 accounts.
		response.JSON().Path("$.users..accountId").Array().ContainsAll(user1.AccountId, user1.AccountId)
	})

	t.Run("invalid email", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    "notan.email",
				"password": "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Email address provided is not valid")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("invalid email weird parser", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    "Barry Gibbs <bg@example.com>",
				"password": "atLeastThisIsAPassword",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Email address provided is not valid")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("password to short", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    "example@example.com",
				"password": "short",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Password must be at least 8 characters")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("trialing login", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.BillingEnabled = true
		// Force the trial to be expired immediately.
		config.Stripe.FreeTrialDays = 4
		e := NewTestApplicationWithConfig(t, config)

		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Object().NotContainsKey("nextUrl")
	})

	t.Run("inactive subscription", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.BillingEnabled = true
		// Force the trial to be expired immediately.
		config.Stripe.FreeTrialDays = -1
		e := NewTestApplicationWithConfig(t, config)

		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusOK)
		AssertSetTokenCookie(t, response)
		response.JSON().Path("$.nextUrl").String().IsEqual("/account/subscribe")
	})

	t.Run("bad password", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    gofakeit.Email(),
				"password": "badPassword",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").IsEqual("invalid email and password")
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
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
				"captcha":  "Believe it or not, I am a valid captcha",
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
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("valid ReCAPTCHA is required")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.POST("/api/authentication/login").
			WithBytes([]byte("{bad json}")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().
			Path("$.error").
			IsEqual("malformed json")
		response.JSON().Object().NotContainsKey("token")
	})

	t.Run("with unverified email", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Email.Enabled = true
		config.Email.Verification.Enabled = true
		config.Email.Verification.TokenLifetime = 5 * time.Second
		config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????/")
		config.Email.Domain = "monetr.mini"
		app, e := NewTestApplicationExWithConfig(t, config)

		MustSendVerificationEmail(t, app, 1)
		email, password := GivenIHaveLogin(t, e)

		response := e.POST("/api/authentication/login").
			WithJSON(map[string]interface{}{
				"email":    email,
				"password": password,
			}).
			Expect()

		response.Status(http.StatusPreconditionRequired)
		response.JSON().Path("$.error").String().IsEqual("email address is not verified")
		response.JSON().Path("$.code").String().IsEqual("EMAIL_NOT_VERIFIED")
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
				WithJSON(map[string]interface{}{
					"email":    email,
					"password": password,
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

		response.Status(http.StatusOK)
		response.Body().IsEmpty()
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

		{ // Register a new user.
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)

			response.JSON().Path("$.nextUrl").String().IsEqual("/setup")
			response.JSON().Path("$.requireVerification").Boolean().IsFalse()
		}

		token := GivenILogin(t, e, registerRequest.Email, registerRequest.Password)
		{ // Get the current user to see what the state of the account is.
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").Number().Gt(0)
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsFalse()
			response.JSON().Path("$.trialingUntil").IsNull()
		}
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
		response.JSON().Path("$.error").String().IsEqual("beta code required for registration")
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
		response.JSON().Path("$.error").String().IsEqual("could not verify beta code: record does not exist")
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
		response.JSON().Path("$.error").IsEqual("Password must be at least 8 characters")
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
		response.JSON().Path("$.error").IsEqual("failed to parse timezone")
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
		response.JSON().Path("$.nextUrl").String().IsEqual("/setup")
		response.JSON().Path("$.requireVerification").Boolean().IsFalse()
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
		response.JSON().
			Path("$.error").
			String().
			IsEqual("valid ReCAPTCHA is required")
	})

	t.Run("invalid json", func(t *testing.T) {
		e := NewTestApplication(t)
		response := e.POST(`/api/authentication/register`).
			WithBytes([]byte("I am not a valid json body")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("invalid JSON body")
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
		}

		{ // Send the same register request again, this time it should result in an error.
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.code").IsEqual("EMAIL_IN_USE")
			response.JSON().Path("$.error").IsEqual("email already in use")
		}
	})

	t.Run("with billing", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.BillingEnabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = 30
		conf.Stripe.InitialPlan = &config.Plan{
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
		{ // Register for a new account
			registerRequest.Email = testutils.GetUniqueEmail(t)
			registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
			registerRequest.FirstName = gofakeit.FirstName()
			registerRequest.LastName = gofakeit.LastName()

			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
			response.JSON().Path("$.nextUrl").String().IsEqual("/setup")
			response.JSON().Path("$.requireVerification").Boolean().IsFalse()
		}

		token := GivenILogin(t, e, registerRequest.Email, registerRequest.Password)
		{ // Get the current user to see what the state of the account is.
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").Number().Gt(0)
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsTrue()
			response.JSON().Path("$.trialingUntil").String().AsDateTime().
				Gt(time.Now().AddDate(0, 0, 29)).
				Lt(time.Now().AddDate(0, 0, 31))
		}
	})

	t.Run("requires email verification", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Email.Enabled = true
		config.Email.Verification.Enabled = true
		config.Email.Verification.TokenLifetime = 5 * time.Second
		config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
		config.Email.Domain = "monetr.mini"
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

		MustSendVerificationEmail(t, app, 1)

		response := e.POST(`/api/authentication/register`).
			WithJSON(registerRequest).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().
			Path("$.message").
			String().
			IsEqual("A verification email has been sent to your email address, please verify your email.")
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
		response.JSON().
			Path("$.error").
			String().
			IsEqual("could not generate email verification token")
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

		MustSendVerificationEmail(t, app, 1)

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().
				Path("$.message").
				String().
				IsEqual("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Now that we have registered using this email. Try to login without verifying.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    registerRequest.Email,
					"password": registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("email address is not verified")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(map[string]interface{}{
					"token": verificationToken,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.nextUrl").String().IsEqual("/login")
			response.JSON().Path("$.message").String().IsEqual("Your email is now verified. Please login.")
		}

		{ // Now try to login AFTER we have verified the email address.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    registerRequest.Email,
					"password": registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusOK)
			AssertSetTokenCookie(t, response)
		}
	})

	t.Run("bad verification token", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

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

		MustSendVerificationEmail(t, app, 1)

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().
				Path("$.message").
				String().
				IsEqual("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(map[string]interface{}{
					"token": verificationToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("Invalid email verification")
		}

		{ // Make sure that even when the verify endpoint fails, that our login is still not verified.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    registerRequest.Email,
					"password": registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("email address is not verified")
		}
	})

	t.Run("expired verification code", func(t *testing.T) {
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

		MustSendVerificationEmail(t, app, 1)

		{
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().
				Path("$.message").
				String().
				IsEqual("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), registerRequest.Email, 1*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			time.Sleep(2 * time.Second) // Make the code expire

			response := e.POST("/api/authentication/verify").
				WithJSON(map[string]interface{}{
					"token": verificationToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("Invalid email verification")
		}

		{ // Make sure that even when the verify endpoint fails, that our login is still not verified.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]interface{}{
					"email":    registerRequest.Email,
					"password": registerRequest.Password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("email address is not verified")
		}
	})

	t.Run("blank verification code", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST("/api/authentication/verify").
			WithJSON(map[string]interface{}{
				"token": "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Token cannot be blank")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST("/api/authentication/verify").
			WithBytes([]byte("{bad json}")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().
			Path("$.error").
			String().
			IsEqual("invalid JSON body")
	})
}

func TestResendVerificationEmail(t *testing.T) {
	config := NewTestApplicationConfig(t)
	config.Email.Enabled = true
	config.Email.Verification.Enabled = true
	config.Email.Verification.TokenLifetime = 2 * time.Second
	config.Email.Verification.TokenSecret = gofakeit.Generate("????????????????????????")
	config.Email.Domain = "monetr.local"

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

		MustSendVerificationEmail(t, app, 2)

		{ // Sign up for monetr, should send 1 verification email
			response := e.POST(`/api/authentication/register`).
				WithJSON(registerRequest).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().
				Path("$.message").
				String().
				IsEqual("A verification email has been sent to your email address, please verify your email.")
			response.JSON().Object().NotContainsKey("token")
		}

		{ // Now try to resend the verification email.
			response := e.POST("/api/authentication/verify/resend").
				WithJSON(map[string]interface{}{
					"email": registerRequest.Email,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty() // This endpoint should not return anything if it succeeds.
		}
	})

	t.Run("non-existent email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

		// Set n to 0, there should not be any emails sent.
		MustSendVerificationEmail(t, app, 0)

		response := e.POST("/api/authentication/verify/resend").
			WithJSON(map[string]interface{}{
				"email": testutils.GetUniqueEmail(t),
			}).
			Expect()

		response.Status(http.StatusOK)
		response.Body().IsEmpty() // Even if the email provided does not exist, don't indicate anything to the client.
	})

	t.Run("blank email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, config)

		// Set n to 0, there should not be any emails sent.
		MustSendVerificationEmail(t, app, 0)

		response := e.POST("/api/authentication/verify/resend").
			WithJSON(map[string]interface{}{
				"email": "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("email must be provided to resend verification link")
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

		MustSendPasswordResetEmail(t, app, 1, email)

		{ // Initiate the password reset email
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}
	})

	t.Run("success for non-existent email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, conf)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = testutils.GetUniqueEmail(t)

		// Because the email does not exist, there should not be any email sent.
		MustSendPasswordResetEmail(t, app, 0)

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}
	})

	t.Run("with unverified email", func(t *testing.T) {
		verificationConf := conf
		verificationConf.Email.Verification.Enabled = true
		verificationConf.Email.Verification.TokenLifetime = time.Second * 10
		verificationConf.Email.Verification.TokenSecret = gofakeit.UUID()
		app, e := NewTestApplicationExWithConfig(t, verificationConf)

		// There should be one email sent for verification
		MustSendVerificationEmail(t, app, 1)

		email, _ := GivenIHaveLogin(t, e)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = email

		// Because they have not verified their email, they should not be able to send a forgot password request. There
		// should be no password reset emails.
		MustSendPasswordResetEmail(t, app, 0)

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("You must verify your email before you can send forgot password requests.")
		}
	})

	t.Run("with verified email", func(t *testing.T) {
		verificationConf := conf
		verificationConf.Email.Verification.Enabled = true
		verificationConf.Email.Verification.TokenLifetime = time.Second * 10
		verificationConf.Email.Verification.TokenSecret = gofakeit.UUID()
		tokenGenerator := verification.NewTokenGenerator(verificationConf.Email.Verification.TokenSecret)
		app, e := NewTestApplicationExWithConfig(t, verificationConf)

		MustSendVerificationEmail(t, app, 1)

		email, _ := GivenIHaveLogin(t, e)

		var resetPasswordRequest struct {
			Email string `json:"email"`
		}
		resetPasswordRequest.Email = email

		{ // Then generate a verification token and try to use it.
			verificationToken, err := tokenGenerator.GenerateToken(context.Background(), email, 10*time.Second)
			assert.NoError(t, err, "must generate verification token")
			assert.NotEmpty(t, verificationToken, "verification token must not be empty")

			response := e.POST("/api/authentication/verify").
				WithJSON(map[string]interface{}{
					"token": verificationToken,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.nextUrl").String().IsEqual("/login")
			response.JSON().Path("$.message").String().IsEqual("Your email is now verified. Please login.")
		}

		// This time since their email is verified, they should receive a forgot password email.
		MustSendPasswordResetEmail(t, app, 1, email)

		{
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(resetPasswordRequest).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}
	})

	t.Run("with bad json body", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, conf)

		// Since it's not a valid request, we should never send an email.
		MustSendPasswordResetEmail(t, app, 0)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithBytes([]byte("not json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("invalid JSON body")
		}
	})

	t.Run("with blank email", func(t *testing.T) {
		app, e := NewTestApplicationExWithConfig(t, conf)

		// Since it's not a valid request, we should never send an email.
		MustSendPasswordResetEmail(t, app, 0)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(map[string]interface{}{
					"email": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Must provide an email address.")
		}
	})

	t.Run("with blank captcha", func(t *testing.T) {
		captchaConf := conf
		captchaConf.ReCAPTCHA.Enabled = true
		captchaConf.ReCAPTCHA.PublicKey = gofakeit.UUID()
		captchaConf.ReCAPTCHA.PrivateKey = gofakeit.UUID()
		captchaConf.ReCAPTCHA.VerifyForgotPassword = true
		app, e := NewTestApplicationExWithConfig(t, captchaConf)

		// Since it's not a valid request, we should never send an email.
		MustSendPasswordResetEmail(t, app, 0)

		{ // Send a request with invalid json body.
			response := e.POST(`/api/authentication/forgot`).
				WithJSON(map[string]interface{}{
					"email":   testutils.GetUniqueEmail(t),
					"captcha": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Must provide a valid ReCAPTCHA.")
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
			response.Body().IsEmpty()
		}

		{ // Try to log in with the old password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": password,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
			response.JSON().Path("$.error").String().IsEqual("invalid email and password")
			response.Cookies().IsEmpty()
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
			response.Body().IsEmpty()
		}

		{ // Try to reset the password with the second token, it should fail.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": "aDifferentPassword",
					"token":    secondToken,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("Password has already been reset, you must request another password reset link.")
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
			response.Body().IsEmpty()
		}

		{ // Try to reset the password with the second token, it should fail.
			response := e.POST(`/api/authentication/reset`).
				WithJSON(map[string]interface{}{
					"password": "aDifferentPassword",
					"token":    token,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().
				Path("$.error").
				String().
				IsEqual("Password has already been reset, you must request another password reset link.")
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
			response.JSON().
				Path("$.error").
				String().
				IsEqual("Failed to validate password reset token")
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
		response.JSON().Path("$.error").String().IsEqual("Failed to verify login for email address: record does not exist")
	})

	t.Run("invalid json", func(t *testing.T) {
		e := NewTestApplicationWithConfig(t, conf)

		response := e.POST(`/api/authentication/reset`).
			WithBytes([]byte("I am not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().
			Path("$.error").
			String().
			IsEqual("invalid JSON body")
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
		response.JSON().Path("$.error").String().IsEqual("Token must be provided to reset password.")
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
		response.JSON().Path("$.error").String().IsEqual("Password must be at least 8 characters long.")
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
		response.JSON().
			Path("$.error").
			String().
			IsEqual("Failed to validate password reset token")
	})
}
