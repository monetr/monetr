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
	"github.com/stretchr/testify/require"
)

func GivenIHaveATransaction(t *testing.T, bankAccount models.BankAccount) models.Transaction {
	transactions := GivenIHaveNTransactions(t, bankAccount, 1)
	require.Len(t, transactions, 1, "must have one transaction")

	return transactions[0]
}

func GivenIHaveNTransactions(t *testing.T, bankAccount models.BankAccount, n int) []models.Transaction {
	require.NotZero(t, bankAccount.BankAccountId, "bank account Id must be included")
	require.NotZero(t, bankAccount.AccountId, "bank account Id must be included")
	require.NotNil(t, bankAccount.Account, "bank account must include account object")

	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(bankAccount.Link.CreatedByUserId, bankAccount.AccountId, db)

	timezone, err := bankAccount.Account.GetTimezone()
	require.NoError(t, err, "must be able to get the timezone from the account")

	transactions := make([]models.Transaction, n)

	for i := 0; i < n; i++ {
		date := util.MidnightInLocal(time.Now(), timezone)

		prefix := gofakeit.RandomString([]string{
			fmt.Sprintf("DEBIT FOR CHECKCARD XXXXXX%s %s", gofakeit.Generate("####"), date.Format("01/02/06")),
			"DEBIT FOR PAYPAL INST XFER CO REF- ",
			"CHECKCARD PURCHASE - ",
		})

		company := gofakeit.Company()
		name := fmt.Sprintf("%s%s", prefix, strings.ToUpper(company))

		transaction := models.Transaction{
			AccountId:                 bankAccount.AccountId,
			Account:                   bankAccount.Account,
			BankAccountId:             bankAccount.BankAccountId,
			BankAccount:               &bankAccount,
			PlaidTransactionId:        gofakeit.UUID(),
			PendingPlaidTransactionId: nil,
			Amount:                    int64(gofakeit.Number(100, 10000)),
			SpendingId:                nil,
			Spending:                  nil,
			SpendingAmount:            nil,
			Categories:                nil,
			OriginalCategories:        nil,
			Date:                      util.MidnightInLocal(time.Now(), timezone),
			AuthorizedDate:            nil,
			Name:                      name,
			CustomName:                nil,
			OriginalName:              name,
			MerchantName:              company,
			OriginalMerchantName:      company,
			IsPending:                 false,
			CreatedAt:                 time.Now(),
		}

		err = repo.CreateTransaction(context.Background(), bankAccount.BankAccountId, &transaction)
		require.NoError(t, err, "must be able to seed transaction")

		transactions[i] = transaction
	}

	return transactions
}

func AssertThatIHaveZeroTransactions(t *testing.T, accountId uint64) {
	db := testutils.GetPgDatabase(t)
	exists, err := db.Model(&models.Transaction{}).Where(`"transaction"."account_id" = ?`, accountId).Exists()
	require.NoError(t, err, "must be able to query transactions successfully")
	if exists {
		panic("account has transactions")
	}
}

func CountTransactions(t *testing.T, accountId uint64) int64 {
	db := testutils.GetPgDatabase(t)
	count, err := db.Model(&models.Transaction{}).Where(`"transaction"."account_id" = ?`, accountId).Count()
	require.NoError(t, err, "must be able to query transactions successfully")

	return int64(count)
}
