package forecast

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestForecasterBase_GetForecast(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 9, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.NewRuleSet(t, 2022, 10, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		spendingRuleTwo := testutils.NewRuleSet(t, 2022, 9, 26, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=26")
		spendingRuleThree := testutils.NewRuleSet(t, 2022, 10, 1, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		now := time.Date(2022, 9, 13, 0, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				RuleSet:         fundingRule,
				ExcludeWeekends: true,
				NextRecurrence:  time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRuleOne,
				SpendingId:     "spnd_1",
			},
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   12354,
				CurrentAmount:  6177,
				NextRecurrence: time.Date(2022, 9, 26, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRuleTwo,
				SpendingId:     "spnd_2",
			},
			{
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   180000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 1, 0, 0, 0, 0, timezone),
				RuleSet:        spendingRuleThree,
				SpendingId:     "spnd_3",
			},
			{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2025, 10, 1, 0, 0, 0, 0, timezone),
				SpendingId:     "spnd_4",
			},
		}

		var firstAverage, secondAverage int64
		{ // Initial
			forecaster := NewForecaster(log, spending, fundingSchedules)
			forecast, err := forecaster.GetForecast(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.NoError(t, err, "should not return an error")
			assert.Greater(t, forecast.StartingBalance, int64(0))
			firstAverage, err = forecaster.GetAverageContribution(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.NoError(t, err, "should not return an error")
		}

		{ // With added expense
			forecaster := NewForecaster(log, append(spending, models.Spending{
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  0,
				NextRecurrence: util.Midnight(now.AddDate(1, 0, 0), timezone),
			}), fundingSchedules)
			forecast, err := forecaster.GetForecast(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.NoError(t, err, "should not return an error")
			assert.Greater(t, forecast.StartingBalance, int64(0))
			secondAverage, err = forecaster.GetAverageContribution(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.NoError(t, err, "should not return an error")
		}
		assert.Greater(t, secondAverage, firstAverage, "should need to contribute more per funding")
	})

	t.Run("dont timeout", func(t *testing.T) {
		// This is part of: https://github.com/monetr/monetr/issues/1243
		// This test previously proved that a timeout bug existed, but now proves that one does not; at least not the one
		// that was originally causing the problem.
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		now := time.Date(2022, 11, 29, 14, 30, 1, 0, timezone).UTC()
		end := time.Date(2022, 12, 2, 0, 0, 0, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				RuleSet:           fundingRule,
				ExcludeWeekends:   true,
				NextRecurrence:    time.Date(2022, 11, 30, 0, 0, 0, 0, timezone),
				FundingScheduleId: "fund_1",
			},
		}
		spending := []models.Spending{
			{
				FundingScheduleId: "fund_1",
				SpendingType:      models.SpendingTypeGoal,
				TargetAmount:      1000,
				CurrentAmount:     0,
				NextRecurrence:    time.Date(2022, 12, 1, 0, 0, 0, 0, timezone),
				RuleSet:           nil,
				SpendingId:        "spnd_1",
			},
		}

		forecaster := NewForecaster(log, spending, fundingSchedules)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		assert.NotPanics(t, func() {
			result, err := forecaster.GetForecast(ctx, now, end, timezone)
			assert.NoError(t, err, "should not return an error")
			assert.NotNil(t, result, "just make sure something is returned, this is to make sure we dont timeout")
		})
	})

	t.Run("midnight goofiness", func(t *testing.T) {
		// This test was part of a bugfix. Previously there was an issue where when we would adjust times to midnight of
		// that day. We would sometimes return a timestamp that was actually the next day. This happened when the timezone
		// we were working in was behind UTC, but the timestamp itself was such that UTC was the next day. This caused the
		// forecaster to miss a funding schedule that was in just a few hours because it believed it had already happened.
		// This test proves that the bug is resolved and if it fails again in the future, that means the timezone bug has
		// been reintroduced somehow.
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2023, 9, 1, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		// Monday August 14th, 2023. Payday is Tuesday August 15th.
		now := time.Date(2023, 8, 14, 19, 30, 1, 0, timezone).UTC()
		// Project 1 month into the future exactly.
		end := time.Date(2023, 9, 14, 19, 30, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				RuleSet:           fundingRule,
				ExcludeWeekends:   true,
				NextRecurrence:    time.Date(2023, 8, 15, 0, 0, 0, 0, timezone),
				FundingScheduleId: "fund_1",
			},
		}
		spending := []models.Spending{
			{
				FundingScheduleId: "fund_1",
				SpendingType:      models.SpendingTypeExpense,
				TargetAmount:      2000,
				CurrentAmount:     0,
				NextRecurrence:    time.Date(2023, 9, 1, 0, 0, 0, 0, timezone),
				RuleSet:           spendingRule,
				SpendingId:        "spnd_1",
			},
		}

		forecaster := NewForecaster(log, spending, fundingSchedules)
		//ctx := context.Background()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := forecaster.GetForecast(ctx, now, end, timezone)
		assert.NoError(t, err, "should not return an error")
		expected := Forecast{
			StartingTime:    now,
			EndingTime:      end,
			StartingBalance: 0,
			EndingBalance:   0,
			Events: []Event{
				{
					Date:         time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
					Delta:        1000,
					Contribution: 1000,
					Transaction:  0,
					Balance:      1000,
					Spending: []SpendingEvent{
						{
							Date:               time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
							TransactionAmount:  0,
							ContributionAmount: 1000,
							RollingAllocation:  1000,
							Funding: []FundingEvent{
								{
									Date:              time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
									OriginalDate:      time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
									WeekendAvoided:    false,
									FundingScheduleId: "fund_1",
								},
							},
							SpendingId: "spnd_1",
						},
					},
					Funding: []FundingEvent{
						{
							Date:              time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
							OriginalDate:      time.Date(2023, 8, 15, 5, 0, 0, 0, time.UTC),
							WeekendAvoided:    false,
							FundingScheduleId: "fund_1",
						},
					},
				},
				{
					Date:         time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
					Delta:        1000,
					Contribution: 1000,
					Transaction:  0,
					Balance:      2000,
					Spending: []SpendingEvent{
						{
							Date:               time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
							TransactionAmount:  0,
							ContributionAmount: 1000,
							RollingAllocation:  2000,
							Funding: []FundingEvent{
								{
									Date:              time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
									OriginalDate:      time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
									WeekendAvoided:    false,
									FundingScheduleId: "fund_1",
								},
							},
							SpendingId: "spnd_1",
						},
					},
					Funding: []FundingEvent{
						{
							Date:              time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
							OriginalDate:      time.Date(2023, 8, 31, 5, 0, 0, 0, time.UTC),
							WeekendAvoided:    false,
							FundingScheduleId: "fund_1",
						},
					},
				},
				{
					Date:         time.Date(2023, 9, 1, 5, 0, 0, 0, time.UTC),
					Delta:        -2000,
					Contribution: 0,
					Transaction:  2000,
					Balance:      0,
					Spending: []SpendingEvent{
						{
							Date:               time.Date(2023, 9, 1, 5, 0, 0, 0, time.UTC),
							TransactionAmount:  2000,
							ContributionAmount: 0,
							RollingAllocation:  0,
							Funding:            []FundingEvent{},
							SpendingId:         "spnd_1",
						},
					},
					Funding: []FundingEvent{},
				},
			},
		}

		expectedJson, _ := json.MarshalIndent(expected, "", "  ")
		resultJson, _ := json.MarshalIndent(result, "", "  ")
		assert.JSONEq(t, string(expectedJson), string(resultJson))
		// assert.EqualValues(t, , result, "expected forecast")
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

		end := funding[0].NextRecurrence
		end = end.AddDate(0, 0, 20)
		assert.Greater(t, end, now, "make sure that our end is actually in the future")

		forecaster := NewForecaster(log, spending, funding)
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()
		result, err := forecaster.GetForecast(ctx, now, end, timezone)
		assert.NoError(t, err, "should not return an error")
		assert.NotNil(t, result, "just make sure something is returned, this is to make sure we dont timeout")
		pretty, err := json.MarshalIndent(result, "", "  ")
		assert.NoError(t, err, "must be able to convert the forecast into a pretty json")
		resultingJson := strings.TrimSpace(string(pretty))
		expectedJson := strings.TrimSpace(string(testutils.Must(t, forecastingFixtureData.ReadFile, "fixtures/elliots-result-20230705.json")))
		// fmt.Println(string(resultingJson))
		assert.Equal(t, expectedJson, resultingJson, "the result should match the saved fixture")
	})

	t.Run("with a completed goal", func(t *testing.T) {
		// This test is part of a bug report: https://github.com/monetr/monetr/issues/1561
		// It proves that when a goal is completed more contributions will not be made to it. The bug showed that if the
		// goal's target date was in the future and the balance of the goal was less than the target (not taking into
		// account the used amount) it would still try to contribute to the goal. This test shows that is no longer the case
		// and that we are doing the correct thing going forward.
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		now := time.Date(2023, 10, 16, 4, 30, 1, 0, timezone).UTC()
		end := time.Date(2023, 11, 1, 5, 30, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)

		fundingSchedules := []models.FundingSchedule{
			{
				RuleSet:           fundingRule,
				ExcludeWeekends:   true,
				NextRecurrence:    time.Date(2023, 10, 31, 0, 0, 0, 0, timezone),
				FundingScheduleId: "fund_1",
			},
		}
		spending := []models.Spending{
			{
				FundingScheduleId: "fund_1",
				SpendingType:      models.SpendingTypeGoal,
				TargetAmount:      2000,
				UsedAmount:        2001,
				CurrentAmount:     0,
				NextRecurrence:    time.Date(2023, 11, 17, 0, 0, 0, 0, timezone),
				SpendingId:        "spnd_1",
			},
		}

		forecaster := NewForecaster(log, spending, fundingSchedules)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := forecaster.GetForecast(ctx, now, end, timezone)
		assert.NoError(t, err, "should not return an error")
		expected := Forecast{
			StartingTime:    now,
			EndingTime:      end,
			StartingBalance: 0,
			EndingBalance:   0,
			Events: []Event{
				{
					Date:         time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
					Delta:        0,
					Contribution: 0,
					Transaction:  0,
					Balance:      0,
					Spending: []SpendingEvent{
						{
							Date:               time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
							TransactionAmount:  0,
							ContributionAmount: 0,
							RollingAllocation:  0,
							Funding: []FundingEvent{
								{
									Date:              time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
									OriginalDate:      time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
									WeekendAvoided:    false,
									FundingScheduleId: "fund_1",
								},
							},
							SpendingId: "spnd_1",
						},
					},
					Funding: []FundingEvent{
						{
							Date:              time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
							OriginalDate:      time.Date(2023, 10, 31, 0, 0, 0, 0, timezone).UTC(),
							WeekendAvoided:    false,
							FundingScheduleId: "fund_1",
						},
					},
				},
			},
		}
		assert.Equal(t, expected, result, "forecast should not have any contributions to the goal")
	})
}
