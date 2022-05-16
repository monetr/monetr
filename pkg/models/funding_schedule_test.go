package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFundingSchedule_CalculateNextOccurrence(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		rule, err := models.NewRule("FREQ=DAILY")
		require.NoError(t, err, "must be able to create a rule")

		originalOccurrence := time.Now().Add(-1 * time.Minute)

		fundingSchedule := models.FundingSchedule{
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
		rule, err := models.NewRule("FREQ=DAILY")
		require.NoError(t, err, "must be able to create a rule")

		originalOccurrence := time.Now().Add(1 * time.Minute)

		fundingSchedule := models.FundingSchedule{
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

	t.Run("calculate on blank next", func(t *testing.T) {
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			Rule:            rule,
			ExcludeWeekends: false,
			LastOccurrence:  nil,
			NextOccurrence:  time.Time{},
		}
		fundingSchedule.CalculateNextOccurrence(context.Background(), time.UTC)
		assert.Greater(t, fundingSchedule.NextOccurrence, time.Now())
	})
}

func TestFundingSchedule_GetNextContributionDateAfter(t *testing.T) {
	t.Run("dont skip weekends", func(t *testing.T) {
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			Rule:            rule,
			ExcludeWeekends: false,
			LastOccurrence:  nil,
			NextOccurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
		next := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should contribute next on sunday the 15th of may")
	})

	t.Run("dont fall on a weekend", func(t *testing.T) {
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			Rule:            rule,
			ExcludeWeekends: true,
			LastOccurrence:  nil,
			NextOccurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC)
		next := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should contribute next on the 13th, which is a friday")
	})

	// This test follows the scenario where a funding schedule was performed early (on friday the 13th) when it would
	// normally have been processed on the 15th (sunday). Because of this, if we try to determine the next occurrence on
	// the saturday before its "real" funding date, we would normally get the wrong day. In a failure path we would get
	// the 15th, which is the next day. But we actually need the 31st because the 15th funding schedule was processed
	// early.
	t.Run("odd early/after calculation", func(t *testing.T) {
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			Rule:            rule,
			ExcludeWeekends: true,
			LastOccurrence:  nil,
			NextOccurrence:  time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 14, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 31, 0, 0, 0, 0, time.UTC)
		next := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should not show the 15th, instead should show the 31st")
	})
}
