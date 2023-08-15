package fixtures

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGivenIHaveATransaction(t *testing.T) {
	user, _ := GivenIHaveABasicAccount(t)
	link := GivenIHaveAPlaidLink(t, user)
	bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

	transaction := GivenIHaveATransaction(t, bankAccount)
	assert.NotZero(t, transaction.TransactionId, "transaction must have been created")
	assert.NotNil(t, transaction.Account, "account must be included on the transaction")
	assert.NotNil(t, transaction.BankAccount, "bank account must be included on the transaction")
	assert.Greater(t, transaction.Amount, int64(0), "amount must be greater than 0")
}

func TestAssertThatIHaveZeroTransactions(t *testing.T) {
	t.Run("no transactions", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		AssertThatIHaveZeroTransactions(t, user.AccountId)
	})

	t.Run("with transactions", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		GivenIHaveATransaction(t, bankAccount)

		assert.Panics(t, func() {
			AssertThatIHaveZeroTransactions(t, user.AccountId)
		})
	})
}

func TestCountTransactions(t *testing.T) {
	t.Run("no transactions", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		assert.EqualValues(t, 0, CountNonDeletedTransactions(t, user.AccountId))
	})

	t.Run("with transactions", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		GivenIHaveATransaction(t, bankAccount)

		assert.EqualValues(t, 1, CountNonDeletedTransactions(t, user.AccountId))
	})
}

func TestCountPendingTransactions(t *testing.T) {
	t.Run("no pending transactions", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		GivenIHaveATransaction(t, bankAccount)

		assert.EqualValues(t, 0, CountPendingTransactions(t, user.AccountId))
	})

	t.Run("one pending transaction", func(t *testing.T) {
		user, _ := GivenIHaveABasicAccount(t)
		link := GivenIHaveAPlaidLink(t, user)
		bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		// Create a non-pending transaction
		GivenIHaveATransaction(t, bankAccount)

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(bankAccount.Link.CreatedByUserId, bankAccount.AccountId, db)

		timezone, err := bankAccount.Account.GetTimezone()
		require.NoError(t, err, "must be able to get the timezone from the account")

		date := util.Midnight(time.Now(), timezone)

		prefix := gofakeit.RandomString([]string{
			fmt.Sprintf("DEBIT FOR CHECKCARD XXXXXX%s %s", gofakeit.Generate("####"), date.Format("01/02/06")),
			"DEBIT FOR PAYPAL INST XFER CO REF- ",
			"CHECKCARD PURCHASE - ",
		})

		company := gofakeit.Company()
		name := fmt.Sprintf("%s%s", prefix, strings.ToUpper(company))

		transaction := models.Transaction{
			AccountId:            bankAccount.AccountId,
			Account:              bankAccount.Account,
			BankAccountId:        bankAccount.BankAccountId,
			BankAccount:          &bankAccount,
			PlaidTransactionId:   gofakeit.UUID(),
			Amount:               int64(gofakeit.Number(100, 10000)),
			Date:                 util.Midnight(time.Now(), timezone),
			Name:                 name,
			OriginalName:         name,
			MerchantName:         company,
			OriginalMerchantName: company,
			IsPending:            true,
			CreatedAt:            time.Now(),
		}

		err = repo.CreateTransaction(context.Background(), bankAccount.BankAccountId, &transaction)
		require.NoError(t, err, "must be able to seed transaction")

		assert.EqualValues(t, 2, CountNonDeletedTransactions(t, user.AccountId))
		assert.EqualValues(t, 1, CountPendingTransactions(t, user.AccountId))
	})
}
