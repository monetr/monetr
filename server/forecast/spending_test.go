package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpendingInstructionBase_GetNextSpendingEventAfter(t *testing.T) {
	t.Run("simple monthly expense", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 10, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: false,
				NextRecurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 3, now, timezone)
		assert.NoError(t, err, "should not return an error")
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
						Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 2500,
				RollingAllocation:  5000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				TransactionAmount:  5000,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
		}, events)
	})

	t.Run("spending was late so balance is full", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 10, 2, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=2")
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: false,
				NextRecurrence:  time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  5000,
				NextRecurrence: time.Date(2024, 3, 2, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		{ // If we do the forecase before the expenses due date, then we will get an event showing it was spent.
			now := time.Date(2024, 3, 1, 0, 0, 1, 0, timezone).UTC()
			events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 4, now, timezone)
			assert.NoError(t, err, "should not return an error")
			for i, item := range events {
				if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
					j, _ := json.MarshalIndent(item, "                        \t", "  ")
					fmt.Println("                        \t" + string(j))
				}
			}
			assert.Equal(t, []SpendingEvent{
				{
					Date:               time.Date(2024, 3, 2, 0, 0, 0, 0, timezone),
					TransactionAmount:  5000,
					ContributionAmount: 0,
					RollingAllocation:  0,
					Funding:            []FundingEvent{},
				},
				{
					Date:               time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
					TransactionAmount:  0,
					ContributionAmount: 2500,
					RollingAllocation:  2500,
					Funding: []FundingEvent{
						{
							Date:           time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
							OriginalDate:   time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
							WeekendAvoided: false,
						},
					},
				},
				{
					Date:               time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
					TransactionAmount:  0,
					ContributionAmount: 2500,
					RollingAllocation:  5000,
					Funding: []FundingEvent{
						{
							Date:           time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
							OriginalDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
							WeekendAvoided: false,
						},
					},
				},
				{
					Date:               time.Date(2024, 4, 2, 0, 0, 0, 0, timezone),
					TransactionAmount:  5000,
					ContributionAmount: 0,
					RollingAllocation:  0,
					Funding:            []FundingEvent{},
				},
			}, events)

		}

		{ // But if it hasn't been spent after its due date then we want the same contributions.
			now := time.Date(2024, 3, 4, 0, 0, 1, 0, timezone).UTC()
			events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 3, now, timezone)
			assert.NoError(t, err, "should not return an error")
			for i, item := range events {
				if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
					j, _ := json.MarshalIndent(item, "                        \t", "  ")
					fmt.Println("                        \t" + string(j))
				}
			}
			assert.Equal(t, []SpendingEvent{
				{
					Date:               time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
					TransactionAmount:  0,
					ContributionAmount: 2500,
					RollingAllocation:  7500,
					Funding: []FundingEvent{
						{
							Date:           time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
							OriginalDate:   time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
							WeekendAvoided: false,
						},
					},
				},
				{
					Date:               time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
					TransactionAmount:  0,
					ContributionAmount: 2500,
					RollingAllocation:  10000,
					Funding: []FundingEvent{
						{
							Date:           time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
							OriginalDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
							WeekendAvoided: false,
						},
					},
				},
				{
					Date:               time.Date(2024, 4, 2, 0, 0, 0, 0, timezone),
					TransactionAmount:  5000,
					ContributionAmount: 0,
					RollingAllocation:  5000,
					Funding:            []FundingEvent{},
				},
			}, events)
		}
	})

	t.Run("spent more frequently than funded", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 9, 16, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
		now := time.Date(2022, 9, 14, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: false,
				NextRecurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   7500,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 7, now, timezone)
		assert.NoError(t, err, "should not return an error")
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
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 9, 23, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 22500,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 10, 7, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
		}, events)
	})

	t.Run("spent more frequently than funded excluding weekends", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 9, 16, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
		now := time.Date(2022, 9, 14, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   7500,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 8, now, timezone)
		assert.NoError(t, err, "should not return an error")
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
				ContributionAmount: 15000,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 9, 16, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 9, 23, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 15000,
				RollingAllocation:  7500,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 9, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2022, 10, 7, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 22500,
				RollingAllocation:  15000,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2022, 10, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: true,
					},
				},
			},
			{
				Date:               time.Date(2022, 10, 21, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  7500,
				Funding:            []FundingEvent{},
			},
			{
				Date:               time.Date(2022, 10, 28, 0, 0, 0, 0, timezone),
				TransactionAmount:  7500,
				ContributionAmount: 0,
				RollingAllocation:  0,
				Funding:            []FundingEvent{},
			},
		}, events)
	})
}

