package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
)

type FundingDay struct {
	Date   time.Time      `json:"date"`
	Events []FundingEvent `json:"events"`
}

type FundingEvent struct {
	Date              time.Time `json:"date"`
	OriginalDate      time.Time `json:"originalDate"`
	WeekendAvoided    bool      `json:"weekendAvoided"`
	FundingScheduleId uint64    `json:"fundingScheduleId"`
}

var (
	_ FundingInstructions = &fundingScheduleBase{}
	_ FundingInstructions = &multipleFundingInstructions{}
)

type FundingInstructions interface {
	GetNFundingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingEvent
	GetNFundingDaysAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingDay
	GetNumberOfFundingDaysBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64
	GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent
	GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent
	GetNextFundingDayAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingDay
}

type fundingScheduleBase struct {
	fundingSchedule models.FundingSchedule
}

func NewFundingScheduleFundingInstructions(fundingSchedule models.FundingSchedule) FundingInstructions {
	return &fundingScheduleBase{
		fundingSchedule: fundingSchedule,
	}
}

func (f *fundingScheduleBase) needsWeekendAdjust(input time.Time) bool {
	if f.fundingSchedule.ExcludeWeekends {
		switch input.Weekday() {
		case time.Sunday, time.Saturday:
			return true
		default:
			return false
		}
	}

	return false
}

func (f *fundingScheduleBase) adjustForWeekendMaybe(input time.Time) (output time.Time, weekendAvoided bool) {
	output = input
	if f.fundingSchedule.ExcludeWeekends {
		switch output.Weekday() {
		case time.Sunday:
			// If it lands on a sunday then subtract 2 days to put the contribution date on a Friday.
			output = output.AddDate(0, 0, -2)
			weekendAvoided = true
		case time.Saturday:
			// If it lands on a sunday then subtract 1 day to put the contribution date on a Friday.
			output = output.AddDate(0, 0, -1)
			weekendAvoided = true
		default:
			weekendAvoided = false
		}
	}
	return output, weekendAvoided
}

func (f *fundingScheduleBase) GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent {
	input = util.MidnightInLocal(input, timezone)
	rule := f.fundingSchedule.Rule.RRule
	var nextContributionDate time.Time
	if f.fundingSchedule.NextOccurrence.IsZero() {
		// Hack to determine the previous contribution date before we figure out the next one.
		rule.DTStart(input.AddDate(-1, 0, 0))
		nextContributionDate = util.MidnightInLocal(rule.Before(input, false), timezone)
	} else {
		rule.DTStart(f.fundingSchedule.NextOccurrence)
		nextContributionDate = util.MidnightInLocal(f.fundingSchedule.NextOccurrence, timezone)
	}
	if input.Before(nextContributionDate) {
		// TODO this is kind of a hack to make the adjusted date and weekend avoided show properly when we are working
		//   off the else clause above. Idk how to improve it right now, but this gives us the correct data.
		realContributionDate := util.MidnightInLocal(rule.After(input, false), timezone)
		return FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			WeekendAvoided:    f.needsWeekendAdjust(realContributionDate),
			Date:              nextContributionDate,
			OriginalDate:      realContributionDate,
		}
	}

	nextContributionRule := rule

	// Force the start of the rule to be the next contribution date. This fixes a bug where the rule would increment
	// properly, but would include the current timestamp in that increment causing incorrect comparisons below. This
	// makes sure that the rule will increment in the user's timezone as intended.
	nextContributionRule.DTStart(nextContributionDate)

	// Keep track of an un-adjusted next contribution date. Because we might subtract days to account for early
	// funding, we need to make sure we are still incrementing relative to the _real_ contribution dates. Not the
	// adjusted ones.
	actualNextContributionDate := nextContributionDate
	weekendAvoided := false
	for !nextContributionDate.After(input) {
		// If the next contribution date is not after now, then increment it.
		nextContributionDate = nextContributionRule.After(actualNextContributionDate, false)
		// Store the real contribution date for later use.
		actualNextContributionDate = nextContributionDate

		// If we are excluding weekends, and the next contribution date falls on a weekend; then we need to adjust the
		// date to the previous business day.
		nextContributionDate, weekendAvoided = f.adjustForWeekendMaybe(nextContributionDate)
		nextContributionDate = util.MidnightInLocal(nextContributionDate, timezone)
	}

	return FundingEvent{
		FundingScheduleId: f.fundingSchedule.FundingScheduleId,
		WeekendAvoided:    weekendAvoided,
		Date:              nextContributionDate,
		OriginalDate:      actualNextContributionDate,
	}
}

func (f *fundingScheduleBase) GetNFundingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingEvent {
	events := make([]FundingEvent, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			events[i] = f.GetNextFundingEventAfter(ctx, input, timezone)
			continue
		}

		events[i] = f.GetNextFundingEventAfter(ctx, events[i-1].Date, timezone)
	}

	return events
}

func (f *fundingScheduleBase) GetNFundingDaysAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingDay {
	days := make([]FundingDay, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			days[i] = f.GetNextFundingDayAfter(ctx, input, timezone)
			continue
		}

		days[i] = f.GetNextFundingDayAfter(ctx, days[i-1].Date, timezone)
	}

	return days
}

