package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank models.BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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
		var bank models.BankAccount

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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
}

func TestPostTransactions(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/bank_accounts/1234/transactions").
			WithCookie(TestCookieName, token).
			WithJSON(models.Transaction{
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
}

func TestPutTransactions(t *testing.T) {
	t.Run("update transaction name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank models.BankAccount
		var originalTransaction, transaction models.Transaction

		{ // Seed the data for the test.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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
			WithJSON(models.Transaction{
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
			WithJSON(models.Transaction{
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
			WithJSON(models.Transaction{
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
			WithJSON(models.Transaction{
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
			WithJSON(models.Transaction{
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
		var bank models.BankAccount
		var originalTransaction, transaction models.Transaction
		now := app.Clock.Now()

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to read the account's timezone")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")

		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		originalTransaction = fixtures.GivenIHaveATransaction(t, app.Clock, bank)
		transaction = originalTransaction
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
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
		spending := testutils.MustInsert(t, models.Spending{
			Name:                   "Spending test",
			SpendingType:           models.SpendingTypeExpense,
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
}
