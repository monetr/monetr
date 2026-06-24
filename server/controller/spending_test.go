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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              gofakeit.Sentence(250),
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			// The schema validates against both the expense and goal variants, so the
			// failure comes back as a oneOf envelope. The expense variant is first.
			response.JSON().Path("$.problems.oneOf[0].name").String().IsEqual("Name must be between 1 and 300 characters")
		}

		{ // Create an expense with a description thats too long
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			// Description is validated as one of a nil value or a valid text field,
			// so a too long description fails both branches inside the expense
			// variant.
			response.JSON().Path("$.problems.oneOf[0].description.oneOf").Array().ConsistsOf(
				"must be nil",
				"Must be between 1 and 300 characters",
			)
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
				WithJSON(map[string]any{
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
			response.JSON().Path("$.error").String().IsEqual("failed to parse request")
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
				WithJSON(map[string]any{
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

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.oneOf[0].name").String().IsEqual("required key is missing")
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.oneOf[0].targetAmount").String().IsEqual("required key is missing")
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      -1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.oneOf[0].targetAmount").String().IsEqual("Target amount must be greater than zero")
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":    "Some Monthly Expense",
					"ruleset": ruleset,
					// A malformed funding schedule Id is now rejected by the schema
					// before we ever look it up, so use a well formed but nonexistent Id
					// to make sure we still exercise the not found path here.
					"fundingScheduleId": NewID[FundingSchedule](),
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
				WithJSON(map[string]any{
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
				WithJSON(map[string]any{
					"name": "Some Monthly Expense",
					// A ruleset is required by the schema for expenses, so include one to
					// make sure we actually get to the past due date check in the
					// controller and not the schema validation.
					"ruleset":           FirstDayOfEveryMonth,
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.oneOf[0].ruleset").String().IsEqual("required key is missing")
		}
	})

	t.Run("missing rule for expense regression 1599", func(t *testing.T) {
		// Pin the timestamp and timezone to a known-bad combo from issue 1599. The
		// old buggy pattern was that util.Midnight would rewind the next recurrence
		// to before now, so the API returned the past-date error instead of the
		// missing-rule error this test is actually checking. The schema validation
		// now runs before any of that date math, so the missing ruleset is caught
		// first regardless of timezone, which is what keeps 1599 fixed.
		t.Setenv("MONETR_TIMESTAMP", "2023-10-31 18:46:01.423737301 +0000 UTC")
		t.Setenv("MONETR_TIMEZONE", "Pacific/Auckland")

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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.oneOf[0].ruleset").String().IsEqual("required key is missing")
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some goal",
					"ruleset":           ruleset,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			// This is a goal with a ruleset, so it fails the goal variant (index 1)
			// of the schema where a ruleset must not be provided.
			response.JSON().Path("$.problems.oneOf[1].ruleset").String().IsEqual("Ruleset cannot be specified for goals")
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
				WithJSON(map[string]any{
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
		ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)
		assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
				WithJSON(map[string]any{
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

	t.Run("rejects auto create transaction on goal", func(t *testing.T) {
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
				WithJSON(map[string]any{
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

		response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                  "Vacation Savings",
				"fundingScheduleId":     fundingScheduleId,
				"targetAmount":          1000,
				"spendingType":          SpendingTypeGoal,
				"nextRecurrence":        app.Clock.Now().Add(30 * 24 * time.Hour),
				"autoCreateTransaction": true,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// A goal cannot specify autoCreateTransaction at all, so the goal variant
		// (index 1) of the schema rejects the key outright.
		response.JSON().Path("$.problems.oneOf[1].autoCreateTransaction").String().IsEqual("key not expected")
	})

	t.Run("rejects auto create transaction on plaid link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                  "Some Monthly Expense",
				"ruleset":               FirstDayOfEveryMonth,
				"fundingScheduleId":     fundingScheduleId,
				"targetAmount":          1000,
				"spendingType":          SpendingTypeExpense,
				"nextRecurrence":        nextRecurrence,
				"autoCreateTransaction": true,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("auto create transaction is only supported for manual links")
	})

	t.Run("creates expense with auto create transaction enabled", func(t *testing.T) {
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
				WithJSON(map[string]any{
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

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                  "Some Monthly Expense",
				"ruleset":               FirstDayOfEveryMonth,
				"fundingScheduleId":     fundingScheduleId,
				"targetAmount":          1000,
				"spendingType":          SpendingTypeExpense,
				"nextRecurrence":        nextRecurrence,
				"autoCreateTransaction": true,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.autoCreateTransaction").Boolean().IsTrue()
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

	t.Run("cant get spending for someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Create a bank account with spending under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to list spending under the other user's bank account
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

	t.Run("cant get someone elses spending by ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var spendingId ID[Spending]

		{ // Create a bank account and spending under one user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.fundingScheduleId").String().Raw())

			spendingId = ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":              "Groceries",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      5000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence": testutils.RuleSetInTimezone(
						t,
						testutils.MustEz(t, user.Account.GetTimezone),
						FirstDayOfEveryMonth,
					).
						After(app.Clock.Now(), false),
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.spendingId").String().Raw())
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to get the spending by ID using the other user's bank account and spending IDs
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		// Now spend the transaction from the expense we just created. The PUT
		// endpoint is gone now so we just PATCH the spendingId onto it.
		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/{transactionId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionId", transaction["transactionId"]).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"spendingId": spendingId.String(),
				}).
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

	t.Run("cant get spending transactions for someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var spendingId ID[Spending]

		{ // Create a bank account and spending under one user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.fundingScheduleId").String().Raw())

			spendingId = ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":              "Utilities",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      8000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence": testutils.RuleSetInTimezone(
						t,
						testutils.MustEz(t, user.Account.GetTimezone),
						FirstDayOfEveryMonth,
					).After(app.Clock.Now(), false),
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.spendingId").String().Raw())
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to get spending transactions using the other user's bank account and spending IDs
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("spending object does not exist")
		}
	})
}

func TestPostSpendingTransfer(t *testing.T) {
	t.Run("move money into spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		{ // Create a deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -10000, // $100
					"isPending":      false,
					"name":           "Deposit",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance + 10000)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance + 10000)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance + 10000)
		}

		var availableBalance, currentBalance, freeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			currentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			freeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		{ // Transfer some money to budget
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000, // $10.00
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(availableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(currentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(freeBalance - 1000)
		}
	})

	t.Run("overdraw free to use", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		{ // Transfer some money to budget
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         startingFreeBalance + 1000,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(startingFreeBalance + 1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(-1000)
		}
	})

	t.Run("between two expenses happy path", func(t *testing.T) {
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
				WithJSON(map[string]any{
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

		{ // Create a deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -10000, // $100
					"isPending":      false,
					"name":           "Deposit",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
		}

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var spendingId ID[Spending]
		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		var spendingIdTwo ID[Spending]
		{ // Create a second expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some other monthly expense",
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
			spendingIdTwo = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Transfer $10 to the first expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance - 1000)
		}

		{ // Transfer $5 to the second expense from the first expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": spendingId,
					"toSpendingId":   spendingIdTwo,
					"amount":         500,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(2)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(500)
			response.JSON().Path("$.spending[1].currentAmount").Number().IsEqual(500)
			// Transfers only affect the free balance, moving between two expenses
			// should not affect the free balance at all.
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance - 1000)
		}

		{ // Retreive the first spending object
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").IsEqual(spendingId)
			// Make sure the first expense has $5 in it.
			response.JSON().Path("$.currentAmount").IsEqual(500)
		}

		{ // Retreive the second spending object
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingIdTwo).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").IsEqual(spendingIdTwo)
			// Make sure the second expense has $5 in it.
			response.JSON().Path("$.currentAmount").IsEqual(500)
		}
	})

	t.Run("between two expenses overdraft", func(t *testing.T) {
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
				WithJSON(map[string]any{
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

		{ // Create a deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -10000, // $100
					"isPending":      false,
					"name":           "Deposit",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
		}

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var spendingId ID[Spending]
		{ // Create an expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		var spendingIdTwo ID[Spending]
		{ // Create a second expense
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some other monthly expense",
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
			spendingIdTwo = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Transfer $10 to the first expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance - 1000)
		}

		{ // Transfer $5 to the second expense from the first expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": spendingId,
					"toSpendingId":   spendingIdTwo,
					"amount":         2000,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Cannot transfer more than is available in source goal/expense")
		}

		{ // Retreive the first spending object
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").IsEqual(spendingId)
			// Make sure the first expense still has $10
			response.JSON().Path("$.currentAmount").IsEqual(1000)
		}

		{ // Retreive the second spending object
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingIdTwo).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spendingId").IsEqual(spendingIdTwo)
			// Make sure the second expense has $0 in it.
			response.JSON().Path("$.currentAmount").IsEqual(0)
		}
	})
}

func TestPutSpending(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Update the spending object
			now := app.Clock.Now()
			timezone := testutils.MustEz(t, user.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some other expense",
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
			// Make sure the name changed!
			response.JSON().Path("$.name").IsEqual("Some other expense")
		}
	})

	t.Run("cannot update another users spending", func(t *testing.T) {
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{
			// Create a new user and try to update the spending object we just
			// created, we should get an error here!
			differentUser, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			differentToken := GivenILogin(t, e, differentUser.Login.Email, password)

			now := app.Clock.Now()
			timezone := testutils.MustEz(t, differentUser.Account.GetTimezone)
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, differentToken).
				WithJSON(map[string]any{
					"name":              "Some other expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to find existing spending: record does not exist")
		}
	})

	t.Run("cant put someone elses spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var spendingId ID[Spending]

		{ // Create a bank account and spending under one user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.fundingScheduleId").String().Raw())

			spendingId = ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":              "Rent",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      100000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence": testutils.RuleSetInTimezone(
						t,
						testutils.MustEz(t, user.Account.GetTimezone),
						FirstDayOfEveryMonth,
					).After(app.Clock.Now(), false),
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.spendingId").String().Raw())
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to update the spending using the other user's bank account and spending IDs
			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "Updated Rent",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to find existing spending: record does not exist")
		}
	})

	t.Run("rejects auto create transaction on goal during update", func(t *testing.T) {
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
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create a goal
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Vacation Savings",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Attempt to enable auto create transaction on the goal
			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":                  "Vacation Savings",
					"fundingScheduleId":     fundingScheduleId,
					"targetAmount":          1000,
					"spendingType":          SpendingTypeGoal,
					"nextRecurrence":        nextRecurrence,
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("auto create transaction is only supported for expenses")
		}
	})

	t.Run("rejects auto create transaction on plaid link during update", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		var spendingId ID[Spending]
		{ // Create an expense on a Plaid link
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Attempt to enable auto create transaction on the Plaid expense
			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":                  "Some Monthly Expense",
					"ruleset":               FirstDayOfEveryMonth,
					"fundingScheduleId":     fundingScheduleId,
					"targetAmount":          1000,
					"spendingType":          SpendingTypeExpense,
					"nextRecurrence":        nextRecurrence,
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("auto create transaction is only supported for manual links")
		}
	})

	t.Run("can toggle auto create transaction on manual link expense", func(t *testing.T) {
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
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		var spendingId ID[Spending]
		{ // Create an expense with auto create transaction off
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsFalse()
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Toggle auto create transaction on
			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":                  "Some Monthly Expense",
					"ruleset":               FirstDayOfEveryMonth,
					"fundingScheduleId":     fundingScheduleId,
					"targetAmount":          1000,
					"spendingType":          SpendingTypeExpense,
					"nextRecurrence":        nextRecurrence,
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsTrue()
		}

		{ // Toggle auto create transaction back off
			response := e.PUT("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":                  "Some Monthly Expense",
					"ruleset":               FirstDayOfEveryMonth,
					"fundingScheduleId":     fundingScheduleId,
					"targetAmount":          1000,
					"spendingType":          SpendingTypeExpense,
					"nextRecurrence":        nextRecurrence,
					"autoCreateTransaction": false,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsFalse()
		}
	})
}

func TestPatchSpending(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Patch the spending object. Unlike the PUT endpoint we do not need to
			// send the entire object, just the fields we want to change. We are not
			// allowed to send the spendingType either since the PATCH schema infers
			// the type from the existing spending object.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "Some other expense",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.nextContributionAmount").Number().Gt(0)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			// Make sure the name changed!
			response.JSON().Path("$.name").IsEqual("Some other expense")
		}
	})

	t.Run("cannot update another users spending", func(t *testing.T) {
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
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{
			// Create a new user and try to update the spending object we just
			// created, we should get an error here!
			differentUser, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			differentToken := GivenILogin(t, e, differentUser.Login.Email, password)

			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, differentToken).
				WithJSON(map[string]any{
					"name": "Some other expense",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to find existing spending: record does not exist")
		}
	})

	t.Run("cant patch someone elses spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var spendingId ID[Spending]

		{ // Create a bank account and spending under one user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.fundingScheduleId").String().Raw())

			spendingId = ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":              "Rent",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      100000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence": testutils.RuleSetInTimezone(
						t,
						testutils.MustEz(t, user.Account.GetTimezone),
						FirstDayOfEveryMonth,
					).After(app.Clock.Now(), false),
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.spendingId").String().Raw())
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to update the spending using the other user's bank account and spending IDs
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "Updated Rent",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to find existing spending: record does not exist")
		}
	})

	t.Run("rejects auto create transaction on goal during update", func(t *testing.T) {
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
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create a goal
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Vacation Savings",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Attempt to enable auto create transaction on the goal. Unlike the PUT
			// endpoint, the PATCH goal schema does not even have an
			// autoCreateTransaction field, so this is rejected at the schema level
			// as an unexpected key rather than with a specific bad request message.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.autoCreateTransaction").String().NotEmpty()
		}
	})

	t.Run("rejects auto create transaction on plaid link during update", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		var spendingId ID[Spending]
		{ // Create an expense on a Plaid link
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Attempt to enable auto create transaction on the Plaid expense
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("auto create transaction is only supported for manual links")
		}
	})

	t.Run("can toggle auto create transaction on manual link expense", func(t *testing.T) {
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
				WithJSON(map[string]any{
					"name":            "Payday",
					"description":     "15th and the Last day of every month",
					"ruleset":         FifthteenthAndLastDayOfEveryMonth,
					"excludeWeekends": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			fundingScheduleId = ID[FundingSchedule](response.JSON().Path("$.fundingScheduleId").String().Raw())
			assert.False(t, fundingScheduleId.IsZero(), "must be able to extract the funding schedule ID")
		}

		now := app.Clock.Now()
		ruleset := testutils.RuleSetInTimezone(t, testutils.MustEz(t, user.Account.GetTimezone), FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(now, false)

		var spendingId ID[Spending]
		{ // Create an expense with auto create transaction off
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsFalse()
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Toggle auto create transaction on
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"autoCreateTransaction": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsTrue()
		}

		{ // Toggle auto create transaction back off
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"autoCreateTransaction": false,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.autoCreateTransaction").Boolean().IsFalse()
		}
	})

	// Everything below here is exercising behavior that is specific to the PATCH
	// endpoint and that the old PUT endpoint did not offer. The whole point of
	// PATCH is that you can send a partial object and only the fields you specify
	// get touched, so most of these are making sure we only change what we were
	// asked to change.

	t.Run("only updates the fields that were provided", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense that we will later partially update
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"description":       "The original description",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			assert.False(t, spendingId.IsZero(), "must be able to extract the spending ID")
		}

		{ // Patch only the name and make sure NOTHING else got touched. This is
			// the core difference between PATCH and PUT, with PUT we would have had
			// to send every single field or it would have been wiped out.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "A brand new name",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").IsEqual("A brand new name")
			// All of these should be exactly what we created the expense with.
			response.JSON().Path("$.description").IsEqual("The original description")
			response.JSON().Path("$.fundingScheduleId").IsEqual(fundingScheduleId)
			response.JSON().Path("$.targetAmount").Number().IsEqual(1000)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(nextRecurrence)
		}

		{ // Re-fetch the spending object to make sure the patch actually persisted
			// to the database and was not just echoed back to us in the response.
			response := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").IsEqual("A brand new name")
			response.JSON().Path("$.description").IsEqual("The original description")
			response.JSON().Path("$.targetAmount").Number().IsEqual(1000)
		}
	})

	t.Run("updating only the target amount recalculates the contribution", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		var originalContribution float64
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
			originalContribution = response.JSON().Path("$.nextContributionAmount").Number().Gt(0).Raw()
		}

		{ // Double the target amount, the contribution should go up to match since
			// we now need to save twice as much by the same date.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"targetAmount": 2000,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.targetAmount").Number().IsEqual(2000)
			response.JSON().Path("$.nextContributionAmount").Number().Gt(originalContribution)
		}
	})

	t.Run("updating only the next recurrence normalizes to midnight", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
		nextRecurrence := ruleset.After(app.Clock.Now(), false)
		// Grab the recurrence after this one so we have a different, valid date to
		// move the spending object to.
		laterRecurrence := ruleset.After(nextRecurrence, false)
		assert.Greater(t, laterRecurrence, nextRecurrence, "the later recurrence should be after the first one")

		var spendingId ID[Spending]
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Move the next recurrence out to the later date.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"nextRecurrence": laterRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			// The recurrence we sent was already midnight in the account timezone so
			// it should round trip exactly.
			response.JSON().Path("$.nextRecurrence").String().AsDateTime(time.RFC3339).IsEqual(laterRecurrence)
		}
	})

	t.Run("can update the ruleset on an expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense that recurs on the first of the month
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Change just the ruleset over to the 15th and last day of the month
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.ruleset").IsEqual(FifthteenthAndLastDayOfEveryMonth)
		}
	})

	t.Run("cannot specify a ruleset on a goal", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create a goal
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Vacation Savings",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Goals do not recur so the goal schema flat out rejects a ruleset.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"ruleset": FirstDayOfEveryMonth,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.ruleset").String().IsEqual("Ruleset cannot be specified for goals")
		}
	})

	t.Run("can pause and unpause a goal", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create a goal that is not paused
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Vacation Savings",
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeGoal,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.isPaused").Boolean().IsFalse()
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Pause the goal
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"isPaused": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.isPaused").Boolean().IsTrue()
		}

		{ // And unpause it again
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"isPaused": false,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.isPaused").Boolean().IsFalse()
		}
	})

	t.Run("cannot pause an expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // isPaused only makes sense for goals, so the expense schema does not even
			// have the field and rejects it as an unexpected key.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"isPaused": true,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.isPaused").String().NotEmpty()
		}
	})

	t.Run("rejects a zero target amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // A target amount of zero is not a positive amount so it gets rejected.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"targetAmount": 0,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.targetAmount").String().NotEmpty()
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // An empty name is not allowed, even though the key itself is optional. If
			// you do send it then it has to be a real name.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().NotEmpty()
		}
	})

	t.Run("an empty patch is a no-op", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Sending an empty body should not blow up and should not change anything
			// since every field on the patch schema is optional.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").IsEqual("Some Monthly Expense")
			response.JSON().Path("$.targetAmount").Number().IsEqual(1000)
			response.JSON().Path("$.ruleset").IsEqual(FirstDayOfEveryMonth)
		}
	})

	t.Run("can update the description", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		var spendingId ID[Spending]
		{ // Create an expense with no description
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":              "Some Monthly Expense",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      1000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence":    nextRecurrence,
				}).
				Expect()

			response.Status(http.StatusOK)
			spendingId = ID[Spending](response.JSON().Path("$.spendingId").String().Raw())
		}

		{ // Add a description to the expense
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"description": "Now it has a description",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.description").IsEqual("Now it has a description")
		}
	})

	// These last few are the more boilerplate-y request validation cases that all
	// of the other PATCH endpoints have. They are not super exciting but they make
	// sure that we are not leaking anything or blowing up on bad input.

	t.Run("invalid bank account Id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
			WithPath("bankAccountId", "bogus_bank_id").
			WithPath("spendingId", "spnd_bogus").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "Does not matter",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
	})

	t.Run("invalid spending Id", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("spendingId", "bogus_spending").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "Does not matter",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid spending Id")
	})

	t.Run("spending does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		// The ID is well formed so it parses fine, but there isnt actually a
		// spending object with this ID so we should get a not found.
		response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("spendingId", "spnd_bogus").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "Does not matter",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to find existing spending: record does not exist")
	})

	t.Run("malformed json", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		// The PATCH endpoint reads the request body AFTER it looks up the spending
		// object, so unlike the PUT we need a real spending to exist or we would
		// just get a 404 before the malformed body is ever evaluated.
		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("spendingId", spendingId).
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I am not really json")).
			Expect()

		response.Status(http.StatusBadRequest)
		// The schema parse path surfaces a decode failure as a generic parse error
		// rather than the PUT's "invalid JSON body".
		response.JSON().Path("$.error").String().IsEqual("failed to parse request")
	})

	t.Run("no authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PATCH(`/api/bank_accounts/bac_bogus/spending/spnd_bogus`).
			WithJSON(map[string]any{
				"name": "Does not matter",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("bad authentication token", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.PATCH(`/api/bank_accounts/bac_bogus/spending/spnd_bogus`).
			WithCookie(TestCookieName, gofakeit.Generate("????????")).
			WithJSON(map[string]any{
				"name": "Does not matter",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("cannot update a read only field", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		{ // The current amount is maintained by monetr based on how much has been
			// contributed and spent, so the patch schema does not expose it and it
			// gets rejected as an unexpected key.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currentAmount": 50000,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currentAmount").String().NotEmpty()
		}
	})

	t.Run("rejects a null name", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		{ // The name key is optional, but if you do send it then it cannot be null.
			// This guards against the merge layer silently dropping a null.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().NotEmpty()
		}
	})

	t.Run("rejects a null target amount", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		{ // A null target amount is not allowed, it has to be a real positive value.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"targetAmount": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.targetAmount").String().NotEmpty()
		}
	})

	t.Run("rejects a null next recurrence", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		{ // A null next recurrence is not allowed when the key is provided.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"nextRecurrence": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.nextRecurrence").String().NotEmpty()
		}
	})

	t.Run("rejects an invalid ruleset", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":    "Payday",
				"ruleset": FifthteenthAndLastDayOfEveryMonth,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.fundingScheduleId").String().Raw())

		nextRecurrence := testutils.RuleSetInTimezone(
			t,
			testutils.MustEz(t, user.Account.GetTimezone),
			FirstDayOfEveryMonth,
		).After(app.Clock.Now(), false)

		spendingId := ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":              "Some Monthly Expense",
				"ruleset":           FirstDayOfEveryMonth,
				"fundingScheduleId": fundingScheduleId,
				"targetAmount":      1000,
				"spendingType":      SpendingTypeExpense,
				"nextRecurrence":    nextRecurrence,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.spendingId").String().Raw())

		{ // Garbage in for the ruleset should be rejected by the ruleset validator.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"ruleset": "this is definitely not a ruleset",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.ruleset").String().IsEqual("Ruleset must be valid")
		}
	})
}

