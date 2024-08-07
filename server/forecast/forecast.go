package forecast

import (
	"context"
	"sort"
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
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
	StartingTime    time.Time `json:"startingTime"`
	EndingTime      time.Time `json:"endingTime"`
	StartingBalance int64     `json:"startingBalance"`
	EndingBalance   int64     `json:"endingBalance"`
	Events          []Event   `json:"events"`
}

type Forecaster interface {
	GetForecast(
		ctx context.Context,
		start, end time.Time,
		timezone *time.Location,
	) (Forecast, error)
	GetAverageContribution(
		ctx context.Context,
		start, end time.Time,
		timezone *time.Location,
	) (int64, error)
	GetNextContribution(
		ctx context.Context,
		start time.Time,
		fundingScheduleId ID[FundingSchedule],
		timezone *time.Location,
	) (int64, error)
}

type forecasterBase struct {
	log            *logrus.Entry
	currentBalance int64
	funding        map[ID[FundingSchedule]]FundingInstructions
	spending       map[ID[Spending]]SpendingInstructions
}

func NewForecaster(log *logrus.Entry, spending []Spending, funding []FundingSchedule) Forecaster {
	forecaster := &forecasterBase{
		log:      log,
		funding:  map[ID[FundingSchedule]]FundingInstructions{},
		spending: map[ID[Spending]]SpendingInstructions{},
	}
	for _, fundingSchedule := range funding {
		forecaster.funding[fundingSchedule.FundingScheduleId] = NewFundingScheduleFundingInstructions(log, fundingSchedule)
	}
	for _, spendingItem := range spending {
		if spendingItem.GetIsPaused() {
			log.WithField("spendingId", spendingItem.SpendingId).
				Debug("spending item will be excluded from forecast because it is paused")
			continue
		}

		fundingInstructions, ok := forecaster.funding[spendingItem.FundingScheduleId]
		if !ok {
			panic("missing funding schedule required by spending object")
		}

		forecaster.spending[spendingItem.SpendingId] = NewSpendingInstructions(
			log,
			spendingItem,
			fundingInstructions,
		)
		forecaster.currentBalance += spendingItem.CurrentAmount
	}

	return forecaster
}

// GetForecast combines the instructions from the spending and funding objects
// and returns a timeline of events that are expected to happen based on those
// instructions. All dates returned by this function will be in UTC. Dates
// returned by other functions may be in the timezone provided by the caller.
// Timezones are converted in this function because they will likely be surfaced
// via an API to a client. For the sake of consistency they should be in UTC.
// Events are returned in order, with the most recent event being first. Objects
// related to a date are sorted in ascending order by their ID.
func (f *forecasterBase) GetForecast(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) (Forecast, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"start":    start,
		"end":      end,
		"timezone": timezone.String(),
	}
	forecast := Forecast{
		StartingTime:    start,
		EndingTime:      end,
		StartingBalance: f.currentBalance,
		EndingBalance:   0,
		Events:          make([]Event, 0),
	}

	// Gather all of our spending events
	events := make([]SpendingEvent, 0)
	for i := range f.spending {
		spending := f.spending[i]
		result, err := spending.GetSpendingEventsBetween(span.Context(), start, end, timezone)
		if err != nil {
			return forecast, err
		}
		events = append(events, result...)
	}

	linq.From(events).
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
			spendingItems := make([]SpendingEvent, len(group.Group))
			fundingMap := map[ID[FundingSchedule]]FundingEvent{}
			var delta int64 = 0
			var transaction int64
			var contribution int64
			for i, item := range group.Group {
				spendingEvent := item.(SpendingEvent)
				delta += spendingEvent.ContributionAmount
				contribution += spendingEvent.ContributionAmount
				delta -= spendingEvent.TransactionAmount
				transaction += spendingEvent.TransactionAmount

				for x, funding := range spendingEvent.Funding {
					funding.Date = funding.Date.UTC()
					funding.OriginalDate = funding.OriginalDate.UTC()
					spendingEvent.Funding[x] = funding
					if _, ok := fundingMap[funding.FundingScheduleId]; !ok {
						fundingMap[funding.FundingScheduleId] = funding
					}
				}
				spendingEvent.Date = spendingEvent.Date.UTC()

				spendingItems[i] = spendingEvent
			}

			fundingItems := make([]FundingEvent, 0, len(fundingMap))
			for _, item := range fundingMap {
				fundingItems = append(fundingItems, item)
			}

			// Sort the items in the result set. This way the output is consistent no matter what. The same inputs will
			// result in the exact same outputs in the exact same output order every time.
			sort.Slice(spendingItems, func(i, j int) bool {
				return spendingItems[i].SpendingId < spendingItems[j].SpendingId
			})
			sort.Slice(fundingItems, func(i, j int) bool {
				return fundingItems[i].FundingScheduleId < fundingItems[j].FundingScheduleId
			})

			return Event{
				Date:         date.UTC(),
				Delta:        delta,
				Balance:      0,
				Transaction:  transaction,
				Contribution: contribution,
				Spending:     spendingItems,
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

	return forecast, nil
}

func (f *forecasterBase) GetAverageContribution(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) (int64, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"start":    start,
		"end":      end,
		"timezone": timezone.String(),
	}
	forecast, err := f.GetForecast(span.Context(), start, end, timezone)
	if err != nil {
		return 0, err
	}

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

	return popularContribution, nil
}

func (f *forecasterBase) GetNextContribution(
	ctx context.Context,
	start time.Time,
	fundingScheduleId ID[FundingSchedule],
	timezone *time.Location,
) (int64, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	end, err := f.funding[fundingScheduleId].GetNextFundingEventAfter(
		span.Context(),
		start,
		timezone,
	)
	if err != nil {
		return 0, err
	}

	span.Data = map[string]interface{}{
		"start":             start,
		"end":               end,
		"fundingScheduleId": fundingScheduleId,
		"timezone":          timezone.String(),
	}

	forecast, err := f.GetForecast(
		span.Context(),
		start, end.Date.AddDate(0, 0, 1),
		timezone,
	)
	if err != nil {
		return 0, err
	}

	for _, event := range forecast.Events {
		if len(event.Funding) == 0 {
			continue
		}

		return event.Contribution, nil
	}

	return 0, nil
}