func TestSpendingInstructionBase_GetSpendingEventsBetween(t *testing.T) {
	t.Run("once a month for a year", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 1, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   1395,
				CurrentAmount:  1395,
				NextRecurrence: time.Date(2022, 1, 8, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		events, err := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.NoError(t, err, "should not return an error")
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
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 1, 7, timezone, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
			},
		)
		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   1395,
				CurrentAmount:  1395,
				NextRecurrence: time.Date(2022, 1, 7, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRule,
			},
			fundingInstructions,
		)

		events, err := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.NoError(t, err, "should not return an error")
		assert.Len(t, events, 45, "should have 45 events")
		for i, item := range events {
			if !assert.GreaterOrEqual(t, item.RollingAllocation, int64(0), "rolling allocation must be greater than zero: [%d] %s", i, item.Date) {
				j, _ := json.MarshalIndent(item, "                        \t", "  ")
				fmt.Println("                        \t" + string(j))
			}
		}
	})

	t.Run("no spending events for paused spending objects", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		now := time.Date(2022, 1, 2, 13, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)
		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 1, 15, 0, 0, 0, 0, timezone),
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

		events, err := spendingInstructions.GetSpendingEventsBetween(context.Background(), now, now.AddDate(1, 0, 0), timezone)
		assert.NoError(t, err, "should not return an error")
		assert.Empty(t, events, "there should be no spending events for paused spending")
	})

	t.Run("spending events infinite loop bug", func(t *testing.T) {
		// This is part of: https://github.com/monetr/monetr/issues/1243
		// Make sure we don't timeout when a goal lands on the funding day.
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		start := time.Date(2022, 12, 1, 0, 0, 0, 0, timezone)
		log := testutils.GetLog(t)

		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 11, 30, 0, 0, 0, 0, timezone),
			},
		)

		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000,
				CurrentAmount:  0,
				NextRecurrence: start,
				RuleSet:        nil,
			},
			fundingInstructions,
		).(*spendingInstructionBase)

		result, err := spendingInstructions.getNextSpendingEventAfter(context.Background(), start, timezone, 0)
		assert.NoError(t, err, "should not return an error")
		assert.Nil(t, result, "result should be nil because the goal is completed as of the start timestamp")
	})

	t.Run("odd goal contribution repro", func(t *testing.T) {
		// This test makes sure that even when we are avoiding weekends that we are
		// staying consistent in how we are contributing to a budget.
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2021, 12, 31, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		now := time.Date(2024, 3, 28, 0, 0, 0, 0, timezone)
		log := testutils.GetLog(t)

		fundingInstructions := NewFundingScheduleFundingInstructions(
			log,
			models.FundingSchedule{
				RuleSet: fundingRule,
				// This is the problem, with exclude weekends set to false the
				// contribution amount is accurate. Ultimately this is because
				// GetFundingEventsBetween does not properly implement excluding
				// weekends.
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2024, 3, 15, 0, 0, 0, 0, timezone),
			},
		)

		spendingInstructions := NewSpendingInstructions(
			log,
			models.Spending{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  48394,
				UsedAmount:     319861,
				NextRecurrence: time.Date(2024, 5, 16, 0, 0, 0, 0, timezone),
				RuleSet:        nil,
			},
			fundingInstructions,
		).(*spendingInstructionBase)

		events, err := spendingInstructions.GetNextNSpendingEventsAfter(context.Background(), 8, now, timezone)
		assert.NoError(t, err, "should not return an error")
		assert.Equal(t, []SpendingEvent{
			{
				Date:               time.Date(2024, 3, 29, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 157936,
				RollingAllocation:  206330,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2024, 3, 29, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, timezone),
						WeekendAvoided: true,
					},
				},
			},
			{
				Date:               time.Date(2024, 4, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 157936,
				RollingAllocation:  364266,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2024, 4, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2024, 4, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2024, 4, 30, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 157936,
				RollingAllocation:  522202,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2024, 4, 30, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2024, 4, 30, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2024, 5, 15, 0, 0, 0, 0, timezone),
				TransactionAmount:  0,
				ContributionAmount: 157937,
				RollingAllocation:  680139,
				Funding: []FundingEvent{
					{
						Date:           time.Date(2024, 5, 15, 0, 0, 0, 0, timezone),
						OriginalDate:   time.Date(2024, 5, 15, 0, 0, 0, 0, timezone),
						WeekendAvoided: false,
					},
				},
			},
			{
				Date:               time.Date(2024, 5, 16, 0, 0, 0, 0, timezone),
				TransactionAmount:  1000000,
				ContributionAmount: 0,
				// TODO This is the same as the used amount, so what gives? Should this
				// be 0? Should it take into account the used amount?
				RollingAllocation: -319861,
				Funding:           []FundingEvent{},
			},
		}, events)
	})
}

