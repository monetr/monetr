package recurring_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring/recurring_jobs"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"go.uber.org/mock/gomock"
)

// givenIHaveAClusterWithTransactions will create a transaction cluster with
// one transaction per amount and date pair, every transaction is a member of
// the cluster.
func givenIHaveAClusterWithTransactions(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	amounts []int64,
	dates []time.Time,
) (models.TransactionCluster, []models.Transaction) {
	require.Len(t, dates, len(amounts), "must have a date for every amount")

	transactions := make([]models.Transaction, len(amounts))
	members := make([]models.ID[models.Transaction], len(amounts))
	for i := range transactions {
		members[i] = models.NewID[models.Transaction]()
	}

	cluster := testutils.MustInsert(t, models.TransactionCluster{
		AccountId:     bankAccount.AccountId,
		BankAccountId: bankAccount.BankAccountId,
		Name:          "GITHUB INC",
		OriginalName:  "GITHUB INC",
		Members:       members,
	})

	for i := range transactions {
		transactions[i] = testutils.MustInsert(t, models.Transaction{
			TransactionId:        members[i],
			AccountId:            bankAccount.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			TransactionClusterId: &cluster.TransactionClusterId,
			Amount:               amounts[i],
			Date:                 dates[i],
			Name:                 "GITHUB INC",
			OriginalName:         "GITHUB INC",
			MerchantName:         "GITHUB INC",
			OriginalMerchantName: "GITHUB INC",
			Source:               models.TransactionSourceUpload,
			CreatedAt:            clock.Now(),
		})
	}

	return cluster, transactions
}

// monthlyDates will return midnight on the 8th of each month starting in
// January 2026. Transaction dates are midnight in the account's timezone.
func monthlyDates(timezone *time.Location, months int) []time.Time {
	result := make([]time.Time, months)
	for i := range result {
		result[i] = time.Date(2026, time.Month(i+1), 8, 0, 0, 0, 0, timezone)
	}
	return result
}

func accountTimezone(t *testing.T, bankAccount models.BankAccount) *time.Location {
	timezone, err := bankAccount.Account.GetTimezone()
	require.NoError(t, err, "must be able to get the timezone from the account")
	return timezone
}

// givenTheAccountTimezoneIs will override the random timezone the fixtures
// give the account.
func givenTheAccountTimezoneIs(
	t *testing.T,
	bankAccount models.BankAccount,
	timezone string,
) {
	_, err := testutils.GetPgDatabase(t).NewUpdate().
		Model(new(models.Account)).
		Set(`"timezone" = ?`, timezone).
		Where(`"account"."account_id" = ?`, bankAccount.AccountId).
		Exec(t.Context())
	require.NoError(t, err, "must be able to set the account's timezone")
}

func repeatAmount(amount int64, count int) []int64 {
	result := make([]int64, count)
	for i := range result {
		result[i] = amount
	}
	return result
}

// runCalculateRecurringTransactions will run the job for the whole bank
// account.
func runCalculateRecurringTransactions(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
) error {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	enqueuer := mockgen.NewMockProcessor(ctrl)
	enqueuer.EXPECT().
		EnqueueAt(
			gomock.Any(),
			mockqueue.EqQueue(recurring_jobs.MatchRecurringTransactionsToSpending),
			gomock.Any(),
			gomock.Eq(recurring_jobs.MatchRecurringTransactionsToSpendingArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
			}),
		).
		Return(nil).
		Times(1)

	context := mockgen.NewMockContext(ctrl)
	context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
	context.EXPECT().Clock().Return(clock).AnyTimes()
	context.EXPECT().DB().Return(testutils.GetPgDatabase(t)).AnyTimes()
	context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
	context.EXPECT().Log().Return(testutils.GetLog(t)).AnyTimes()

	return recurring_jobs.CalculateRecurringTransactions(
		mockqueue.NewMockContext(context),
		recurring_jobs.CalculateRecurringTransactionsArguments{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
		},
	)
}

func readRecurringByCluster(
	t *testing.T,
	clock clock.Clock,
	cluster models.TransactionCluster,
) []models.TransactionRecurring {
	repo := repository.NewRepositoryFromSession(
		clock,
		"user_system",
		cluster.AccountId,
		testutils.GetPgDatabase(t),
		testutils.GetLog(t),
	)
	result, err := repo.GetTransactionRecurringByCluster(
		t.Context(),
		cluster.BankAccountId,
		cluster.TransactionClusterId,
	)
	require.NoError(t, err, "must be able to read recurring transactions")
	return result
}

