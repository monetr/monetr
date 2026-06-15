package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(10)
	})

	t.Run("pagination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 70)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // First page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(25)
		}

		{ // Second page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 25).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(25)
		}

		{ // Third page
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 50).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(20)
		}
	})

	t.Run("cant get transactions for someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Create a bank account with transactions under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 5)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to list transactions under the other user's bank account
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}
	})
}

func TestPostTransactions(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/bank_accounts/1234/transactions").
			WithCookie(TestCookieName, token).
			WithJSON(Transaction{
				BankAccountId: "bac_bogus",
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
		response.JSON().Path("$.error").IsEqual("must specify a valid bank account Id")
	})

	t.Run("name is required", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"amount": 1200,
				"date":   app.Clock.Now(),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().IsEqual(map[string]any{
			"error": "Invalid request",
			"problems": map[string]any{
				"name": "required key is missing",
			},
		})
	})

	t.Run("date is required", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":   "Foobar",
				"amount": 100,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().IsEqual(map[string]any{
			"error": "Invalid request",
			"problems": map[string]any{
				"date": "required key is missing",
			},
		})
	})

	t.Run("amount is required", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "Foobar",
				"date": app.Clock.Now(),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().IsEqual(map[string]any{
			"error": "Invalid request",
			"problems": map[string]any{
				"amount": "required key is missing",
			},
		})
	})

	t.Run("bogus spending object", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":   "Foobar",
				"date":   app.Clock.Now(),
				"amount": 100,
				// A properly shaped spending Id that does not exist. The ID validation now
				// enforces a length, so a short stub like "spnd_bogus" gets rejected as a
				// bad request before we ever try to look the spending up. We want the not
				// found path at the lookup instead.
				"spendingId": "spnd_01hy4rfqk8z4xv1c2v44cf6abc",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").IsEqual("Could not get spending provided for transaction: record does not exist")
	})

	t.Run("adjusts balance", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 100)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 100)
		}

		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -200, // Earned $2
					"isPending":      false,
					"name":           "I earned some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			// Balance should have gone up by $2, so we should be $1 higher from when
			// we started.
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance + 100)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance + 100)
		}
	})

	t.Run("does not adjust balance", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": false,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			// Make sure that if adjusts balance is false then we do not modify the
			// actual balances for the account.
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
		}
	})

	t.Run("cannot create transactions for Plaid links", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": false,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Cannot create transactions for non-manual links")
		}
	})

	t.Run("nil fields", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		{ // Seed the data for the test.
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Now create our transaction and have it linked to our expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": false,
					"spendingId":     nil,
					"merchantName":   nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Object().Keys().IsEqual([]string{"balance", "transaction"})
		}
	})

	t.Run("will use a spending object", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		{ // Seed the data for the test.
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").String().IsASCII()
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		var spendingId ID[Spending]
		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Now create our transaction and have it linked to our expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": false,
					"spendingId":     spendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.spending.spendingId").String().NotEmpty()
		}
	})

	t.Run("create a transaction for a non-usd bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		{ // Register a new user
			email := testutils.GetUniqueEmail(t)
			password := gofakeit.Password(true, true, true, true, false, 32)
			response := e.POST(`/api/authentication/register`).
				WithJSON(map[string]any{
					"email":     email,
					"password":  password,
					"firstName": gofakeit.FirstName(),
					"lastName":  gofakeit.LastName(),
					// Create an account with a non-default locale such that the currency
					// code should be different.
					"locale":   "ja_JP",
					"timezone": "Asia/Tokyo",
				}).
				Expect()

			response.Status(http.StatusOK)
			token = GivenILogin(t, e, email, password)
		}

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "Manual Link",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		var bankAccountId ID[BankAccount]
		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           BankAccountStatusActive,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.availableBalance").Number().IsEqual(100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(100)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
			response.JSON().Path("$.mask").String().IsEqual("1234")
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEqual("PERSONAL CHECKING")
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
			bankAccountId = ID[BankAccount](response.JSON().Path("$.bankAccountId").String().Raw())
		}

		{ // Create a transaction which should have the same currency as the bank.
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
		}
	})
}

