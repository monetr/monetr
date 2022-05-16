package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
)

func TestPostFundingSchedules(t *testing.T) {
	t.Run("create a basic funding schedule", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":        "Payday",
				"description": "15th and the Last day of every month",
				"rule":        "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().Equal(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().DateTime(time.RFC3339).Gt(time.Now())
		response.JSON().Path("$.excludeWeekends").Boolean().False()
	})

	t.Run("create a funding schedule with excluded weekends", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":            "Payday",
				"description":     "15th and the Last day of every month",
				"rule":            "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
				"excludeWeekends": true,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().Equal(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().DateTime(time.RFC3339).Gt(time.Now())
		response.JSON().Path("$.excludeWeekends").Boolean().True()
	})

	t.Run("create a funding schedule that respects the provided next occurrence", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		nextFriday := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;BYDAY=FR").After(time.Now(), false)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		nextFriday = util.MidnightInLocal(nextFriday, timezone)
		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":           "Payday",
				"description":    "Every other friday",
				"rule":           "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR",
				"nextOccurrence": nextFriday,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().Equal(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().DateTime(time.RFC3339).Gt(time.Now())
		response.JSON().Path("$.nextOccurrence").String().DateTime(time.RFC3339).Equal(nextFriday)
		response.JSON().Path("$.excludeWeekends").Boolean().False()
	})

	t.Run("cannot create a duplicate name", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create the initial funding schedule.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":        "Payday",
					"description": "15th and the Last day of every month",
					"rule":        "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			response.JSON().Path("$.bankAccountId").Number().Equal(bank.BankAccountId)
			response.JSON().Path("$.nextOccurrence").String().DateTime(time.RFC3339).Gt(time.Now())
			response.JSON().Path("$.excludeWeekends").Boolean().False()
		}

		{ // Then try to create another one with the same name.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":        "Payday",
					"description": "Every other friday",
					"rule":        "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").Equal("failed to create funding schedule: a similar object already exists")
		}
	})

	t.Run("requires a name", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"rule": "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("funding schedule must have a name")
	})

	t.Run("requires a valid bank account Id", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/0/funding_schedules").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":        "Payday",
				"description": "15th and the Last day of every month",
				"rule":        "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("must specify valid bank account Id")
	})

	t.Run("invalid json", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("malformed JSON: invalid character 'o' in literal null (expecting 'u')")
	})
}
