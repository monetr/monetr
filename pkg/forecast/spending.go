package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/teambition/rrule-go"
)

type SpendingEvent struct {
	Date               time.Time      `json:"date"`
	TransactionAmount  int64          `json:"transactionAmount"`
	ContributionAmount int64          `json:"contributionAmount"`
	RollingAllocation  int64          `json:"rollingAllocation"`
	Funding            []FundingEvent `json:"funding"`
	SpendingId         uint64         `json:"spendingId"`
}

var (
	_ SpendingInstructions = &spendingInstructionBase{}
)

type SpendingInstructions interface {
	GetNextNSpendingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []SpendingEvent
	GetSpendingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []SpendingEvent
}

type spendingInstructionBase struct {
	log      *logrus.Entry
	spending models.Spending
	funding  FundingInstructions
}

func NewSpendingInstructions(log *logrus.Entry, spending models.Spending, fundingInstructions FundingInstructions) SpendingInstructions {
	return &spendingInstructionBase{
		log:      log,
		spending: spending,
		funding:  fundingInstructions,
	}
}

func (s *spendingInstructionBase) GetSpendingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []SpendingEvent {
	events := make([]SpendingEvent, 0)

	log := s.log.
		WithContext(ctx).
		WithFields(logrus.Fields{
			"start":    start,
			"end":      end,
			"timezone": timezone.String(),
		})
	for i := 0; ; i++ {
		ilog := log.WithField("i", i)
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				ilog.
					WithError(err).
					Error("timed out while trying to determine spending events between dates")
				crumbs.Error(ctx, "Timed out while trying to determine spending events between dates", "forecast", map[string]interface{}{
					"start":    start,
					"end":      end,
					"timezone": timezone.String(),
					"i":        i,
				})
				panic(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return events
		default:
			// Do nothing
		}

		var event *SpendingEvent
		afterDate := start
		allocation := s.spending.CurrentAmount
		if i > 0 {
			afterDate = events[i-1].Date
			allocation = events[i-1].RollingAllocation
		}

		ilog = ilog.WithField("after", start)
		event = s.getNextSpendingEventAfter(ctx, afterDate, timezone, allocation)

		// No event returned means there are no more.
		if event == nil {
			ilog.Trace("no more spending events to calculate")
			break
		}

		if event.Date.After(end) {
			ilog.Trace("calculated next spending event, but it happens after the end window, discarding and exiting calculation")
			break
		}

		// This should not happen, and to some degree there are now tests to prove this. But if it does happen that means
		// there has been a regression. Send something to sentry with some contextual data so it can be diagnosted.
		if !event.Date.After(afterDate) {
			ilog.Error("calculated a spending event that does not come after the after date specified! there is a bug somewhere!!!")
			crumbs.IndicateBug(ctx, "Calculated a spending event that does not come after the after date specified", map[string]interface{}{
				"spending":   s.spending,
				"afterDate":  afterDate,
				"start":      start,
				"end":        end,
				"allocation": allocation,
				"i":          i,
				"event":      event,
				"timezone":   timezone.String(),
				"count":      len(events),
			})
			panic("calculated a spending event that does not come after the after date specified")
		}

		ilog.Trace("calculated next spending event, adding to return set")

		events = append(events, *event)
	}

	log.WithField("count", len(events)).Trace("returning calculated events")

	return events
}

func (s spendingInstructionBase) GetNextNSpendingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) []SpendingEvent {
	events := make([]SpendingEvent, 0, n)
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				panic(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return events
		default:
			// Do nothing
		}

		var event *SpendingEvent
		if i == 0 {
			event = s.getNextSpendingEventAfter(ctx, input, timezone, s.spending.CurrentAmount)
		} else {
			event = s.getNextSpendingEventAfter(ctx, events[i-1].Date, timezone, events[i-1].RollingAllocation)
		}

		// No event returned means there are no more.
		if event == nil {
			break
		}

		events = append(events, *event)
	}

	return events
}

func (s *spendingInstructionBase) GetRecurrencesBetween(ctx context.Context, start, end time.Time, timezone *time.Location) []time.Time {
	switch s.spending.SpendingType {
	case models.SpendingTypeExpense:
		rule := s.spending.RecurrenceRule.RRule
		if s.spending.DateStarted.IsZero() {
			dtMidnight := util.MidnightInLocal(start, timezone)
			rule.DTStart(dtMidnight)
		} else {
			dateStarted := s.spending.DateStarted
			corrected := dateStarted.In(timezone)
			rule.DTStart(corrected)
		}
		// This little bit is really confusing. Basically we want to know how many times this spending boi happens
		// before the specified end date. This can include the start date, but we want to exclude the end date. This is
		// because this function is **INTENDED** to be called with the start being now or the next funding event, and
		// end being the next funding event immediately after that. We can't control what happens after the later
		// funding event, so we need to know how much will be spent before then, so we know how much to allocate.
		items := rule.Between(start, end.Add(-1 * time.Second), true)
		return items
	case models.SpendingTypeGoal:
		if s.spending.NextRecurrence.After(start) && s.spending.NextRecurrence.Before(end) {
			return []time.Time{s.spending.NextRecurrence}
		}
		fallthrough
	default:
		return nil
	}
}

