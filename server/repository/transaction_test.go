package repository_test

import (
	"context"
	"testing"

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
	t.Run("non-pending", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		checkingAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		plaidTransaction := models.PlaidTransaction{
			AccountId:          repo.AccountId(),
			PlaidBankAccountId: *checkingAccount.PlaidBankAccountId,
			PlaidId:            gofakeit.UUID(),
			PendingPlaidId:     nil,
			Categories: []string{
				"Fast Food",
			},
			Date:           clock.Now(),
			AuthorizedDate: nil,
			Name:           "Wendy's",
			MerchantName:   "Wendy's",
			Amount:         1594,
			Currency:       "USD",
			IsPending:      false,
			CreatedAt:      clock.Now(),
			DeletedAt:      nil,
		}
		assert.NoError(t, repo.CreatePlaidTransactions(context.Background(), &plaidTransaction))

		transaction := models.Transaction{
			AccountId:          repo.AccountId(),
			BankAccountId:      checkingAccount.BankAccountId,
			PlaidTransactionId: &plaidTransaction.PlaidTransactionId,
			Amount:             1594,
			Categories: []string{
				"Fast Food",
			},
			Date:                 clock.Now(),
			Name:                 "Wendy's",
			OriginalName:         "Wendy's",
			MerchantName:         "Wendy's",
			OriginalMerchantName: "Wendy's",
			IsPending:            false,
			Source:               models.TransactionSourcePlaid,
			CreatedAt:            clock.Now(),
		}

		require.NoError(t, repo.CreateTransaction(context.Background(), transaction.BankAccountId, &transaction), "must create transaction")

		byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(context.Background(),
			checkingAccount.LinkId,
			[]string{
				plaidTransaction.PlaidId,
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.NotEmpty(t, byPlaidTransaction)
		assert.Equal(t, transaction.TransactionId, byPlaidTransaction[0].TransactionId)
	})

	t.Run("with pending", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		checkingAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		pendingPlaidTransaction := models.PlaidTransaction{
			AccountId:          repo.AccountId(),
			PlaidBankAccountId: *checkingAccount.PlaidBankAccountId,
			PlaidId:            gofakeit.UUID(),
			PendingPlaidId:     nil,
			Categories: []string{
				"Fast Food",
			},
			Date:           clock.Now(),
			AuthorizedDate: nil,
			Name:           "Wendy's",
			MerchantName:   "Wendy's",
			Amount:         1594,
			Currency:       "USD",
			IsPending:      true,
			CreatedAt:      clock.Now(),
			DeletedAt:      nil,
		}
		assert.NoError(t, repo.CreatePlaidTransactions(context.Background(), &pendingPlaidTransaction))

		plaidTransaction := models.PlaidTransaction{
			AccountId:          repo.AccountId(),
			PlaidBankAccountId: *checkingAccount.PlaidBankAccountId,
			PlaidId:            gofakeit.UUID(),
			PendingPlaidId:     &pendingPlaidTransaction.PlaidId,
			Categories: []string{
				"Fast Food",
			},
			Date:           clock.Now(),
			AuthorizedDate: nil,
			Name:           "Wendy's",
			MerchantName:   "Wendy's",
			Amount:         1594,
			Currency:       "USD",
			IsPending:      false,
			CreatedAt:      clock.Now(),
			DeletedAt:      nil,
		}
		assert.NoError(t, repo.CreatePlaidTransactions(context.Background(), &plaidTransaction))

		transaction := models.Transaction{
			AccountId:                 repo.AccountId(),
			BankAccountId:             checkingAccount.BankAccountId,
			PlaidTransactionId:        &plaidTransaction.PlaidTransactionId,
			PendingPlaidTransactionId: &pendingPlaidTransaction.PlaidTransactionId,
			Amount:                    1594,
			Categories: []string{
				"Fast Food",
			},
			Date:                 clock.Now(),
			Name:                 "Wendy's",
			OriginalName:         "Wendy's",
			MerchantName:         "Wendy's",
			OriginalMerchantName: "Wendy's",
			IsPending:            false,
			Source:               models.TransactionSourcePlaid,
			CreatedAt:            clock.Now(),
		}

		require.NoError(t, repo.CreateTransaction(context.Background(), transaction.BankAccountId, &transaction), "must create transaction")

		{ // Query by the non pending ID
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(context.Background(),
				checkingAccount.LinkId,
				[]string{
					plaidTransaction.PlaidId,
				},
			)
			assert.NoError(t, err, "should be able to retrieve transactions successfully")
			assert.NotEmpty(t, byPlaidTransaction)
			assert.Len(t, byPlaidTransaction, 1)
			assert.Equal(t, transaction.TransactionId, byPlaidTransaction[0].TransactionId)
		}

		{ // And query by the pending transaction ID.
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(context.Background(),
				checkingAccount.LinkId,
				[]string{
					pendingPlaidTransaction.PlaidId,
				},
			)
			assert.NoError(t, err, "should be able to retrieve transactions successfully")
			assert.NotEmpty(t, byPlaidTransaction)
			assert.Len(t, byPlaidTransaction, 1)
			assert.Equal(t, transaction.TransactionId, byPlaidTransaction[0].TransactionId)
		}

		{ // And query by both!
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(context.Background(),
				checkingAccount.LinkId,
				[]string{
					plaidTransaction.PlaidId,
					pendingPlaidTransaction.PlaidId,
				},
			)
			assert.NoError(t, err, "should be able to retrieve transactions successfully")
			assert.NotEmpty(t, byPlaidTransaction)
			assert.Len(t, byPlaidTransaction, 1)
			assert.Equal(t, transaction.TransactionId, byPlaidTransaction[0].TransactionId)
		}
	})
}

func TestRepositoryBase_GetTransactionsByLunchFlowId(t *testing.T) {
	t.Run("does not return transactions from another bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		firstLink := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		firstBankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &firstLink)
		secondLink := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		secondBankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &secondLink)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		lunchFlowId := gofakeit.UUID()
		lunchFlowTransactions := []models.LunchFlowTransaction{
			{
				AccountId:              repo.AccountId(),
				LunchFlowBankAccountId: *firstBankAccount.LunchFlowBankAccountId,
				LunchFlowId:            lunchFlowId,
				Merchant:               "Wendy's",
				Description:            "Wendy's",
				Date:                   clock.Now(),
				Currency:               "USD",
				Amount:                 1594,
				IsPending:              false,
				CreatedAt:              clock.Now(),
			},
		}
		require.NoError(t, repo.CreateLunchFlowTransactions(context.Background(), lunchFlowTransactions))

		transaction := models.Transaction{
			AccountId:              repo.AccountId(),
			BankAccountId:          firstBankAccount.BankAccountId,
			LunchFlowTransactionId: &lunchFlowTransactions[0].LunchFlowTransactionId,
			Amount:                 1594,
			Categories: []string{
				"Fast Food",
			},
			Date:                 clock.Now(),
			Name:                 "Wendy's",
			OriginalName:         "Wendy's",
			MerchantName:         "Wendy's",
			OriginalMerchantName: "Wendy's",
			IsPending:            false,
			Source:               models.TransactionSourceLunchFlow,
			CreatedAt:            clock.Now(),
		}
		require.NoError(t, repo.CreateTransaction(context.Background(), transaction.BankAccountId, &transaction), "must create transaction")

		byLunchFlowId, err := repo.GetTransactionsByLunchFlowId(
			context.Background(),
			secondBankAccount.BankAccountId,
			[]string{
				lunchFlowId,
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.Empty(t, byLunchFlowId, "must not return a transaction that belongs to another bank account")
	})
}
