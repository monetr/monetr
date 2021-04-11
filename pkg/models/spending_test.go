package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
