package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type FundingEvent struct {
	Date              time.Time           `json:"date"`
	OriginalDate      time.Time           `json:"originalDate"`
	WeekendAvoided    bool                `json:"weekendAvoided"`
	FundingScheduleId ID[FundingSchedule] `json:"fundingScheduleId"`
}

var (
	_ FundingInstructions = &fundingScheduleBase{}
	_ FundingInstructions = &multipleFundingInstructions{}
)

type FundingInstructions interface {
	GetNFundingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) ([]FundingEvent, error)
	GetNumberOfFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) (int64, error)
	GetNextFundingEventAfter(ctx context.Context, input time.Time, timezone *time.Location) (FundingEvent, error)
	GetFundingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) ([]FundingEvent, error)
}

type fundingScheduleBase struct {
	log             *logrus.Entry
	ruleset         *RuleSet
	fundingSchedule FundingSchedule
}

func NewFundingScheduleFundingInstructions(
	log *logrus.Entry,
	fundingSchedule FundingSchedule,
) FundingInstructions {
	return &fundingScheduleBase{
		log:             log,
		ruleset:         fundingSchedule.RuleSet.Clone(),
		fundingSchedule: fundingSchedule,
	}
}

func (f *fundingScheduleBase) GetNextFundingEventAfter(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
) (FundingEvent, error) {
	input = util.Midnight(input, timezone)
	rule := f.ruleset
	// This does not change the timezone or the date start of the ruleset. It just corrects it. The date start is
	// normally stored in UTC so this just adjusts it to be the user's current timezone.
	rule.DTStart(rule.GetDTStart().In(timezone))
	var nextContributionDate time.Time
	if f.fundingSchedule.NextRecurrence.IsZero() {
		nextContributionDate = util.Midnight(rule.Before(input, false), timezone)
	} else {
		nextContributionDate = util.Midnight(f.fundingSchedule.NextRecurrence, timezone)
	}
	if input.Before(nextContributionDate) {
		// If now is before the already established next occurrence, then just return that.
		// This might be goofy if we want to test stuff in the distant past?
		return FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			WeekendAvoided:    false,
			Date:              nextContributionDate,
			OriginalDate:      nextContributionDate,
		}, nil
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
				return FundingEvent{}, errors.WithStack(err)
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
	}, nil
}

func (f *fundingScheduleBase) GetNFundingEventsAfter(
	ctx context.Context,
	n int,
	input time.Time,
	timezone *time.Location,
) ([]FundingEvent, error) {
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
				crumbs.Error(ctx, "Timed out while trying to determine N funding events after", "forecast", map[string]any{
					"n":        n,
					"input":    input,
					"timezone": timezone.String(),
					"i":        i,
				})
				return events, errors.WithStack(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "funding", nil)
			return events, nil
		default:
			// Do nothing
		}

		var err error
		if i == 0 {
			events[i], err = f.GetNextFundingEventAfter(ctx, input, timezone)
			if err != nil {
				return events, err
			}
			continue
		}

		events[i], err = f.GetNextFundingEventAfter(ctx, events[i-1].Date, timezone)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (f *fundingScheduleBase) GetFundingEventsBetween(
	ctx context.Context,
	start,
	end time.Time,
	timezone *time.Location,
) ([]FundingEvent, error) {
	rule := f.fundingSchedule.RuleSet.Set
	// Make sure that the rule is using the timezone of the dates provided. This is an easy way to force that.
	// We also need to truncate the hours on the start time. To make sure that we are operating relative to
	// midnight.
	rule.DTStart(rule.GetDTStart().In(timezone))

	// TODO Technically we should add 2 days to the "end" here if exclude weekend
	// is on. Lets say the end is the 30th, but the rule recurres on the 31st. If
	// the 31st is a sunday then the actual recurrence is the 29th and would fall
	// within the range. But because of the window here it wouldn't be included
	// properly.
	items := rule.Between(start, end, true)

	events := make([]FundingEvent, 0, len(items))
	for i := range items {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return nil, errors.WithStack(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "funding", nil)
			return events, nil
		default:
			// Do nothing
		}

		date := items[i]
		nextContributionDate := date
		weekendAvoided := false
		// If we are excluding weekends then do that. But don't do it if the actual
		// next recurrence for the funding schedule is hard set on a weekend. If its
		// hard set, then just use that date.
		if f.fundingSchedule.ExcludeWeekends && !date.Equal(f.fundingSchedule.NextRecurrence) {
			switch date.Weekday() {
			case time.Sunday:
				nextContributionDate = nextContributionDate.AddDate(0, 0, -2)
				weekendAvoided = true
			case time.Saturday:
				nextContributionDate = nextContributionDate.AddDate(0, 0, -1)
				weekendAvoided = true
			default:
				weekendAvoided = false
			}

			// If we have adjusted for a weekend and it was before the start time then
			// don't include it in the result set.
			if nextContributionDate.Before(start) {
				continue
			}
		}

		events = append(events, FundingEvent{
			FundingScheduleId: f.fundingSchedule.FundingScheduleId,
			Date:              nextContributionDate,
			OriginalDate:      date,
			WeekendAvoided:    weekendAvoided,
		})
	}
	return events, nil
}

func (f *fundingScheduleBase) GetNumberOfFundingEventsBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) (int64, error) {
	events, err := f.GetFundingEventsBetween(ctx, start, end, timezone)
	return int64(len(events)), err
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

func (m *multipleFundingInstructions) GetNFundingEventsAfter(
	ctx context.Context,
	n int,
	input time.Time,
	timezone *time.Location,
) ([]FundingEvent, error) {
	events := make([]FundingEvent, n)
	var err error
	for i := 0; i < n; i++ {
		if i == 0 {
			events[i], err = m.GetNextFundingEventAfter(ctx, input, timezone)
			if err != nil {
				return nil, err
			}
			continue
		}

		events[i], err = m.GetNextFundingEventAfter(ctx, events[i-1].Date, timezone)
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

func (m *multipleFundingInstructions) GetFundingEventsBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) ([]FundingEvent, error) {
	result := make([]FundingEvent, 0)
	for _, instruction := range m.instructions {
		events, err := instruction.GetFundingEventsBetween(ctx, start, end, timezone)
		if err != nil {
			return nil, err
		}
		result = append(result, events...)
	}

	return result, nil
}

func (m *multipleFundingInstructions) GetNumberOfFundingEventsBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) (int64, error) {
	events, err := m.GetFundingEventsBetween(ctx, start, end, timezone)
	return int64(len(events)), err
}

func (m *multipleFundingInstructions) GetNextFundingEventAfter(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
) (FundingEvent, error) {
	var earliest FundingEvent
	var err error
	for _, instruction := range m.instructions {
		if earliest.Date.IsZero() {
			earliest, err = instruction.GetNextFundingEventAfter(ctx, input, timezone)
			if err != nil {
				return earliest, err
			}
			continue
		}

		// If one of our instructions happens before the earliest one we've seen, then use that one instead.
		next, err := instruction.GetNextFundingEventAfter(ctx, input, timezone)
		if err != nil {
			return next, err
		}
		if next.Date.Before(earliest.Date) {
			earliest = next
		}
	}

	if earliest.Date.IsZero() {
		return earliest, errors.New("the earliest next contribution cannot be zero, something is wrong with the provided instructions")
	}

	return earliest, nil
}
