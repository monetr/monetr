package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestFundingSchedule_CalculateNextOccurrence(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=DAILY", clock.Now())

		originalOccurrence := clock.Now().Add(-1 * time.Minute)

		fundingSchedule := models.FundingSchedule{
			AccountId:      "acct_1234",
			BankAccountId:  "bac_1234",
			Name:           "Testing #123",
			Description:    t.Name(),
			RuleSet:        rule,
			LastRecurrence: nil,
			NextRecurrence: originalOccurrence,
		}

		assert.Nil(t, fundingSchedule.LastRecurrence, "last occurrence should still be nil")

		ok := fundingSchedule.CalculateNextOccurrence(context.Background(), clock.Now(), time.Local)
		assert.True(t, ok, "should calculate next occurrence")
		assert.NotNil(t, fundingSchedule.LastRecurrence, "last occurrence should no longer be nil")
		assert.Equal(t, originalOccurrence.Unix(), fundingSchedule.LastRecurrence.Unix(), "last occurrence should match original")
		assert.Greater(t, fundingSchedule.NextRecurrence.Unix(), originalOccurrence.Unix(), "next occurrence should be in the future relative to the last occurrence")
	})

	t.Run("would skip", func(t *testing.T) {
		clock := clock.NewMock()
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=DAILY", clock.Now())

		originalOccurrence := clock.Now().Add(1 * time.Minute)

		fundingSchedule := models.FundingSchedule{
			AccountId:      "acct_1234",
			BankAccountId:  "bac_1234",
			Name:           "Testing #123",
			Description:    t.Name(),
			RuleSet:        rule,
			LastRecurrence: nil,
			NextRecurrence: originalOccurrence,
		}

		assert.Nil(t, fundingSchedule.LastRecurrence, "last occurrence should still be nil")

		ok := fundingSchedule.CalculateNextOccurrence(context.Background(), clock.Now(), time.Local)
		assert.False(t, ok, "next occurrence should not be calculated")
		assert.Equal(t, originalOccurrence.Unix(), fundingSchedule.NextRecurrence.Unix(), "next occurrence should not have changed")
	})

	t.Run("calculate on blank next", func(t *testing.T) {
		clock := clock.NewMock()
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())
		rule.DTStart(rule.GetDTStart().AddDate(0, -1, 0))

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			RuleSet:         rule,
			ExcludeWeekends: false,
			LastRecurrence:  nil,
			NextRecurrence:  time.Time{},
		}
		fundingSchedule.CalculateNextOccurrence(context.Background(), clock.Now(), time.UTC)
		assert.Greater(t, fundingSchedule.NextRecurrence, clock.Now())
	})
}

func TestFundingSchedule_GetNextContributionDateAfter(t *testing.T) {
	t.Run("dont skip weekends", func(t *testing.T) {
		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			RuleSet:         rule,
			ExcludeWeekends: false,
			LastRecurrence:  nil,
			NextRecurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
		}

		expected := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
		next, _ := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should contribute next on sunday the 15th of may")
	})

	t.Run("dont fall on a weekend #1", func(t *testing.T) {
		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			RuleSet:         rule,
			ExcludeWeekends: true,
			LastRecurrence:  nil,
			NextRecurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
		}

		expected := time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC)
		next, _ := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should contribute next on the 13th, which is a friday")
	})

	t.Run("dont fall on a weekend #2", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// 12:23:10 AM on September 31st 2022.
		now := time.Date(2022, 9, 31, 0, 23, 10, 0, timezone)
		rule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			RuleSet:         rule,
			ExcludeWeekends: true,
			LastRecurrence:  nil,
			NextRecurrence:  time.Date(2022, 9, 31, 0, 0, 0, 0, timezone),
		}

		expected := time.Date(2022, 10, 14, 0, 0, 0, 0, timezone)
		next, _ := fundingSchedule.GetNextContributionDateAfter(now, timezone)
		assert.Equal(t, expected, next, "should contribute next on the 14th, which is a friday")

		fundingSchedule.ExcludeWeekends = false
		expected = time.Date(2022, 10, 15, 0, 0, 0, 0, timezone)
		next, _ = fundingSchedule.GetNextContributionDateAfter(now, timezone)
		assert.Equal(t, expected, next, "should contribute next on the 15th when we are not excluding weekends")
	})

	// This test follows the scenario where a funding schedule was performed early (on friday the 13th) when it would
	// normally have been processed on the 15th (sunday). Because of this, if we try to determine the next occurrence on
	// the saturday before its "real" funding date, we would normally get the wrong day. In a failure path we would get
	// the 15th, which is the next day. But we actually need the 31st because the 15th funding schedule was processed
	// early.
	t.Run("odd early-after calculation", func(t *testing.T) {
		now := time.Date(2022, 5, 14, 0, 0, 0, 0, time.UTC)
		rule := testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)

		fundingSchedule := models.FundingSchedule{
			Name:            "Bogus",
			RuleSet:         rule,
			ExcludeWeekends: true,
			LastRecurrence:  nil,
			NextRecurrence:  time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC),
		}

		expected := time.Date(2022, 5, 31, 0, 0, 0, 0, time.UTC)
		next, _ := fundingSchedule.GetNextContributionDateAfter(now, time.UTC)
		assert.Equal(t, expected, next, "should not show the 15th, instead should show the 31st")
	})

	t.Run("prevent regression calculating midnight", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// 1 Second after midnight in timezone on last funding day. But in UTC because that's the server timezone.
		now := time.Date(2022, 9, 15, 0, 0, 1, 0, timezone).UTC()
		rule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)
		next := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)
		expected := time.Date(2022, 9, 30, 0, 0, 0, 0, timezone)
		fundingSchedule := models.FundingSchedule{
			Name:            "Payday",
			RuleSet:         rule,
			ExcludeWeekends: false,
			LastRecurrence:  nil,
			NextRecurrence:  next,
		}
		nextFundingOccurrence, _ := fundingSchedule.GetNextContributionDateAfter(now, timezone)
		assert.Equal(t, expected, nextFundingOccurrence, "should be on friday the 30th of september next")
	})
}
