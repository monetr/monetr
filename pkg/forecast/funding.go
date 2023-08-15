package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/sirupsen/logrus"
)

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
	GetNumberOfFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64
	GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent
	GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent
}

type fundingScheduleBase struct {
	log             *logrus.Entry
	fundingSchedule models.FundingSchedule
}

func NewFundingScheduleFundingInstructions(log *logrus.Entry, fundingSchedule models.FundingSchedule) FundingInstructions {
	return &fundingScheduleBase{
		log:             log,
		fundingSchedule: fundingSchedule,
	}
}

func (f *fundingScheduleBase) GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent {
	input = util.Midnight(input, timezone)
	rule := f.fundingSchedule.Rule.RRule
	var nextContributionDate time.Time
	if f.fundingSchedule.NextOccurrence.IsZero() {
		// Hack to determine the previous contribution date before we figure out the next one.
		if f.fundingSchedule.DateStarted.IsZero() {
			rule.DTStart(input.AddDate(-1, 0, 0))
		} else {
			dateStarted := f.fundingSchedule.DateStarted
			corrected := dateStarted.In(timezone)
			rule.DTStart(corrected)
		}
		nextContributionDate = util.Midnight(rule.Before(input, false), timezone)
	} else {
		// If we have the date started defined on the funding schedule. Then use that so we can see the past and the future.
		if f.fundingSchedule.DateStarted.IsZero() {
			rule.DTStart(f.fundingSchedule.NextOccurrence)
		} else {
			dateStarted := f.fundingSchedule.DateStarted
			corrected := dateStarted.In(timezone)
			rule.DTStart(corrected)
		}
		nextContributionDate = util.Midnight(f.fundingSchedule.NextOccurrence, timezone)
	}
	if input.Before(nextContributionDate) {
		// If now is before the already established next occurrence, then just return that.
		// This might be goofy if we want to test stuff in the distant past?
		return FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			WeekendAvoided:    false,
			Date:              nextContributionDate,
			OriginalDate:      nextContributionDate,
		}
	}

	nextContributionRule := rule

	// Force the start of the rule to be the next contribution date. This fixes a bug where the rule would increment
	// properly, but would include the current timestamp in that increment causing incorrect comparisons below. This
	// makes sure that the rule will increment in the user's timezone as intended.
	//nextContributionRule.DTStart(nextContributionDate)

	// Keep track of an un-adjusted next contribution date. Because we might subtract days to account for early
	// funding, we need to make sure we are still incrementing relative to the _real_ contribution dates. Not the
	// adjusted ones.
	actualNextContributionDate := nextContributionDate
	weekendAvoided := false
AfterLoop:
	for !nextContributionDate.After(input) {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				panic(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "funding", nil)
			break AfterLoop
		default:
			// Do nothing
		}

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
				weekendAvoided = true
			case time.Saturday:
				// If it lands on a sunday then subtract 1 day to put the contribution date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -1)
				weekendAvoided = true
			default:
				weekendAvoided = false
			}
		}

		nextContributionDate = util.Midnight(nextContributionDate, timezone)
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
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				f.log.
					WithContext(ctx).
					WithError(err).
					WithFields(logrus.Fields{
						"n":        n,
						"input":    input,
						"timezone": timezone.String(),
						"i":        i,
					}).
					Error("timed out while trying to determine N funding events after")
				crumbs.Error(ctx, "Timed out while trying to determine N funding events after", "forecast", map[string]interface{}{
					"n":        n,
					"input":    input,
					"timezone": timezone.String(),
					"i":        i,
				})
				panic(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "funding", nil)
			return events
		default:
			// Do nothing
		}

		if i == 0 {
			events[i] = f.GetNextFundingEventAfter(ctx, input, timezone)
			continue
		}

		events[i] = f.GetNextFundingEventAfter(ctx, events[i-1].Date, timezone)
	}

	return events
}

func (f *fundingScheduleBase) GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent {
	rule := f.fundingSchedule.Rule.RRule
	// Make sure that the rule is using the timezone of the dates provided. This is an easy way to force that.
	// We also need to truncate the hours on the start time. To make sure that we are operating relative to
	// midnight.
	if f.fundingSchedule.DateStarted.IsZero() {
		dtStart := util.Midnight(start, timezone)
		rule.DTStart(dtStart)
	} else {
		dateStarted := f.fundingSchedule.DateStarted
		corrected := dateStarted.In(timezone)
		rule.DTStart(corrected)
	}
	items := rule.Between(start, end, true)
	events := make([]FundingEvent, len(items))
	for i, item := range items {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				panic(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "funding", nil)
			return events
		default:
			// Do nothing
		}

		// TODO Implement the skip weekends here too.
		events[i] = FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			Date:              item,
			OriginalDate:      item,
		}
	}
	return events
}

func (f *fundingScheduleBase) GetNumberOfFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64 {
	return int64(len(f.GetFundingEventsBetween(ctx, start, end, timezone)))
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

func NewMultipleFundingInstructions(instructions []FundingInstructions) FundingInstructions {
	return &multipleFundingInstructions{
		instructions: instructions,
	}
}

func (m *multipleFundingInstructions) GetNFundingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []FundingEvent {
	events := make([]FundingEvent, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			events[i] = m.GetNextFundingEventAfter(ctx, input, timezone)
			continue
		}

		events[i] = m.GetNextFundingEventAfter(ctx, events[i-1].Date, timezone)
	}

	return events
}

func (m *multipleFundingInstructions) GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []FundingEvent {
	result := make([]FundingEvent, 0)
	for _, instruction := range m.instructions {
		result = append(result, instruction.GetFundingEventsBetween(ctx, start, end, timezone)...)
	}

	return result
}

func (m *multipleFundingInstructions) GetNumberOfFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) int64 {
	return int64(len(m.GetFundingEventsBetween(ctx, start, end, timezone)))
}

func (m *multipleFundingInstructions) GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) FundingEvent {
	var earliest FundingEvent
	for _, instruction := range m.instructions {
		if earliest.Date.IsZero() {
			earliest = instruction.GetNextFundingEventAfter(ctx, input, timezone)
			continue
		}

		// If one of our instructions happens before the earliest one we've seen, then use that one instead.
		if next := instruction.GetNextFundingEventAfter(ctx, input, timezone); next.Date.Before(earliest.Date) {
			earliest = next
		}
	}

	if earliest.Date.IsZero() {
		panic("the earliest next contribution cannot be zero, something is wrong with the provided instructions")
	}

	return earliest
}
