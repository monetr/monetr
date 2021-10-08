package repository

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRepositoryBase_GetTransactionsByPlaidTransactionId(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestAuthenticatedRepository(t)

		bankAccounts, err := repo.GetBankAccounts(context.Background())
		require.NoError(t, err, "must be able to retrieve bank accounts")

		var checkingAccount models.BankAccount
		for _, bankAccount := range bankAccounts {
			if bankAccount.SubType == "checking" {
				checkingAccount = bankAccount
				break
			}
		}

		transaction := models.Transaction{
			AccountId:          repo.AccountId(),
			BankAccountId:      checkingAccount.BankAccountId,
			PlaidTransactionId: gofakeit.UUID(),
			Amount:             499,
			Categories: []string{
				"Fast Food",
			},
			OriginalCategories: []string{
				"Fast Food",
			},
			Date:                 time.Now(),
			AuthorizedDate:       nil,
			Name:                 "Wendy's",
			OriginalName:         "Wendy's",
			MerchantName:         "Wendy's",
			OriginalMerchantName: "Wendy's",
			IsPending:            true,
			CreatedAt:            time.Now(),
		}

		require.NoError(t, repo.CreateTransaction(context.Background(), transaction.BankAccountId, &transaction), "must create transaction")

		byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(context.Background(),
			checkingAccount.LinkId,
			[]string{
				transaction.PlaidTransactionId,
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.NotEmpty(t, byPlaidTransaction)
	})
}
