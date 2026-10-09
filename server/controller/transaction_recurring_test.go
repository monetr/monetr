package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestGetRecurringTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(1)
		response.JSON().Path("$[0].transactionRecurringId").IsEqual(recurring.TransactionRecurringId)
		response.JSON().Path("$[0]").Object().NotContainsKey("transactionCluster")
	})

	t.Run("pagination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		for range 30 {
			GivenIHaveATransactionRecurring(t, bank)
		}
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // First page, uses the default limit of 25
			response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(25)
		}

		{ // Second page
			response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("offset", 25).
				WithQuery("limit", 25).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(5)
		}

		{ // Smaller limit
			response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
				WithPath("bankAccountId", bank.BankAccountId).
				WithQuery("limit", 10).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(10)
		}
	})

	t.Run("ended ones come last", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		ended := GivenIHaveATransactionRecurring(t, bank)
		ended.Ended = true
		ended.Next = time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
		testutils.MustDBUpdate(t, &ended)
		active := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$[0].transactionRecurringId").IsEqual(active.TransactionRecurringId)
		response.JSON().Path("$[1].transactionRecurringId").IsEqual(ended.TransactionRecurringId)
	})

	t.Run("invalid limit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithQuery("limit", 101).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Limit cannot be greater than 100")
	})

	t.Run("filter by direction and ended", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		charge := GivenIHaveATransactionRecurring(t, bank)
		credit := GivenIHaveATransactionRecurring(t, bank)
		credit.Direction = CreditDirection
		testutils.MustDBUpdate(t, &credit)
		ended := GivenIHaveATransactionRecurring(t, bank)
		ended.Ended = true
		testutils.MustDBUpdate(t, &ended)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithQuery("direction", "debit").
			WithQuery("ended", "false").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(1)
		response.JSON().Path("$[0].transactionRecurringId").IsEqual(charge.TransactionRecurringId)
	})

	t.Run("invalid direction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithQuery("direction", "sideways").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Direction must be debit or credit")
	})

	t.Run("invalid ended", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithQuery("ended", "maybe").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Ended must be true or false")
	})

	t.Run("cant get someone elses recurring", func(t *testing.T) {
		app, e := NewTestApplication(t)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		otherLink := fixtures.GivenIHaveAManualLink(t, app.Clock, otherUser)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &otherLink, DepositoryBankAccountType, CheckingBankAccountSubType)
		GivenIHaveATransactionRecurring(t, otherBank)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", otherBank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().IsEmpty()
	})
}

