package forecast

import (
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
)

type FundingEvent struct {
	Date              time.Time
	FundingScheduleId uint64
}

var (
	_ FundingInstructions = &fundingScheduleBase{}
	_ FundingInstructions = &multipleFundingInstructions{}
)

type FundingInstructions interface {
	GetNFundingEventsAfter(n int, input time.Time, timezone *time.Location) []FundingEvent
	GetNumberOfFundingEventsBetween(start, end time.Time, timezone *time.Location) int64
	GetNextFundingEventAfter(input time.Time, timezone *time.Location) FundingEvent
	GetFundingEventsBetween(start, end time.Time, timezone *time.Location) []FundingEvent
}

type fundingScheduleBase struct {
	fundingSchedule models.FundingSchedule
}

func NewFundingScheduleFundingInstructions(fundingSchedule models.FundingSchedule) FundingInstructions {
	return &fundingScheduleBase{
		fundingSchedule: fundingSchedule,
	}
}

func (f *fundingScheduleBase) GetNextFundingEventAfter(input time.Time, timezone *time.Location) FundingEvent {
	input = input.In(timezone)
	rule := f.fundingSchedule.Rule.RRule
	var nextContributionDate time.Time
	if !f.fundingSchedule.NextOccurrence.IsZero() {
		nextContributionDate = util.MidnightInLocal(f.fundingSchedule.NextOccurrence, timezone)
	} else {
		// Hack to determine the previous contribution date before we figure out the next one.
		rule.DTStart(input.AddDate(-1, 0, 0))
		nextContributionDate = util.MidnightInLocal(rule.Before(input, false), timezone)
	}
	if input.Before(nextContributionDate) {
		// If now is before the already established next occurrence, then just return that.
		// This might be goofy if we want to test stuff in the distant past?
		return FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			Date:              nextContributionDate,
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
	for !nextContributionDate.After(input) {
		// If the next contribution date is not after now, then increment it.
		nextContributionDate = nextContributionRule.After(actualNextContributionDate, false)
		// Store the real contribution date for later use.
		actualNextContributionDate = nextContributionDate

		// If we are excluding weekends, and the next contribution date falls on a weekend; then we need to adjust the
		// date to the previous business day.
		if f.fundingSchedule.ExcludeWeekends {
			switch nextContributionDate.Weekday() {
			case time.Sunday:
				// If it lands on a sunday then subtract 2 days to put the contribution date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -2)
			case time.Saturday:
				// If it lands on a sunday then subtract 1 day to put the contribution date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -1)
			}
		}

		nextContributionDate = util.MidnightInLocal(nextContributionDate, timezone)
	}

	return FundingEvent{
		FundingScheduleId: f.fundingSchedule.FundingScheduleId,
		Date:              nextContributionDate,
	}
}

func (f *fundingScheduleBase) GetNFundingEventsAfter(n int, input time.Time, timezone *time.Location) []FundingEvent {
	events := make([]FundingEvent, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			events[i] = f.GetNextFundingEventAfter(input, timezone)
			continue
		}

		events[i] = f.GetNextFundingEventAfter(events[i-1].Date, timezone)
	}

	return events
}

func (f *fundingScheduleBase) GetFundingEventsBetween(start, end time.Time, timezone *time.Location) []FundingEvent {
	rule := f.fundingSchedule.Rule.RRule
	// Make sure that the rule is using the timezone of the dates provided. This is an easy way to force that.
	// We also need to truncate the hours on the start time. To make sure that we are operating relative to
	// midnight.
	dtStart := util.MidnightInLocal(start, timezone)
	rule.DTStart(dtStart)
	items := rule.Between(start, end, true)
	events := make([]FundingEvent, len(items))
	for i, item := range items {
		events[i] = FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			Date:              item,
		}
	}
	return events
}

func (f *fundingScheduleBase) GetNumberOfFundingEventsBetween(start, end time.Time, timezone *time.Location) int64 {
	return int64(len(f.GetFundingEventsBetween(start, end, timezone)))
}

// multipleFundingInstructions isn't in use yet. It's kind of a proof of concept that with the funding instruction
// interface, it is easy to wrap multiple funding schedules to appear as one. But it fails to address some things that I
// want to be able to offer.
//   - How does one differentiate between the multiple funding schedules? I could see a case where person A contributes X
//     amount, and person B contributes Y amount. For us to do this we would need to know which schedule we are processing
//     on a given day. But also, what if both of those schedules fall on the same day?
type multipleFundingInstructions struct {
	instructions []FundingInstructions
}

func (m *multipleFundingInstructions) GetNFundingEventsAfter(n int, input time.Time, timezone *time.Location) []FundingEvent {
	events := make([]FundingEvent, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			events[i] = m.GetNextFundingEventAfter(input, timezone)
			continue
		}

		events[i] = m.GetNextFundingEventAfter(events[i-1].Date, timezone)
	}

	return events
}

func (m *multipleFundingInstructions) GetFundingEventsBetween(start, end time.Time, timezone *time.Location) []FundingEvent {
	result := make([]FundingEvent, 0)
	for _, instruction := range m.instructions {
		result = append(result, instruction.GetFundingEventsBetween(start, end, timezone)...)
	}

	return result
}

func (m *multipleFundingInstructions) GetNumberOfFundingEventsBetween(start, end time.Time, timezone *time.Location) int64 {
	return int64(len(m.GetFundingEventsBetween(start, end, timezone)))
}

func (m *multipleFundingInstructions) GetNextFundingEventAfter(input time.Time, timezone *time.Location) FundingEvent {
	var earliest FundingEvent
	for _, instruction := range m.instructions {
		if earliest.Date.IsZero() {
			earliest = instruction.GetNextFundingEventAfter(input, timezone)
			continue
		}

		// If one of our instructions happens before the earliest one we've seen, then use that one instead.
		if next := instruction.GetNextFundingEventAfter(input, timezone); next.Date.Before(earliest.Date) {
			earliest = next
		}
	}

	if earliest.Date.IsZero() {
		panic("the earliest next contribution cannot be zero, something is wrong with the provided instructions")
	}

	return earliest
}