func readTransactions(
	t *testing.T,
	transactions []models.Transaction,
) map[models.ID[models.Transaction]]models.Transaction {
	ids := make([]models.ID[models.Transaction], len(transactions))
	for i := range transactions {
		ids[i] = transactions[i].TransactionId
	}

	var result []models.Transaction
	err := testutils.GetPgDatabase(t).NewSelect().
		Model(&result).
		Where(`"transaction"."transaction_id" IN (?)`, bun.List(ids)).
		Scan(t.Context())
	require.NoError(t, err, "must be able to read transactions")

	byId := make(map[models.ID[models.Transaction]]models.Transaction, len(result))
	for _, item := range result {
		byId[item.TransactionId] = item
	}
	return byId
}

func TestCalculateRecurringTransactions(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		cluster, transactions := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(800, 6),
			monthlyDates(accountTimezone(t, bankAccount), 6),
		)

		err := runCalculateRecurringTransactions(t, clock, bankAccount)
		require.NoError(t, err, "must be able to calculate recurring transactions")

		recurring := readRecurringByCluster(t, clock, cluster)
		require.Len(t, recurring, 1, "should have a single recurring transaction for the debits")
		assert.Equal(t, models.DebitDirection, recurring[0].Direction, "should be a debit")
		assert.Equal(t, models.MonthlyWindowType, recurring[0].Window, "should be monthly")
		assert.EqualValues(t, 800, recurring[0].LastAmount, "last amount should match")
		assert.False(t, recurring[0].Ended, "should not have ended")

		for _, item := range readTransactions(t, transactions) {
			require.NotNil(t, item.TransactionRecurringId, "every member should point at the recurring transaction")
			assert.Equal(t, recurring[0].TransactionRecurringId, *item.TransactionRecurringId, "should be the same recurring transaction")
		}
	})

	t.Run("running again", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		cluster, _ := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(800, 6),
			monthlyDates(accountTimezone(t, bankAccount), 6),
		)

		require.NoError(t, runCalculateRecurringTransactions(t, clock, bankAccount), "first run must succeed")
		first := readRecurringByCluster(t, clock, cluster)
		require.Len(t, first, 1, "should have a single recurring transaction")

		clock.Add(24 * time.Hour)
		require.NoError(t, runCalculateRecurringTransactions(t, clock, bankAccount), "second run must succeed")
		second := readRecurringByCluster(t, clock, cluster)
		require.Len(t, second, 1, "should still have a single recurring transaction")
		assert.Equal(t, first[0].TransactionRecurringId, second[0].TransactionRecurringId, "should keep the same ID")
		assert.True(t, second[0].UpdatedAt.After(first[0].UpdatedAt), "should have been updated")
	})

	t.Run("keeps linked expense", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
		spendingRule := testutils.Must(t, models.NewRuleSet, "DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Github",
			TargetAmount:      800,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.After(clock.Now(), false),
			CreatedAt:         clock.Now(),
		})
		cluster, _ := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(800, 6),
			monthlyDates(accountTimezone(t, bankAccount), 6),
		)

		require.NoError(t, runCalculateRecurringTransactions(t, clock, bankAccount), "first run must succeed")
		first := readRecurringByCluster(t, clock, cluster)
		require.Len(t, first, 1, "should have a single recurring transaction")
		first[0].SpendingId = &spending.SpendingId
		testutils.MustDBUpdate(t, &first[0])

		clock.Add(24 * time.Hour)
		require.NoError(t, runCalculateRecurringTransactions(t, clock, bankAccount), "second run must succeed")
		second := readRecurringByCluster(t, clock, cluster)
		require.Len(t, second, 1, "should still have a single recurring transaction")
		assert.True(t, second[0].UpdatedAt.After(first[0].UpdatedAt), "should have been recalculated")
		require.NotNil(t, second[0].SpendingId, "the linked expense must survive recalculation")
		assert.Equal(t, spending.SpendingId, *second[0].SpendingId, "should still be linked to the same expense")
	})

	t.Run("direction no longer recurs", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		// Six monthly charges and a single refund, the refund alone is not
		// enough to recur.
		timezone := accountTimezone(t, bankAccount)
		cluster, transactions := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			append(repeatAmount(800, 6), -800),
			append(monthlyDates(timezone, 6), time.Date(2026, 3, 12, 0, 0, 0, 0, timezone)),
		)
		refund := transactions[6]

		// Pretend the refunds used to recur and the refund was a member.
		stale := testutils.MustInsert(t, models.TransactionRecurring{
			AccountId:            bankAccount.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			TransactionClusterId: cluster.TransactionClusterId,
			Direction:            models.CreditDirection,
			Window:               models.MonthlyWindowType,
			RuleSet: testutils.Must(
				t,
				models.NewRuleSet,
				"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=12",
			),
			First:      time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			Last:       time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
			Next:       time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
			Confidence: 0.8,
			Amounts:    map[int64]int{-800: 3},
			LastAmount: -800,
		})
		_, err := testutils.GetPgDatabase(t).NewUpdate().
			Model(new(models.Transaction)).
			Set(`"transaction_recurring_id" = ?`, stale.TransactionRecurringId).
			Where(`"transaction"."transaction_id" = ?`, refund.TransactionId).
			Exec(t.Context())
		require.NoError(t, err, "must be able to link the refund to the stale recurring transaction")

		err = runCalculateRecurringTransactions(t, clock, bankAccount)
		require.NoError(t, err, "must be able to calculate recurring transactions")

		recurring := readRecurringByCluster(t, clock, cluster)
		require.Len(t, recurring, 1, "only the debits should still recur")
		assert.Equal(t, models.DebitDirection, recurring[0].Direction, "should be the debits")
		assert.NotEqual(t, stale.TransactionRecurringId, recurring[0].TransactionRecurringId, "the stale credit should be gone")

		updated := readTransactions(t, transactions)
		assert.Nil(t, updated[refund.TransactionId].TransactionRecurringId, "the refund should not point at anything")
	})

	t.Run("invalid timezone", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		cluster, _ := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(800, 6),
			monthlyDates(time.UTC, 6),
		)

		givenTheAccountTimezoneIs(t, bankAccount, "Not/A_Timezone")

		err := runCalculateRecurringTransactions(t, clock, bankAccount)
		require.NoError(t, err, "an invalid timezone should not fail the job")

		recurring := readRecurringByCluster(t, clock, cluster)
		assert.Len(t, recurring, 1, "should still detect the recurring transaction")
	})

	t.Run("cluster without transactions", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		cluster, _ := givenIHaveAClusterWithTransactions(t, clock, bankAccount, nil, nil)

		err := runCalculateRecurringTransactions(t, clock, bankAccount)
		require.NoError(t, err, "must be able to calculate an empty cluster")
		assert.Empty(t, readRecurringByCluster(t, clock, cluster), "nothing should recur")
	})

	t.Run("bank account without clusters", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		err := runCalculateRecurringTransactions(t, clock, bankAccount)
		assert.NoError(t, err, "nothing to calculate is not an error")
	})

	t.Run("every cluster", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		timezone := accountTimezone(t, bankAccount)
		first, _ := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(800, 6),
			monthlyDates(timezone, 6),
		)
		second, _ := givenIHaveAClusterWithTransactions(
			t,
			clock,
			bankAccount,
			repeatAmount(2600, 6),
			monthlyDates(timezone, 6),
		)

		err := runCalculateRecurringTransactions(t, clock, bankAccount)
		require.NoError(t, err, "must be able to calculate recurring transactions")

		assert.Len(t, readRecurringByCluster(t, clock, first), 1, "the first cluster should recur")
		assert.Len(t, readRecurringByCluster(t, clock, second), 1, "the second cluster should recur")
	})

	t.Run("enqueues auto assign", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		cluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)[0]
		recurring.SpendingId = &expense.SpendingId
		recurring.AutoAssign = true
		testutils.MustDBUpdate(t, &recurring)

		// A new charge shows up in the cluster, this is the only transaction that
		// is being added to the recurring transaction on the next run.
		clock.Add(8 * 24 * time.Hour)
		transaction := testutils.MustInsert(t, models.Transaction{
			TransactionId:        models.NewID[models.Transaction](),
			AccountId:            bankAccount.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			TransactionClusterId: &cluster.TransactionClusterId,
			Amount:               800,
			Date:                 time.Date(2026, 7, 8, 0, 0, 0, 0, accountTimezone(t, bankAccount)),
			Name:                 "GITHUB INC",
			OriginalName:         "GITHUB INC",
			MerchantName:         "GITHUB INC",
			OriginalMerchantName: "GITHUB INC",
			Source:               models.TransactionSourceUpload,
			CreatedAt:            clock.Now(),
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.AutoAssignRecurringTransactions),
				gomock.Any(),
				gomock.Eq(recurring_jobs.AutoAssignRecurringTransactionsArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
					TransactionIds: []models.ID[models.Transaction]{
						transaction.TransactionId,
					},
				}),
			).
			Return(nil).
			Times(1)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.MatchRecurringTransactionsToSpending),
				gomock.Any(),
				gomock.Eq(recurring_jobs.MatchRecurringTransactionsToSpendingArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(testutils.GetPgDatabase(t)).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Log().Return(testutils.GetLog(t)).AnyTimes()

		err := recurring_jobs.CalculateRecurringTransactions(
			mockqueue.NewMockContext(context),
			recurring_jobs.CalculateRecurringTransactionsArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
			},
		)
		require.NoError(t, err, "must be able to calculate recurring transactions")

		stored := readTransactions(t, []models.Transaction{transaction})[transaction.TransactionId]
		require.NotNil(t, stored.TransactionRecurringId, "new charge should be part of the recurring transaction")
		assert.Equal(t, recurring.TransactionRecurringId, *stored.TransactionRecurringId, "new charge should be in the existing recurring transaction")
		assert.Nil(t, stored.SpendingId, "calculating should not spend anything, the auto assign job does that")
	})
}
