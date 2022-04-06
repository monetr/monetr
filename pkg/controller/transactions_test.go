package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGetTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		var token string
		var bank models.BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t)
			link := fixtures.GivenIHaveAPlaidLink(t, user)
			bank = fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().Equal(10)
	})

	t.Run("pagination", func(t *testing.T) {
		e := NewTestApplication(t)
		var token string
		var bank models.BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t)
			link := fixtures.GivenIHaveAPlaidLink(t, user)
			bank = fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, bank, 70)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // First page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().Equal(25)
		}

		{ // Second page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 25).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().Equal(25)
		}

		{ // Third page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 50).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().Equal(20)
		}
	})
}

func TestPostTransactions(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/bank_accounts/1234/transactions").
			WithCookie(TestCookieName, token).
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

func TestPutTransactions(t *testing.T) {
	t.Run("update transaction name", func(t *testing.T) {
		e := NewTestApplication(t)
		var token string
		var bank models.BankAccount
		var originalTransaction, transaction models.Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t)
			link := fixtures.GivenIHaveAPlaidLink(t, user)
			bank = fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, bank)
			transaction = originalTransaction

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		transaction.Name = "A More Friendly Name"
		assert.NotEqual(t, originalTransaction.Name, transaction.Name, "make sure the names dont somehow match")

		response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", transaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(transaction).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").Number().Equal(transaction.TransactionId)
		response.JSON().Path("$.transaction.name").String().Equal(transaction.Name)
		response.JSON().Path("$.transaction.name").String().NotEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.originalName").String().Equal(originalTransaction.Name)
		response.JSON().Object().NotContainsKey("spending") // Should not be present for non-balance updates.
	})

	t.Run("transaction does not exist", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", 1234).
			WithPath("transactionId", 1234).
			WithCookie(TestCookieName, token).
			WithJSON(models.Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().Equal("failed to retrieve existing transaction for update: record does not exist")
	})

	t.Run("invalid bank account Id", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/00000/transactions/1234`).
			WithCookie(TestCookieName, token).
			WithJSON(models.Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("must specify a valid bank account Id")
	})

	t.Run("invalid transaction Id", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/1234/transactions/0000`).
			WithCookie(TestCookieName, token).
			WithJSON(models.Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("must specify a valid transaction Id")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/1234/transactions/1234`).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I am not really json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().Equal("malformed JSON: invalid character 'I' looking for beginning of value")
	})

	t.Run("no authentication token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.PUT(`/api/bank_accounts/1234/transactions/1234`).
			WithJSON(models.Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("token must be provided")
	})

	t.Run("bad authentication token", func(t *testing.T) {
		e := NewTestApplication(t)

		response := e.PUT(`/api/bank_accounts/1234/transactions/1234`).
			WithCookie(TestCookieName, gofakeit.Generate("????????")).
			WithJSON(models.Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("failed to validate token: token contains an invalid number of segments")
	})
}
