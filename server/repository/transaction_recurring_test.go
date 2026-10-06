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

func givenIHaveARecurringTransaction(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	transactionRecurringId *models.ID[models.TransactionRecurring],
	date time.Time,
) models.Transaction {
	return testutils.MustInsert(t, models.Transaction{
		AccountId:              bankAccount.AccountId,
		BankAccountId:          bankAccount.BankAccountId,
		TransactionRecurringId: transactionRecurringId,
		Amount:                 800,
		Date:                   date,
		Name:                   "Github",
		OriginalName:           "Github",
		Source:                 models.TransactionSourceUpload,
		CreatedAt:              clock.Now(),
	})
}

func givenIHaveAnExpense(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	fundingSchedule *models.FundingSchedule,
	name string,
) models.Spending {
	spendingRule := testutils.Must(t, models.NewRuleSet, "DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15")
	return testutils.MustInsert(t, models.Spending{
		AccountId:         bankAccount.AccountId,
		BankAccountId:     bankAccount.BankAccountId,
		FundingScheduleId: fundingSchedule.FundingScheduleId,
		SpendingType:      models.SpendingTypeExpense,
		Name:              name,
		TargetAmount:      800,
		RuleSet:           spendingRule,
		NextRecurrence:    spendingRule.After(clock.Now(), false),
		CreatedAt:         clock.Now(),
	})
}

func TestRepositoryBase_GetTransactionRecurringById(t *testing.T) {
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

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")

		result, err := repo.GetTransactionRecurringById(t.Context(), bankAccount.BankAccountId, recurring[0].TransactionRecurringId)
		assert.NoError(t, err, "must be able to read the recurring transaction")
		require.NotNil(t, result, "result must not be nil")
		assert.Equal(t, recurring[0].TransactionRecurringId, result.TransactionRecurringId, "ID should match")
		assert.Equal(t, cluster.TransactionClusterId, result.TransactionClusterId, "cluster should match")
		assert.Equal(t, models.DebitDirection, result.Direction, "direction should match")
		assert.EqualValues(t, 800, result.LastAmount, "last amount should match")
	})

	t.Run("includes the linked spending", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")

		result, err := repo.GetTransactionRecurringById(t.Context(), bankAccount.BankAccountId, recurring[0].TransactionRecurringId)
		assert.NoError(t, err, "must be able to read the recurring transaction")
		require.NotNil(t, result, "result must not be nil")
		assert.Nil(t, result.Spending, "spending should be nil when nothing is linked yet")

		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")
		recurring[0].SpendingId = &spending.SpendingId
		err = repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &recurring[0])
		require.NoError(t, err, "must be able to link the spending")

		result, err = repo.GetTransactionRecurringById(t.Context(), bankAccount.BankAccountId, recurring[0].TransactionRecurringId)
		assert.NoError(t, err, "must be able to read the recurring transaction")
		require.NotNil(t, result, "result must not be nil")
		require.NotNil(t, result.Spending, "spending should be included now that one is linked")
		assert.Equal(t, spending.SpendingId, result.Spending.SpendingId, "should be the linked spending")
	})

	t.Run("does not exist", func(t *testing.T) {
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

		result, err := repo.GetTransactionRecurringById(t.Context(), bankAccount.BankAccountId, models.NewID[models.TransactionRecurring]())
		assert.Error(t, err, "should fail to read a recurring transaction that does not exist")
		assert.Nil(t, result, "result should be nil")
	})

	t.Run("wrong bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		otherBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.SavingsBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")

		result, err := repo.GetTransactionRecurringById(t.Context(), otherBankAccount.BankAccountId, recurring[0].TransactionRecurringId)
		assert.Error(t, err, "should not find the recurring transaction under another bank account")
		assert.Nil(t, result, "result should be nil")
	})

	t.Run("cannot read another account's recurring transaction", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")

		otherRepo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		result, err := otherRepo.GetTransactionRecurringById(t.Context(), bankAccount.BankAccountId, recurring[0].TransactionRecurringId)
		assert.Error(t, err, "should not be able to read another account's recurring transaction")
		assert.Nil(t, result, "result should be nil")
	})
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

