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

// givenIHaveAClusterWithTransactions creates a transaction cluster with one
// transaction per amount and date pair, every transaction is a member of the
// cluster.
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

// monthlyDates returns midnight on the 8th of each month starting in January
// 2026. Transaction dates are midnight in the account's timezone.
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

// givenTheAccountTimezoneIs overrides the random timezone the fixtures give the
// account.
func givenTheAccountTimezoneIs(
	t *testing.T,
	bankAccount models.BankAccount,
	timezone string,
) {
	_, err := testutils.GetPgDatabase(t).NewUpdate().
		Model(&models.Account{}).
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

func runCalculateRecurringTransactions(
	t *testing.T,
	clock clock.Clock,
	cluster models.TransactionCluster,
) error {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := mockgen.NewMockContext(ctrl)
	context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
	context.EXPECT().Clock().Return(clock).AnyTimes()
	context.EXPECT().DB().Return(testutils.GetPgDatabase(t)).AnyTimes()
	context.EXPECT().Log().Return(testutils.GetLog(t)).AnyTimes()

	return recurring_jobs.CalculateRecurringTransactions(
		mockqueue.NewMockContext(context),
		recurring_jobs.CalculateRecurringTransactionsArguments{
			AccountId:            cluster.AccountId,
			BankAccountId:        cluster.BankAccountId,
			TransactionClusterId: cluster.TransactionClusterId,
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
	t.Run("creates the recurring transaction and links its members", func(t *testing.T) {
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

		err := runCalculateRecurringTransactions(t, clock, cluster)
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

	t.Run("running again keeps the same recurring transaction", func(t *testing.T) {
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

		require.NoError(t, runCalculateRecurringTransactions(t, clock, cluster), "first run must succeed")
		first := readRecurringByCluster(t, clock, cluster)
		require.Len(t, first, 1, "should have a single recurring transaction")

		clock.Add(24 * time.Hour)
		require.NoError(t, runCalculateRecurringTransactions(t, clock, cluster), "second run must succeed")
		second := readRecurringByCluster(t, clock, cluster)
		require.Len(t, second, 1, "should still have a single recurring transaction")
		assert.Equal(t, first[0].TransactionRecurringId, second[0].TransactionRecurringId, "should keep the same ID")
		assert.True(t, second[0].UpdatedAt.After(first[0].UpdatedAt), "should have been updated")
	})

	t.Run("removes a direction that no longer recurs", func(t *testing.T) {
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
			Model(&models.Transaction{}).
			Set(`"transaction_recurring_id" = ?`, stale.TransactionRecurringId).
			Where(`"transaction"."transaction_id" = ?`, refund.TransactionId).
			Exec(t.Context())
		require.NoError(t, err, "must be able to link the refund to the stale recurring transaction")

		err = runCalculateRecurringTransactions(t, clock, cluster)
		require.NoError(t, err, "must be able to calculate recurring transactions")

		recurring := readRecurringByCluster(t, clock, cluster)
		require.Len(t, recurring, 1, "only the debits should still recur")
		assert.Equal(t, models.DebitDirection, recurring[0].Direction, "should be the debits")
		assert.NotEqual(t, stale.TransactionRecurringId, recurring[0].TransactionRecurringId, "the stale credit should be gone")

		updated := readTransactions(t, transactions)
		assert.Nil(t, updated[refund.TransactionId].TransactionRecurringId, "the refund should not point at anything")
	})

	t.Run("falls back to UTC for an invalid timezone", func(t *testing.T) {
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

		err := runCalculateRecurringTransactions(t, clock, cluster)
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

		err := runCalculateRecurringTransactions(t, clock, cluster)
		require.NoError(t, err, "must be able to calculate an empty cluster")
		assert.Empty(t, readRecurringByCluster(t, clock, cluster), "nothing should recur")
	})
}