func TestDeleteSpending(t *testing.T) {
	t.Run("delete spending happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		{ // Create a deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -10000, // $100
					"isPending":      false,
					"name":           "Deposit",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance + 10000)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance + 10000)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance + 10000)
		}

		var availableBalance, currentBalance, freeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			currentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			freeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		{ // Transfer some money to budget
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000, // $10.00
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(availableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(currentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(freeBalance - 1000)
			response.JSON().Path("$.balance.expenses").Number().IsEqual(1000)
		}

		{ // Delete the spending object
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // Check to make sure the balance goes back to what it was
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.available").Number().IsEqual(availableBalance)
			response.JSON().Path("$.current").Number().IsEqual(currentBalance)
			// Make sure the amount allocated is returned to free when a spending
			// object is deleted.
			response.JSON().Path("$.free").Number().IsEqual(freeBalance)
		}
	})

	t.Run("delete spending that was used on a transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		var startingAvailableBalance, startingCurrentBalance, startingFreeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			startingAvailableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			startingCurrentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			startingFreeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		var fundingScheduleId ID[FundingSchedule]
		{ // Create the funding schedule
			response := e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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
			ruleset := testutils.RuleSetInTimezone(t, timezone, FirstDayOfEveryMonth)
			nextRecurrence := ruleset.After(now, false)
			assert.Greater(t, nextRecurrence, now, "first of the next month should be relative to now")

			response := e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
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

		{ // Create a deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         -10000, // $100
					"isPending":      false,
					"name":           "Deposit",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(startingAvailableBalance + 10000)
			response.JSON().Path("$.balance.current").Number().IsEqual(startingCurrentBalance + 10000)
			response.JSON().Path("$.balance.free").Number().IsEqual(startingFreeBalance + 10000)
		}

		var availableBalance, currentBalance, freeBalance int64
		{
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			currentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			freeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		{ // Transfer some money to budget
			response := e.POST("/api/bank_accounts/{bankAccountId}/spending/transfer").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"fromSpendingId": nil,
					"toSpendingId":   spendingId,
					"amount":         1000, // $10.00
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.spending").Array().Length().IsEqual(1)
			response.JSON().Path("$.spending[0].currentAmount").Number().IsEqual(1000)
			// Transfers only affect the free balance
			response.JSON().Path("$.balance.available").Number().IsEqual(availableBalance)
			response.JSON().Path("$.balance.current").Number().IsEqual(currentBalance)
			response.JSON().Path("$.balance.free").Number().IsEqual(freeBalance - 1000)
			response.JSON().Path("$.balance.expenses").Number().IsEqual(1000)
		}

		{ // Retrieve updated balances
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			currentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			freeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		{ // Create a transaction
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount":         1000, // $100
					"isPending":      false,
					"name":           "Spending from my budget",
					"date":           app.Clock.Now(), // Should use midnight, but idc
					"adjustsBalance": true,
					"spendingId":     spendingId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transaction.transactionId").String().NotEmpty()
			response.JSON().Path("$.balance.available").Number().IsEqual(availableBalance - 1000)
			response.JSON().Path("$.balance.current").Number().IsEqual(currentBalance - 1000)
			response.JSON().Path("$.balance.free").Number().IsEqual(freeBalance) // Doesn't change when spent from
		}

		{ // Retrieve updated balances
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			availableBalance = int64(response.JSON().Path("$.available").Number().Gt(0).Raw())
			currentBalance = int64(response.JSON().Path("$.current").Number().Gt(0).Raw())
			freeBalance = int64(response.JSON().Path("$.free").Number().Gt(0).Raw())
		}

		{ // Delete the spending object
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // Check to make sure the balance goes back to what it was
			response := e.GET("/api/bank_accounts/{bankAccountId}/balances").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.available").Number().IsEqual(availableBalance)
			response.JSON().Path("$.current").Number().IsEqual(currentBalance)
			// Make sure the amount allocated is returned to free when a spending
			// object is deleted.
			response.JSON().Path("$.free").Number().IsEqual(freeBalance)
		}
	})

	t.Run("cant delete someone elses spending", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var spendingId ID[Spending]

		{ // Create a bank account and spending under one user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			fundingScheduleId := ID[FundingSchedule](e.POST("/api/bank_accounts/{bankAccountId}/funding_schedules").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":    "Payday",
					"ruleset": FifthteenthAndLastDayOfEveryMonth,
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.fundingScheduleId").String().Raw())

			spendingId = ID[Spending](e.POST("/api/bank_accounts/{bankAccountId}/spending").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, tok).
				WithJSON(map[string]any{
					"name":              "Groceries",
					"ruleset":           FirstDayOfEveryMonth,
					"fundingScheduleId": fundingScheduleId,
					"targetAmount":      5000,
					"spendingType":      SpendingTypeExpense,
					"nextRecurrence": testutils.RuleSetInTimezone(
						t,
						testutils.MustEz(t, user.Account.GetTimezone),
						FirstDayOfEveryMonth,
					).After(app.Clock.Now(), false),
				}).
				Expect().
				Status(http.StatusOK).
				JSON().Path("$.spendingId").String().Raw())
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to delete the spending using the other user's bank account and spending IDs
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("spendingId", spendingId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("spending object does not exist")
		}
	})
}
