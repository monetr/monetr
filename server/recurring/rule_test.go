package recurring

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFundedInTime will check that every transaction from the start of the
// rule onward has an occurrence on the same day or up to maxDays before it. So
// something funded by the rule would never be funded late.
func assertFundedInTime(
	t *testing.T,
	ruleset *models.RuleSet,
	timezone *time.Location,
	transactions []models.Transaction,
	maxDays int,
) {
	// Same as funding schedules, the rule needs to be in the account's timezone
	// so it stays on midnight across daylight savings time.
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

// inTimezone will return a copy of the ruleset with its DTSTART in the
// specified timezone, the same way funding schedules evaluate their rules.
func inTimezone(ruleset *models.RuleSet, timezone *time.Location) *models.RuleSet {
	rule := ruleset.Clone()
	rule.DTStart(rule.GetDTStart().In(timezone))
	return rule
}

func TestGenerateRuleSet(t *testing.T) {
	central, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err, "must be able to load the timezone")

	t.Run("monthly past weekends", func(t *testing.T) {
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

	t.Run("first of the month", func(t *testing.T) {
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

	t.Run("capped at the 28th", func(t *testing.T) {
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

	t.Run("twice a month", func(t *testing.T) {
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

	t.Run("twice a month other days", func(t *testing.T) {
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

	t.Run("twice a month first and fifteenth", func(t *testing.T) {
		// Both days are in the first half of the month, this used to leave the second
		// half empty and generate BYMONTHDAY=1,0 which rrule rejects.
		// See: https://monetr.sentry.io/issues/7780473773/
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        -500000,
				Date:          time.Date(2026, 4, 1, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_7",
				Amount:        -500000,
				Date:          time.Date(2026, 4, 15, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 1st and the 15th of the month")
	})

	t.Run("twice a month full year", func(t *testing.T) {
		// Payroll on the 15th and the last day of the month for all of 2025, on the
		// business day before when either one lands on a weekend. The 15th moves to
		// the 13th or 14th but those days shouldn't change where the month is split.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2025, 1, 31, 0, 0, 0, 0, central),
			},
			// 15th was a Saturday
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2025, 2, 14, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -500000,
				Date:          time.Date(2025, 2, 28, 0, 0, 0, 0, central),
			},
			// 15th was a Saturday
			{
				TransactionId: "txn_4",
				Amount:        -500000,
				Date:          time.Date(2025, 3, 14, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_5",
				Amount:        -500000,
				Date:          time.Date(2025, 3, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        -500000,
				Date:          time.Date(2025, 4, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_7",
				Amount:        -500000,
				Date:          time.Date(2025, 4, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_8",
				Amount:        -500000,
				Date:          time.Date(2025, 5, 15, 0, 0, 0, 0, central),
			},
			// 31st was a Saturday
			{
				TransactionId: "txn_9",
				Amount:        -500000,
				Date:          time.Date(2025, 5, 30, 0, 0, 0, 0, central),
			},
			// 15th was a Sunday
			{
				TransactionId: "txn_10",
				Amount:        -500000,
				Date:          time.Date(2025, 6, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_11",
				Amount:        -500000,
				Date:          time.Date(2025, 6, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_12",
				Amount:        -500000,
				Date:          time.Date(2025, 7, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_13",
				Amount:        -500000,
				Date:          time.Date(2025, 7, 31, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_14",
				Amount:        -500000,
				Date:          time.Date(2025, 8, 15, 0, 0, 0, 0, central),
			},
			// 31st was a Sunday
			{
				TransactionId: "txn_15",
				Amount:        -500000,
				Date:          time.Date(2025, 8, 29, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_16",
				Amount:        -500000,
				Date:          time.Date(2025, 9, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_17",
				Amount:        -500000,
				Date:          time.Date(2025, 9, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_18",
				Amount:        -500000,
				Date:          time.Date(2025, 10, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_19",
				Amount:        -500000,
				Date:          time.Date(2025, 10, 31, 0, 0, 0, 0, central),
			},
			// 15th was a Saturday
			{
				TransactionId: "txn_20",
				Amount:        -500000,
				Date:          time.Date(2025, 11, 14, 0, 0, 0, 0, central),
			},
			// 30th was a Sunday
			{
				TransactionId: "txn_21",
				Amount:        -500000,
				Date:          time.Date(2025, 11, 28, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_22",
				Amount:        -500000,
				Date:          time.Date(2025, 12, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_23",
				Amount:        -500000,
				Date:          time.Date(2025, 12, 31, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 15th and the last day of the month")
	})

	t.Run("twice a month mostly the 14th", func(t *testing.T) {
		// Made up so the shifted day is the most common one, the rule should land on
		// it instead of the 15th.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 14, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_4",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 14, 0, 0, 0, 0, central),
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
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=14,-1", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 14th and the last day of the month")
	})

	t.Run("twice a month mostly the 13th", func(t *testing.T) {
		// Same as above but with the 13th.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 13, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_3",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, central),
			},
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
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		require.NoError(t, err, "must be able to generate a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=13,-1", ruleset.GetRRule().OrigOptions.RRuleString(), "should be the 13th and the last day of the month")
	})

	t.Run("twice a month single day", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        -500000,
				Date:          time.Date(2026, 1, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500000,
				Date:          time.Date(2026, 2, 15, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_2",
				Amount:        -500000,
				Date:          time.Date(2026, 3, 15, 0, 0, 0, 0, central),
			},
		}

		ruleset, err := GenerateRuleSet(15, transactions, central)
		assert.EqualError(t, err, "cannot generate a twice a month ruleset from a single day of the month", "should not generate a twice a month rule from one day")
		assert.Nil(t, ruleset, "should not return a ruleset")
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

	t.Run("every other week", func(t *testing.T) {
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

	t.Run("quarterly", func(t *testing.T) {
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

	t.Run("dtstart stored as utc", func(t *testing.T) {
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
	t.Run("smaller day", func(t *testing.T) {
		assert.True(t, earlierDay(14, 15), "smaller day should be earlier")
	})

	t.Run("larger day", func(t *testing.T) {
		assert.False(t, earlierDay(16, 15), "larger day should not be earlier")
	})

	t.Run("same day", func(t *testing.T) {
		assert.False(t, earlierDay(15, 15), "same day should not be earlier")
	})

	t.Run("before the last day", func(t *testing.T) {
		assert.True(t, earlierDay(31, -1), "any day should be earlier than the last day")
	})

	t.Run("last day", func(t *testing.T) {
		assert.False(t, earlierDay(-1, 28), "the last day should never be earlier")
	})

	t.Run("last day twice", func(t *testing.T) {
		assert.False(t, earlierDay(-1, -1), "the last day should not be earlier than itself")
	})
}

func TestMostCommonDay(t *testing.T) {
	t.Run("highest count wins", func(t *testing.T) {
		assert.Equal(t, 15, mostCommonDay(map[int]int{
			14: 1,
			15: 3,
			16: 2,
		}), "should pick the day with the highest count")
	})

	t.Run("earlier day wins a tie", func(t *testing.T) {
		// Run it a few times since map iteration order is random, the result
		// must not depend on which day is seen first.
		for range 20 {
			assert.Equal(t, 14, mostCommonDay(map[int]int{
				14: 2,
				15: 2,
			}), "earlier day should win the tie")
		}
	})

	t.Run("last day loses a tie", func(t *testing.T) {
		for range 20 {
			assert.Equal(t, 30, mostCommonDay(map[int]int{
				-1: 2,
				30: 2,
			}), "last day should lose the tie")
		}
	})
}
