package models

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/util"
	"github.com/stretchr/testify/assert"
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

func TestSpending_CalculateNextContribution(t *testing.T) {
	// This might eventually become obsolete, but it covers a bug scenario I discovered while working on institutions.
	t.Run("next funding in the past", func(t *testing.T) {
		today := util.MidnightInLocal(time.Now(), time.UTC)
		tomorrow := util.MidnightInLocal(time.Now().Add(25*time.Hour), time.UTC)
		assert.True(t, tomorrow.After(today), "tomorrow timestamp must come after today's")

		spending := Spending{
			SpendingType:   SpendingTypeExpense,
			TargetAmount:   100,
			CurrentAmount:  0,
			NextRecurrence: tomorrow,
		}

		rule, err := NewRule("FREQ=WEEKLY;INTERVAL=2;BYDAY=FR") // Every other friday
		assert.NoError(t, err, "must be able to parse the rrule")

		err = spending.CalculateNextContribution(context.Background(), time.UTC.String(), today, rule)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.EqualValues(t, spending.TargetAmount, spending.NextContributionAmount, "next contribution should be the entire amount")
	})
}
