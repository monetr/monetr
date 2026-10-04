package recurring

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFundedInTime checks that every transaction from the start of the rule
// onward has an occurrence on the same day or up to maxDays before it. So
// something funded by the rule would never be funded late.
func assertFundedInTime(
	t *testing.T,
	ruleset *models.RuleSet,
	timezone *time.Location,
	transactions []models.Transaction,
	maxDays int,
) {
	t.Helper()
	// Same as funding schedules, the rule needs to be in the account's timezone so
	// it stays on midnight across daylight savings time.
	ruleset = inTimezone(ruleset, timezone)
	for _, txn := range transactions {
		if txn.Date.Before(ruleset.GetDTStart()) {
			continue
		}
		previous := ruleset.Before(txn.Date, true)
		require.Falsef(t, previous.IsZero(), "%s must have an occurrence before it", txn.Date)
		assert.Falsef(t, previous.After(txn.Date), "%s must not be funded late, funded on %s", txn.Date, previous)
		assert.LessOrEqualf(t, txn.Date.Sub(previous).Hours()/24, float64(maxDays), "%s must not be funded more than %d days early, funded on %s", txn.Date, maxDays, previous)
	}
}

// inTimezone returns a copy of the ruleset with its DTSTART in the specified
// timezone, the same way funding schedules evaluate their rules.
func inTimezone(ruleset *models.RuleSet, timezone *time.Location) *models.RuleSet {
	rule := ruleset.Clone()
	rule.DTStart(rule.GetDTStart().In(timezone))
	return rule
}