func TestSpendingInstructionBase_GetNextInflowEventAfter(t *testing.T) {
	t.Run("next fuunding in the past updated", func(t *testing.T) {
		log := testutils.GetLog(t)
		today := util.Midnight(time.Now(), time.UTC)
		dayAfterTomorrow := util.Midnight(today.Add(48*time.Hour), time.UTC)
		dayAfterDayAfterTomorrow := util.Midnight(time.Now().Add(72*time.Hour), time.UTC)
		assert.True(t, dayAfterDayAfterTomorrow.After(today), "dayAfterDayAfterTomorrow timestamp must come after today's")

		ruleset := testutils.RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", today)
		spending := models.Spending{
			SpendingType:   models.SpendingTypeGoal,
			TargetAmount:   100,
			CurrentAmount:  0,
			NextRecurrence: dayAfterDayAfterTomorrow,
		}

		fundingInstructions := NewFundingScheduleFundingInstructions(log, models.FundingSchedule{
			FundingScheduleId: "fund_bogus",
			Name:              "Bogus Funding Schedule",
			Description:       "Bogus",
			RuleSet:           ruleset,
			NextRecurrence:    dayAfterTomorrow,
		})
		spendingInstructions := NewSpendingInstructions(log, spending, fundingInstructions)

		event, err := spendingInstructions.GetNextInflowEventAfter(
			context.Background(),
			time.Now(),
			time.UTC,
		)
		assert.NoError(t, err, "must be able to calculate the next spending event")
		// TODO Need to add support for "is behind" in the forecasting code.
		assert.EqualValues(t, spending.TargetAmount, event.ContributionAmount, "next contribution should be the entire amount")
	})

	t.Run("next funding in the past is behind", func(t *testing.T) {
		log := testutils.GetLog(t)
		today := util.Midnight(time.Now(), time.UTC)
		dayAfterTomorrow := util.Midnight(time.Now().Add(48*time.Hour), time.UTC)
		assert.True(t, dayAfterTomorrow.After(today), "dayAfterTomorrow timestamp must come after today's")
		ruleset := testutils.RuleToSet(t, time.UTC, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", today)

		spending := models.Spending{
			SpendingType:   models.SpendingTypeGoal,
			TargetAmount:   100,
			CurrentAmount:  0,
			NextRecurrence: dayAfterTomorrow,
		}

		fundingInstructions := NewFundingScheduleFundingInstructions(log, models.FundingSchedule{
			FundingScheduleId: "fund_bogus",
			Name:              "Bogus Funding Schedule",
			Description:       "Bogus",
			RuleSet:           ruleset,
			NextRecurrence:    dayAfterTomorrow.AddDate(0, 0, 1),
		})
		spendingInstructions := NewSpendingInstructions(log, spending, fundingInstructions)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()
		event, err := spendingInstructions.GetNextInflowEventAfter(
			ctx,
			time.Now(),
			time.UTC,
		)
		assert.NoError(t, err, "must be able to calculate the next spending event")
		assert.Nil(t, event, "the goal finishes before the next funding date")
	})

	t.Run("timezone near midnight", func(t *testing.T) {
		log := testutils.GetLog(t)
		// This test is here to prove a fix for https://github.com/monetr/monetr/issues/937
		// Basically, we want to make sure that if the user is close to their
		// funding schedule such that the server is already on the next day; that we
		// still calculate everything correctly. This was not happening, this test
		// accompanies a fix to remedy that situation.
		centralTimezone, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load timezone")
		now := time.Date(2022, 06, 14, 22, 37, 43, 0, centralTimezone)
		nextFunding := time.Date(2022, 06, 15, 0, 0, 0, 0, centralTimezone)
		nextRecurrence := time.Date(2022, 7, 8, 0, 0, 0, 0, centralTimezone)

		fundingRule := testutils.RuleToSet(t, centralTimezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", now)
		spendingRule := testutils.RuleToSet(t, centralTimezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8", now)

		spending := models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   25000,
			CurrentAmount:  6960,
			RuleSet:        spendingRule,
			NextRecurrence: nextRecurrence.UTC(),
		}

		funding := models.FundingSchedule{
			FundingScheduleId: "fund_bogus",
			Name:              "Bogus Funding Schedule",
			Description:       "Bogus",
			RuleSet:           fundingRule,
			NextRecurrence:    nextFunding.UTC(),
		}

		fundingInstructions := NewFundingScheduleFundingInstructions(log, funding)
		spendingInstructions := NewSpendingInstructions(log, spending, fundingInstructions)

		event, err := spendingInstructions.GetNextInflowEventAfter(
			context.Background(),
			now,
			centralTimezone,
		)
		assert.NoError(t, err, "must be able to calculate the next contribution even with a past funding date")
		assert.EqualValues(t, 9020, event.ContributionAmount, "next contribution amount should be half of the total needed to reach the target")
	})
}
