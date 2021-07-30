package models

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestFundingSchedule_CalculateNextOccurrence(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		rule, err := NewRule("FREQ=DAILY")
		require.NoError(t, err, "must be able to create a rule")

		originalOccurrence := time.Now().Add(-1 * time.Minute)

		fundingSchedule := FundingSchedule{
			AccountId:      1234,
			BankAccountId:  1234,
			Name:           "Testing #123",
			Description:    t.Name(),
			Rule:           rule,
			LastOccurrence: nil,
			NextOccurrence: originalOccurrence,
		}

		assert.Nil(t, fundingSchedule.LastOccurrence, "last occurrence should still be nil")

		ok := fundingSchedule.CalculateNextOccurrence(context.Background(), time.Local)
		assert.True(t, ok, "should calculate next occurrence")
		assert.NotNil(t, fundingSchedule.LastOccurrence, "last occurrence should no longer be nil")
		assert.Equal(t, originalOccurrence.Unix(), fundingSchedule.LastOccurrence.Unix(), "last occurrence should match original")
		assert.Greater(t, fundingSchedule.NextOccurrence.Unix(), originalOccurrence.Unix(), "next occurrence should be in the future relative to the last occurrence")
	})

	t.Run("would skip", func(t *testing.T) {
		rule, err := NewRule("FREQ=DAILY")
		require.NoError(t, err, "must be able to create a rule")

		originalOccurrence := time.Now().Add(1 * time.Minute)

		fundingSchedule := FundingSchedule{
			AccountId:      1234,
			BankAccountId:  1234,
			Name:           "Testing #123",
			Description:    t.Name(),
			Rule:           rule,
			LastOccurrence: nil,
			NextOccurrence: originalOccurrence,
		}

		assert.Nil(t, fundingSchedule.LastOccurrence, "last occurrence should still be nil")

		ok := fundingSchedule.CalculateNextOccurrence(context.Background(), time.Local)
		assert.False(t, ok, "next occurrence should not be calculated")
		assert.Equal(t, originalOccurrence.Unix(), fundingSchedule.NextOccurrence.Unix(), "next occurrence should not have changed")
	})
}
