package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

func TestSpending_GetProgressAmount(t *testing.T) {
	t.Run("expense", func(t *testing.T) {
		expense := Spending{
			SpendingType:  SpendingTypeExpense,
			TargetAmount:  10000,
			CurrentAmount: 5000,
		}

		progress := expense.GetProgressAmount()
		assert.EqualValues(t, 5000, progress, "progress should be 5000")
	})

	t.Run("goal normal", func(t *testing.T) {
		goal := Spending{
			SpendingType:  SpendingTypeGoal,
			TargetAmount:  10000,
			CurrentAmount: 5000,
			UsedAmount:    0,
		}

		progress := goal.GetProgressAmount()
		assert.EqualValues(t, 5000, progress, "progress should be 5000")
	})

	t.Run("goal with used", func(t *testing.T) {
		goal := Spending{
			SpendingType:  SpendingTypeGoal,
			TargetAmount:  10000,
			CurrentAmount: 4000,
			UsedAmount:    1000,
		}

		progress := goal.GetProgressAmount()
		assert.EqualValues(t, 5000, progress, "progress should be 5000")
	})
}

// Same as testutils.NewRuleSet because of cyclical imports.
func RuleToSet(t *testing.T, timezone *time.Location, ruleString string, potentialNow ...time.Time) *RuleSet {
	rule, err := rrule.StrToRRule(ruleString)
	require.NoError(t, err, "must be able to parse rule string")

	var now time.Time
	if len(potentialNow) == 1 {
		now = potentialNow[0]
	} else if len(potentialNow) > 1 {
		panic("can only provide a single now")
	} else {
		now = time.Now()
	}
	rule.DTStart(now)

	after := rule.After(now, false)
	dtstart := util.Midnight(after, timezone)

	ruleSetString := fmt.Sprintf(
		"DTSTART:%s\nRRULE:%s",
		dtstart.UTC().Format("20060102T150405Z"),
		ruleString,
	)

	set, err := NewRuleSet(ruleSetString)
	require.NoError(t, err, "must be able to parse rule and start into ruleset: %s", ruleSetString)

	return set
}

func GiveMeAFundingSchedule(nextContributionDate time.Time, ruleset *RuleSet) *FundingSchedule {
	return &FundingSchedule{
		FundingScheduleId: 12345,
		Name:              "Bogus Funding Schedule",
		Description:       "Bogus",
		RuleSet:           ruleset,
		NextOccurrence:    nextContributionDate,
	}
}

