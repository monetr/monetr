package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func givenIHaveATransactionRecurring(
	t *testing.T,
	bank BankAccount,
) TransactionRecurring {
	cluster := testutils.MustInsert(t, TransactionCluster{
		AccountId:     bank.AccountId,
		BankAccountId: bank.BankAccountId,
		Name:          "Github",
		OriginalName:  "Github",
		Members:       []ID[Transaction]{NewID[Transaction]()},
	})

	return testutils.MustInsert(t, TransactionRecurring{
		AccountId:            bank.AccountId,
		BankAccountId:        bank.BankAccountId,
		TransactionClusterId: cluster.TransactionClusterId,
		Direction:            DebitDirection,
		Window:               MonthlyWindowType,
		RuleSet: testutils.Must(
			t,
			NewRuleSet,
			"DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15",
		),
		First:      time.Date(2026, 1, 15, 6, 0, 0, 0, time.UTC),
		Last:       time.Date(2026, 3, 15, 5, 0, 0, 0, time.UTC),
		Next:       time.Date(2026, 4, 15, 5, 0, 0, 0, time.UTC),
		Confidence: 0.9,
		Amounts:    map[int64]int{800: 3},
		LastAmount: 800,
	})
}

func givenIHaveASpending(
	t *testing.T,
	clock clock.Clock,
	fundingSchedule *FundingSchedule,
	spendingType SpendingType,
	name string,
) Spending {
	var ruleset *RuleSet
	nextRecurrence := clock.Now().AddDate(0, 1, 0)
	if spendingType == SpendingTypeExpense {
		ruleset = testutils.Must(t, NewRuleSet, "DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15")
		nextRecurrence = ruleset.After(clock.Now(), false)
	}

	return testutils.MustInsert(t, Spending{
		AccountId:         fundingSchedule.AccountId,
		BankAccountId:     fundingSchedule.BankAccountId,
		FundingScheduleId: fundingSchedule.FundingScheduleId,
		SpendingType:      spendingType,
		Name:              name,
		TargetAmount:      800,
		RuleSet:           ruleset,
		NextRecurrence:    nextRecurrence,
		CreatedAt:         clock.Now(),
	})
}

func givenIHaveACreditTransactionRecurring(
	t *testing.T,
	bank BankAccount,
) TransactionRecurring {
	recurring := givenIHaveATransactionRecurring(t, bank)
	recurring.Direction = CreditDirection
	testutils.MustDBUpdate(t, &recurring)
	return recurring
}

func TestGetRecurringTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := givenIHaveATransactionRecurring(t, bank)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/recurring").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(1)
		response.JSON().Path("$[0].transactionRecurringId").IsEqual(recurring.TransactionRecurringId)
		response.JSON().Path("$[0].transactionCluster.name").IsEqual("Github")
		response.JSON().Path("$[0].transactionCluster").Object().NotContainsKey("members")
		response.JSON().Path("$[0].transactionCluster.debug").IsNull()
	})

	t.Run("pagination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		for range 30 {
			givenIHaveATransactionRecurring(t, bank)
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
		ended := givenIHaveATransactionRecurring(t, bank)
		ended.Ended = true
		ended.Next = time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
		testutils.MustDBUpdate(t, &ended)
		active := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("limit cannot be greater than 100")
	})

	t.Run("cant get recurring for someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		otherLink := fixtures.GivenIHaveAManualLink(t, app.Clock, otherUser)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &otherLink, DepositoryBankAccountType, CheckingBankAccountSubType)
		givenIHaveATransactionRecurring(t, otherBank)

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
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		recurring := givenIHaveATransactionRecurring(t, bank)
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

	t.Run("cannot read someone else's recurring transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bank BankAccount
		var recurring TransactionRecurring
		{ // Seed the recurring transaction under the first account.
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			recurring = givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
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
		response.JSON().Path("$.error").String().IsEqual("must specify a valid recurring transaction Id")
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
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.spending.spendingId").IsEqual(spending.SpendingId)
		response.JSON().Path("$.fundingScheduleId").IsNull()
		response.JSON().Path("$.lastAmount").IsEqual(800)

		stored := testutils.MustDBRead(t, recurring)
		if assert.NotNil(t, stored.SpendingId, "spending ID must be persisted") {
			assert.Equal(t, spending.SpendingId, *stored.SpendingId)
		}
		assert.WithinDuration(t, app.Clock.Now(), stored.UpdatedAt, time.Second, "updated at should be bumped")

		// And the spending object itself no longer says anything about recurring.
		spendingResponse := e.GET("/api/bank_accounts/{bankAccountId}/spending/{spendingId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("spendingId", spending.SpendingId).
			WithCookie(TestCookieName, token).
			Expect()
		spendingResponse.Status(http.StatusOK)
		spendingResponse.JSON().Object().NotContainsKey("transactionRecurringId")
	})

	t.Run("links a funding schedule", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := givenIHaveACreditTransactionRecurring(t, bank)
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
		if assert.NotNil(t, stored.FundingScheduleId, "funding schedule ID must be persisted") {
			assert.Equal(t, fundingSchedule.FundingScheduleId, *stored.FundingScheduleId)
		}
	})

	t.Run("with an api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		assert.Nil(t, stored.SpendingId, "spending ID must be cleared")
	})

	t.Run("re-points to another expense", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		first := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		second := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github Copilot")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.spending.spendingId").IsEqual(second.SpendingId)
	})

	t.Run("clears auto matched when the user changes the link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		first := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		second := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github Copilot")
		recurring := givenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &first.SpendingId
		recurring.AutoMatched = true
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		e.GET("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			Expect().
			Status(http.StatusOK).
			JSON().Path("$.autoMatched").IsEqual(true)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/recurring/{transactionRecurringId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionRecurringId", recurring.TransactionRecurringId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"spendingId": second.SpendingId,
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.autoMatched").IsEqual(false)
		stored := testutils.MustDBRead(t, recurring)
		assert.False(t, stored.AutoMatched, "a link the user chose must not be marked as auto matched")
	})

	t.Run("an expense can only be linked once", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		linked := givenIHaveATransactionRecurring(t, bank)
		linked.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &linked)
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("failed to update recurring transaction: a similar object already exists")

		stored := testutils.MustDBRead(t, recurring)
		assert.Nil(t, stored.SpendingId, "the second recurring transaction must not be linked")
	})

	t.Run("rejects an expense on a credit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveACreditTransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("spending can only be linked to a debit recurring transaction")
	})

	t.Run("rejects a funding schedule on a debit", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("funding schedules can only be linked to a credit recurring transaction")
	})

	t.Run("rejects a goal", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		goal := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeGoal, "Vacation")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("only expenses can be linked to a recurring transaction")
	})

	t.Run("expense does not exist", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("could not find spending specified: record does not exist")
	})

	t.Run("expense from another bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, SavingsBankAccountSubType)
		otherFundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &otherBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, otherFundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("could not find spending specified: record does not exist")
	})

	t.Run("funding schedule from another bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		otherBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, SavingsBankAccountSubType)
		otherFundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &otherBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		recurring := givenIHaveACreditTransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("could not find funding schedule specified: record does not exist")
	})

	t.Run("rejects other fields", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		recurring := givenIHaveATransactionRecurring(t, bank)
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
		response.JSON().Path("$.error").IsEqual("Invalid request")

		stored := testutils.MustDBRead(t, recurring)
		assert.EqualValues(t, 800, stored.LastAmount, "last amount must not change")
	})

	t.Run("deleting the expense clears the link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &bank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
		recurring := givenIHaveATransactionRecurring(t, bank)
		recurring.SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		token := GivenILogin(t, e, user.Login.Email, password)

		{
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

	t.Run("cannot patch someone else's recurring transaction", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var bank BankAccount
		var recurring TransactionRecurring
		{ // Seed the recurring transaction under the first account.
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			recurring = givenIHaveATransactionRecurring(t, bank)
		}

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		myBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, app.Clock, &myBank, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
		spending := givenIHaveASpending(t, app.Clock, fundingSchedule, SpendingTypeExpense, "Github")
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
		assert.Nil(t, stored.SpendingId, "someone else's recurring transaction must not be touched")
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