func (s *spendingInstructionBase) getNextSpendingEventAfter(ctx context.Context, input time.Time, timezone *time.Location, balance int64) *SpendingEvent {
	// If the spending object is paused then there wont be any events for it at all.
	if s.spending.IsPaused {
		return nil
	}

	input = util.MidnightInLocal(input, timezone)

	var rule *rrule.RRule
	if s.spending.RecurrenceRule != nil {
		// This is terrible and I hate it :tada:
		rule = &(*s.spending.RecurrenceRule).RRule
	}

	nextRecurrence := util.MidnightInLocal(s.spending.NextRecurrence, timezone)
	switch s.spending.SpendingType {
	case models.SpendingTypeOverflow:
		return nil
	case models.SpendingTypeGoal:
		// If we are working with a goal and it has already "completed" then there is nothing more to do, no more events
		// will come up for this spending object.
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			return nil
		}
	case models.SpendingTypeExpense:
		if rule == nil {
			panic("expense spending type must have a recurrence rule!")
		}

		// If we are working with a spending object, but the next recurrence is before our start time. Then figure out
		// what the next recurrence would be after the start time.
		if s.spending.DateStarted.IsZero() {
			rule.DTStart(nextRecurrence)
		} else {
			dateStarted := s.spending.DateStarted
			corrected := dateStarted.In(timezone)
			rule.DTStart(corrected)
		}
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			nextRecurrence = rule.After(input, false)
		}
	}

	var fundingFirst, fundingSecond FundingEvent
	{ // Get our next two funding events
		fundingEvents := s.funding.GetNFundingEventsAfter(ctx, 2, input, timezone)
		if len(fundingEvents) != 2 {
			// TODO, if there are multiple funding schedules and they land on the same day, this will happen.
			panic("invalid number of funding events returned;")
		}

		fundingFirst, fundingSecond = fundingEvents[0], fundingEvents[1]
	}

	// The number of times this item will be spent before it receives funding again. This is considered the current
	// funding period. This is used to determine if the spending is currently behind. As the total amount that will be
	// spent must be <= the amount currently allocated to this spending item. If it is not then there will not be enough
	// funds to cover each spending event between now and the next funding event.
	eventsBeforeFirst := int64(len(s.GetRecurrencesBetween(ctx, input, fundingFirst.Date, timezone)))
	// The number of times this item will be spent in the subsequent funding period. This is used to determine how much
	// needs to be allocated at the beginning of the next funding period.
	eventsBeforeSecond := int64(len(s.GetRecurrencesBetween(ctx, fundingFirst.Date, fundingSecond.Date, timezone)))

	// The amount of funds needed for each individual spending event.
	perSpendingAmount := s.spending.TargetAmount
	// The amount of funds currently allocated towards this spending item. This is not increased until the next funding
	// event, or the user transfers funds to this spending item.

	event := SpendingEvent{
		Date:               time.Time{},
		TransactionAmount:  0,
		ContributionAmount: 0,
		RollingAllocation:  balance,
		Funding:            make([]FundingEvent, 0),
		SpendingId:         s.spending.SpendingId,
	}

	// The total contribution amount is the amount of money that needs to be allocated to this spending item during the
	// next funding event in order to cover all the spending events that will happen between then and the subsequent
	// funding event.
	var totalContributionAmount int64

	// If there are spending events in the next funding period then we need to make sure that we calculate for those.
	if eventsBeforeSecond > 0 {
		// We need to subtract the spending that will happen before the next period though.
		// We have $5 allocated but between now and the next funding we need to spend $5. So we cannot take the $5 we
		// currently have into account when we calculate how much will be needed for the next funding event.
		amountAfterCurrentSpending := myownsanity.Max(0, balance-(perSpendingAmount*eventsBeforeFirst))
		// The total amount we need is determined by how many times we will need the target amount during the next period
		// between funding events multiplied by how much each spending event costs.
		// If the current spending object is over-allocated for this funding period and the next funding period then
		// this can result in a negative contribution amount. Because we would be subtracting more than the calculated
		// amount that we need.
		nextSpendingPeriodTotal := perSpendingAmount * eventsBeforeSecond
		// By taking the min of the amount we will have allocated and the amount needed. We can safely arrive at a 0
		// contribution amount when we are over-allocated.
		totalContributionAmount = nextSpendingPeriodTotal - myownsanity.Min(amountAfterCurrentSpending, nextSpendingPeriodTotal)
	} else {
		// Otherwise we can simply look at how much we need vs how much we already have.
		amountNeeded := myownsanity.Max(0, perSpendingAmount-balance)
		// And how many times we will have a funding event before our due date.
		numberOfContributions := s.funding.GetNumberOfFundingEventsBetween(ctx, input, nextRecurrence, timezone)
		// Then determine how much we would need at each of those funding events.
		totalContributionAmount = amountNeeded / myownsanity.Max(1, numberOfContributions)
	}

	switch {
	case fundingFirst.Date.Before(nextRecurrence):
		// The next event will be a contribution.
		event.Date = fundingFirst.Date
		event.ContributionAmount = totalContributionAmount
		event.Funding = []FundingEvent{
			fundingFirst,
		}
		event.RollingAllocation = event.RollingAllocation + totalContributionAmount
	case nextRecurrence.Before(fundingFirst.Date):
		// The next event will be a transaction.
		event.Date = nextRecurrence
		event.TransactionAmount = s.spending.TargetAmount
		// NOTE At the time of writing this, event.RollingAllocation is not being defined anywhere. But this is
		// ultimately what the math will end up being once it is defined, and we calculate the effects of a transaction.
		event.RollingAllocation = event.RollingAllocation - s.spending.TargetAmount
	case nextRecurrence.Equal(fundingFirst.Date):
		// The next event will be both a contribution and a transaction.
		event.Date = nextRecurrence
		event.ContributionAmount = totalContributionAmount
		event.TransactionAmount = s.spending.TargetAmount
		// NOTE At the time of writing this, event.RollingAllocation is not being defined anywhere. But this is
		// ultimately what the math will end up being once it is defined, and we calculate the effects of a transaction.
		event.RollingAllocation = (event.RollingAllocation + totalContributionAmount) - s.spending.TargetAmount
		event.Funding = []FundingEvent{
			fundingFirst,
		}
	}

	return &event
}
