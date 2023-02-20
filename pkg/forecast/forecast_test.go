package forecast

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestForecasterBase_GetForecast(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		spendingRuleTwo := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=26")
		spendingRuleThree := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				Rule:            fundingRule,
				ExcludeWeekends: true,
				NextOccurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleOne,
				SpendingId:     1,
			},
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   12354,
				CurrentAmount:  6177,
				NextRecurrence: time.Date(2022, 9, 26, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleTwo,
				SpendingId:     2,
			},
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   180000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleThree,
				SpendingId:     3,
			},
			{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2025, 10, 1, 0, 0, 0, 0, timezone),
				SpendingId:     4,
			},
		}

		var firstAverage, secondAverage int64
		{ // Initial
			forecaster := NewForecaster(log, spending, fundingSchedules)
			forecast := forecaster.GetForecast(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.Greater(t, forecast.StartingBalance, int64(0))
			firstAverage = forecaster.GetAverageContribution(context.Background(), now, now.AddDate(0, 1, 4), timezone)
		}

		{ // With added expense
			forecaster := NewForecaster(log, append(spending, models.Spending{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  0,
				NextRecurrence: util.MidnightInLocal(now.AddDate(1, 0, 0), timezone),
			}), fundingSchedules)
			forecast := forecaster.GetForecast(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.Greater(t, forecast.StartingBalance, int64(0))
			secondAverage = forecaster.GetAverageContribution(context.Background(), now, now.AddDate(0, 1, 4), timezone)
		}
		assert.Greater(t, secondAverage, firstAverage, "should need to contribute more per funding")
	})

	t.Run("dont timeout", func(t *testing.T) {
		// This is part of: https://github.com/monetr/monetr/issues/1243
		// This test previously proved that a timeout bug existed, but now proves that one does not; at least not the one
		// that was originally causing the problem.
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 11, 29, 14, 30, 1, 0, timezone).UTC()
		end := time.Date(2022, 12, 2, 0, 0, 0, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				Rule:              fundingRule,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 11, 30, 0, 0, 0, 0, timezone),
				FundingScheduleId: 1,
			},
		}
		spending := []models.Spending{
			{
				FundingScheduleId: 1,
				SpendingType:      models.SpendingTypeGoal,
				TargetAmount:      1000,
				CurrentAmount:     0,
				NextRecurrence:    time.Date(2022, 12, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule:    nil,
				SpendingId:        1,
			},
		}

		forecaster := NewForecaster(log, spending, fundingSchedules)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		assert.NotPanics(t, func() {
			result := forecaster.GetForecast(ctx, now, end, timezone)
			assert.NotNil(t, result, "just make sure something is returned, this is to make sure we dont timeout")
		})
	})
}
