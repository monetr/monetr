package forecast

import (
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/monetr/monetr/pkg/models"
)

type Event struct {
	Date     time.Time `json:"date"`
	Delta    int64 `json:"delta"`
	Balance  int64 `json:"balance"`
	Spending []SpendingEvent `json:"spending"`
	Funding  []FundingEvent `json:"funding"`
}

type Forecast struct {
	StartingBalance int64 `json:"startingBalance"`
	EndingBalance   int64 `json:"endingBalance"`
	Events          []Event `json:"events"`
}

type Forecaster interface {
	GetForecast(start, end time.Time, timezone *time.Location) Forecast
}

type forecasterBase struct {
	currentBalance int64
	funding        map[uint64]FundingInstructions
	spending       map[uint64]SpendingInstructions
}

func NewForecaster(spending []models.Spending, funding []models.FundingSchedule) Forecaster {
	forecaster := &forecasterBase{
		funding:  map[uint64]FundingInstructions{},
		spending: map[uint64]SpendingInstructions{},
	}
	for _, fundingSchedule := range funding {
		forecaster.funding[fundingSchedule.FundingScheduleId] = NewFundingScheduleFundingInstructions(fundingSchedule)
	}
	for _, spendingItem := range spending {
		fundingInstructions, ok := forecaster.funding[spendingItem.FundingScheduleId]
		if !ok {
			panic("missing funding schedule required by spending object")
		}

		forecaster.spending[spendingItem.SpendingId] = NewSpendingInstructions(spendingItem, fundingInstructions)
		forecaster.currentBalance += spendingItem.GetProgressAmount()
	}

	return forecaster
}

func (f *forecasterBase) GetForecast(start, end time.Time, timezone *time.Location) Forecast {
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
			return linq.From(spending.GetSpendingEventsBetween(start, end, timezone))
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
			for i, item := range group.Group {
				spendingEvent := item.(SpendingEvent)
				delta += spendingEvent.ContributionAmount
				delta -= spendingEvent.TransactionAmount

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
				Date:     date,
				Delta:    delta,
				Balance:  0,
				Spending: items,
				Funding:  fundingItems,
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
		forecast.EndingBalance = forecast.Events[len(forecast.Events) - 1].Balance
	}

	return forecast
}
