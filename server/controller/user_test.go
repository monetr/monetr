package controller_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_stripe"
	"github.com/monetr/monetr/server/security"
	"github.com/stretchr/testify/assert"
	"github.com/xlzd/gotp"
)

func TestMe(t *testing.T) {
	t.Run("authenticated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Then retrieve "me".
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").String().IsASCII()
			response.JSON().Path("$.user.login.loginId").String().IsASCII()
			response.JSON().Path("$.user.login").Object().Keys().IsEqualUnordered([]string{
				// Make sure that no additional fields are exposed ever on the login
				// object. There are some sensitive fields on the login record that
				// should never be readable via the API. This helps make sure those stay
				// out of API responses.
				"loginId",
				"email",
				"firstName",
				"lastName",
				"passwordResetAt",
				"isEmailVerified",
				"emailVerifiedAt",
				"totpEnabledAt",
			})
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsFalse()
			response.JSON().Path("$.trialingUntil").IsNull()
			// Should not have the nextUrl key when billing is not enabled.
			response.JSON().Object().NotContainsKey("nextUrl")
		}
	})

	t.Run("authenticated pending MFA", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		// Then configure the login fixture with TOTP.
		_ = fixtures.GivenIHaveTOTPForLogin(t, app.Clock, user.Login)

		var token string
		{ // Send the initial request and make sure it responds with the error.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": password,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("login requires MFA")
			response.JSON().Path("$.code").String().IsEqual("MFA_REQUIRED")
			token = AssertSetTokenCookie(t, response)
		}

		{ // Then retrieve "me".
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").String().IsASCII()
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsFalse()
			response.JSON().Path("$.trialingUntil").IsNull()
			response.JSON().Path("$.nextUrl").IsEqual("/login/multifactor")
		}
	})

	t.Run("bad token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.GET(`/api/users/me`).
			WithCookie(TestCookieName, gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("no token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.GET(`/api/users/me`).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("billing enabled - on trial", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = 15
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}
		app, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)
		{ // Then retrieve "me".
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").String().IsASCII()
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsTrue()
			response.JSON().Path("$.trialingUntil").String().AsDateTime().
				Gt(app.Clock.Now().AddDate(0, 0, 14)).
				Lt(app.Clock.Now().AddDate(0, 0, 16))
			// Should not have the nextUrl key when billing is not enabled.
			response.JSON().Object().NotContainsKey("nextUrl")
		}
	})

	t.Run("billing enabled - trial expired", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		// Can force a trial to be expired immediately by setting a negative free trial days.
		conf.Stripe.FreeTrialDays = -1
		conf.Stripe.InitialPlan = &config.Plan{
			StripePriceId: mock_stripe.FakeStripePriceId(t),
		}
		_, e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)
		{ // Then retrieve "me".
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").String().IsASCII()
			response.JSON().Path("$.isActive").Boolean().IsFalse()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsFalse()
			response.JSON().Path("$.trialingUntil").String().AsDateTime().
				Lt(time.Now())
			// The me endpoint should return a next url when billing is expired.
			response.JSON().Path("$.nextUrl").String().IsEqual("/account/subscribe")
		}
	})
}

