package forecast

import (
	"context"
	"encoding/json"
	"fmt"
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

		fundingSchedules := []models.FundingSchedule{
			{
				FundingScheduleId: 1,
				Rule:              fundingRule,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 9, 15, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   5000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 8, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleOne,
				SpendingId:     1,
			},
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   12354,
				CurrentAmount:  6177,
				NextRecurrence: time.Date(2022, 9, 26, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleTwo,
				SpendingId:     2,
			},
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   180000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 10, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleThree,
				SpendingId:     3,
			},
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
				SpendingType:   models.SpendingTypeGoal,
				TargetAmount:   1000000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2025, 10, 1, 0, 0, 0, 0, timezone),
				SpendingId:     4,
			},
		}

		var firstAverage, secondAverage int64
		{ // Initial
			forecaster := NewForecaster(spending, fundingSchedules)
			forecast := forecaster.GetForecast(context.Background(), now, now.AddDate(0, 1, 4), timezone)
			assert.Greater(t, forecast.StartingBalance, int64(0))
			firstAverage = forecaster.GetAverageContribution(context.Background(), now, now.AddDate(0, 1, 4), timezone)
		}

		{ // With added expense
			forecaster := NewForecaster(append(spending, models.Spending{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
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

	t.Run("simple expense monthly, one funding schedule", func(t *testing.T) {
		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 11, 3, 0, 0, 1, 0, timezone).UTC()

		fundingSchedules := []models.FundingSchedule{
			{
				FundingScheduleId: 1,
				Rule:              fundingRule,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 11, 15, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   100000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 12, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleOne,
				SpendingId:     1,
			},
		}

		forecaster := NewForecaster(spending, fundingSchedules)
		start, end := now, now.AddDate(0, 1, 0)
		forecast := forecaster.GetForecast(context.Background(), start, end, timezone)

		{ // First event should be on the 15th, and should be a funding for $500
			event := forecast.Events[0]
			assert.EqualValues(t, 50000, event.Contribution, "should have contributed $500")
			assert.EqualValues(t, 50000, event.Balance, "should have a balance of $500")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.Equal(t, "2022-11-15", event.Date.Format("2006-01-02"), "should be funded a bit on the 15th of november")
		}

		{ // Second event should also be funding for $500, but will have the balance from the previous event.
			event := forecast.Events[1]
			assert.EqualValues(t, 50000, event.Contribution, "should have contributed $500")
			assert.EqualValues(t, 100000, event.Balance, "should have a balance of $1000")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.Equal(t, "2022-11-30", event.Date.Format("2006-01-02"), "should be funded a bit on the 30th of november")
		}

		{ // Third event should be spending the entire amount of the expense.
			event := forecast.Events[2]
			assert.Zero(t, event.Contribution, "should not have contributed anything")
			assert.Zero(t, event.Balance, "should not have anything left after spending")
			assert.EqualValues(t, 100000, event.Transaction, "should have spend the $1000 expense")
			assert.Equal(t, "2022-12-01", event.Date.Format("2006-01-02"), "on the 1st of december it should be spent")
		}
	})

	t.Run("simple monthly expense, split funding", func(t *testing.T) {
		fundingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15")
		fundingRuleTwo := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1")
		spendingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 11, 3, 0, 0, 1, 0, timezone).UTC()

		fundingSchedules := []models.FundingSchedule{
			{
				FundingScheduleId: 1,
				Rule:              fundingRuleOne,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 11, 15, 0, 0, 0, 0, timezone),
			},
			{
				FundingScheduleId: 2,
				Rule:              fundingRuleTwo,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 11, 30, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
					{
						FundingScheduleId: 2,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   100000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 12, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleOne,
				SpendingId:     1,
			},
		}

		forecaster := NewForecaster(spending, fundingSchedules)
		start, end := now, now.AddDate(0, 1, 0)
		forecast := forecaster.GetForecast(context.Background(), start, end, timezone)

		{ // First event should be on the 15th, and should be a funding for $500
			event := forecast.Events[0]
			assert.EqualValues(t, 50000, event.Contribution, "should have contributed $500")
			assert.EqualValues(t, 50000, event.Balance, "should have a balance of $500")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 1, event.Funding[0].FundingScheduleId, "should be the first funding schedule")
			assert.Equal(t, "2022-11-15", event.Date.Format("2006-01-02"), "should be funded a bit on the 15th of november")
		}

		{ // Second event should also be funding for $500, but will have the balance from the previous event.
			event := forecast.Events[1]
			assert.EqualValues(t, 50000, event.Contribution, "should have contributed $500")
			assert.EqualValues(t, 100000, event.Balance, "should have a balance of $1000")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 2, event.Funding[0].FundingScheduleId, "should be the second funding schedule")
			assert.Equal(t, "2022-11-30", event.Date.Format("2006-01-02"), "should be funded a bit on the 30th of november")
		}

		{ // Third event should be spending the entire amount of the expense.
			event := forecast.Events[2]
			assert.Zero(t, event.Contribution, "should not have contributed anything")
			assert.Zero(t, event.Balance, "should not have anything left after spending")
			assert.EqualValues(t, 100000, event.Transaction, "should have spend the $1000 expense")
			assert.Empty(t, event.Funding, "should not have any funding events")
			assert.Equal(t, "2022-12-01", event.Date.Format("2006-01-02"), "on the 1st of december it should be spent")
		}
	})

	t.Run("simple monthly expense, split funding overlap", func(t *testing.T) {
		fundingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		fundingRuleTwo := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
		spendingRuleOne := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		now := time.Date(2022, 10, 3, 0, 0, 1, 0, timezone).UTC()

		fundingSchedules := []models.FundingSchedule{
			{
				FundingScheduleId: 1,
				Rule:              fundingRuleOne,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 10, 14, 0, 0, 0, 0, timezone),
			},
			{
				FundingScheduleId: 2,
				Rule:              fundingRuleTwo,
				ExcludeWeekends:   true,
				NextOccurrence:    time.Date(2022, 10, 7, 0, 0, 0, 0, timezone),
			},
		}
		spending := []models.Spending{
			{
				SpendingFunding: []models.SpendingFunding{
					{
						FundingScheduleId: 1,
					},
					{
						FundingScheduleId: 2,
					},
				},
				SpendingType:   models.SpendingTypeExpense,
				TargetAmount:   100000,
				CurrentAmount:  0,
				NextRecurrence: time.Date(2022, 11, 1, 0, 0, 0, 0, timezone),
				RecurrenceRule: spendingRuleOne,
				SpendingId:     1,
			},
		}

		forecaster := NewForecaster(spending, fundingSchedules)
		start, end := now, now.AddDate(0, 1, 0)
		forecast := forecaster.GetForecast(context.Background(), start, end, timezone)

		j, _ := json.MarshalIndent(forecast, "", "  ")
		fmt.Println(string(j))

		{ // First event should be on the 7th, funded by only the weekly funding schedule.
			event := forecast.Events[0]
			assert.EqualValues(t, 20000, event.Contribution, "should have contributed $200")
			assert.EqualValues(t, 20000, event.Balance, "should have a balance of $200")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 2, event.Funding[0].FundingScheduleId, "should be the weekly funding schedule")
			assert.Len(t, event.Funding, 1, "should only have a single funding event")
			assert.Equal(t, "2022-10-07", event.Date.Format("2006-01-02"), "should be on the 7th of october")
		}

		{ // Second event should be on the 14th, excluding the weekend.
			event := forecast.Events[1]
			assert.EqualValues(t, 20000, event.Contribution, "should have contributed $200")
			assert.EqualValues(t, 40000, event.Balance, "should have a balance of $400 now")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.Len(t, event.Funding, 2, "should have two funding events overlap")
			assert.Equal(t, "2022-10-14", event.Date.Format("2006-01-02"), "should be on the 14th of october")
		}

		{
			event := forecast.Events[2]
			assert.EqualValues(t, 20000, event.Contribution, "should have contributed $200")
			assert.EqualValues(t, 60000, event.Balance, "should have a balance of $600 now")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 2, event.Funding[0].FundingScheduleId, "should be the weekly funding schedule")
			assert.Len(t, event.Funding, 1, "should be only the friday funding event")
			assert.Equal(t, "2022-10-21", event.Date.Format("2006-01-02"), "should be on the 21st of october")
		}

		{
			event := forecast.Events[3]
			assert.EqualValues(t, 20000, event.Contribution, "should have contributed $200")
			assert.EqualValues(t, 80000, event.Balance, "should have a balance of $800 now")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 2, event.Funding[0].FundingScheduleId, "should be the weekly funding schedule")
			assert.Len(t, event.Funding, 1, "should be only the friday funding event")
			assert.Equal(t, "2022-10-28", event.Date.Format("2006-01-02"), "should be on the 28th of october")
		}

		{
			event := forecast.Events[4]
			assert.EqualValues(t, 20000, event.Contribution, "should have contributed $200")
			assert.EqualValues(t, 100000, event.Balance, "should have a balance of $1000 now")
			assert.Zero(t, event.Transaction, "should not be spending anything yet")
			assert.EqualValues(t, 1, event.Funding[0].FundingScheduleId, "should be the semi-monthly funding schedule")
			assert.Len(t, event.Funding, 1, "should be only the semi-monthly funding event")
			assert.Equal(t, "2022-10-31", event.Date.Format("2006-01-02"), "should be on the 31st of october")
		}

		{
			event := forecast.Events[5]
			assert.Zero(t, event.Contribution, "nothing should be contributed on this day")
			assert.Zero(t, event.Balance, "balance should now be zero")
			assert.EqualValues(t, 100000, event.Transaction, "should have spent $1000 now")
			assert.Empty(t, event.Funding, "should not have any funding events today")
			assert.Equal(t, "2022-11-01", event.Date.Format("2006-01-02"), "should be on the 31st of october")
		}
	})
}
