package controller_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_stripe"
)

func TestMe(t *testing.T) {
	t.Run("authenticated", func(t *testing.T) {
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
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
			response.JSON().Path("$.user.userId").Number().Gt(0)
			response.JSON().Path("$.isActive").Boolean().IsTrue()
			response.JSON().Path("$.hasSubscription").Boolean().IsFalse()
			response.JSON().Path("$.isTrialing").Boolean().IsFalse()
			response.JSON().Path("$.trialingUntil").IsNull()
			// Should not have the nextUrl key when billing is not enabled.
			response.JSON().Object().NotContainsKey("nextUrl")
		}
	})

	t.Run("bad token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.GET(`/api/users/me`).
			WithCookie(TestCookieName, gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("no token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.GET(`/api/users/me`).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("billing enabled - on trial", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.BillingEnabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		conf.Stripe.FreeTrialDays = 15
		conf.Stripe.InitialPlan = &config.Plan{
			Visible:       true,
			StripePriceId: mock_stripe.FakeStripePriceId(t),
			Default:       true,
		}
		e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)
		{ // Then retrieve "me".
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
				Gt(time.Now().AddDate(0, 0, 14)).
				Lt(time.Now().AddDate(0, 0, 16))
			// Should not have the nextUrl key when billing is not enabled.
			response.JSON().Object().NotContainsKey("nextUrl")
		}
	})

	t.Run("billing enabled - trial expired", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Stripe.Enabled = true
		conf.Stripe.BillingEnabled = true
		conf.Stripe.APIKey = gofakeit.UUID()
		// Can force a trial to be expired immediately by setting a negative free trial days.
		conf.Stripe.FreeTrialDays = -1
		conf.Stripe.InitialPlan = &config.Plan{
			Visible:       true,
			StripePriceId: mock_stripe.FakeStripePriceId(t),
			Default:       true,
		}
		e := NewTestApplicationWithConfig(t, conf)

		token := GivenIHaveToken(t, e)
		{ // Then retrieve "me".
			response := e.GET(`/api/users/me`).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.user").Object().NotEmpty()
			response.JSON().Path("$.user.userId").Number().Gt(0)
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
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)
		newPassword := gofakeit.Generate("????????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"currentPassword": currentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.Body().Empty()
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
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword, // Old password now.
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
		}

		{ // Login to the user with their new password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
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
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)
		wrongCurrentPassword := gofakeit.Generate("????????")
		newPassword := gofakeit.Generate("????????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"currentPassword": wrongCurrentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
			response.JSON().Path("$.error").String().Equal("current password provided is not correct")
		}

		{ // Make sure that even though the change password request failed that the password really didn't change.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
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
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": newPassword,
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
		}
	})

	t.Run("bad new password", func(t *testing.T) {
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)
		newPassword := gofakeit.Generate("????")

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"currentPassword": currentPassword,
					"newPassword":     newPassword,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("new password is not valid")
		}
	})

	t.Run("current and new passwords match", func(t *testing.T) {
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"currentPassword": currentPassword,
					"newPassword":     currentPassword,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("new password must be different from the current password")
		}
	})

	t.Run("bad json body", func(t *testing.T) {
		e := NewTestApplication(t)
		user, currentPassword := fixtures.GivenIHaveABasicAccount(t)

		var token string
		{ // Login to the user with their current password.
			response := e.POST(`/api/authentication/login`).
				WithJSON(map[string]interface{}{
					"email":    user.Login.Email,
					"password": currentPassword,
				}).
				Expect()

			response.Status(http.StatusOK)
			// Then make sure we get a token back and that it is valid.
			token = AssertSetTokenCookie(t, response)
		}

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithBytes([]byte("i am not valid json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().Equal("invalid JSON body")
		}
	})

	t.Run("token for a non-existent user", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		e := NewTestApplicationWithConfig(t, conf)
		token := GenerateToken(t, conf, math.MaxUint64, math.MaxUint64, math.MaxUint64)

		bogusCurrentPassword := gofakeit.Generate("????????")
		bogusNewPassword := gofakeit.Generate("????????")

		{ // Change the user's password.
			response := e.PUT(`/api/users/security/password`).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"currentPassword": bogusCurrentPassword,
					"newPassword":     bogusNewPassword,
				}).
				Expect()

			response.Status(http.StatusInternalServerError)
			response.JSON().Path("$.error").String().Equal("failed to retrieve current user details")
		}
	})

	t.Run("bad token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.PUT(`/api/users/security/password`).
			WithCookie(TestCookieName, gofakeit.UUID()).
			WithJSON(map[string]interface{}{
				"currentPassword": gofakeit.Generate("????????"),
				"newPassword":     gofakeit.Generate("????????"),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("no token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.PUT(`/api/users/security/password`).
			WithJSON(map[string]interface{}{
				"currentPassword": gofakeit.Generate("????????"),
				"newPassword":     gofakeit.Generate("????????"),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
