package fixtures

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/require"
)

func GivenIHaveATransactionName(t *testing.T, clock clock.Clock) (name, company string) {
	date := util.Midnight(clock.Now(), time.UTC)
	prefix := gofakeit.RandomString([]string{
		fmt.Sprintf("DEBIT FOR CHECKCARD XXXXXX1234 %s", date.Format("01/02/06")),
		"DEBIT FOR PAYPAL INST XFER CO REF- ",
		"CHECKCARD PURCHASE - ",
		"POS Debit - Visa Check Card 1234 -\n\t\t\t\t\t",
		"POS Debit - 1234 -\n\t\t\t\t\t",
		"POS Debit - Visa Check Card 1234 -\n\t\t\t\t\tPWP*",
		"POS Debit - 1234 -\n\t\t\t\t\tPWP*",
		"ACH Transaction - ",
	})

	company = gofakeit.Company()
	name = fmt.Sprintf("%s%s", prefix, strings.ToUpper(company))
	require.NotEmpty(t, name, "transaction name cannot be empty")

	return name, company
}

func GivenIHaveATransaction(t *testing.T, clock clock.Clock, bankAccount BankAccount) Transaction {
	transactions := GivenIHaveNTransactions(t, clock, bankAccount, 1)
	require.Len(t, transactions, 1, "must have one transaction")

	return transactions[0]
}

func GivenIHaveNTransactions(t *testing.T, clock clock.Clock, bankAccount BankAccount, n int) []Transaction {
	require.NotZero(t, bankAccount.BankAccountId, "bank account Id must be included")
	require.NotZero(t, bankAccount.AccountId, "bank account Id must be included")
	require.NotNil(t, bankAccount.Account, "bank account must include account object")

	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(
		clock,
		bankAccount.Link.CreatedBy,
		bankAccount.AccountId,
		db,
		log,
	)

	timezone, err := bankAccount.Account.GetTimezone()
	require.NoError(t, err, "must be able to get the timezone from the account")

	transactions := make([]Transaction, n)

	for i := 0; i < n; i++ {
		date := util.Midnight(clock.Now(), timezone)
		name, company := GivenIHaveATransactionName(t, clock)
		amount := int64(gofakeit.Number(100, 10000))

		var source TransactionSource = TransactionSourceUpload
		var plaidTransaction *PlaidTransaction
		if plaidBankAccount := bankAccount.PlaidBankAccount; plaidBankAccount != nil {
			plaidTransaction = &PlaidTransaction{
				AccountId:          bankAccount.AccountId,
				PlaidBankAccountId: plaidBankAccount.PlaidBankAccountId,
				PlaidId:            gofakeit.UUID(),
				PendingPlaidId:     nil,
				Categories:         nil,
				Date:               date,
				AuthorizedDate:     nil,
				Name:               name,
				MerchantName:       company,
				Amount:             amount,
				Currency:           "USD",
				IsPending:          false,
				CreatedAt:          clock.Now().UTC(),
				DeletedAt:          nil,
			}
			source = TransactionSourcePlaid

			require.NoError(
				t,
				repo.CreatePlaidTransactions(t.Context(), plaidTransaction),
				"must be able to seed transaction",
			)
		}

		transaction := Transaction{
			AccountId:                 bankAccount.AccountId,
			Account:                   bankAccount.Account,
			BankAccountId:             bankAccount.BankAccountId,
			BankAccount:               &bankAccount,
			PendingPlaidTransactionId: nil,
			Amount:                    amount,
			SpendingId:                nil,
			Spending:                  nil,
			SpendingAmount:            nil,
			Categories:                nil,
			Date:                      util.Midnight(clock.Now(), timezone),
			Name:                      name,
			OriginalName:              name,
			MerchantName:              company,
			OriginalMerchantName:      company,
			IsPending:                 false,
			Source:                    source,
			CreatedAt:                 clock.Now(),
		}
		if plaidTransaction != nil {
			transaction.PlaidTransactionId = &plaidTransaction.PlaidTransactionId
			transaction.PlaidTransaction = plaidTransaction
		}

		err = repo.CreateTransaction(t.Context(), bankAccount.BankAccountId, &transaction)
		require.NoError(t, err, "must be able to seed transaction")

		transactions[i] = transaction
	}

	return transactions
}

func AssertThatIHaveZeroTransactions(t *testing.T, accountId ID[Account]) {
	db := testutils.GetPgDatabase(t)
	exists, err := db.Model(&Transaction{}).Where(`"transaction"."account_id" = ?`, accountId).Exists()
	require.NoError(t, err, "must be able to query transactions successfully")
	if exists {
		panic("account has transactions")
	}
}

func CountNonDeletedTransactions(t *testing.T, accountId ID[Account]) int64 {
	db := testutils.GetPgDatabase(t)
	count, err := db.Model(&Transaction{}).
		Where(`"transaction"."account_id" = ?`, accountId).
		Where(`"transaction"."deleted_at" IS NULL`).
		Count()
	require.NoError(t, err, "must be able to query transactions successfully")

	return int64(count)
}

func CountAllTransactions(t *testing.T, accountId ID[Account]) int64 {
	db := testutils.GetPgDatabase(t)
	count, err := db.Model(&Transaction{}).
		Where(`"transaction"."account_id" = ?`, accountId).
		Count()
	require.NoError(t, err, "must be able to query transactions successfully")

	return int64(count)
}

func CountPendingTransactions(t *testing.T, accountId ID[Account]) int64 {
	db := testutils.GetPgDatabase(t)
	count, err := db.Model(&Transaction{}).
		Where(`"transaction"."account_id" = ?`, accountId).
		Where(`"transaction"."is_pending" = ?`, true).
		Where(`"transaction"."deleted_at" IS NULL`).
		Count()
	require.NoError(t, err, "must be able to query transactions successfully")

	return int64(count)
}
