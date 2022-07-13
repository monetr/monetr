//go:build icons && simple_icons
package controller_test

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/pkg/internal/fixtures"
)

func TestSearchIcon(t *testing.T) {
	t.Run("get amazon icon", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/icons/search").
			WithQuery("name", "amazon").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.svg").String().NotEmpty()
	})
}
