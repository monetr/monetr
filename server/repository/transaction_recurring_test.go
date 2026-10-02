package repository_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func givenIHaveATransactionCluster(
	t *testing.T,
	bankAccount models.BankAccount,
) models.TransactionCluster {
	return testutils.MustInsert(t, models.TransactionCluster{
		AccountId:     bankAccount.AccountId,
		BankAccountId: bankAccount.BankAccountId,
		Name:          "Github",
		OriginalName:  "Github",
		Members:       []models.ID[models.Transaction]{models.NewID[models.Transaction]()},
	})
}

func newTransactionRecurring(
	t *testing.T,
	cluster models.TransactionCluster,
	direction models.Direction,
	amount int64,
) models.TransactionRecurring {
	first := time.Date(2026, 1, 15, 6, 0, 0, 0, time.UTC)
	last := time.Date(2026, 3, 15, 5, 0, 0, 0, time.UTC)
	return models.TransactionRecurring{
		TransactionClusterId: cluster.TransactionClusterId,
		Direction:            direction,
		Window:               models.MonthlyWindowType,
		RuleSet: testutils.Must(
			t,
			models.NewRuleSet,
			"DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15",
		),
		First:      first,
		Last:       last,
		Next:       time.Date(2026, 4, 15, 5, 0, 0, 0, time.UTC),
		Confidence: 0.9,
		Amounts:    map[int64]int{amount: 3},
		LastAmount: amount,
	}
}

func TestRepositoryBase_GetTransactionRecurringByCluster(t *testing.T) {
	t.Run("no recurring transactions", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		result, err := repo.GetTransactionRecurringByCluster(t.Context(), bankAccount.BankAccountId, cluster.TransactionClusterId)
		assert.NoError(t, err, "must be able to read recurring transactions")
		assert.Empty(t, result, "there should be no recurring transactions")
	})

	t.Run("only the specified cluster", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		otherCluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
			newTransactionRecurring(t, cluster, models.CreditDirection, -500),
			newTransactionRecurring(t, otherCluster, models.DebitDirection, 1200),
		})
		require.NoError(t, err, "must be able to create recurring transactions")

		result, err := repo.GetTransactionRecurringByCluster(t.Context(), bankAccount.BankAccountId, cluster.TransactionClusterId)
		assert.NoError(t, err, "must be able to read recurring transactions")
		require.Len(t, result, 2, "should only return the recurring transactions for the cluster")
		for _, item := range result {
			assert.Equal(t, cluster.TransactionClusterId, item.TransactionClusterId, "cluster should match")
		}
	})
}

func TestRepositoryBase_UpsertTransactionRecurring(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		items := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items)
		assert.NoError(t, err, "must be able to create recurring transaction")
		assert.False(t, items[0].TransactionRecurringId.IsZero(), "id must be assigned")
		assert.Equal(t, user.AccountId, items[0].AccountId, "account id must be set from session")
		assert.Equal(t, bankAccount.BankAccountId, items[0].BankAccountId, "bank account id must be set")

		stored := testutils.MustDBRead(t, items[0])
		assert.Equal(t, models.DebitDirection, stored.Direction, "direction should match")
		assert.Equal(t, models.MonthlyWindowType, stored.Window, "window should match")
		assert.Equal(t, items[0].RuleSet.String(), stored.RuleSet.String(), "ruleset should match")
		assert.Equal(t, map[int64]int{800: 3}, stored.Amounts, "amounts should match")
		assert.EqualValues(t, 800, stored.LastAmount, "last amount should match")
	})

	t.Run("updates existing for the same cluster and direction", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		original := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, original)
		require.NoError(t, err, "must be able to create recurring transaction")

		// A new recurring transaction for the same cluster and direction, it has no
		// ID so it would be a new record if it weren't for the unique constraint.
		updated := newTransactionRecurring(t, cluster, models.DebitDirection, 1000)
		updated.Ended = true
		err = repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.TransactionRecurring{updated})
		assert.NoError(t, err, "must be able to update recurring transaction")

		result, err := repo.GetTransactionRecurringByCluster(t.Context(), bankAccount.BankAccountId, cluster.TransactionClusterId)
		require.NoError(t, err, "must be able to read recurring transactions")
		require.Len(t, result, 1, "there should still only be one recurring transaction")
		assert.Equal(t, original[0].TransactionRecurringId, result[0].TransactionRecurringId, "should keep the existing id")
		assert.EqualValues(t, 1000, result[0].LastAmount, "last amount should be updated")
		assert.Equal(t, map[int64]int{1000: 3}, result[0].Amounts, "amounts should be updated")
		assert.True(t, result[0].Ended, "ended should be updated")
	})

	t.Run("each direction is separate", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		})
		require.NoError(t, err, "must be able to create debit recurring transaction")

		err = repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.CreditDirection, -500),
		})
		require.NoError(t, err, "must be able to create credit recurring transaction")

		result, err := repo.GetTransactionRecurringByCluster(t.Context(), bankAccount.BankAccountId, cluster.TransactionClusterId)
		require.NoError(t, err, "must be able to read recurring transactions")
		assert.Len(t, result, 2, "there should be one recurring transaction per direction")
	})
}

func TestRepositoryBase_DeleteTransactionRecurring(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		items := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
			newTransactionRecurring(t, cluster, models.CreditDirection, -500),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items)
		require.NoError(t, err, "must be able to create recurring transactions")

		err = repo.DeleteTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.ID[models.TransactionRecurring]{
			items[0].TransactionRecurringId,
		})
		assert.NoError(t, err, "must be able to delete recurring transaction")

		testutils.MustDBNotExist(t, items[0])
		testutils.MustDBExist(t, items[1])
	})

	t.Run("nothing to delete", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		err := repo.DeleteTransactionRecurring(t.Context(), bankAccount.BankAccountId, nil)
		assert.NoError(t, err, "deleting nothing should not fail")
	})
}
