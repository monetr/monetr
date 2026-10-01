package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
)

func TestGetTransactionClusters(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			for _, transaction := range fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 3) {
				testutils.MustInsert(t, TransactionCluster{
					AccountId:     bank.AccountId,
					BankAccountId: bank.BankAccountId,
					Name:          transaction.Name,
					OriginalName:  transaction.OriginalName,
					Members:       []ID[Transaction]{transaction.TransactionId},
				})
			}

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(3)
		response.JSON().Path("$[0].bankAccountId").String().IsEqual(bank.BankAccountId.String())
		response.JSON().Path("$[0]").Object().NotContainsKey("members")
	})

	t.Run("pagination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			for _, transaction := range fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 30) {
				testutils.MustInsert(t, TransactionCluster{
					AccountId:     bank.AccountId,
					BankAccountId: bank.BankAccountId,
					Name:          transaction.Name,
					OriginalName:  transaction.OriginalName,
					Members:       []ID[Transaction]{transaction.TransactionId},
				})
			}

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // First page, uses the default limit of 25
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(25)
		}

		{ // Second page
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 25).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(5)
		}

		{ // Smaller limit
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("limit", 10).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(10)
		}
	})

	t.Run("no clusters returns an empty array", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test, but don't create any clusters.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().IsEmpty()
	})

	t.Run("cant get clusters for someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Create a bank account with clusters under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			transaction := fixtures.GivenIHaveATransaction(t, app.Clock, bank)
			testutils.MustInsert(t, TransactionCluster{
				AccountId:     bank.AccountId,
				BankAccountId: bank.BankAccountId,
				Name:          transaction.Name,
				OriginalName:  transaction.OriginalName,
				Members:       []ID[Transaction]{transaction.TransactionId},
			})
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().IsEmpty()
	})

	t.Run("invalid bank account Id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", "bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
	})

	t.Run("invalid pagination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Limit too small
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("limit", 0).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("limit must be at least 1")
		}

		{ // Limit too large
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("limit", 101).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("limit cannot be greater than 100")
		}

		{ // Negative offset
			response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", -1).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("offset cannot be less than 0")
		}
	})

	t.Run("with a valid api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var cluster TransactionCluster

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			transaction := fixtures.GivenIHaveATransaction(t, app.Clock, bank)
			cluster = testutils.MustInsert(t, TransactionCluster{
				AccountId:     bank.AccountId,
				BankAccountId: bank.BankAccountId,
				Name:          transaction.Name,
				OriginalName:  transaction.OriginalName,
				Members:       []ID[Transaction]{transaction.TransactionId},
			})

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", bank.BankAccountId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(1)
		response.JSON().Path("$[0].transactionClusterId").String().IsEqual(cluster.TransactionClusterId.String())
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.GET("/api/bank_accounts/{bankAccountId}/similar").
			WithPath("bankAccountId", "bac_fake").
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
