package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/fixtures"
)

func TestGetAccountSettings(t *testing.T) {
	t.Run("retrieve account settings", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/account/settings").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.maxSafeToSpend.enabled").Boolean().False()
		response.JSON().Path("$.maxSafeToSpend.maximum").Number().Equal(0)
	})

	t.Run("unauthorized", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.GET("/api/account/settings").
			WithCookie(TestCookieName, gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("token is not valid")
	})
}
