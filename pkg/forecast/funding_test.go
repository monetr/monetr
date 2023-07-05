package forecast

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFundingScheduleBase_GetNextContributionDateAfter(t *testing.T) {
	t.Run("dont skip weekends", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		next := instructions.GetNextFundingEventAfter(context.Background(), now, time.UTC)
		assert.Equal(t, expected, next.Date, "should contribute next on sunday the 15th of may")
	})

	t.Run("dont fall on a weekend #1", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 4, 30, 0, 0, 0, 0, time.UTC),
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC)
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		next := instructions.GetNextFundingEventAfter(context.Background(), now, time.UTC)
		assert.Equal(t, expected, next.Date, "should contribute next on the 13th, which is a friday")
	})

	t.Run("dont fall on a weekend #2", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")

		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 9, 31, 0, 0, 0, 0, timezone),
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}

		// 12:23:10 AM on September 31st 2022.
		now := time.Date(2022, 9, 31, 0, 23, 10, 0, timezone)
		expected := time.Date(2022, 10, 14, 0, 0, 0, 0, timezone)
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		next := instructions.GetNextFundingEventAfter(context.Background(), now, timezone)
		assert.Equal(t, expected, next.Date, "should contribute next on the 14th, which is a friday")

		fundingSchedule.ExcludeWeekends = false
		expected = time.Date(2022, 10, 15, 0, 0, 0, 0, timezone)
		instructions = NewFundingScheduleFundingInstructions(log, fundingSchedule)
		next = instructions.GetNextFundingEventAfter(context.Background(), now, timezone)
		assert.Equal(t, expected, next.Date, "should contribute next on the 15th when we are not excluding weekends")
	})

	// This test follows the scenario where a funding schedule was performed early (on friday the 13th) when it would
	// normally have been processed on the 15th (sunday). Because of this, if we try to determine the next occurrence on
	// the saturday before its "real" funding date, we would normally get the wrong day. In a failure path we would get
	// the 15th, which is the next day. But we actually need the 31st because the 15th funding schedule was processed
	// early.
	t.Run("odd early/after calculation", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 5, 13, 0, 0, 0, 0, time.UTC),
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		now := time.Date(2022, 5, 14, 0, 0, 0, 0, time.UTC)
		expected := time.Date(2022, 5, 31, 0, 0, 0, 0, time.UTC)
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		next := instructions.GetNextFundingEventAfter(context.Background(), now, time.UTC)
		assert.Equal(t, expected, next.Date, "should not show the 15th, instead should show the 31st")
	})

	t.Run("prevent regression calculating midnight", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		next := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)
		expected := time.Date(2022, 9, 30, 0, 0, 0, 0, timezone)
		// 1 Second after midnight in timezone on last funding day. But in UTC because that's the server timezone.
		now := time.Date(2022, 9, 15, 0, 0, 1, 0, timezone).UTC()
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  next,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextFundingOccurrence := instructions.GetNextFundingEventAfter(context.Background(), now, timezone)
		assert.Equal(t, expected, nextFundingOccurrence.Date, "should be on friday the 30th of september next")
	})

	t.Run("next funding is empty", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		expected := time.Date(2022, 9, 30, 0, 0, 0, 0, timezone)
		// 1 Second after midnight in timezone on last funding day. But in UTC because that's the server timezone.
		now := time.Date(2022, 9, 15, 0, 0, 1, 0, timezone).UTC()
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  time.Time{},
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextFundingOccurrence := instructions.GetNextFundingEventAfter(context.Background(), now, timezone)
		assert.Equal(t, expected, nextFundingOccurrence.Date, "should be on friday the 30th of september next")
	})

	t.Run("we are before the next funding", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		next := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)
		// 1 Second after midnight in timezone on last funding day. But in UTC because that's the server timezone.
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  next,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextFundingOccurrence := instructions.GetNextFundingEventAfter(context.Background(), now, timezone)
		assert.Equal(t, next, nextFundingOccurrence.Date, "should be on friday the 30th of september next")
	})
}

func TestFundingScheduleBase_GetNContributionDatesAfter(t *testing.T) {
	t.Run("get next 2 funding dates", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		expected := []FundingEvent{
			{
				Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
		}
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  expected[0].Date,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextN := instructions.GetNFundingEventsAfter(context.Background(), 2, now, timezone)
		assert.Equal(t, expected, nextN, "should have the next two be the 15th and the 30th of september")
	})

	t.Run("get next 4 funding dates", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		expected := []FundingEvent{
			{
				Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 10, 31, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 10, 31, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
		}
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  expected[0].Date,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextN := instructions.GetNFundingEventsAfter(context.Background(), 4, now, timezone)
		assert.Equal(t, expected, nextN)
	})

	t.Run("get next 4 funding dates excluding weekends", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		expected := []FundingEvent{
			{
				Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
			{
				Date:           time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
				WeekendAvoided: true, // Exclude weekend, moved back one day.
			},
			{
				Date:           time.Date(2022, 10, 31, 0, 0, 0, 0, timezone),
				OriginalDate:   time.Date(2022, 10, 31, 0, 0, 0, 0, timezone),
				WeekendAvoided: false,
			},
		}
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: true,
			NextOccurrence:  expected[0].Date,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}
		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		nextN := instructions.GetNFundingEventsAfter(context.Background(), 4, now, timezone)
		assert.Equal(t, expected, nextN)
	})
}

func TestFundingScheduleBase_GetNumberOfContributionsBetween(t *testing.T) {
	t.Run("september to christmas", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// Between September 13th, and December 25th
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		end := time.Date(2022, 12, 25, 0, 0, 1, 0, timezone).UTC()
		next := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  next,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}

		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		count := instructions.GetNumberOfFundingEventsBetween(context.Background(), now, end, timezone)
		assert.EqualValues(t, 7, count, "should have seven contributions")
	})

	t.Run("one year twice a month", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// Between September 13th, and December 25th
		now := time.Date(2022, 1, 2, 0, 0, 1, 0, timezone).UTC()
		end := time.Date(2023, 1, 1, 0, 0, 0, 0, timezone).UTC()
		next := time.Date(2022, 1, 15, 0, 0, 0, 0, timezone)
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  next,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}

		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		count := instructions.GetNumberOfFundingEventsBetween(context.Background(), now, end, timezone)
		assert.EqualValues(t, 24, count, "should have 24 contributions")
	})

	t.Run("one year twice a month", func(t *testing.T) {
		log := testutils.GetLog(t)
		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// Between September 13th, and December 25th
		now := time.Date(2022, 1, 2, 0, 0, 1, 0, timezone).UTC()
		end := time.Date(2023, 1, 1, 0, 0, 0, 0, timezone).UTC()
		next := time.Date(2022, 1, 15, 0, 0, 0, 0, timezone)
		fundingSchedule := models.FundingSchedule{
			Rule:            rule,
			ExcludeWeekends: false,
			NextOccurrence:  next,
			DateStarted:     time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
		}

		instructions := NewFundingScheduleFundingInstructions(log, fundingSchedule)
		count := instructions.GetNumberOfFundingEventsBetween(context.Background(), now, end, timezone)
		assert.EqualValues(t, 12, count, "should have 12 contributions")
	})
}