func TestGetRecurringTransaction(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionRecurringId").IsEqual(recurring.TransactionRecurringId)
		response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
		response.JSON().Path("$.transactionClusterId").IsEqual(recurring.TransactionClusterId)
		response.JSON().Path("$.direction").IsEqual(DebitDirection)
		response.JSON().Path("$.window").IsEqual(MonthlyWindowType)
		response.JSON().Path("$.lastAmount").IsEqual(800)
	})

	t.Run("with an api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionRecurringId").IsEqual(recurring.TransactionRecurringId)
	})

	t.Run("does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", NewID[TransactionRecurring]()).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
	})

	t.Run("cant read someone elses recurring", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bank BankAccount
		var recurring TransactionRecurring
		{ // Seed the recurring transaction under the first account
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			recurring = GivenIHaveATransactionRecurring(t, bank)
		}

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
	})

	t.Run("invalid bank account id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", "not_a_bank_account").
			WithPath("transactionRecurringId", NewID[TransactionRecurring]()).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Must specify a valid bank account Id")
	})

	t.Run("invalid recurring transaction id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", "not_a_recurring_id").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Must specify a valid recurring transaction Id")
	})

	t.Run("invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", "bac_fake").
			WithPath("transactionRecurringId", "txrc_fake").
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestPatchRecurringTransaction(t *testing.T) {
	t.Run("links an expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)
		app.Clock.Add(time.Hour)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionRecurringId").IsEqual(recurring.TransactionRecurringId)
		response.JSON().Path("$.spendingId").IsEqual(spending.SpendingId)
		response.JSON().Object().NotContainsKey("spending")
		response.JSON().Path("$.fundingScheduleId").IsNull()
		response.JSON().Path("$.lastAmount").IsEqual(800)

		stored := testutils.MustDBRead(t, recurring)
		if assert.NotNil(t, stored.SpendingId, "spending id should be stored") {
			assert.Equal(t, spending.SpendingId, *stored.SpendingId, "should store the linked expense")
		}
		assert.WithinDuration(t, app.Clock.Now(), stored.UpdatedAt, time.Second, "updated at should be bumped")

		{ // The spending object itself shouldn't say anything about recurring
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spending.SpendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Object().NotContainsKey("transactionRecurringId")
		}
	})

	t.Run("links a funding schedule", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = CreditDirection
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"fundingScheduleId": fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").IsEqual(fundingSchedule.FundingScheduleId)
		response.JSON().Path("$.fundingSchedule.fundingScheduleId").IsEqual(fundingSchedule.FundingScheduleId)
		response.JSON().Path("$.spendingId").IsNull()

		stored := testutils.MustDBRead(t, recurring)
		if assert.NotNil(t, stored.FundingScheduleId, "funding schedule id should be stored") {
			assert.Equal(t, fundingSchedule.FundingScheduleId, *stored.FundingScheduleId, "should store the linked funding schedule")
		}
	})

	t.Run("with an api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsEqual(spending.SpendingId)
	})

	t.Run("clears the link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": nil,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsNull()
		response.JSON().Object().NotContainsKey("spending")

		stored := testutils.MustDBRead(t, recurring)
		assert.Nil(t, stored.SpendingId, "spending id should be cleared")
	})

	t.Run("change the linked expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		first := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		second := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github Copilot")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &first.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": second.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsEqual(second.SpendingId)
	})

	t.Run("clears auto matched", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		first := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		second := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github Copilot")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &first.SpendingId
		recurring.AutoMatched = true
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Make sure it starts out auto matched
			response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionRecurringId", recurring.TransactionRecurringId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoMatched").Boolean().IsTrue()
		}

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": second.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.autoMatched").Boolean().IsFalse()

		stored := testutils.MustDBRead(t, recurring)
		assert.False(t, stored.AutoMatched, "link the user picked should not be auto matched")
	})

	t.Run("same expense keeps auto matched", func(t *testing.T) {
		// The spending ID is a pointer, sending the same expense back gives us a
		// different pointer to the same value. That should not count as the user
		// changing the expense.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		recurring.AutoMatched = true
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
				"autoAssign": true,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsEqual(spending.SpendingId)
		response.JSON().Path("$.autoMatched").Boolean().IsTrue()
		response.JSON().Path("$.autoAssign").Boolean().IsTrue()

		stored := testutils.MustDBRead(t, recurring)
		assert.True(t, stored.AutoMatched, "should still be auto matched")
	})

	t.Run("turns on auto assign", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"autoAssign": true,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsEqual(spending.SpendingId)
		response.JSON().Path("$.autoAssign").Boolean().IsTrue()

		stored := testutils.MustDBRead(t, recurring)
		assert.True(t, stored.AutoAssign, "auto assign should be stored")
	})

	t.Run("auto assign without an expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"autoAssign": true,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Must have a spending to auto assign to")

		stored := testutils.MustDBRead(t, recurring)
		assert.False(t, stored.AutoAssign, "auto assign should not be stored")
	})

	t.Run("auto assign on a credit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = CreditDirection
		recurring.FundingScheduleId = &fundingSchedule.FundingScheduleId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"autoAssign": true,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Cannot auto assign recurring funding at this time")

		stored := testutils.MustDBRead(t, recurring)
		assert.False(t, stored.AutoAssign, "auto assign should not be stored")
	})

	t.Run("expense already linked", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		linked := GivenIHaveATransactionRecurring(t, bank)
		linked.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &linked)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("failed to update recurring transaction: a similar object already exists")

		stored := testutils.MustDBRead(t, recurring)
		assert.Nil(t, stored.SpendingId, "second recurring should not be linked")
	})

	t.Run("expense on a credit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = CreditDirection
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Spending can only be linked to a debit recurring transaction")
	})

	t.Run("funding schedule on a debit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"fundingScheduleId": fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Funding schedules can only be linked to a credit recurring transaction")
	})

	t.Run("cant link a goal", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		goal := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeGoal, "Vacation")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": goal.SpendingId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Only expenses can be linked to a recurring transaction")
	})

	t.Run("expense does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": NewID[Spending](),
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Could not find spending specified: record does not exist")
	})

	t.Run("expense from another bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, SavingsBankAccountSubType)
		otherFundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &otherBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, otherFundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Could not find spending specified: record does not exist")
	})

	t.Run("funding schedule from another bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, SavingsBankAccountSubType)
		otherFundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &otherBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = CreditDirection
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"fundingScheduleId": otherFundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Could not find funding schedule specified: record does not exist")
	})

	t.Run("rejects other fields", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"lastAmount": 1,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")

		stored := testutils.MustDBRead(t, recurring)
		assert.EqualValues(t, 800, stored.LastAmount, "last amount should not change")
	})

	t.Run("deleting expense clears link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Delete the expense
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spending.SpendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.spendingId").IsNull()
		response.JSON().Object().NotContainsKey("spending")
	})

	t.Run("cant patch someone elses recurring", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bank BankAccount
		var recurring TransactionRecurring
		{ // Seed the recurring transaction under the first account
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			recurring = GivenIHaveATransactionRecurring(t, bank)
		}

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		myBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &myBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": spending.SpendingId,
			}).
			Expect()

		response.Status(http.StatusNotFound)

		stored := testutils.MustDBRead(t, recurring)
		assert.Nil(t, stored.SpendingId, "someone elses recurring should not change")
	})

	t.Run("invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", "bac_fake").
			WithPath("transactionRecurringId", "txrc_fake").
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			WithJSON(map[string]any{
				"spendingId": nil,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestDeleteRecurringTransaction(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)
		app.Clock.Add(time.Hour)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)

		stored := testutils.MustDBRead(t, recurring)
		if assert.NotNil(t, stored.DeletedAt, "deleted at should be stored") {
			assert.WithinDuration(t, app.Clock.Now(), *stored.DeletedAt, time.Second, "deleted at should be now")
		}

		{ // It shouldn't show up in the list anymore
			response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(0)
		}
	})

	t.Run("clears the expense link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := GivenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		recurring.AutoMatched = true
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)

		stored := testutils.MustDBRead(t, recurring)
		assert.NotNil(t, stored.DeletedAt, "should be deleted")
		assert.Nil(t, stored.SpendingId, "expense link should be cleared")
		assert.False(t, stored.AutoMatched, "should not be auto matched anymore")

		{ // The expense should be free to link to something else now
			other := GivenIHaveATransactionRecurring(t, bank)
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionRecurringId", other.TransactionRecurringId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": spending.SpendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").IsEqual(spending.SpendingId)
		}
	})

	t.Run("clears the funding schedule link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = CreditDirection
		recurring.FundingScheduleId = &fundingSchedule.FundingScheduleId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)

		stored := testutils.MustDBRead(t, recurring)
		assert.NotNil(t, stored.DeletedAt, "should be deleted")
		assert.Nil(t, stored.FundingScheduleId, "funding schedule link should be cleared")
	})

	t.Run("already deleted", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.DeletedAt = new(app.Clock.Now())
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Recurring transaction is already deleted")
	})

	t.Run("cant delete someone elses recurring", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bank BankAccount
		var recurring TransactionRecurring
		{ // Seed the recurring transaction under the first account
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			recurring = GivenIHaveATransactionRecurring(t, bank)
		}

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)

		stored := testutils.MustDBRead(t, recurring)
		assert.Nil(t, stored.DeletedAt, "someone elses recurring should not be deleted")
	})
}