func TestSpending_CalculateNextContribution(t *testing.T) {
	t.Run("next funding in the past updated", func(t *testing.T) {
		today := util.Midnight(time.Now(), time.UTC)
		dayAfterTomorrow := util.Midnight(today.Add(48*time.Hour), time.UTC)
		dayAfterDayAfterTomorrow := util.Midnight(time.Now().Add(72*time.Hour), time.UTC)
		assert.True(t, dayAfterDayAfterTomorrow.After(today), "dayAfterDayAfterTomorrow timestamp must come after today's")

		ruleset := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", today)
		spending := Spending{
			SpendingType:   SpendingTypeGoal,
			TargetAmount:   100,
			CurrentAmount:  0,
			NextRecurrence: dayAfterDayAfterTomorrow,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(dayAfterTomorrow, ruleset),
			time.Now(),
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind because it will be funded before it is spent")
		assert.EqualValues(t, spending.TargetAmount, spending.NextContributionAmount, "next contribution should be the entire amount")
	})

	// This might eventually become obsolete, but it covers a bug scenario I discovered while working on institutions.

	t.Run("next funding in the past is behind", func(t *testing.T) {
		today := util.Midnight(time.Now(), time.UTC)
		dayAfterTomorrow := util.Midnight(time.Now().Add(48*time.Hour), time.UTC)
		assert.True(t, dayAfterTomorrow.After(today), "dayAfterTomorrow timestamp must come after today's")
		ruleset := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", today)

		spending := Spending{
			SpendingType:   SpendingTypeGoal,
			TargetAmount:   100,
			CurrentAmount:  0,
			NextRecurrence: dayAfterTomorrow,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(dayAfterTomorrow.Add(25*time.Hour), ruleset),
			time.Now(),
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.EqualValues(t, spending.TargetAmount, spending.NextContributionAmount, "next contribution should be the entire amount")
		assert.True(t, spending.IsBehind, "spending should be behind since it will not be funded until after it is needed")
	})

	t.Run("timezone near midnight", func(t *testing.T) {
		// This test is here to prove a fix for https://github.com/monetr/monetr/issues/937
		// Basically, we want to make sure that if the user is close to their funding schedule such that the server is
		// already on the next day; that we still calculate everything correctly. This was not happening, this test
		// accompanies a fix to remedy that situation.
		central, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		now := time.Date(2022, 06, 14, 22, 37, 43, 0, central)
		nextFunding := time.Date(2022, 06, 15, 0, 0, 0, 0, central)
		nextRecurrence := time.Date(2022, 7, 8, 0, 0, 0, 0, central)

		fundingRule := RuleToSet(t, central, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)
		spendingRule := RuleToSet(t, central, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   25000,
			CurrentAmount:  6960,
			RuleSet:        spendingRule,
			NextRecurrence: nextRecurrence.UTC(),
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			central.String(),
			GiveMeAFundingSchedule(nextFunding.UTC(), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.EqualValues(t, 9020, spending.NextContributionAmount, "next contribution amount should be half of the total needed to reach the target")
	})

	t.Run("spend from a day early falls behind", func(t *testing.T) {
		// This test is a placeholder for now. Recurring expenses are not always consistent about the day they come in.
		// If we try to process an expense early then it can sometimes misrepresent that expense for the next cycle. In
		// this example, the expense was processed on April 9th, whereas it is due on the 10th. Because the expense is
		// funded on the 15th and last day of each month, the contribution code believes that this expense has fallen
		// behind, when in fact it has not. This test simply proves this behavior for now, until I find a way to fix it.
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 10, 0, 0, 0, 0, time.UTC)

		// This is spent every month on the 10th.
		spendingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10", now)

		// Contribute to the spending object on the 15th and last day of every month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   1500,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.True(t, spending.IsBehind, "should be behind")
		assert.EqualValues(t, 1500, spending.NextContributionAmount, "should try to contribute the entire amount")
	})

	t.Run("spent the day after properly recalculates", func(t *testing.T) {
		// This test is similar to the one above, but makes sure that if the spending is calculated the day after it is
		// due, that it will recalculate properly.
		now := time.Date(2022, 4, 11, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 10, 0, 0, 0, 0, time.UTC)
		subsequentDueDate := time.Date(2022, 5, 10, 0, 0, 0, 0, time.UTC)

		// This is spent every month on the 10th.
		spendingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10", now)

		// Contribute to the spending object on the 15th and last day of every month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   1500,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.Equal(t, subsequentDueDate, spending.NextRecurrence, "subsequent due date should be a month in the future")
		assert.EqualValues(t, 750, spending.NextContributionAmount, "next contribution should be half")
	})

	t.Run("more frequent than funding", func(t *testing.T) {
		now := time.Date(2022, 4, 14, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 15, 0, 0, 0, 0, time.UTC)

		// We need to spend this expense every Friday.
		spendingRule := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", now)

		// Contribute to the spending object on the 15th and last day of every month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   1500,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 3000, spending.NextContributionAmount, "next contribution amount should be more than the target to account for frequency")
	})

	t.Run("more frequent, odd scenario", func(t *testing.T) {
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 11, 0, 0, 0, 0, time.UTC)

		// We need to spend this expense every Friday.
		spendingRule := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", now)

		// Contribute to the spending object on the 15th and last day of every month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  5000, // We have enough to cover the 11th, but not subsequent ones.
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}
		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 10000, spending.NextContributionAmount, "we should allocate the target * 2 to cover the next spending period for the expense")

		// What if we spend it on the 12th?
		now = time.Date(2022, 4, 12, 12, 0, 0, 0, time.UTC)
		// This will make it evaluate how much it needs to allocate for the next two instances of the expense.
		spending.CurrentAmount = 0
		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 10000, spending.NextContributionAmount, "we will need to contribute twice the target on the 15th in order to fulfill the expense")
		expectedNextRecurrence := time.Date(2022, 4, 18, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextRecurrence, spending.NextRecurrence, "should then be needed next on the 18th")
	})

	t.Run("more frequent over-allocated", func(t *testing.T) {
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 11, 0, 0, 0, 0, time.UTC)

		// We need to spend this expense every Friday.
		spendingRule := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", now)

		// Contribute to the spending object on the 15th and last day of every month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  16000, // Allocate enough for this funding period and the next one.
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}
		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 0, spending.NextContributionAmount, "because we are over-allocated we should not need to fund this spending at all this next period")
	})

	t.Run("goal over allocated", func(t *testing.T) {
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 11, 0, 0, 0, 0, time.UTC)

		// But we can only contribute to the expense twice a month.
		fundingRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeGoal,
			TargetAmount:   5000,
			CurrentAmount:  16000, // Allocate enough for this funding period and the next one.
			NextRecurrence: nextDueDate,
		}
		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingRule.After(now, false), fundingRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 0, spending.NextContributionAmount, "because we are beyond our target amount for the goal there is nothing more to contribute")
	})

	t.Run("dont spend more frequent expense before funding", func(t *testing.T) {
		// This tests what happens if we need $50 every monday, but one monday we don't spend it. What happens on the
		// next funding.
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, time.UTC)
		nextDueDate := time.Date(2022, 4, 11, 0, 0, 0, 0, time.UTC)

		spendingRule := RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", now)
		contributionRule := RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  5000, // We have enough to cover the 11th, but not subsequent ones.
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err := spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(contributionRule.After(now, false), contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 10000, spending.NextContributionAmount, "we will need to allocate the target * 2 for the 2 spending events in the second funding period")

		// Now if the 15th (payday) comes and we still have not spent this expense. We need to calculate how much more
		// we need.
		now = time.Date(2022, 4, 15, 0, 0, 0, 0, time.UTC)
		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(contributionRule.After(now, false), contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		// Because we are calculating the next contribution, on the day of that contribution it thinks that a
		// contribution cannot be made before the next X recurrences. Because of this the expense has fallen behind.
		assert.True(t, spending.IsBehind, "should be behind because this funding period needs $100 but only $50 is allocated to the spending object")
		assert.EqualValues(t, 10000, spending.NextContributionAmount, "we still need to allocate 2x the target for the second funding period")
		expectedNextRecurrence := time.Date(2022, 4, 18, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextRecurrence, spending.NextRecurrence, "should then be needed next on the 18th")
	})

	t.Run("generic monthly expense - central time", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		// Now is before the pay day, so there should be two contributions.
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, location)
		nextDueDate := time.Date(2022, 5, 2, 0, 0, 0, 0, location)

		spendingRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=2", now)
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(contributionRule.After(now, false), contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 2500, spending.NextContributionAmount, "give that there will be 2 contributions, half of the target amount should be allocated on the next contribution")
	})

	t.Run("generic monthly expense stale", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		// Now is before the pay day, so there should be two contributions.
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, location)
		nextDueDate := time.Date(2022, 4, 2, 0, 0, 0, 0, location)

		spendingRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=2", now)
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  100, // We have enough to cover the 11th, but not subsequent ones.
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(contributionRule.After(now, false), contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 2450, spending.NextContributionAmount, "little bit less than half per contribution")
	})

	t.Run("yearly generic", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		// Now is before the pay day, so there should be two contributions.
		now := time.Date(2022, 4, 9, 12, 0, 0, 0, location)
		nextDueDate := time.Date(2023, 1, 1, 0, 0, 0, 0, location)

		spendingRule := RuleToSet(t, location, "FREQ=YEARLY;INTERVAL=1;BYMONTH=2;BYMONTHDAY=1", now)
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   20000,
			CurrentAmount:  1454,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(contributionRule.After(now, false), contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 1030, spending.NextContributionAmount, "little bit less than half per contribution")
	})

	t.Run("contribution rule is today", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		// 3 minutes after the funding schedule is set to contribute.
		now := time.Date(2022, 4, 1, 0, 3, 0, 0, location)
		fundingDate := time.Date(2022, 4, 1, 0, 0, 0, 0, location)
		nextDueDate := time.Date(2023, 1, 1, 0, 0, 0, 0, location)

		spendingRule := RuleToSet(t, location, "FREQ=YEARLY;INTERVAL=1;BYMONTH=2;BYMONTHDAY=1", now)
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   20000,
			CurrentAmount:  1454,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			time.UTC.String(),
			GiveMeAFundingSchedule(fundingDate, contributionRule),
			now,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.False(t, spending.IsBehind, "should not be behind")
		assert.EqualValues(t, 1030, spending.NextContributionAmount, "little bit less than half per contribution")
	})

	t.Run("expense recurs on funding date", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		now := time.Date(2022, 9, 27, 14, 32, 0, 0, location)
		nextFundingDate := time.Date(2022, 9, 30, 0, 0, 0, 0, location)
		nextDueDate := time.Date(2022, 10, 15, 0, 0, 0, 0, location)

		spendingRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", now)
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   25000,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
			RuleSet:        spendingRule,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			location.String(),
			GiveMeAFundingSchedule(nextFundingDate, contributionRule),
			now,
		)
		assert.EqualValues(t, 12500, spending.NextContributionAmount, "should be half of the target amount")
	})

	t.Run("goal is due on a funding date", func(t *testing.T) {
		location, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		now := time.Date(2022, 1, 1, 14, 0, 0, 0, location)
		nextFundingDate := time.Date(2022, 1, 15, 0, 0, 0, 0, location)
		nextDueDate := time.Date(2022, 12, 31, 0, 0, 0, 0, location)

		// But we can only contribute to the goal twice a month.
		contributionRule := RuleToSet(t, location, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		spending := Spending{
			SpendingType:   SpendingTypeGoal,
			TargetAmount:   12000,
			CurrentAmount:  0,
			NextRecurrence: nextDueDate,
		}

		err = spending.CalculateNextContribution(
			context.Background(),
			location.String(),
			GiveMeAFundingSchedule(nextFundingDate, contributionRule),
			now,
		)
		assert.EqualValues(t, 500, spending.NextContributionAmount, "should be 12000/24")
	})
}