func TestRepositoryBase_UpdateTransactionRecurring(t *testing.T) {
	t.Run("sets and clears the links", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

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

		before := testutils.MustDBRead(t, items[0])
		clock.Add(time.Hour)

		items[0].SpendingId = &spending.SpendingId
		err = repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0])
		require.NoError(t, err, "must be able to link the spending")
		items[1].FundingScheduleId = &fundingSchedule.FundingScheduleId
		err = repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[1])
		require.NoError(t, err, "must be able to link the funding schedule")

		debit := testutils.MustDBRead(t, items[0])
		require.NotNil(t, debit.SpendingId, "spending should be linked")
		assert.Equal(t, spending.SpendingId, *debit.SpendingId, "should be linked to the spending")
		assert.Nil(t, debit.FundingScheduleId, "debit should not have a funding schedule")
		assert.True(t, debit.UpdatedAt.After(before.UpdatedAt), "updated at should be bumped")
		credit := testutils.MustDBRead(t, items[1])
		require.NotNil(t, credit.FundingScheduleId, "funding schedule should be linked")
		assert.Equal(t, fundingSchedule.FundingScheduleId, *credit.FundingScheduleId, "should be linked to the funding schedule")
		assert.Nil(t, credit.SpendingId, "credit should not have a spending")

		items[0].SpendingId = nil
		err = repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0])
		require.NoError(t, err, "must be able to clear the spending")
		debit = testutils.MustDBRead(t, items[0])
		assert.Nil(t, debit.SpendingId, "spending should be cleared")
	})

	t.Run("recalculation keeps the links", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

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
		items[0].SpendingId = &spending.SpendingId
		require.NoError(t, repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0]), "must be able to link the spending")
		items[1].FundingScheduleId = &fundingSchedule.FundingScheduleId
		require.NoError(t, repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[1]), "must be able to link the funding schedule")

		// The recalculation job builds fresh recurring transactions that know
		// nothing about the links, upserting them must not wipe the links out.
		err = repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 1000),
			newTransactionRecurring(t, cluster, models.CreditDirection, -600),
		})
		require.NoError(t, err, "must be able to recalculate recurring transactions")

		debit := testutils.MustDBRead(t, items[0])
		assert.EqualValues(t, 1000, debit.LastAmount, "recalculation should have updated the debit")
		require.NotNil(t, debit.SpendingId, "spending link should survive recalculation")
		assert.Equal(t, spending.SpendingId, *debit.SpendingId, "should still be linked to the spending")
		credit := testutils.MustDBRead(t, items[1])
		assert.EqualValues(t, -600, credit.LastAmount, "recalculation should have updated the credit")
		require.NotNil(t, credit.FundingScheduleId, "funding schedule link should survive recalculation")
		assert.Equal(t, fundingSchedule.FundingScheduleId, *credit.FundingScheduleId, "should still be linked to the funding schedule")
	})

	t.Run("deleting the spending clears the link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

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
		require.NoError(t, repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items), "must be able to create recurring transaction")
		items[0].SpendingId = &spending.SpendingId
		require.NoError(t, repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0]), "must be able to link the spending")

		require.NoError(t, repo.DeleteSpending(t.Context(), bankAccount.BankAccountId, spending.SpendingId), "must be able to delete the spending")

		debit := testutils.MustDBRead(t, items[0])
		assert.Nil(t, debit.SpendingId, "link should be cleared once the spending is gone")
	})

	t.Run("database enforces direction", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

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
		require.NoError(t, repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items), "must be able to create recurring transactions")

		items[0].FundingScheduleId = &fundingSchedule.FundingScheduleId
		err := repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0])
		assert.Error(t, err, "a debit must not be linked to a funding schedule")

		items[1].SpendingId = &spending.SpendingId
		err = repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[1])
		assert.Error(t, err, "a credit must not be linked to a spending")
	})

	t.Run("a spending can only be linked once", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		otherCluster := givenIHaveATransactionCluster(t, bankAccount)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		items := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
			newTransactionRecurring(t, otherCluster, models.DebitDirection, 800),
		}
		require.NoError(t, repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items), "must be able to create recurring transactions")

		items[0].SpendingId = &spending.SpendingId
		require.NoError(t, repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[0]), "must be able to link the spending")
		items[1].SpendingId = &spending.SpendingId
		err := repo.UpdateTransactionRecurring(t.Context(), bankAccount.BankAccountId, &items[1])
		assert.Error(t, err, "the same spending must not be linked to a second recurring transaction")
	})

	t.Run("cannot update another account's recurring transaction", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spending := givenIHaveAnExpense(t, clock, bankAccount, fundingSchedule, "Github")

		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		otherLink := fixtures.GivenIHaveAManualLink(t, clock, otherUser)
		otherBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &otherLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		otherCluster := givenIHaveATransactionCluster(t, otherBankAccount)

		otherRepo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)
		otherItems := []models.TransactionRecurring{
			newTransactionRecurring(t, otherCluster, models.DebitDirection, 800),
		}
		require.NoError(t, otherRepo.UpsertTransactionRecurring(t.Context(), otherBankAccount.BankAccountId, otherItems), "must be able to create the other recurring transaction")

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)
		target := otherItems[0]
		target.SpendingId = &spending.SpendingId
		for _, bankAccountId := range []models.ID[models.BankAccount]{
			otherBankAccount.BankAccountId,
			bankAccount.BankAccountId,
		} {
			err := repo.UpdateTransactionRecurring(t.Context(), bankAccountId, &target)
			assert.NoError(t, err, "updating a recurring transaction that isn't in the account should do nothing")
		}

		stored := testutils.MustDBRead(t, otherItems[0])
		assert.Nil(t, stored.SpendingId, "the other account's recurring transaction must not be touched")
	})
}

