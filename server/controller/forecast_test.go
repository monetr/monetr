package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/models"
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
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
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
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
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
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			WithJSON(map[string]any{
				"fundingScheduleId": "fund_" + gofakeit.UUID(),
			}).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}
