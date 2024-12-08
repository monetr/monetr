package controller_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestPostSpending(t *testing.T) {
	t.Run("create an expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
		}
	})

	t.Run("name and description too long", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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

		{ // Create an expense with a name thats too long
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
					"name":              gofakeit.Sentence(250),
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Name must not be longer than 250 characters")
		}

		{ // Create an expense with a description thats too long
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
					"name":              "Name is fine",
					"description":       gofakeit.Sentence(250),
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Description must not be longer than 250 characters")
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			now := app.Clock.Now()
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", "bogus_bank_id").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": math.MaxInt32,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    now.AddDate(0, 0, 1),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", "bac_bogus").
				WithCookie(TestCookieName, token).
				WithBytes([]byte("im not json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("invalid JSON body")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("spending must have a name")
		}
	})

	t.Run("missing target amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

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
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("target amount must be greater than 0")
		}
	})

	t.Run("negative target amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

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
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      -1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("target amount must be greater than 0")
		}
	})

	t.Run("invalid funding schedule", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
					"ruleset":           ruleset,
					"fundingScheduleId": "fund_bogus",
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("could not find funding schedule specified: record does not exist")
		}
	})

	t.Run("due date in the past", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := app.Clock.Now()
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    now.AddDate(0, 0, -1),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("next due date cannot be in the past")
		}
	})

	t.Run("missing rule for expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

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
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("recurrence rule must be specified for expenses")
		}
	})

	t.Run("included rule for goal", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.Must(t, NewRuleSet, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some goal",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("recurrence rule cannot be specified for goals")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		now := time.Now()
		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		ruleset := testutils.Must(t, NewRuleSet, FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)
		assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
		nextRecurrence = util.Midnight(nextRecurrence, timezone)

		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
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
		}

		{ // Try to create another expense with the same name
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("failed to create spending: a similar object already exists")
		}
	})
}

func TestGetSpending(t *testing.T) {
	t.Run("list spending objects", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
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
					"ruleset":           ruleset,
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
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.NotZero(t, spendingId, "must be able to extract the spending ID")
		}

		{ // List the spending we've created
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$[0].fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$[0].spendingId").IsEqual(spendingId)
			response.JSON().Path("$").Array().Length().IsEqual(1)
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", "bogus_bank_id").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
		}
	})
}

func TestGetSpendingByID(t *testing.T) {
	t.Run("retrieve single spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

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
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
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
					"ruleset":           ruleset,
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
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.NotZero(t, spendingId, "must be able to extract the spending ID")
		}

		{ // List the spending we've created
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.spendingId").IsEqual(spendingId)
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", "bogus_bank_id").
				WithPath("spendingId", 1234).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
		}
	})

	t.Run("invalid spending ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", "bogus_spending").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("must specify a valid spending Id")
		}
	})

	t.Run("non-existant spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", "spnd_bogus").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("could not retrieve spending: record does not exist")
		}
	})
}

func TestGetSpendingTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token = GivenILogin(t, e, user.Login.Email, password)
		{ // Seed the data for the test.
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(10)

		transactionResponse := response.JSON().Array().Value(0)
		// Make sure there is not already a spending object on the transaction.
		transactionResponse.Path("$.spendingId").IsNull()
		transaction := transactionResponse.Object().Raw()

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
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Before we use the expense, check to make sure there are no transactions.
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		// Now spend the transaction from the expense we just created.
		transaction["spendingId"] = spendingId.String()
		{
			response := e.PUT("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction["transactionId"]).
				WithCookie(TestCookieName, token).
				WithJSON(transaction).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").IsEqual(transaction["transactionId"])
			response.JSON().Path("$.transaction.spendingId").IsEqual(spendingId.String())
		}

		// Now query transactions for the spending object and we should see the
		// transaction we used above.
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].transactionId").IsEqual(transaction["transactionId"])
		}
	})
}
