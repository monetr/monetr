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
	require.NotZero(t, bankAccount.BankAccountId, "bank account Id must be included")
	require.NotZero(t, bankAccount.AccountId, "bank account Id must be included")
	require.NotNil(t, bankAccount.Account, "bank account must include account object")

	timezone, err := bankAccount.Account.GetTimezone()
	require.NoError(t, err, "must be able to get the timezone from the account")

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

	db := testutils.GetTestDatabase(t)
	repo := repository.NewRepositoryFromSession(bankAccount.Link.CreatedByUserId, bankAccount.AccountId, db)

	err = repo.CreateTransaction(context.Background(), bankAccount.BankAccountId, &transaction)
	require.NoError(t, err, "must be able to seed transaction")

	return transaction
}
