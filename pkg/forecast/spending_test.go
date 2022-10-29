package forecast

import (
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

		events := spendingInstructions.GetNextNSpendingEventsAfter(3, now, timezone)
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 2500,
				RollingAllocation:  2500,
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 2500,
				RollingAllocation:  5000,
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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

		events := spendingInstructions.GetNextNSpendingEventsAfter(7, now, timezone)
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 22500,
				RollingAllocation:  22500,
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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

		events := spendingInstructions.GetNextNSpendingEventsAfter(7, now, timezone)
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 22500,
				RollingAllocation:  22500,
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided:    false,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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
				Funding:            []FundingEvent{
					{
						Date:              time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
						OriginalDate:      time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided:    true,
						FundingScheduleId: 0,
					},
				},
				SpendingId:         0,
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
