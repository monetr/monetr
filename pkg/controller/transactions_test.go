package controller_test

import (
	"github.com/monetr/monetr/pkg/models"
	"net/http"
	"testing"
)

func TestPostTransactions(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/bank_accounts/1234/transactions").
			WithHeader("M-Token", token).
			WithJSON(models.Transaction{
				BankAccountId: 1234,
				SpendingId:    nil,
				Categories: []string{
					"Things",
				},
				Name:         "I spent money",
				MerchantName: "A place",
				IsPending:    false,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("cannot create transactions for non-manual links")
	})
}