func TestGenerateRuleSet(t *testing.T) {
	central, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err, "must be able to load the timezone")

	t.Run("monthly that slips past weekends", func(t *testing.T) {
		// Billed on the 18th, but when the 18th is on a weekend it posts on the
		// following Monday instead.
		transactions := []models.Transaction{
			// 18th was a Sunday
			{
				TransactionId: "txn_0",
				Amount:        26000,
				Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        26000,
				Date:          time.Date(2026, 2, 18, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        26000,
				Date:          time.Date(2026, 3, 18, 0, 0, 0, 0, central),
			},
			// 18th was a Saturday
			{
				TransactionId: "txn_3",
				Amount:        26000,
				Date:          time.Date(2026, 4, 20, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        26000,
				Date:          time.Date(2026, 5, 18, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        26000,
				Date:          time.Date(2026, 6, 18, 0, 0, 0, 0, central),
			},
			// 18th was a Saturday
			{
				TransactionId: "txn_6",
				Amount:        26000,
				Date:          time.Date(2026, 7, 20, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(30, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=18", ruleset.GetRRule().OrigOptions.RRuleString(), "should use the earliest day, not the day after the weekend")
		assertFundedInTime(t, ruleset, central, transactions, 3)
	})

	t.Run("first of the month that posts the business day before", func(t *testing.T) {
		// A scheduled ACH on the 1st, which posts on the previous business day when
		// the 1st is a weekend or a holiday. March 1st 2026 was a Sunday, so it
		// posted on Friday February 27th, which is the earliest it can ever post.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        4000,
				Date:          time.Date(2025, 12, 1, 0, 0, 0, 0, central),
			},
			// January 1st holiday
			{
				TransactionId: "txn_1",
				Amount:        4000,
				Date:          time.Date(2025, 12, 31, 0, 0, 0, 0, central),
			},
			// February 1st was a Sunday
			{
				TransactionId: "txn_2",
				Amount:        4000,
				Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, central),
			},
			// March 1st was a Sunday
			{
				TransactionId: "txn_3",
				Amount:        4000,
				Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        4000,
				Date:          time.Date(2026, 4, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        4000,
				Date:          time.Date(2026, 5, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        4000,
				Date:          time.Date(2026, 6, 1, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(30, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=27", ruleset.GetRRule().OrigOptions.RRuleString(), "should use the earliest the transaction has ever posted")
		assertFundedInTime(t, ruleset, central, transactions, 5)
	})

	t.Run("last day of the month is capped at the 28th", func(t *testing.T) {
		// RRULE skips any month that doesn't have the day, so anything after the
		// 28th would miss February entirely.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -109,
				Date:          time.Date(2025, 10, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -109,
				Date:          time.Date(2025, 11, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -109,
				Date:          time.Date(2025, 12, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -109,
				Date:          time.Date(2026, 1, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        -109,
				Date:          time.Date(2026, 2, 28, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        -109,
				Date:          time.Date(2026, 3, 31, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(30, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28", ruleset.GetRRule().OrigOptions.RRuleString(), "should cap the day at the 28th")
		assertFundedInTime(t, ruleset, central, transactions, 3)
	})

	t.Run("twice a month uses the most common days", func(t *testing.T) {
		// Payroll on the 15th and the last day of the month, on the business day
		// before when either one lands on a weekend.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 15, 0, 0, 0, 0, central),
			},
			// 31st was a Saturday
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, central),
			},
			// 15th was a Sunday
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 13, 0, 0, 0, 0, central),
			},
			// 28th was a Saturday
			{
				TransactionId: "txn_3",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, central),
			},
			// 15th was a Sunday
			{
				TransactionId: "txn_4",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        -500000,
				Date:          time.Date(2026, 4, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_7",
				Amount:        -500000,
				Date:          time.Date(2026, 4, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_8",
				Amount:        -500000,
				Date:          time.Date(2026, 5, 15, 0, 0, 0, 0, central),
			},
			// 31st was a Sunday
			{
				TransactionId: "txn_9",
				Amount:        -500000,
				Date:          time.Date(2026, 5, 29, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_10",
				Amount:        -500000,
				Date:          time.Date(2026, 6, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_11",
				Amount:        -500000,
				Date:          time.Date(2026, 6, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_12",
				Amount:        -500000,
				Date:          time.Date(2026, 7, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_13",
				Amount:        -500000,
				Date:          time.Date(2026, 7, 31, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 15th and the last day of the month")
	})

	t.Run("twice a month on other days", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        3092,
				Date:          time.Date(2026, 1, 9, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        3092,
				Date:          time.Date(2026, 1, 26, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        3092,
				Date:          time.Date(2026, 2, 9, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        3092,
				Date:          time.Date(2026, 2, 26, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        3092,
				Date:          time.Date(2026, 3, 9, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        3092,
				Date:          time.Date(2026, 3, 26, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        3092,
				Date:          time.Date(2026, 4, 9, 0, 0, 0, 0, central),
			},
			// 26th was a Sunday
			{
				TransactionId: "txn_7",
				Amount:        3092,
				Date:          time.Date(2026, 4, 27, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=9,26", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the most common day in each half of the month")
	})

	t.Run("weekly", func(t *testing.T) {
		// Crosses the end of daylight savings time on November 5th.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        5525,
				Date:          time.Date(2023, 10, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        5525,
				Date:          time.Date(2023, 10, 20, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        5525,
				Date:          time.Date(2023, 10, 27, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        5525,
				Date:          time.Date(2023, 11, 3, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        5525,
				Date:          time.Date(2023, 11, 10, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        5525,
				Date:          time.Date(2023, 11, 17, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(7, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", ruleset.GetRRule().OrigOptions.RRuleString(), "should be every friday")
		assertFundedInTime(t, ruleset, central, transactions, 0)
	})

	t.Run("every other week keeps its phase", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -200000,
				Date:          time.Date(2026, 1, 2, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -200000,
				Date:          time.Date(2026, 1, 16, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -200000,
				Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -200000,
				Date:          time.Date(2026, 2, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        -200000,
				Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(14, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", ruleset.GetRRule().OrigOptions.RRuleString(), "should be every other friday")

		// The next occurrence must keep landing on the same weeks as the
		// transactions, not the weeks in between.
		next := inTimezone(ruleset, central).After(time.Date(2026, 2, 27, 0, 0, 0, 0, central), false)
		assert.Equal(t, time.Date(2026, 3, 13, 0, 0, 0, 0, central), next, "next occurrence should be two weeks after the last transaction")
		assertFundedInTime(t, ruleset, central, transactions, 0)
	})

	t.Run("quarterly keeps its phase", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        14925,
				Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        14925,
				Date:          time.Date(2025, 4, 10, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        14925,
				Date:          time.Date(2025, 7, 11, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        14925,
				Date:          time.Date(2025, 10, 10, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        14925,
				Date:          time.Date(2026, 1, 12, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(90, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=10", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 10th every three months")

		next := inTimezone(ruleset, central).After(time.Date(2026, 1, 12, 0, 0, 0, 0, central), false)
		assert.Equal(t, time.Date(2026, 4, 10, 0, 0, 0, 0, central), next, "next occurrence should be in april, not february or march")
		assertFundedInTime(t, ruleset, central, transactions, 2)
	})

	t.Run("yearly", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        5985,
				Date:          time.Date(2024, 4, 23, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        5985,
				Date:          time.Date(2025, 4, 23, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        7188,
				Date:          time.Date(2026, 4, 23, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(365, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=YEARLY;INTERVAL=1;BYMONTH=4;BYMONTHDAY=23", ruleset.GetRRule().OrigOptions.RRuleString(), "should be april 23rd every year")
		assertFundedInTime(t, ruleset, central, transactions, 0)
	})

	t.Run("dtstart is midnight in the timezone stored as utc", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        1073,
				Date:          time.Date(2026, 1, 18, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        1073,
				Date:          time.Date(2026, 2, 18, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        1073,
				Date:          time.Date(2026, 3, 18, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(30, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC), ruleset.GetDTStart(), "should start at midnight central on the first of the month")
		assert.Contains(t, ruleset.String(), "DTSTART:20260101T060000Z", "should be stored in utc without a tzid")
	})

	t.Run("unsupported frequency", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        100,
				Date:          time.Date(2026, 1, 1, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(45, transactions, central)
		assert.Error(t, err, "should not generate a ruleset for an unknown frequency")
		assert.Nil(t, ruleset, "should not return a ruleset")
	})

	t.Run("no transactions", func(t *testing.T) {
		ruleset, err := GenerateRuleSet(30, nil, central)
		assert.Error(t, err, "should not generate a ruleset without transactions")
		assert.Nil(t, ruleset, "should not return a ruleset")
	})
}

func TestEarlierDay(t *testing.T) {
	cases := []struct {
		name     string
		a        int
		b        int
		expected bool
	}{
		{
			name:     "smaller day is earlier",
			a:        14,
			b:        15,
			expected: true,
		},
		{
			name:     "larger day is not earlier",
			a:        16,
			b:        15,
			expected: false,
		},
		{
			name:     "same day is not earlier",
			a:        15,
			b:        15,
			expected: false,
		},
		{
			name:     "any day is earlier than the last day",
			a:        31,
			b:        -1,
			expected: true,
		},
		{
			name:     "the last day is never earlier",
			a:        -1,
			b:        28,
			expected: false,
		},
		{
			name:     "the last day is not earlier than itself",
			a:        -1,
			b:        -1,
			expected: false,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			assert.Equal(t, item.expected, earlierDay(item.a, item.b))
		})
	}
}

func TestMostCommonDay(t *testing.T) {
	t.Run("highest count wins", func(t *testing.T) {
		assert.Equal(t, 15, mostCommonDay(map[int]int{
			14: 1,
			15: 3,
			16: 2,
		}))
	})

	t.Run("earlier day wins a tie", func(t *testing.T) {
		// Run it a few times since map iteration order is random, the result
		// must not depend on which day is seen first.
		for range 20 {
			assert.Equal(t, 14, mostCommonDay(map[int]int{
				14: 2,
				15: 2,
			}))
		}
	})

	t.Run("last day loses a tie", func(t *testing.T) {
		for range 20 {
			assert.Equal(t, 30, mostCommonDay(map[int]int{
				-1: 2,
				30: 2,
			}))
		}
	})
}
