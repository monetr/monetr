package controller_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestPostFundingSchedules(t *testing.T) {
	t.Run("create a basic funding schedule", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":        "Payday",
				"description": "15th and the Last day of every month",
				"ruleset":     FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().AsDateTime(time.RFC3339).Gt(app.Clock.Now())
		response.JSON().Path("$.excludeWeekends").Boolean().IsFalse()
	})

	t.Run("create a funding schedule with excluded weekends", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().AsDateTime(time.RFC3339).Gt(app.Clock.Now())
		response.JSON().Path("$.excludeWeekends").Boolean().IsTrue()
	})

	t.Run("create a funding schedule that respects the provided next occurrence", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		rule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;BYDAY=FR")
		rule.DTStart(util.Midnight(app.Clock.Now().In(timezone).Add(-30*24*time.Hour), timezone)) // Force the Rule to be in the correct TZ.
		nextFriday := rule.After(app.Clock.Now(), false)
		assert.Greater(t, nextFriday, app.Clock.Now(), "next friday should be in the future relative to now")
		nextFriday = util.Midnight(nextFriday, timezone)

		ruleset := testutils.NewRuleSet(t, nextFriday.Year(), int(nextFriday.Month()), nextFriday.Day(), timezone, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":           "Payday",
				"description":    "Every other friday",
				"ruleset":        ruleset,
				"nextOccurrence": nextFriday,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
		response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
		response.JSON().Path("$.nextOccurrence").String().AsDateTime(time.RFC3339).Gt(app.Clock.Now())
		response.JSON().Path("$.nextOccurrence").String().AsDateTime(time.RFC3339).IsEqual(nextFriday)
		response.JSON().Path("$.excludeWeekends").Boolean().IsFalse()
	})

	t.Run("cannot create a duplicate name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create the initial funding schedule.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":        "Payday",
					"description": "15th and the Last day of every month",
					"ruleset":     FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.nextOccurrence").String().AsDateTime(time.RFC3339).Gt(app.Clock.Now())
			response.JSON().Path("$.excludeWeekends").Boolean().IsFalse()
		}

		{ // Then try to create another one with the same name.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":        "Payday",
					"description": "First Day Of Every Month",
					"ruleset":     FirstDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("failed to create funding schedule: a similar object already exists")
		}
	})

	t.Run("requires a name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("funding schedule must have a name")
	})

	t.Run("requires a valid bank account Id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/0/funding_schedules").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"name":        "Payday",
				"description": "15th and the Last day of every month",
				"rule":        FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("must specify a valid bank account Id")
	})

	t.Run("invalid json", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("invalid JSON body")
	})
}

func TestPutFundingSchedules(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingSchedule.Name = "This is an updated name"

		response := e.PUT("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", fundingSchedule.BankAccountId).
			WithPath("fundingScheduleId", fundingSchedule.FundingScheduleId).
			WithJSON(fundingSchedule).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingSchedule.name").String().IsEqual(fundingSchedule.Name)
		response.JSON().Path("$.spending").IsArray()
		response.JSON().Path("$.spending").Array().IsEmpty()
	})

	t.Run("updates a spending object", func(t *testing.T) {
		app, e := NewTestApplication(t)
		now := app.Clock.Now()
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		var fundingScheduleId uint64
		{ // Create the funding schedule
			fundingRule := testutils.Must(t, models.NewRuleSet, FifthteenthAndLastDayOfEveryMonth)
			fundingRule.DTStart(util.Midnight(fundingRule.GetDTStart(), timezone)) // Force the Rule to be in the correct TZ.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         fundingRule,
					"excludeWeekends": true,
					"nextOccurrence":  fundingRule.After(now, false),
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		var spendingId uint64
		{ // Create an expense
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.Must(t, models.NewRuleSet, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			nextRecurrence = util.Midnight(nextRecurrence, timezone)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").Number().Gt(0)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
			spendingId = uint64(response.JSON().Path("$.spendingId").Number().Raw())
			assert.NotZero(t, spendingId, "must be able to extract the spending ID")
		}

		{ // Now update the rule on the funding schedule and the next occurrence
			newFundingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", app.Clock.Now())

			next := util.Midnight(newFundingRule.After(now, false), timezone)
			response := e.PUT("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("fundingScheduleId", fundingScheduleId).
				WithJSON(map[string]interface{}{
					"name":            "Payday",
					"description":     "Every friday",
					"ruleset":         newFundingRule,
					"excludeWeekends": false,
					"nextOccurrence":  next,
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.fundingSchedule.name").String().IsEqual("Payday")
			response.JSON().Path("$.fundingSchedule.nextOccurrence").String().AsDateTime(time.RFC3339).IsEqual(next)
			response.JSON().Path("$.spending").IsArray()
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].spendingId").Number().IsEqual(spendingId)
		}
	})
}

func TestDeleteFundingSchedules(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", fundingSchedule.BankAccountId).
			WithPath("fundingScheduleId", fundingSchedule.FundingScheduleId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.Body().IsEmpty()
	})

	t.Run("funding schedule is in use", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", app.Clock.Now())
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
		}

		{ // Then try to delete the funding schedule
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("fundingScheduleId", fundingScheduleId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Cannot delete a funding schedule with goals or expenses associated with it")
		}
	})

	t.Run("funding schedule does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.DELETE("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("fundingScheduleId", math.MaxInt64).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("cannot remove funding schedule, it does not exist")
	})
}

func TestGetFundingSchedulesByID(t *testing.T) {
	t.Run("should be able to retrieve an owned schedule by ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule.
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":        "Payday",
					"description": "15th and the Last day of every month",
					"ruleset":     FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)

			// Save the ID of the created funding schedule so we can use it below.
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("fundingScheduleId", fundingScheduleId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.fundingScheduleId").Number().IsEqual(fundingScheduleId)
		response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
	})

	t.Run("cannot read someone else's funding schedule", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bankAccountId, fundingScheduleId uint64
		{ // Create the funding schedule under the first account.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank := fixtures.GivenIHaveABankAccount(
				t,
				app.Clock,
				&link,
				models.DepositoryBankAccountType,
				models.CheckingBankAccountSubType,
			)
			token := GivenILogin(t, e, user.Login.Email, password)

			{ // Create the funding schedule.
				response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
					WithPath("bankAccountId", bank.BankAccountId).
					WithCookie(TestCookieName, token).
					WithJSON(map[string]interface{}{
						"name":        "Payday",
						"description": "15th and the Last day of every month",
						"ruleset":     FifthteenthAndLastDayOfEveryMonth,
					}).
					Expect()

				response.Status(http.StatusOK)
				response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
				response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)

				// Save the ID of the created funding schedule so we can use it below.
				fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
				bankAccountId = bank.BankAccountId
			}

			{ // Try to read it as the owning user, just to make sure it does work.
				response := e.GET("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
					WithPath("bankAccountId", bankAccountId).
					WithPath("fundingScheduleId", fundingScheduleId).
					WithCookie(TestCookieName, token).
					Expect()

				response.Status(http.StatusOK)
				response.JSON().Path("$.fundingScheduleId").Number().IsEqual(fundingScheduleId)
				response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			}
		}

		{ // Then try to read the funding schedule under another account.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token := GivenILogin(t, e, user.Login.Email, password)

			response := e.GET("/api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}").
				WithPath("bankAccountId", bankAccountId).
				WithPath("fundingScheduleId", fundingScheduleId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve funding schedule: record does not exist")
		}
	})
}
