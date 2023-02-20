package forecast

import (
	"context"
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/sirupsen/logrus"
)

type Event struct {
	Date         time.Time       `json:"date"`
	Delta        int64           `json:"delta"`
	Contribution int64           `json:"contribution"`
	Transaction  int64           `json:"transaction"`
	Balance      int64           `json:"balance"`
	Spending     []SpendingEvent `json:"spending"`
	Funding      []FundingEvent  `json:"funding"`
}

type Forecast struct {
	StartingBalance int64   `json:"startingBalance"`
	EndingBalance   int64   `json:"endingBalance"`
	Events          []Event `json:"events"`
}

type Forecaster interface {
	GetForecast(ctx context.Context, start, end time.Time, timezone *time.Location) Forecast
	GetAverageContribution(ctx context.Context, start, end time.Time, timezone *time.Location) int64
	GetNextContribution(ctx context.Context, start time.Time, fundingScheduleId uint64, timezone *time.Location) int64
}

type forecasterBase struct {
	log            *logrus.Entry
	currentBalance int64
	funding        map[uint64]FundingInstructions
	spending       map[uint64]SpendingInstructions
}

func NewForecaster(log *logrus.Entry, spending []models.Spending, funding []models.FundingSchedule) Forecaster {
	forecaster := &forecasterBase{
		log:      log,
		funding:  map[uint64]FundingInstructions{},
		spending: map[uint64]SpendingInstructions{},
	}
	for _, fundingSchedule := range funding {
		forecaster.funding[fundingSchedule.FundingScheduleId] = NewFundingScheduleFundingInstructions(log, fundingSchedule)
	}
	for _, spendingItem := range spending {
		fundingInstructions, ok := forecaster.funding[spendingItem.FundingScheduleId]
		if !ok {
			panic("missing funding schedule required by spending object")
		}

		forecaster.spending[spendingItem.SpendingId] = NewSpendingInstructions(
			log,
			spendingItem,
			fundingInstructions,
		)
		forecaster.currentBalance += spendingItem.GetProgressAmount()
	}

	return forecaster
}

func (f *forecasterBase) GetForecast(ctx context.Context, start, end time.Time, timezone *time.Location) Forecast {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"start":    start,
		"end":      end,
		"timezone": timezone.String(),
	}
	forecast := Forecast{
		StartingBalance: f.currentBalance,
		EndingBalance:   0,
		Events:          make([]Event, 0),
	}

	linq.From(f.spending).
		SelectT(func(item linq.KeyValue) SpendingInstructions {
			return item.Value.(SpendingInstructions)
		}).
		SelectManyT(func(spending SpendingInstructions) linq.Query {
			return linq.From(spending.GetSpendingEventsBetween(span.Context(), start, end, timezone))
		}).
		GroupByT(
			func(item SpendingEvent) int64 {
				return item.Date.Unix()
			},
			func(item SpendingEvent) SpendingEvent {
				return item
			},
		).
		SelectT(func(group linq.Group) Event {
			date := time.Unix(group.Key.(int64), 0)
			items := make([]SpendingEvent, len(group.Group))
			fundingMap := map[uint64]FundingEvent{}
			var delta int64 = 0
			var transaction int64
			var contribution int64
			for i, item := range group.Group {
				spendingEvent := item.(SpendingEvent)
				delta += spendingEvent.ContributionAmount
				contribution += spendingEvent.ContributionAmount
				delta -= spendingEvent.TransactionAmount
				transaction += spendingEvent.TransactionAmount

				for _, funding := range spendingEvent.Funding {
					if _, ok := fundingMap[funding.FundingScheduleId]; !ok {
						fundingMap[funding.FundingScheduleId] = funding
					}
				}

				items[i] = spendingEvent
			}

			fundingItems := make([]FundingEvent, 0, len(fundingMap))
			for _, item := range fundingMap {
				fundingItems = append(fundingItems, item)
			}

			return Event{
				Date:         date,
				Delta:        delta,
				Balance:      0,
				Transaction:  transaction,
				Contribution: contribution,
				Spending:     items,
				Funding:      fundingItems,
			}
		}).
		OrderByT(func(event Event) int64 {
			return event.Date.Unix()
		}).
		ToSlice(&forecast.Events)

	for i, event := range forecast.Events {
		var previousBalance int64 = 0
		if i == 0 {
			previousBalance = forecast.StartingBalance
		} else {
			previousBalance = forecast.Events[i-1].Balance
		}

		forecast.Events[i].Balance = previousBalance + event.Delta
	}

	if len(forecast.Events) > 0 {
		forecast.EndingBalance = forecast.Events[len(forecast.Events)-1].Balance
	}

	return forecast
}

func (f *forecasterBase) GetAverageContribution(ctx context.Context, start, end time.Time, timezone *time.Location) int64 {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"start":    start,
		"end":      end,
		"timezone": timezone.String(),
	}
	forecast := f.GetForecast(span.Context(), start, end, timezone)
	contributionAmounts := map[int64]int64{}
	for _, event := range forecast.Events {
		if event.Contribution == 0 {
			continue
		}

		contributionAmounts[event.Contribution] += 1
	}
	var popularContribution, popularCount int64
	for contribution, contributionCount := range contributionAmounts {
		if contributionCount > popularCount {
			popularContribution = contribution
			popularCount = contributionCount
		}
	}

	return popularContribution
}

func (f *forecasterBase) GetNextContribution(ctx context.Context, start time.Time, fundingScheduleId uint64, timezone *time.Location) int64 {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	end := f.funding[fundingScheduleId].GetNextFundingEventAfter(span.Context(), start, timezone)

	span.Data = map[string]interface{}{
		"start":             start,
		"end":               end,
		"fundingScheduleId": fundingScheduleId,
		"timezone":          timezone.String(),
	}

	forecast := f.GetForecast(span.Context(), start, end.Date.AddDate(0, 0, 1), timezone)
	for _, event := range forecast.Events {
		if len(event.Funding) == 0 {
			continue
		}

		return event.Contribution
	}

	return 0
}