func TestPutTransactions(t *testing.T) {
	t.Run("update transaction name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction, transaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
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
		response.JSON().Path("$.transaction.transactionId").IsEqual(transaction.TransactionId)
		response.JSON().Path("$.transaction.name").String().IsEqual(transaction.Name)
		response.JSON().Path("$.transaction.name").String().NotEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.originalName").String().IsEqual(originalTransaction.Name)
		response.JSON().Object().NotContainsKey("spending") // Should not be present for non-balance updates.
	})

	t.Run("transaction does not exist", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", "bac_bogus").
			WithPath("transactionId", "txn_bogus").
			WithCookie(TestCookieName, token).
			WithJSON(Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
	})

	t.Run("invalid bank account Id", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/1234/transactions/txn_bogus`).
			WithCookie(TestCookieName, token).
			WithJSON(Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
	})

	t.Run("invalid transaction Id word", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, token).
			WithJSON(Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
	})

	t.Run("malformed json", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I am not really json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("invalid JSON body")
	})

	t.Run("no authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PUT(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithJSON(Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("bad authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PUT(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, gofakeit.Generate("????????")).
			WithJSON(Transaction{
				Name:   "PayPal",
				Amount: 1243,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("spend from an expense with more than the transaction amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction, transaction Transaction
		now := app.Clock.Now()

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		transaction = originalTransaction
		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			Name:                   "Payday",
			Description:            "Whenever I get paid",
			RuleSet:                fundingRule,
			ExcludeWeekends:        true,
			WaitForDeposit:         false,
			EstimatedDeposit:       nil,
			LastRecurrence:         nil,
			NextRecurrence:         fundingRule.After(now, false),
			NextRecurrenceOriginal: fundingRule.After(now, false),
		})

		// Create the spending object we want to test spending from, specifically make it so that the spending object has
		// more funds in it than the transaction. Also make the contribution amount the equivalent to the spending amount,
		// this way we can assert easily that the contribution amount changes when the spending object is used. It will
		// always be less than the the target amount because we are never spending the entire amount. It could be zero but
		// it can never be equal to the spending amount.
		spending := testutils.MustInsert(t, Spending{
			Name:                   "Spending test",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           transaction.Amount * 2,
			CurrentAmount:          transaction.Amount * 2,
			NextContributionAmount: transaction.Amount * 2,
			NextRecurrence:         spendingRuleOne.After(now, false),
			RuleSet:                spendingRuleOne,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})

		token = GivenILogin(t, e, user.Login.Email, password)

		// Spend the transaction from the spending object we created.
		transaction.SpendingId = &spending.SpendingId

		response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", transaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(transaction).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").IsEqual(transaction.TransactionId)
		response.JSON().Path("$.transaction.spendingId").IsEqual(*transaction.SpendingId)
		response.JSON().Path("$.transaction.spendingAmount").IsEqual(transaction.Amount)
		// Make sure we spent from the right spending object.
		response.JSON().Path("$.spending[0].spendingId").IsEqual(spending.SpendingId)
		// And make sure we spent the amount we wanted.
		response.JSON().Path("$.spending[0].currentAmount").IsEqual(spending.CurrentAmount - transaction.Amount)
		// Make sure the next contribution gets recalculated.
		response.JSON().Path("$.spending[0].nextContributionAmount").Number().Lt(spending.NextContributionAmount)
	})

	t.Run("update preserves lunch flow transaction id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction, transaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
			transaction = originalTransaction

			// Create a lunch flow transaction record and associate it with the
			// transaction to simulate a transaction that was synced from Lunch Flow.
			lunchFlowTransaction := testutils.MustInsert(t, LunchFlowTransaction{
				AccountId:              user.AccountId,
				LunchFlowBankAccountId: *bank.LunchFlowBankAccountId,
				LunchFlowId:            "lf_test_1234",
				Merchant:               originalTransaction.MerchantName,
				Description:            originalTransaction.Name,
				Date:                   originalTransaction.Date,
				Currency:               "USD",
				Amount:                 originalTransaction.Amount,
				IsPending:              false,
			})
			originalTransaction.LunchFlowTransactionId = &lunchFlowTransaction.LunchFlowTransactionId
			testutils.MustDBUpdate(t, &originalTransaction)
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
		response.JSON().Path("$.transaction.transactionId").IsEqual(transaction.TransactionId)
		response.JSON().Path("$.transaction.name").String().IsEqual(transaction.Name)
		response.JSON().Path("$.transaction.name").String().NotEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.originalName").String().IsEqual(originalTransaction.Name)

		// Verify the lunch flow transaction ID is preserved by reading directly
		// from the database, since it is not included in the JSON response.
		updatedTransaction := testutils.MustRetrieve(t, Transaction{
			TransactionId: transaction.TransactionId,
			AccountId:     bank.AccountId,
			BankAccountId: bank.BankAccountId,
		})
		assert.NotNil(t, updatedTransaction.LunchFlowTransactionId, "lunch flow transaction ID must be preserved after update")
		assert.Equal(t, originalTransaction.LunchFlowTransactionId, updatedTransaction.LunchFlowTransactionId, "lunch flow transaction ID must not change")
	})

	t.Run("update does not overwrite spending amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction, transaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			// Set an initial spending amount on the transaction directly to simulate
			// a transaction that was already spent from an expense.
			originalTransaction.SpendingAmount = myownsanity.Pointer(int64(500))
			testutils.MustDBUpdate(t, &originalTransaction)
			transaction = originalTransaction

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		// Attempt to override the spending amount via user input.
		transaction.SpendingAmount = myownsanity.Pointer(int64(9999))
		assert.NotEqual(t, *originalTransaction.SpendingAmount, *transaction.SpendingAmount, "make sure the spending amounts dont somehow match")

		response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", transaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(transaction).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").IsEqual(transaction.TransactionId)
		// Make sure the spending amount was not overwritten by user input.
		response.JSON().Path("$.transaction.spendingAmount").IsEqual(*originalTransaction.SpendingAmount)
	})

	t.Run("cant put someone elses transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var transaction Transaction

		{ // Create a bank account with a transaction under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			transaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to update the transaction
			response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "Updated Name",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
		}
	})

	t.Run("cannot update fields that shouldn't be updated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction, transaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
			transaction = originalTransaction

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		assert.Nil(t, originalTransaction.DeletedAt, "deleted at should be nil to start with")

		{ // Update the transaction
			now := app.Clock.Now()
			transaction.Source = "other"
			transaction.CreatedAt = now
			transaction.DeletedAt = myownsanity.Pointer(now)

			response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(transaction).
				Expect()

			response.Status(http.StatusOK)
			// Make sure we cannot update any fields that aren't meant to be updated.
			response.JSON().Path("$.transaction.transactionId").IsEqual(transaction.TransactionId)
			response.JSON().Path("$.transaction.source").IsEqual(originalTransaction.Source)
			response.JSON().Path("$.transaction.createdAt").String().AsDateTime(time.RFC3339).IsEqual(originalTransaction.CreatedAt)
			// Make sure that we did not actaully set the deleted at timestamp
			response.JSON().Path("$.transaction.deletedAt").IsNull()
		}

		{ // Make sure that we can still see the transaction
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)

			// Make sure that our original transaction is still visible in the
			// response. Previously this transaction would have been hidden if we did
			// a PUT to the deleted at.
			response.JSON().Path("$[0].transactionId").IsEqual(transaction.TransactionId)
		}
	})
}

func TestPatchTransaction(t *testing.T) {
	t.Run("update transaction name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newName := "A More Friendly Name"
		assert.NotEqual(t, originalTransaction.Name, newName, "make sure the names dont somehow match")

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": newName,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").IsEqual(originalTransaction.TransactionId)
		response.JSON().Path("$.transaction.name").String().IsEqual(newName)
		response.JSON().Path("$.transaction.name").String().NotEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.originalName").String().IsEqual(originalTransaction.Name)
		// Unlike the PUT, the PATCH always returns a spending array. It should be
		// empty here because we did not touch anything that affects an expense.
		response.JSON().Path("$.spending").Array().IsEmpty()
	})

	t.Run("update merchant name on a non manual link", func(t *testing.T) {
		// This one has no PUT equivalent on purpose. We just added merchantName to
		// the non-manual patch schema so make sure a Plaid transaction can actually
		// have its display merchant name updated.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newMerchantName := "A Cleaner Merchant Name"
		assert.NotEqual(t, originalTransaction.MerchantName, newMerchantName, "make sure the merchant names dont somehow match")

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"merchantName": newMerchantName,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.merchantName").String().IsEqual(newMerchantName)
		// The original merchant name from Plaid should be left alone.
		response.JSON().Path("$.transaction.originalMerchantName").String().IsEqual(originalTransaction.OriginalMerchantName)
	})

	t.Run("transaction does not exist", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", "bac_bogus").
			WithPath("transactionId", "txn_bogus").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "PayPal",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
	})

	t.Run("invalid bank account Id", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PATCH(`/api/bank_accounts/1234/transactions/txn_bogus`).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "PayPal",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		// The PATCH handler capitalizes this message where the PUT one does not.
		response.JSON().Path("$.error").String().IsEqual("Must specify a valid bank account Id")
	})

	t.Run("invalid transaction Id word", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PATCH(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "PayPal",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
	})

	t.Run("malformed json", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var transaction Transaction

		{ // Seed the data for the test.
			// The PATCH endpoint reads the request body AFTER it looks up the
			// transaction, so unlike the PUT we need a real transaction to exist or
			// we would just get a 404 before the malformed body is ever evaluated.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			transaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", transaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I am not really json")).
			Expect()

		response.Status(http.StatusBadRequest)
		// The schema parse path surfaces a decode failure as a generic parse error
		// rather than the PUT's "invalid JSON body".
		response.JSON().Path("$.error").String().IsEqual("failed to parse request")
	})

	t.Run("no authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PATCH(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithJSON(map[string]any{
				"name": "PayPal",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("bad authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PATCH(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, gofakeit.Generate("????????")).
			WithJSON(map[string]any{
				"name": "PayPal",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("spend from an expense with more than the transaction amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction
		now := app.Clock.Now()

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			Name:                   "Payday",
			Description:            "Whenever I get paid",
			RuleSet:                fundingRule,
			ExcludeWeekends:        true,
			WaitForDeposit:         false,
			EstimatedDeposit:       nil,
			LastRecurrence:         nil,
			NextRecurrence:         fundingRule.After(now, false),
			NextRecurrenceOriginal: fundingRule.After(now, false),
		})

		// Create the spending object we want to test spending from, specifically
		// make it so that the spending object has more funds in it than the
		// transaction. Also make the contribution amount the equivalent to the
		// spending amount, this way we can assert easily that the contribution
		// amount changes when the spending object is used. It will always be less
		// than the the target amount because we are never spending the entire
		// amount. It could be zero but it can never be equal to the spending
		// amount.
		spending := testutils.MustInsert(t, Spending{
			Name:                   "Spending test",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           originalTransaction.Amount * 2,
			CurrentAmount:          originalTransaction.Amount * 2,
			NextContributionAmount: originalTransaction.Amount * 2,
			NextRecurrence:         spendingRuleOne.After(now, false),
			RuleSet:                spendingRuleOne,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})

		token = GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").IsEqual(originalTransaction.TransactionId)
		response.JSON().Path("$.transaction.spendingId").IsEqual(spending.SpendingId)
		response.JSON().Path("$.transaction.spendingAmount").IsEqual(originalTransaction.Amount)
		// Make sure we spent from the right spending object.
		response.JSON().Path("$.spending[0].spendingId").IsEqual(spending.SpendingId)
		// And make sure we spent the amount we wanted.
		response.JSON().Path("$.spending[0].currentAmount").IsEqual(spending.CurrentAmount - originalTransaction.Amount)
		// Make sure the next contribution gets recalculated.
		response.JSON().Path("$.spending[0].nextContributionAmount").Number().Lt(spending.NextContributionAmount)
	})

	t.Run("update preserves lunch flow transaction id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			// Create a lunch flow transaction record and associate it with the
			// transaction to simulate a transaction that was synced from Lunch Flow.
			lunchFlowTransaction := testutils.MustInsert(t, LunchFlowTransaction{
				AccountId:              user.AccountId,
				LunchFlowBankAccountId: *bank.LunchFlowBankAccountId,
				LunchFlowId:            "lf_test_1234",
				Merchant:               originalTransaction.MerchantName,
				Description:            originalTransaction.Name,
				Date:                   originalTransaction.Date,
				Currency:               "USD",
				Amount:                 originalTransaction.Amount,
				IsPending:              false,
			})
			originalTransaction.LunchFlowTransactionId = &lunchFlowTransaction.LunchFlowTransactionId
			testutils.MustDBUpdate(t, &originalTransaction)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newName := "A More Friendly Name"
		assert.NotEqual(t, originalTransaction.Name, newName, "make sure the names dont somehow match")

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": newName,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.transactionId").IsEqual(originalTransaction.TransactionId)
		response.JSON().Path("$.transaction.name").String().IsEqual(newName)
		response.JSON().Path("$.transaction.name").String().NotEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.originalName").String().IsEqual(originalTransaction.Name)

		// Verify the lunch flow transaction ID is preserved by reading directly
		// from the database, since it is not included in the JSON response.
		updatedTransaction := testutils.MustRetrieve(t, Transaction{
			TransactionId: originalTransaction.TransactionId,
			AccountId:     bank.AccountId,
			BankAccountId: bank.BankAccountId,
		})
		assert.NotNil(t, updatedTransaction.LunchFlowTransactionId, "lunch flow transaction ID must be preserved after update")
		assert.Equal(t, originalTransaction.LunchFlowTransactionId, updatedTransaction.LunchFlowTransactionId, "lunch flow transaction ID must not change")
	})

	t.Run("update does not overwrite spending amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			// Set an initial spending amount on the transaction directly to simulate
			// a transaction that was already spent from an expense.
			originalTransaction.SpendingAmount = myownsanity.Pointer(int64(500))
			testutils.MustDBUpdate(t, &originalTransaction)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		// The PUT silently ignored an attempt to set the spending amount. The PATCH
		// schema goes a step further and rejects the field outright since it is not
		// something a client is ever allowed to set directly.
		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingAmount": 9999,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.spendingAmount").String().IsEqual("key not expected")
	})

	t.Run("cant patch someone elses transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var transaction Transaction

		{ // Create a bank account with a transaction under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			transaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to update the transaction
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "Updated Name",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing transaction for update: record does not exist")
		}
	})

	t.Run("cannot update fields that shouldn't be updated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		assert.Nil(t, originalTransaction.DeletedAt, "deleted at should be nil to start with")

		{ // Try to update the protected fields
			// Where the PUT would silently drop these, the PATCH schema rejects each
			// of them as an unexpected key so the request never even gets applied.
			now := app.Clock.Now()
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", originalTransaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"source":    "other",
					"createdAt": now,
					"deletedAt": now,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.source").String().IsEqual("key not expected")
			response.JSON().Path("$.problems.createdAt").String().IsEqual("key not expected")
			response.JSON().Path("$.problems.deletedAt").String().IsEqual("key not expected")
		}

		{ // Make sure that we can still see the transaction
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)

			// Make sure that our original transaction is still visible in the
			// response and was not soft deleted by the rejected patch.
			response.JSON().Path("$[0].transactionId").IsEqual(originalTransaction.TransactionId)
		}
	})

	t.Run("manual link can patch just the name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		// On a manual link the amount IS editable, but a patch that only wants to
		// change the name should not be forced to also send the amount. This is the
		// core partial update behavior, and it is what the amount
		// Required(Optional) fix enables.
		newName := "Renamed Manual Transaction"
		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": newName,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.name").String().IsEqual(newName)
		// The amount must be left exactly as it was since we did not send it.
		response.JSON().Path("$.transaction.amount").IsEqual(originalTransaction.Amount)
	})

	t.Run("manual link can change the amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newAmount := int64(81234)
		assert.NotEqual(t, originalTransaction.Amount, newAmount, "make sure the amounts dont somehow match")

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"amount": newAmount,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.amount").IsEqual(newAmount)
	})

	t.Run("manual link can change the date", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newDate := originalTransaction.Date.AddDate(0, 0, -3)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"date": newDate.Format(time.RFC3339),
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.date").String().AsDateTime(time.RFC3339).IsEqual(newDate)
	})

	t.Run("manual link can set is pending to true", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		assert.False(t, originalTransaction.IsPending, "transaction should not be pending to start with")

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"isPending": true,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.isPending").Boolean().IsTrue()
	})

	t.Run("manual link can set is pending back to false", func(t *testing.T) {
		// This is the regression test for the is pending schema fix. false is the
		// zero value for a bool, so a validation.Required rule would have rejected
		// this perfectly valid attempt to mark a transaction as no longer pending.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			// Flip it to pending first so that setting it back to false is a real
			// change we can assert on.
			originalTransaction.IsPending = true
			testutils.MustDBUpdate(t, &originalTransaction)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"isPending": false,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.isPending").Boolean().IsFalse()
	})

	t.Run("manual link rejects an amount of zero", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"amount": 0,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// The Amount rule runs before NotEq and rejects zero first, so this is the
		// message that actually surfaces.
		response.JSON().Path("$.problems.amount").String().IsEqual("Amount cannot be zero")
	})

	t.Run("manual link rejects an invalid date", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"date": "not a real date",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.date").String().IsEqual("Date must be in a valid format")
	})

	t.Run("non manual link cannot change the amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"amount": 5000,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.amount").String().IsEqual("key not expected")
	})

	t.Run("non manual link cannot change the date", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"date": originalTransaction.Date.Format(time.RFC3339),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.date").String().IsEqual("key not expected")
	})

	t.Run("non manual link cannot change is pending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"isPending": true,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.isPending").String().IsEqual("key not expected")
	})

	t.Run("name that is too long is rejected", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": gofakeit.Sentence(250),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.name").String().IsEqual("Name must be between 1 and 300 characters")
	})

	t.Run("name cannot be blanked out", func(t *testing.T) {
		// cleanStrings trims the body before validation, so a whitespace only name
		// collapses to an empty string and Name's Required rule rejects it.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "   ",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
	})

	t.Run("name cannot be null", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": nil,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
	})

	t.Run("merchant name that is too long is rejected", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"merchantName": gofakeit.Sentence(250),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.merchantName").String().IsEqual("Must be between 1 and 300 characters")
	})

	t.Run("invalid spending Id is rejected", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": "not_a_real_id",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// spendingId is a one of null or a valid id, so the failure serializes
		// under a oneOf key with the reason each branch was rejected.
		response.JSON().Path("$.problems.spendingId.oneOf").Array().ContainsAll("id does not match format spnd_...")
	})

	t.Run("patch with only the name leaves the other fields untouched", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test. Use a manual link so that amount, date and
			// is pending are all things the schema WOULD allow us to change, which
			// makes leaving them untouched a meaningful assertion.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		newName := "Only The Name Changed"
		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": newName,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.name").String().IsEqual(newName)
		// Everything we did not send should be exactly as it was.
		response.JSON().Path("$.transaction.merchantName").String().IsEqual(originalTransaction.MerchantName)
		response.JSON().Path("$.transaction.amount").IsEqual(originalTransaction.Amount)
		response.JSON().Path("$.transaction.date").String().AsDateTime(time.RFC3339).IsEqual(originalTransaction.Date)
		response.JSON().Path("$.transaction.isPending").Boolean().IsEqual(originalTransaction.IsPending)
	})

	t.Run("empty patch is a no op", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.name").String().IsEqual(originalTransaction.Name)
		response.JSON().Path("$.transaction.amount").IsEqual(originalTransaction.Amount)
	})

	t.Run("remove an expense by setting spending id to null", func(t *testing.T) {
		app, e := NewTestApplication(t)
		now := app.Clock.Now()
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		transaction := fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			Name:                   "Payday",
			Description:            "Whenever I get paid",
			RuleSet:                fundingRule,
			ExcludeWeekends:        true,
			WaitForDeposit:         false,
			EstimatedDeposit:       nil,
			LastRecurrence:         nil,
			NextRecurrence:         fundingRule.After(now, false),
			NextRecurrenceOriginal: fundingRule.After(now, false),
		})
		// Give the expense more than enough to cover the transaction so that we can
		// clearly assert removing the transaction puts the funds right back.
		originalCurrentAmount := transaction.Amount * 2
		spending := testutils.MustInsert(t, Spending{
			Name:                   "Groceries",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           transaction.Amount * 2,
			CurrentAmount:          originalCurrentAmount,
			NextContributionAmount: 0,
			NextRecurrence:         spendingRule.After(now, false),
			RuleSet:                spendingRule,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})

		token := GivenILogin(t, e, user.Login.Email, password)

		{ // First assign the transaction to the expense.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": spending.SpendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.spendingId").IsEqual(spending.SpendingId)
			// The expense should have been charged for the transaction.
			response.JSON().Path("$.spending[0].currentAmount").IsEqual(originalCurrentAmount - transaction.Amount)
		}

		{ // Now remove the expense by setting the spending id back to null.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			// The transaction should no longer be associated with the expense.
			response.JSON().Path("$.transaction.spendingId").IsNull()
			// And the money should have been returned to the expense in full.
			response.JSON().Path("$.spending[0].spendingId").IsEqual(spending.SpendingId)
			response.JSON().Path("$.spending[0].currentAmount").IsEqual(originalCurrentAmount)
		}
	})

	t.Run("move a transaction from one expense to another", func(t *testing.T) {
		app, e := NewTestApplication(t)
		now := app.Clock.Now()
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		transaction := fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			Name:                   "Payday",
			Description:            "Whenever I get paid",
			RuleSet:                fundingRule,
			ExcludeWeekends:        true,
			WaitForDeposit:         false,
			EstimatedDeposit:       nil,
			LastRecurrence:         nil,
			NextRecurrence:         fundingRule.After(now, false),
			NextRecurrenceOriginal: fundingRule.After(now, false),
		})
		expenseAStart := transaction.Amount * 2
		expenseBStart := transaction.Amount * 3
		expenseA := testutils.MustInsert(t, Spending{
			Name:                   "Groceries",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           transaction.Amount * 4,
			CurrentAmount:          expenseAStart,
			NextContributionAmount: 0,
			NextRecurrence:         spendingRule.After(now, false),
			RuleSet:                spendingRule,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})
		expenseB := testutils.MustInsert(t, Spending{
			Name:                   "Dining Out",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           transaction.Amount * 4,
			CurrentAmount:          expenseBStart,
			NextContributionAmount: 0,
			NextRecurrence:         spendingRule.After(now, false),
			RuleSet:                spendingRule,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})

		token := GivenILogin(t, e, user.Login.Email, password)

		{ // First put the transaction on expense A.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": expenseA.SpendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.spendingId").IsEqual(expenseA.SpendingId)
		}

		{ // Now move it over to expense B.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction.TransactionId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": expenseB.SpendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.spendingId").IsEqual(expenseB.SpendingId)
			// Both expenses are affected by the move so both should come back.
			response.JSON().Path("$.spending").Array().Length().IsEqual(2)
		}

		// Verify the final balances directly from the database so we do not have to
		// depend on the ordering of the spending array in the response. Expense A
		// should be made whole again and expense B should be charged.
		updatedA := testutils.MustRetrieve(t, Spending{
			SpendingId:    expenseA.SpendingId,
			AccountId:     bank.AccountId,
			BankAccountId: bank.BankAccountId,
		})
		assert.EqualValues(t, expenseAStart, updatedA.CurrentAmount, "expense A should be fully refunded after the move")
		updatedB := testutils.MustRetrieve(t, Spending{
			SpendingId:    expenseB.SpendingId,
			AccountId:     bank.AccountId,
			BankAccountId: bank.BankAccountId,
		})
		assert.EqualValues(t, expenseBStart-transaction.Amount, updatedB.CurrentAmount, "expense B should be charged after the move")
	})

	t.Run("cannot spend from a deposit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		now := app.Clock.Now()
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		transaction := fixtures.GivenIHaveATransaction(t, app.Clock, bank)

		// Turn the transaction into a deposit. In monetr deposits are stored as
		// negative amounts, which is what IsAddition keys off of.
		transaction.Amount = -1 * transaction.Amount
		testutils.MustDBUpdate(t, &transaction)

		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			Name:                   "Payday",
			Description:            "Whenever I get paid",
			RuleSet:                fundingRule,
			ExcludeWeekends:        true,
			WaitForDeposit:         false,
			EstimatedDeposit:       nil,
			LastRecurrence:         nil,
			NextRecurrence:         fundingRule.After(now, false),
			NextRecurrenceOriginal: fundingRule.After(now, false),
		})
		spending := testutils.MustInsert(t, Spending{
			Name:                   "Groceries",
			SpendingType:           SpendingTypeExpense,
			TargetAmount:           100000,
			CurrentAmount:          100000,
			NextContributionAmount: 0,
			NextRecurrence:         spendingRule.After(now, false),
			RuleSet:                spendingRule,
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			CreatedAt:              now,
		})

		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", transaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("cannot specify a spent from on a deposit")
	})

	t.Run("changing the amount on a manual transaction does not adjust the balance", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var originalTransaction Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var availableBefore, currentBefore float64
		{ // Capture the starting balances.
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBefore = response.JSON().Path("$.available").Number().Raw()
			currentBefore = response.JSON().Path("$.current").Number().Raw()
		}

		newAmount := originalTransaction.Amount + 5000
		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", originalTransaction.TransactionId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"amount": newAmount,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transaction.amount").IsEqual(newAmount)
		// The balance should be untouched. When monetr eventually recalculates the
		// balance on an amount change this assertion is expected to change.
		response.JSON().Path("$.balance.available").Number().IsEqual(availableBefore)
		response.JSON().Path("$.balance.current").Number().IsEqual(currentBefore)
	})
}

func TestDeleteTransactions(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		var transactionId ID[Transaction]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         200, // Spent $2
					"isPending":      false,
					"name":           "I spent money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 200)
			transactionId = ID[Transaction](response.JSON().Path("$.transaction.transactionId").String().Raw())
		}

		{ // Check that we can see the transaction in the API response.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].transactionId").String().IsEqual(transactionId.String())
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the transaction we just created
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 200)
		}

		{ // Check that the transaction is now missing.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Make sure the balance has not been changed
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.current").Number().IsEqual(startingCurrentBalance - 200)
		}
	})

	t.Run("happy path with balance adjustment", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		var transactionId ID[Transaction]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         200, // Spent $2
					"isPending":      false,
					"name":           "I spent money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 200)
			transactionId = ID[Transaction](response.JSON().Path("$.transaction.transactionId").String().Raw())
		}

		{ // Check that we can see the transaction in the API response.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].transactionId").String().IsEqual(transactionId.String())
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the transaction we just created
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithQuery("adjusts_balance", "true").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
		}

		{ // Check that the transaction is now missing.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Make sure the balance has not been changed
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			// Balances should not have the $2 deduction anymore
			response.JSON().Path("$.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.current").Number().IsEqual(startingCurrentBalance)
		}
	})

	t.Run("happy path with spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		{ // Seed the data for the test.
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":        "Payday",
					"description": "15th and the Last day of every month",
					"ruleset":     FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
		}

		var spendingId ID[Spending]
		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Transfer funds to the spending object
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000, // $10.00
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
		}

		var transactionId ID[Transaction]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         200, // Spent $2
					"isPending":      false,
					"name":           "I spent money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
					"spendingId":     spendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 200)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
			response.JSON().Path("$.spending.spendingId").IsEqual(spendingId)
			response.JSON().Path("$.spending.currentAmount").IsEqual(800)
			transactionId = ID[Transaction](response.JSON().Path("$.transaction.transactionId").String().Raw())
		}

		{ // Check that we can see the transaction in the API response.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].transactionId").String().IsEqual(transactionId.String())
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the transaction we just created
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithQuery("adjusts_balance", "true").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			// When we delete a transaction that was spent from an expense, and we are
			// adjusting balances. We should remove the transactions impact on the
			// available and current balances (so they should return to the balances
			// seen before the txn was created). We should still see a deduction of
			// $10 though from free because the funds from the transaction would have
			// been returned to the spending object's allocation.
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
			response.JSON().Path("$.spending[0].spendingId").IsEqual(spendingId)
			response.JSON().Path("$.spending[0].currentAmount").IsEqual(1000)
		}

		{ // Check that the transaction is now missing.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Make sure the balance has not been changed
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			// Balances should not have the $2 deduction anymore
			response.JSON().Path("$.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.current").Number().IsEqual(startingCurrentBalance)
		}

		{ // Because this was soft deleted, it can still be read by ID.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transactionId").IsEqual(transactionId)
		}
	})

	t.Run("hard delete with spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)

		{ // Seed the data for the test.
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		var startingAvailableBalance, startingCurrentBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":        "Payday",
					"description": "15th and the Last day of every month",
					"ruleset":     FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
		}

		var spendingId ID[Spending]
		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Transfer funds to the spending object
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000, // $10.00
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
		}

		var transactionId ID[Transaction]
		{
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         200, // Spent $2
					"isPending":      false,
					"name":           "I spent money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
					"spendingId":     spendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance - 200)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance - 200)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
			response.JSON().Path("$.spending.spendingId").IsEqual(spendingId)
			response.JSON().Path("$.spending.currentAmount").IsEqual(800)
			transactionId = ID[Transaction](response.JSON().Path("$.transaction.transactionId").String().Raw())
		}

		{ // Check that we can see the transaction in the API response.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].transactionId").String().IsEqual(transactionId.String())
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the transaction we just created
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithQuery("adjusts_balance", "true").
				WithQuery("soft", "false").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			// When we delete a transaction that was spent from an expense, and we are
			// adjusting balances. We should remove the transactions impact on the
			// available and current balances (so they should return to the balances
			// seen before the txn was created). We should still see a deduction of
			// $10 though from free because the funds from the transaction would have
			// been returned to the spending object's allocation.
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingAvailableBalance - 1000)
			response.JSON().Path("$.spending[0].spendingId").IsEqual(spendingId)
			response.JSON().Path("$.spending[0].currentAmount").IsEqual(1000)
		}

		{ // Check that the transaction is now missing.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Make sure the balance has not been changed
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			// Balances should not have the $2 deduction anymore
			response.JSON().Path("$.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.current").Number().IsEqual(startingCurrentBalance)
		}

		{ // Because this was soft deleted, it can still be read by ID.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transactionId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
		}
	})

	t.Run("no authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.DELETE(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithJSON(map[string]any{
				"adjustsBalance": false,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("bad authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.DELETE(`/api/bank_accounts/bac_bogus/transactions/txn_bogus`).
			WithCookie(TestCookieName, gofakeit.Generate("????????")).
			WithJSON(map[string]any{
				"adjustsBalance": false,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("transaction does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionId", "txn_bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Failed to find transaction to be removed: record does not exist")
	})

	t.Run("bank account does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var token string

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", "bac_bogus").
			WithPath("transactionId", "txn_bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Cannot delete transactions for non-manual links")
	})

	t.Run("plaid link doesnt allow transaction deletion", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var token string
		var bank BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		txn := fixtures.GivenIHaveATransaction(t, app.Clock, bank)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
			WithPath("bankAccountId", txn.BankAccountId).
			WithPath("transactionId", txn.TransactionId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Cannot delete transactions for non-manual links")
	})
}
