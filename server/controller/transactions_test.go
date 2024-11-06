package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
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
					"bankAccountId":  bank.BankAccountId,
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
					"bankAccountId":  bank.BankAccountId,
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
					"bankAccountId":  bank.BankAccountId,
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
					"bankAccountId":  bank.BankAccountId,
					"amount":         100, // $1
					"isPending":      false,
					"name":           "I spent some money",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": false,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
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
				WithJSON(map[string]interface{}{
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
			ruleset := testutils.Must(t, NewRuleSet, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
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
					"bankAccountId":  bank.BankAccountId,
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
}
