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
		fundingInstructions := NewFundingScheduleFundingInstructions(models.FundingSchedule{
			Rule:            fundingRule,
			ExcludeWeekends: false,
			NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
		})
		spendingInstructions := NewSpendingInstructions(models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  0,
			NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
			RecurrenceRule: spendingRule,
		}, fundingInstructions)

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
		fundingInstructions := NewFundingScheduleFundingInstructions(models.FundingSchedule{
			Rule:            fundingRule,
			ExcludeWeekends: false,
			NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
		})
		spendingInstructions := NewSpendingInstructions(models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   7500,
			CurrentAmount:  0,
			NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
			RecurrenceRule: spendingRule,
		}, fundingInstructions)

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
		fundingInstructions := NewFundingScheduleFundingInstructions(models.FundingSchedule{
			Rule:            fundingRule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
		})
		spendingInstructions := NewSpendingInstructions(models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   7500,
			CurrentAmount:  0,
			NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
			RecurrenceRule: spendingRule,
		}, fundingInstructions)

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
		fundingInstructions := NewFundingScheduleFundingInstructions(models.FundingSchedule{
			Rule:            fundingRule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
		})
		spendingInstructions := NewSpendingInstructions(models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   1395,
			CurrentAmount:  1395,
			NextRecurrence: time.Date(2022, 1, 8, 0, 0, 0, 0, timezone),
			RecurrenceRule: spendingRule,
		}, fundingInstructions)

		events := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		// Should have 36 events, 12 spending events and 24 funding events.
		assert.Len(t, events, 12 + 24, "should have 36 events")
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
		fundingInstructions := NewFundingScheduleFundingInstructions(models.FundingSchedule{
			Rule:            fundingRule,
			ExcludeWeekends: true,
			NextOccurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
		})
		spendingInstructions := NewSpendingInstructions(models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   1395,
			CurrentAmount:  1395,
			NextRecurrence: time.Date(2022, 1, 7, 0, 0, 0, 0, timezone),
			RecurrenceRule: spendingRule,
		}, fundingInstructions)

		events := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.Len(t, events, 45, "should have 45 events")
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
	})
}
