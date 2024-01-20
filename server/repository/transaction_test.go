package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryBase_GetTransactionsByPlaidTransactionId(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		checkingAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

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
