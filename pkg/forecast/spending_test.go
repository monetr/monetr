package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestSpendingInstructionBase_GetNextSpendingEventAfter(t *testing.T) {
	t.Run("simple monthly expense", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: false,
				NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRule,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 3, now, timezone)
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 2500,
				RollingAllocation:  2500,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 2500,
				RollingAllocation:  5000,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				TransactionAmount:  5000,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
		}, events)
	})

	t.Run("spent more frequently than funded", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 14, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: false,
				NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   7500,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRule,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 7, now, timezone)
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 22500,
				RollingAllocation:  22500,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  15000,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 9, 23, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 10, 7, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
		}, events)
	})

	t.Run("spent more frequently than funded excluding weekends", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 14, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: true,
				NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   7500,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRule,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 7, now, timezone)
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 22500,
				RollingAllocation:  22500,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  15000,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 9, 23, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 10, 7, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:              time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    true,
						FundingScheduleId: 0,
					},
				},
				SpendingId: 0,
			},
			{
				Date:               time.Date(2022, 10, 21, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
				SpendingId:         0,
			},
		}, events)
	})
}

func TestSpendingInstructionBase_GetSpendingEventsBetween(t *testing.T) {
	t.Run("once a month for a year", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: true,
				NextOccurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   1395,
				CurrentAmount:  1395,
				NextRecurrence: time.Date(2022, 1, 8, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRule,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		// Should have 36 events, 12 spending events and 24 funding events.
		assert.Len(t, events, 12+24, "should have 36 events")
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
	})

	t.Run("every other week for a year", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: true,
				NextOccurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   1395,
				CurrentAmount:  1395,
				NextRecurrence: time.Date(2022, 1, 7, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRule,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.Len(t, events, 45, "should have 45 events")
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
	})

	t.Run("no spending events for paused spending objects", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:            fundingRule,
				ExcludeWeekends: true,
				NextOccurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   10000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2023, 1, 3, 0, 0, 0, 0, timezone),
				IsPaused:       true,
			},
			fundingInstructions,
		)

		events := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.Empty(t, events, "there should be no spending events for paused spending")
	})

	t.Run("spending events infinite loop bug", func(t *testing.T) {
		// This is part of: https://github.com/monetr/monetr/issues/1243
		// Make sure we don't timeout when a goal lands on the funding day.
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		// now := time.Date(2022, 11, 29, 14, 30, 1, 0, timezone).UTC()
		start := time.Date(2022, 12, 1, 0, 0, 0, 0, timezone)
		// end := time.Date(2022, 12, 2, 0, 0, 0, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				Rule:              fundingRule,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 11, 30, 0, 0, 0, 0, timezone),
				FundingScheduleId: 1,
			},
		)

		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				FundingScheduleId: 1,
				SpendingType:      models.SpendingTypeGoal,
				TargetAmount:      1000,
				CurrentAmount:     0,
				NextRecurrence:    start,
				RecurrenceRule:    nil,
				SpendingId:        1,
			},
			fundingInstructions,
		).(*spendingInstructionBase)

		result := spendingInstructions.getNextSpendingEventAfter(context.Background(), start, timezone, 0)
		assert.Nil(t, result, "result should be nil because the goal is completed as of the start timestamp")
	})
}
