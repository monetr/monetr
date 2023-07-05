package forecast

import (
	"context"
	"encoding/json"
	"strings"
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
				DateStarted:       time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
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
				DateStarted:       time.Date(2022, 1, 1, 0, 0, 0, 0, timezone),
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

	t.Run("with elliot fixtures 20230705", func(t *testing.T) {
		funding := make([]models.FundingSchedule, 0)
		spending := make([]models.Spending, 0)

		{ // Read fixture data into the test.
			fundingJson := testutils.Must(t, forecastingFixtureData.ReadFile, "fixtures/elliots-funding-20230705.json")
			spendingJson := testutils.Must(t, forecastingFixtureData.ReadFile, "fixtures/elliots-spending-20230705.json")
			testutils.MustUnmarshalJSON(t, fundingJson, &funding)
			testutils.MustUnmarshalJSON(t, spendingJson, &spending)
			assert.NotEmpty(t, funding, "must have funding schedules loaded")
			assert.NotEmpty(t, spending, "must have spending data loaded")
		}

		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2023, 07, 05, 15, 9, 0, 0, timezone).UTC()
		log := testutils.GetLog(t)

		end := funding[0].NextOccurrence
		end = end.AddDate(0, 0, 20)
		assert.Greater(t, end, now, "make sure that our end is actually in the future")

		forecaster := NewForecaster(log, spending, funding)
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()
		result := forecaster.GetForecast(ctx, now, end, timezone)
		assert.NotNil(t, result, "just make sure something is returned, this is to make sure we dont timeout")
		pretty, err := json.MarshalIndent(result, "", "  ")
		assert.NoError(t, err, "must be able to convert the forecast into a pretty json")
		resultingJson := strings.TrimSpace(string(pretty))
		expectedJson := strings.TrimSpace(string(testutils.Must(t, forecastingFixtureData.ReadFile, "fixtures/elliots-result-20230705.json")))
		assert.Equal(t, expectedJson, resultingJson, "the result should match the saved fixture")
	})
}