func (f *fundingScheduleBase) GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent {
	rule := f.fundingSchedule.Rule.RRule
	// Make sure that the rule is using the timezone of the dates provided. This is an easy way to force that.
	// We also need to truncate the hours on the start time. To make sure that we are operating relative to
	// midnight.
	dtStart := util.MidnightInLocal(start, timezone)
	rule.DTStart(dtStart)
	// We need this to be inclusive because we want to account for the scenario when we could contribute to a spending
	// object on the day it is actually due. However, we do not want to include the start date as it should have been
	// accounted for elsewhere (hard to explain). So we do inclusive here, but we exclude any instance that happens on
	// or before the provided start date below.
	items := rule.Between(start, end, true)
	events := make([]FundingEvent, 0, len(items))
	for _, item := range items {
		adjustedDate, avoidedWeekend := f.adjustForWeekendMaybe(item)
		// If we are on the funding day then we need to exclude it _after_ the weekend has been adjusted for.
		if adjustedDate.Equal(start) || adjustedDate.Before(start) {
			continue
		}
		events = append(events, FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			Date:              adjustedDate,
			OriginalDate:      item,
			WeekendAvoided:    avoidedWeekend,
		})
	}
	return events
}

func (f *fundingScheduleBase) GetFundingDaysBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingDay {
	// Because a single funding schedule cannot have multiple events on a single day we can streamline this quite a bit.
	events := f.GetFundingEventsBetween(ctx, start, end, timezone)
	days := make([]FundingDay, len(events))
	for i, event := range events {
		days[i] = FundingDay{
			Date: event.Date,
			Events: []FundingEvent{
				event,
			},
		}
	}
	return days
}

func (f *fundingScheduleBase) GetNumberOfFundingDaysBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64 {
	return int64(len(f.GetFundingDaysBetween(ctx, start, end, timezone)))
}

func (f *fundingScheduleBase) GetNextFundingDayAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingDay {
	event := f.GetNextFundingEventAfter(ctx, input, timezone)
	return FundingDay{
		Date: event.Date,
		Events: []FundingEvent{
			event,
		},
	}
}

type multipleFundingInstructions struct {
	instructions []FundingInstructions
}

func NewMultipleFundingInstructions(instructions []FundingInstructions) FundingInstructions {
	myownsanity.Assert(len(instructions) > 0, "Must provide at least a single funding instruction")
	return &multipleFundingInstructions{
		instructions: instructions,
	}
}

func (m *multipleFundingInstructions) GetNFundingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingEvent {
	panic("GetNFundingEventsAfter should not be used with the multiple funding instructions interface, use GetNFundingDaysAfter")
}

func (m *multipleFundingInstructions) GetNFundingDaysAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingDay {
	days := make([]FundingDay, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			days[i] = m.GetNextFundingDayAfter(ctx, input, timezone)
			continue
		}

		days[i] = m.GetNextFundingDayAfter(ctx, days[i-1].Date, timezone)
	}

	return days
}

func (m *multipleFundingInstructions) GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent {
	result := make([]FundingEvent, 0)
	for _, instruction := range m.instructions {
		result = append(result, instruction.GetFundingEventsBetween(ctx, start, end, timezone)...)
	}

	return result
}

// GetFundingDaysBetween returns an array of funding days for the range provided. The array is returned unordered and
// must be sorted by the caller if it is important.
func (m *multipleFundingInstructions) GetFundingDaysBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingDay {
	events := m.GetFundingEventsBetween(ctx, start, end, timezone)
	result := map[int64][]FundingEvent{}
	for _, event := range events {
		date := event.Date.Unix()
		items, ok := result[date]
		if !ok {
			items = make([]FundingEvent, 0)
		}
		items = append(items, event)
		result[date] = items
	}
	days := make([]FundingDay, 0, len(result))
	for date, items := range result {
		days = append(days, FundingDay{
			Date:   time.Unix(date, 0),
			Events: items,
		})
	}
	return days
}

func (m *multipleFundingInstructions) GetNumberOfFundingDaysBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64 {
	days := m.GetFundingDaysBetween(ctx, start, end, timezone)
	return int64(len(days))
}

func (m *multipleFundingInstructions) GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent {
	panic("GetNextFundingEventAfter should not be used with the multiple funding instructions interface, use GetNextFundingDayAfter instead")
}

func (m *multipleFundingInstructions) GetNextFundingDayAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingDay {
	var earliest time.Time
	result := make([]FundingEvent, 0, len(m.instructions))
	for _, instruction := range m.instructions {
		next := instruction.GetNextFundingEventAfter(ctx, input, timezone)

		switch {
		case earliest.IsZero():
			earliest = next.Date
			fallthrough // Add the current one to the array.
		case next.Date.Equal(earliest):
			result = append(result, next)
		case next.Date.Before(earliest):
			// If we find one before the earliest we've seen then we need to clear the result array and only add this one.
			earliest = next.Date
			result = []FundingEvent{
				next,
			}
			continue
		}
	}

	myownsanity.Assert(!earliest.IsZero(), "The earliest next contribution cannot be zero, something is wrong with the provided funding instructions.")

	return FundingDay{
		Date:   earliest,
		Events: result,
	}
}
