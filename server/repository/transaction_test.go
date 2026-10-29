package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v7"
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
		assert.NoError(t, repo.CreatePlaidTransactions(t.Context(), &plaidTransaction))

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

		require.NoError(t, repo.CreateTransaction(t.Context(), transaction.BankAccountId, &transaction), "must create transaction")

		byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(t.Context(),
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
		assert.NoError(t, repo.CreatePlaidTransactions(t.Context(), &pendingPlaidTransaction))

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
		assert.NoError(t, repo.CreatePlaidTransactions(t.Context(), &plaidTransaction))

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

		require.NoError(t, repo.CreateTransaction(t.Context(), transaction.BankAccountId, &transaction), "must create transaction")

		{ // Query by the non pending ID
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(t.Context(),
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
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(t.Context(),
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
			byPlaidTransaction, err := repo.GetTransactionsByPlaidTransactionId(t.Context(),
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
		require.NoError(t, repo.CreateLunchFlowTransactions(t.Context(), lunchFlowTransactions))

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
		require.NoError(t, repo.CreateTransaction(t.Context(), transaction.BankAccountId, &transaction), "must create transaction")

		byLunchFlowId, err := repo.GetTransactionsByLunchFlowId(
			t.Context(),
			secondBankAccount.BankAccountId,
			[]string{
				lunchFlowId,
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.Empty(t, byLunchFlowId, "must not return a transaction that belongs to another bank account")
	})
}

func TestRepositoryBase_GetTransactionsBySimpleFinId(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetTransactionsBySimpleFinId(
			t.Context(),
			bankAccount.BankAccountId,
			[]string{},
		)
		assert.NoError(t, err, "should not return an error for empty input")
		assert.NotNil(t, result, "should return an empty map, not nil")
		assert.Empty(t, result, "should not return any transactions")
	})

	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		simpleFinTransactions := []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1235",
				Description:            "Arby's",
				Amount:                 -1099,
				Posted:                 clock.Now(),
			},
		}
		require.NoError(t, repo.CreateSimpleFinTransactions(t.Context(), simpleFinTransactions))

		first := models.Transaction{
			AccountId:              repo.AccountId(),
			BankAccountId:          bankAccount.BankAccountId,
			SimpleFinTransactionId: &simpleFinTransactions[0].SimpleFinTransactionId,
			Amount:                 1594,
			Date:                   clock.Now(),
			Name:                   "Wendy's",
			OriginalName:           "Wendy's",
			IsPending:              false,
			Source:                 models.TransactionSourceSimpleFin,
			CreatedAt:              clock.Now(),
		}
		require.NoError(t, repo.CreateTransaction(t.Context(), first.BankAccountId, &first), "must create first transaction")

		second := models.Transaction{
			AccountId:              repo.AccountId(),
			BankAccountId:          bankAccount.BankAccountId,
			SimpleFinTransactionId: &simpleFinTransactions[1].SimpleFinTransactionId,
			Amount:                 1099,
			Date:                   clock.Now(),
			Name:                   "Arby's",
			OriginalName:           "Arby's",
			IsPending:              false,
			Source:                 models.TransactionSourceSimpleFin,
			CreatedAt:              clock.Now(),
		}
		require.NoError(t, repo.CreateTransaction(t.Context(), second.BankAccountId, &second), "must create second transaction")

		result, err := repo.GetTransactionsBySimpleFinId(
			t.Context(),
			bankAccount.BankAccountId,
			[]string{
				"TRN-1234",
				"TRN-1235",
				"TRN-9999",
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.Len(t, result, 2, "should only return the transactions that exist")
		assert.Equal(t, first.TransactionId, result["TRN-1234"].TransactionId, "first transaction should be keyed by its simplefin Id")
		assert.Equal(t, second.TransactionId, result["TRN-1235"].TransactionId, "second transaction should be keyed by its simplefin Id")
		assert.NotContains(t, result, "TRN-9999", "should not return a transaction that does not exist")
	})

	t.Run("includes deleted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		simpleFinTransactions := []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
		}
		require.NoError(t, repo.CreateSimpleFinTransactions(t.Context(), simpleFinTransactions))

		transaction := models.Transaction{
			AccountId:              repo.AccountId(),
			BankAccountId:          bankAccount.BankAccountId,
			SimpleFinTransactionId: &simpleFinTransactions[0].SimpleFinTransactionId,
			Amount:                 1594,
			Date:                   clock.Now(),
			Name:                   "Wendy's",
			OriginalName:           "Wendy's",
			IsPending:              false,
			Source:                 models.TransactionSourceSimpleFin,
			CreatedAt:              clock.Now(),
		}
		require.NoError(t, repo.CreateTransaction(t.Context(), transaction.BankAccountId, &transaction), "must create transaction")
		require.NoError(t, repo.SoftDeleteTransaction(t.Context(), transaction.BankAccountId, transaction.TransactionId), "must soft delete transaction")

		result, err := repo.GetTransactionsBySimpleFinId(
			t.Context(),
			bankAccount.BankAccountId,
			[]string{
				"TRN-1234",
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.Len(t, result, 1, "should return the deleted transaction")
		assert.Equal(t, transaction.TransactionId, result["TRN-1234"].TransactionId, "should return the right transaction")
		assert.NotNil(t, result["TRN-1234"].DeletedAt, "transaction should still be deleted")
	})

	t.Run("does not return transactions from another bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		firstBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		secondBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		simpleFinTransactions := []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *firstBankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
		}
		require.NoError(t, repo.CreateSimpleFinTransactions(t.Context(), simpleFinTransactions))

		transaction := models.Transaction{
			AccountId:              repo.AccountId(),
			BankAccountId:          firstBankAccount.BankAccountId,
			SimpleFinTransactionId: &simpleFinTransactions[0].SimpleFinTransactionId,
			Amount:                 1594,
			Date:                   clock.Now(),
			Name:                   "Wendy's",
			OriginalName:           "Wendy's",
			IsPending:              false,
			Source:                 models.TransactionSourceSimpleFin,
			CreatedAt:              clock.Now(),
		}
		require.NoError(t, repo.CreateTransaction(t.Context(), transaction.BankAccountId, &transaction), "must create transaction")

		result, err := repo.GetTransactionsBySimpleFinId(
			t.Context(),
			secondBankAccount.BankAccountId,
			[]string{
				"TRN-1234",
			},
		)
		assert.NoError(t, err, "should be able to retrieve transactions successfully")
		assert.Empty(t, result, "must not return a transaction that belongs to another bank account")
	})
}