func TestRepositoryBase_UpdateTransactionRecurringIds(t *testing.T) {
	t.Run("sets and clears the recurring id", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 2)

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
		require.NoError(t, err, "must be able to create recurring transaction")

		err = repo.UpdateTransactionRecurringIds(t.Context(), bankAccount.BankAccountId, []models.Transaction{
			{
				TransactionId:          transactions[0].TransactionId,
				TransactionRecurringId: &items[0].TransactionRecurringId,
			},
			{
				TransactionId:          transactions[1].TransactionId,
				TransactionRecurringId: &items[0].TransactionRecurringId,
			},
		})
		assert.NoError(t, err, "must be able to set the recurring id")

		first := testutils.MustDBRead(t, transactions[0])
		require.NotNil(t, first.TransactionRecurringId, "recurring id should be set")
		assert.Equal(t, items[0].TransactionRecurringId, *first.TransactionRecurringId, "should point at the recurring transaction")
		assert.Equal(t, transactions[0].Name, first.Name, "no other columns should be written")
		assert.Equal(t, transactions[0].Amount, first.Amount, "no other columns should be written")

		err = repo.UpdateTransactionRecurringIds(t.Context(), bankAccount.BankAccountId, []models.Transaction{
			{
				TransactionId:          transactions[0].TransactionId,
				TransactionRecurringId: nil,
			},
		})
		assert.NoError(t, err, "must be able to clear the recurring id")

		first = testutils.MustDBRead(t, transactions[0])
		assert.Nil(t, first.TransactionRecurringId, "recurring id should be cleared")
		second := testutils.MustDBRead(t, transactions[1])
		require.NotNil(t, second.TransactionRecurringId, "the other transaction should not be changed")
		assert.Equal(t, items[0].TransactionRecurringId, *second.TransactionRecurringId, "the other transaction should still point at the recurring transaction")
	})

	t.Run("cannot update another account's transactions", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		// The account the repository is for.
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		// Another account entirely, with its own recurring transaction that its
		// transactions point at.
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		otherLink := fixtures.GivenIHaveAManualLink(t, clock, otherUser)
		otherBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &otherLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		otherCluster := givenIHaveATransactionCluster(t, otherBankAccount)
		otherTransactions := fixtures.GivenIHaveNTransactions(t, clock, otherBankAccount, 2)

		otherRepo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)
		otherItems := []models.TransactionRecurring{
			newTransactionRecurring(t, otherCluster, models.DebitDirection, 800),
		}
		err := otherRepo.UpsertTransactionRecurring(t.Context(), otherBankAccount.BankAccountId, otherItems)
		require.NoError(t, err, "must be able to create the other recurring transaction")
		err = otherRepo.UpdateTransactionRecurringIds(t.Context(), otherBankAccount.BankAccountId, []models.Transaction{
			{
				TransactionId:          otherTransactions[0].TransactionId,
				TransactionRecurringId: &otherItems[0].TransactionRecurringId,
			},
			{
				TransactionId:          otherTransactions[1].TransactionId,
				TransactionRecurringId: &otherItems[0].TransactionRecurringId,
			},
		})
		require.NoError(t, err, "must be able to set the other account's recurring ids")

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
		err = repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, items)
		require.NoError(t, err, "must be able to create recurring transaction")

		// Try to clear one of the other account's transactions, and point the other
		// at this account's recurring transaction. Both with the other account's
		// bank account and with this account's bank account.
		for _, bankAccountId := range []models.ID[models.BankAccount]{
			otherBankAccount.BankAccountId,
			bankAccount.BankAccountId,
		} {
			err = repo.UpdateTransactionRecurringIds(t.Context(), bankAccountId, []models.Transaction{
				{
					TransactionId:          otherTransactions[0].TransactionId,
					TransactionRecurringId: nil,
				},
				{
					TransactionId:          otherTransactions[1].TransactionId,
					TransactionRecurringId: &items[0].TransactionRecurringId,
				},
			})
			assert.NoError(t, err, "updating transactions that aren't in the account should do nothing")
		}

		for _, txn := range otherTransactions {
			stored := testutils.MustDBRead(t, txn)
			require.NotNil(t, stored.TransactionRecurringId, "the other account's transaction must not be cleared")
			assert.Equal(t, otherItems[0].TransactionRecurringId, *stored.TransactionRecurringId, "the other account's transaction must still point at its own recurring transaction")
		}
	})
}

