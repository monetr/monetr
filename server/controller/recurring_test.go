package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
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
