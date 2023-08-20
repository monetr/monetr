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

func TestPostSpending(t *testing.T) {
	t.Run("create an expense", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
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
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			now := time.Now()
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", "bogus_bank_id").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": math.MaxInt32,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    now.AddDate(0, 0, 1),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", 1234).
				WithCookie(TestCookieName, token).
				WithBytes([]byte("im not json")).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("invalid JSON body")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("spending must have a name")
		}
	})

	t.Run("missing target amount", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": fundingScheduleId,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("target amount must be greater than 0")
		}
	})

	t.Run("negative target amount", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      -1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("target amount must be greater than 0")
		}
	})

	t.Run("invalid funding schedule", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": math.MaxInt32,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("could not find funding schedule specified: record does not exist")
		}
	})

	t.Run("due date in the past", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    now.AddDate(0, 0, -1),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("next due date cannot be in the past")
		}
	})

	t.Run("missing rule for expense", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("recurrence rule must be specified for expenses")
		}
	})

	t.Run("included rule for goal", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some goal",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("recurrence rule cannot be specified for goals")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		now := time.Now()
		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
		nextRecurrence := rule.After(now, false)
		assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
		nextRecurrence = util.Midnight(nextRecurrence, timezone)

		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
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
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
		}

		{ // Try to create another expense with the same name
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      models.SpendingTypeExpense,
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
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		var spendingId uint64
		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
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
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			spendingId = uint64(response.JSON().Path("$.spendingId").Number().Raw())
			assert.NotZero(t, spendingId, "must be able to extract the spending ID")
		}

		{ // List the spending we've created
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$[0].fundingScheduleId").Number().IsEqual(fundingScheduleId)
			response.JSON().Path("$[0].spendingId").Number().IsEqual(spendingId)
			response.JSON().Path("$").Array().Length().IsEqual(1)
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
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
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId uint64
		{ // Create the funding schedule
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
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().Gt(0)
			fundingScheduleId = uint64(response.JSON().Path("$.fundingScheduleId").Number().Raw())
			assert.NotZero(t, fundingScheduleId, "must be able to extract the funding schedule ID")
		}

		var spendingId uint64
		{ // Create an expense
			now := time.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
			rule.DTStart(util.Midnight(now, timezone)) // Force the Rule to be in the correct TZ.
			nextRecurrence := rule.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")
			nextRecurrence = util.Midnight(nextRecurrence, timezone)

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]interface{}{
					"name":              "Some Monthly Expense",
					"recurrenceRule":    "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
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
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			spendingId = uint64(response.JSON().Path("$.spendingId").Number().Raw())
			assert.NotZero(t, spendingId, "must be able to extract the spending ID")
		}

		{ // List the spending we've created
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").Number().IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").Number().IsEqual(fundingScheduleId)
			response.JSON().Path("$.spendingId").Number().IsEqual(spendingId)
		}
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", "bogus_bank_id").
				WithPath("spendingId", 1234).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Must specify a valid bank account ID")
		}
	})

	t.Run("invalid spending ID", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", 1234).
				WithPath("spendingId", "bogus_spending").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Must specify a valid spending ID")
		}
	})

	t.Run("non-existant spending", func(t *testing.T) {
		e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bank := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Create an expense
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", math.MaxInt32).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("could not retrieve spending: record does not exist")
		}
	})
}