func TestChangePassword(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		newPassword := gofakeit.Generate("????????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		MustSendPasswordChangedEmail(t, app, 1, user.Login.Email)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentPassword": currentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		// This is just here to make sure that the current token still works after changing the password. If this test
		// ever fails at this point then that means the behavior has changed and tokens become invalidated upon changing
		// a password.
		{
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
		}

		{ // Try to authenticate the user with the old password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword, // Old password now.
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
		}

		{ // Login to the user with their new password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": newPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			AssertSetTokenCookie(t, response)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		wrongCurrentPassword := gofakeit.Generate("????????")
		newPassword := gofakeit.Generate("????????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		MustSendPasswordChangedEmail(t, app, 0)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentPassword": wrongCurrentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
			response.JSON().Path("$.error").String().IsEqual("current password provided is not correct")
		}

		{ // Make sure that even though the change password request failed that the password really didn't change.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword, // Old password now.
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			AssertSetTokenCookie(t, response)
		}

		{ // Make super duper extra sure that we cannot authenticate with the "new password".
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": newPassword,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
		}
	})

	t.Run("bad new password", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		newPassword := gofakeit.Generate("????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		MustSendPasswordChangedEmail(t, app, 0)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentPassword": currentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("new password is not valid")
		}
	})

	t.Run("current and new passwords match", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		MustSendPasswordChangedEmail(t, app, 0)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentPassword": currentPassword,
					"newPassword":     currentPassword,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("new password must be different from the current password")
		}
	})

	t.Run("bad json body", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		MustSendPasswordChangedEmail(t, app, 0)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithBytes([]byte("i am not valid json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("invalid JSON body")
		}
	})

	t.Run("token for a non-existent user", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		app, e := NewTestApplicationWithConfig(t, conf)
		token, err := app.Tokens.Create(
			10*time.Minute,
			security.Claims{
				Scope:        security.AuthenticatedScope,
				EmailAddress: gofakeit.Email(),
				UserId:       "user_bogus",
				AccountId:    "acct_bogus",
				LoginId:      "lgn_bogus",
			},
		)
		assert.NoError(t, err, "should not have an error generating a bogus token")

		bogusCurrentPassword := gofakeit.Generate("????????")
		bogusNewPassword := gofakeit.Generate("????????")

		MustSendPasswordChangedEmail(t, app, 0)

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentPassword": bogusCurrentPassword,
					"newPassword":     bogusNewPassword,
				}).
				Expect()

			response.Status(http.StatusInternalServerError)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve current user details")
		}
	})

	t.Run("bad token", func(t *testing.T) {
		app, e := NewTestApplication(t)

		MustSendPasswordChangedEmail(t, app, 0)

		response := e.PUT(`/api/users/security/password`).
			WithCookie(TestCookieName, gofakeit.UUID()).
			WithJSON(map[string]any{
				"currentPassword": gofakeit.Generate("????????"),
				"newPassword":     gofakeit.Generate("????????"),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("no token", func(t *testing.T) {
		app, e := NewTestApplication(t)

		MustSendPasswordChangedEmail(t, app, 0)

		response := e.PUT(`/api/users/security/password`).
			WithJSON(map[string]any{
				"currentPassword": gofakeit.Generate("????????"),
				"newPassword":     gofakeit.Generate("????????"),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestSetupTOTP(t *testing.T) {
	t.Run("setup twice", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		var uri string
		{ // Start the TOTP setup process.
			response := e.POST(`/api/users/security/totp/setup`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.uri").String().NotEmpty()
			response.JSON().Path("$.recoveryCodes").Array().NotEmpty()
			uri = response.JSON().Path("$.uri").String().Raw()
			// Don't do anything with this response
		}

		{ // Validate the TOTP URI
			parsed, err := url.Parse(uri)
			assert.NoError(t, err, "Must be able to parse the TOTP URI")
			assert.NotNil(t, parsed)
			assert.Equal(t, "otpauth", parsed.Scheme)
			assert.Equal(t, "totp", parsed.Host)
			assert.Equal(t, "monetr", parsed.Query().Get("issuer"))
			assert.NotEmpty(t, parsed.Query().Get("secret"), "must have a secret in the URI")
		}

		{ // Start the TOTP setup process again.
			response := e.POST(`/api/users/security/totp/setup`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.uri").String().NotEmpty()
			response.JSON().Path("$.recoveryCodes").Array().NotEmpty()
			secondUri := response.JSON().Path("$.uri").String().Raw()
			assert.NotEqual(t, uri, secondUri, "Must generate a new URI each time")
		}
	})

	t.Run("already setup", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		_ = fixtures.GivenIHaveTOTPForLogin(t, app.Clock, user.Login)

		{ // Start the TOTP setup process.
			response := e.POST(`/api/users/security/totp/setup`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusInternalServerError)
			response.JSON().Path("$.error").String().IsEqual("Failed to setup TOTP")
		}
	})

	t.Run("no token", func(t *testing.T) {
		_, e := NewTestApplication(t)
		response := e.POST(`/api/users/security/totp/setup`).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestConfirmTOTP(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		var uri string
		{ // Start the TOTP setup process.
			response := e.POST(`/api/users/security/totp/setup`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.uri").String().NotEmpty()
			response.JSON().Path("$.recoveryCodes").Array().NotEmpty()
			uri = response.JSON().Path("$.uri").String().Raw()
		}

		var totp *gotp.TOTP
		{ // Validate the TOTP URI
			parsed, err := url.Parse(uri)
			assert.NoError(t, err, "Must be able to parse the TOTP URI")
			assert.NotNil(t, parsed)
			assert.Equal(t, "otpauth", parsed.Scheme)
			assert.Equal(t, "totp", parsed.Host)
			assert.Equal(t, "monetr", parsed.Query().Get("issuer"))
			secret := parsed.Query().Get("secret")
			assert.NotEmpty(t, secret, "must have a secret in the URI")
			totp = gotp.NewDefaultTOTP(secret)
		}

		{ // Enable TOTP for the current login
			response := e.POST(`/api/users/security/totp/confirm`).
				WithJSON(map[string]any{
					"totp": totp.AtTime(app.Clock.Now()),
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // This should not return an MFA required code.
			response := e.POST("/api/authentication/login").
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusPreconditionRequired)
			response.JSON().Path("$.error").String().IsEqual("login requires MFA")
			response.JSON().Path("$.code").String().IsEqual("MFA_REQUIRED")
			response.JSON().Path("$.nextUrl").IsEqual("/login/multifactor")
		}
	})

	t.Run("cant confirm twice", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]any{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		var uri string
		{ // Start the TOTP setup process.
			response := e.POST(`/api/users/security/totp/setup`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.uri").String().NotEmpty()
			response.JSON().Path("$.recoveryCodes").Array().NotEmpty()
			uri = response.JSON().Path("$.uri").String().Raw()
		}

		var totp *gotp.TOTP
		{ // Validate the TOTP URI
			parsed, err := url.Parse(uri)
			assert.NoError(t, err, "Must be able to parse the TOTP URI")
			assert.NotNil(t, parsed)
			assert.Equal(t, "otpauth", parsed.Scheme)
			assert.Equal(t, "totp", parsed.Host)
			assert.Equal(t, "monetr", parsed.Query().Get("issuer"))
			secret := parsed.Query().Get("secret")
			assert.NotEmpty(t, secret, "must have a secret in the URI")
			totp = gotp.NewDefaultTOTP(secret)
		}

		{ // Enable TOTP for the current login
			response := e.POST(`/api/users/security/totp/confirm`).
				WithJSON(map[string]any{
					"totp": totp.AtTime(app.Clock.Now()),
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // Should fail to confirm the second time
			response := e.POST(`/api/users/security/totp/confirm`).
				WithJSON(map[string]any{
					"totp": totp.AtTime(app.Clock.Now()),
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Failed to enable TOTP")
		}
	})

	t.Run("no token", func(t *testing.T) {
		_, e := NewTestApplication(t)
		response := e.POST(`/api/users/security/totp/confirm`).
			WithJSON(map[string]any{
				"totp": "123456",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