func TestRepositoryBase_GetTransactionsForRecurring(t *testing.T) {
	t.Run("only returns members newest first", func(t *testing.T) {
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

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")
		recurringId := &recurring[0].TransactionRecurringId

		january := givenIHaveARecurringTransaction(t, clock, bankAccount, recurringId, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
		march := givenIHaveARecurringTransaction(t, clock, bankAccount, recurringId, time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		february := givenIHaveARecurringTransaction(t, clock, bankAccount, recurringId, time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC))
		// not part of the recurring transaction at all
		givenIHaveARecurringTransaction(t, clock, bankAccount, nil, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
		// part of it but deleted so it shouldnt come back
		deleted := givenIHaveARecurringTransaction(t, clock, bankAccount, recurringId, time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC))
		deletedAt := clock.Now()
		deleted.DeletedAt = &deletedAt
		testutils.MustDBUpdate(t, &deleted)

		result, err := repo.GetTransactionsForRecurring(t.Context(), bankAccount.BankAccountId, *recurringId, 100, 0)
		assert.NoError(t, err, "must be able to read the transactions for the recurring transaction")
		require.Len(t, result, 3, "should only have the members that are not deleted")
		assert.Equal(t, march.TransactionId, result[0].TransactionId, "newest should be first")
		assert.Equal(t, february.TransactionId, result[1].TransactionId, "then the one before it")
		assert.Equal(t, january.TransactionId, result[2].TransactionId, "oldest should be last")

		result, err = repo.GetTransactionsForRecurring(t.Context(), bankAccount.BankAccountId, *recurringId, 1, 1)
		assert.NoError(t, err, "must be able to page through the transactions")
		require.Len(t, result, 1, "should respect the limit")
		assert.Equal(t, february.TransactionId, result[0].TransactionId, "should respect the offset")
	})

	t.Run("cannot read another bank account's transactions", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		otherBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.SavingsBankAccountSubType)
		cluster := givenIHaveATransactionCluster(t, bankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		recurring := []models.TransactionRecurring{
			newTransactionRecurring(t, cluster, models.DebitDirection, 800),
		}
		err := repo.UpsertTransactionRecurring(t.Context(), bankAccount.BankAccountId, recurring)
		require.NoError(t, err, "must be able to create recurring transactions")
		givenIHaveARecurringTransaction(t, clock, bankAccount, &recurring[0].TransactionRecurringId, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))

		result, err := repo.GetTransactionsForRecurring(t.Context(), otherBankAccount.BankAccountId, recurring[0].TransactionRecurringId, 100, 0)
		assert.NoError(t, err, "should not fail just because nothing matched")
		assert.Empty(t, result, "should not return transactions from another bank account")
	})
}
