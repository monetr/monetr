package controller_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/stretchr/testify/assert"
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

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		rule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;BYDAY=FR")
		rule.DTStart(util.MidnightInLocal(time.Now().In(timezone).Add(-30*24*time.Hour), timezone)) // Force the Rule to be in the correct TZ.
		nextFriday := rule.After(time.Now(), false)
		assert.Greater(t, nextFriday, time.Now(), "next friday should be in the future relative to now")
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

func TestPutFundingSchedules(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingSchedule.Name = "This is an updated name"

		response := e.PUT("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", fundingSchedule.BankAccountId).
			WithPath("fundingScheduleId", fundingSchedule.FundingScheduleId).
			WithJSON(fundingSchedule).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.name").Equal(fundingSchedule.Name)
	})
}

func TestDeleteFundingSchedules(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", fundingSchedule.BankAccountId).
			WithPath("fundingScheduleId", fundingSchedule.FundingScheduleId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.Body().Empty()
	})

	t.Run("funding schedule does not exist", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("fundingScheduleId", math.MaxInt64).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().Equal("cannot remove funding schedule, it does not exist")
	})
}
