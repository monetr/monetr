package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestGetForecast(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The forecast endpoint lives on the billedKeyOrToken route group so it
		// accepts an API key. It needs a bank account belonging to the authenticated
		// account. We seed the account and bank account via fixtures, log in as that
		// same account to obtain a token, and then mint the API key with that token
		// so the key belongs to the SAME account that owns the bank account.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET(`/api/bank_accounts/{bankAccountId}/forecast`).
			WithPath("bankAccountId", bank.BankAccountId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()
		response.Status(http.StatusOK)
		// A brand new bank account has no spending, so the forecast has no events but
		// the response is still a well formed forecast object.
		response.JSON().Object().ContainsKey("events")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs, so the bank account path value is irrelevant here.
		_, e := NewTestApplication(t)

		response := e.GET(`/api/bank_accounts/{bankAccountId}/forecast`).
			WithPath("bankAccountId", "bac_"+gofakeit.UUID()).
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}

func TestPostForecastNewSpending(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The new spending forecast endpoint lives on the billedKeyOrToken route
		// group so it accepts an API key. It needs a bank account and a real funding
		// schedule (the forecaster panics if a spending object references a funding
		// schedule that does not exist). We seed everything for one account and mint
		// the API key with a token for that same account.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		// Forecast a simple goal funded by the real funding schedule. A goal must not
		// carry a recurrence rule, only a future due date.
		nextRecurrence := app.Clock.Now().AddDate(0, 6, 0)
		response := e.POST(`/api/bank_accounts/{bankAccountId}/forecast/spending`).
			WithPath("bankAccountId", bank.BankAccountId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"fundingScheduleId": fundingSchedule.FundingScheduleId,
				"spendingType":      models.SpendingTypeGoal,
				"targetAmount":      10000,
				"currentAmount":     0,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect()
		response.Status(http.StatusOK)
		response.JSON().Object().ContainsKey("estimatedCost")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.POST(`/api/bank_accounts/{bankAccountId}/forecast/spending`).
			WithPath("bankAccountId", "bac_"+gofakeit.UUID()).
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			WithJSON(map[string]any{
				"fundingScheduleId": "fund_" + gofakeit.UUID(),
				"spendingType":      models.SpendingTypeGoal,
				"targetAmount":      10000,
				"currentAmount":     0,
			}).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}

func TestPostForecastNextFunding(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The next funding forecast endpoint lives on the billedKeyOrToken route
		// group so it accepts an API key. It looks the funding schedule up in the
		// database, so it must actually exist. We seed a bank account and funding
		// schedule for one account and mint the API key with a token for that same
		// account.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.POST(`/api/bank_accounts/{bankAccountId}/forecast/next_funding`).
			WithPath("bankAccountId", bank.BankAccountId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"fundingScheduleId": fundingSchedule.FundingScheduleId,
			}).
			Expect()
		response.Status(http.StatusOK)
		response.JSON().Object().ContainsKey("nextContribution")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.POST(`/api/bank_accounts/{bankAccountId}/forecast/next_funding`).
			WithPath("bankAccountId", "bac_"+gofakeit.UUID()).
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			WithJSON(map[string]any{
				"fundingScheduleId": "fund_" + gofakeit.UUID(),
			}).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}

func TestQueryInitialFundingForRecurringTransaction(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		app.Clock.Set(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		// $8 on the 15th of every month
		recurring := GivenIHaveATransactionRecurring(t, bank)

		var fundingScheduleId models.ID[models.FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": EveryFriday,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = models.ID[models.FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		// The last charge was on the 15th and there are 5 fridays before the next
		// one on November 15th, so each friday puts aside $1.60. Two of them have
		// happened since the charge, the 16th and the 23rd.
		response := e.Request("QUERY", "/api/bank_accounts/{bankAccountId}/forecast/initial").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"transactionRecurringId": recurring.TransactionRecurringId,
				"fundingScheduleId":      fundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.currentAmount").Number().IsEqual(320)
	})

	t.Run("funding schedule created today", func(t *testing.T) {
		// Someone who just signed up won't have a funding schedule that started in
		// the past, their first payday is still coming up. But they have still been
		// paid before, so the funding schedule should be treated like it has always
		// existed.
		app, e := NewTestApplication(t)
		app.Clock.Set(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		// $8 on the 15th of every month
		recurring := GivenIHaveATransactionRecurring(t, bank)
		// Every other friday, the fixture starts it on the first one after now
		// which is October 30th. This can't be created through the API right now,
		// creating a funding schedule with no occurrence before now panics.
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", false)

		// Carried back, the paydays are the 16th, the 30th and November 13th. So
		// each one puts aside a third of the $8 and only the 16th has happened
		// since the charge on the 15th.
		response := e.Request("QUERY", "/api/bank_accounts/{bankAccountId}/forecast/initial").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"transactionRecurringId": recurring.TransactionRecurringId,
				"fundingScheduleId":      fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.currentAmount").Number().IsEqual(266)
	})

	t.Run("recurring paycheck", func(t *testing.T) {
		app, e := NewTestApplication(t)
		app.Clock.Set(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", false)
		recurring := GivenIHaveATransactionRecurring(t, bank)
		recurring.Direction = models.CreditDirection
		testutils.MustDBUpdate(t, &recurring)

		response := e.Request("QUERY", "/api/bank_accounts/{bankAccountId}/forecast/initial").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"transactionRecurringId": recurring.TransactionRecurringId,
				"fundingScheduleId":      fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Recurring transaction must be a debit")
	})

	t.Run("missing recurring transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		app.Clock.Set(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", false)

		response := e.Request("QUERY", "/api/bank_accounts/{bankAccountId}/forecast/initial").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"fundingScheduleId": fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.transactionRecurringId").String().IsEqual("required key is missing")
	})

	t.Run("cant use someone elses recurring transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		app.Clock.Set(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", false)

		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		otherLink := fixtures.GivenIHaveAManualLink(t, app.Clock, otherUser)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &otherLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		otherRecurring := GivenIHaveATransactionRecurring(t, otherBank)

		response := e.Request("QUERY", "/api/bank_accounts/{bankAccountId}/forecast/initial").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"transactionRecurringId": otherRecurring.TransactionRecurringId,
				"fundingScheduleId":      fundingSchedule.FundingScheduleId,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Failed to retrieve recurring transaction: record does not exist")
	})
}
