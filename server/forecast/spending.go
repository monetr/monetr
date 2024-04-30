package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/teambition/rrule-go"
)

type SpendingEvent struct {
	// Date is the timestamp of the event. There will only be a single spending
	// event per spending object per day max.
	Date time.Time `json:"date"`
	// TransactionAmount represents the amount of money removed from this spending
	// object's allocation at this time.
	TransactionAmount int64 `json:"transactionAmount"`
	// ContributionAmount represents the amount of money contributed to this
	// spending object at this time.
	ContributionAmount int64 `json:"contributionAmount"`
	// OverspendAmount represents the excess amount spent on an event beyond what
	// was currently allocated to the spending object at that time.
	OverspendAmount int64 `json:"overspendAmount"`
	// RollingAllocation represents the amount of funds allocated towards this
	// spending object at this moment in time. This is after contribution and
	// spending has been taken into account (in that order) for this moment in
	// time.
	RollingAllocation int64 `json:"rollingAllocation"`
	// IsBehind will be true if there was not enough funds allocated to cover this
	// spending objects target amount at this time.
	IsBehind bool `json:"isBehind"`
	// Funding is an array of funding events that contributed to this event's
	// contribution amount.
	Funding []FundingEvent `json:"funding"`
	// SpendingId is the unique identifier of this spending object.
	SpendingId uint64 `json:"spendingId"`
}

var (
	_ SpendingInstructions = &spendingInstructionBase{}
)

type SpendingInstructions interface {
	// GetNextNSpendingEventsAfter will return an array of events for this
	// spending object after the specified timestamp. If the spending object is a
	// goal then fewer than n events may be returned depending on its current
	// progress. For expense spending objects n events will always be returned.
	GetNextNSpendingEventsAfter(
		ctx context.Context,
		n int,
		input time.Time,
		timezone *time.Location,
	) ([]SpendingEvent, error)
	// GetSpendingEventsBetween will return all of the events for this spending
	// object between the two timestamps provided.
	GetSpendingEventsBetween(
		ctx context.Context,
		start, end time.Time,
		timezone *time.Location,
	) ([]SpendingEvent, error)
	// GetNextContributionEvent will return the next contribution that will be
	// made to this spending object if there is one. Goals may not have a next
	// contribution if the goal is complete or if the goal is past its target
	// date. In that case the returned event will have the input date and a $0
	// contribution.
	GetNextContributionEvent(
		ctx context.Context,
		input time.Time,
		timezone *time.Location,
	) (SpendingEvent, error)
}

type spendingInstructionBase struct {
	log      *logrus.Entry
	spending models.Spending
	funding  FundingInstructions
}

func NewSpendingInstructions(
	log *logrus.Entry,
	spending models.Spending,
	fundingInstructions FundingInstructions,
) SpendingInstructions {
	return &spendingInstructionBase{
		log:      log,
		spending: spending,
		funding:  fundingInstructions,
	}
}

func (s *spendingInstructionBase) GetSpendingEventsBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) ([]SpendingEvent, error) {
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

				return events, errors.WithStack(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return events, nil
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

	return events, nil
}

func (s spendingInstructionBase) GetNextNSpendingEventsAfter(
	ctx context.Context,
	n int,
	input time.Time,
	timezone *time.Location,
) ([]SpendingEvent, error) {
	events := make([]SpendingEvent, 0, n)
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return events, errors.WithStack(err)
			}

			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return events, nil
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

	return events, nil
}

func (s spendingInstructionBase) GetNextContributionEvent(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
) (SpendingEvent, error) {
	balance := s.spending.CurrentAmount
	fundingEvents := s.funding.GetNFundingEventsAfter(ctx, 1, input, timezone)
	if len(fundingEvents) == 0 {
		return SpendingEvent{
			Date:               input,
			TransactionAmount:  0,
			ContributionAmount: 0,
			RollingAllocation:  balance,
			Funding:            fundingEvents,
			SpendingId:         s.spending.SpendingId,
		}, nil
	}

	end := fundingEvents[0].Date
	start := input

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return SpendingEvent{}, errors.WithStack(err)
			}

			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return SpendingEvent{}, nil
		default:
			// Do nothing
		}

		event := s.getNextSpendingEventAfter(ctx, start, timezone, balance)
		if event == nil || event.Date.After(end) {
			return SpendingEvent{
				Date:               input,
				TransactionAmount:  0,
				ContributionAmount: 0,
				RollingAllocation:  balance,
				Funding:            fundingEvents,
				SpendingId:         s.spending.SpendingId,
			}, nil
		}

		if event.ContributionAmount > 0 {
			return *event, nil
		}

		// Bump the timestamp and do it again
		start = event.Date
	}
}

func (s *spendingInstructionBase) getRecurrencesBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) []time.Time {
	switch s.spending.SpendingType {
	case models.SpendingTypeExpense:
		rule := s.spending.RuleSet.Set
		rule.DTStart(rule.GetDTStart().In(timezone))

		// This little bit is really confusing. Basically we want to know how many
		// times this spending boi happens before the specified end date. This can
		// include the start date, but we want to exclude the end date. This is
		// because this function is **INTENDED** to be called with the start being
		// now or the next funding event, and end being the next funding event
		// immediately after that. We can't control what happens after the later
		// funding event, so we need to know how much will be spent before then, so
		// we know how much to allocate.
		items := rule.Between(start, end.Add(-1*time.Second), true)
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

func (s *spendingInstructionBase) getNextSpendingEventAfter(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
	balance int64,
) *SpendingEvent {
	// If the spending object is paused then there wont be any events for it at
	// all.
	if s.spending.IsPaused {
		return nil
	}

	input = util.Midnight(input, timezone)

	var rule *rrule.Set
	if s.spending.RuleSet != nil {
		// This is terrible and I hate it :tada:
		rule = &(*s.spending.RuleSet).Set
	}

	nextRecurrence := util.Midnight(s.spending.NextRecurrence, timezone)
	switch s.spending.SpendingType {
	case models.SpendingTypeOverflow:
		return nil
	case models.SpendingTypeGoal:
		// If we are working with a goal and it has already "completed" then there
		// is nothing more to do, no more events will come up for this spending
		// object.
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			return nil
		}
	case models.SpendingTypeExpense:
		if rule == nil {
			panic("expense spending type must have a recurrence rule!")
		}

		rule.DTStart(rule.GetDTStart().In(timezone))
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			nextRecurrence = rule.After(input, false)
		}
	}

	var fundingFirst, fundingSecond FundingEvent
	{ // Get our next two funding events
		fundingEvents := s.funding.GetNFundingEventsAfter(ctx, 2, input, timezone)
		if len(fundingEvents) != 2 {
			// TODO, if there are multiple funding schedules and they land on the same
			// day, this will happen.
			panic("invalid number of funding events returned;")
		}

		fundingFirst, fundingSecond = fundingEvents[0], fundingEvents[1]
	}

	// The number of times this item will be spent before it receives funding
	// again. This is considered the current funding period. This is used to
	// determine if the spending is currently behind. As the total amount that
	// will be spent must be <= the amount currently allocated to this spending
	// item. If it is not then there will not be enough funds to cover each
	// spending event between now and the next funding event.
	eventsBeforeFirst := int64(len(s.getRecurrencesBetween(ctx, input, fundingFirst.Date, timezone)))
	// The number of times this item will be spent in the subsequent funding
	// period. This is used to determine how much needs to be allocated at the
	// beginning of the next funding period.
	eventsBeforeSecond := int64(len(s.getRecurrencesBetween(ctx, fundingFirst.Date, fundingSecond.Date, timezone)))

	// The amount of funds needed for each individual spending event.
	var perSpendingAmount int64
	switch s.spending.SpendingType {
	case models.SpendingTypeExpense:
		perSpendingAmount = s.spending.TargetAmount
	case models.SpendingTypeGoal:
		// If we are working with a goal then we need to subtract the amount we have
		// already used from the goal. This is because a goal could have its funds
		// spent from it throughout the life of the goal. But we don't want to
		// change the target. We assume that spending from a goal is progress
		// towards that goal. Basically for a completed goal we don't need to make
		// any contributions to it.
		perSpendingAmount = myownsanity.Max(s.spending.TargetAmount-s.spending.UsedAmount, 0)
	}
	// The amount of funds currently allocated towards this spending item. This is
	// not increased until the next funding event, or the user transfers funds to
	// this spending item.

	event := SpendingEvent{
		Date:               time.Time{},
		TransactionAmount:  0,
		ContributionAmount: 0,
		RollingAllocation:  balance,
		Funding:            make([]FundingEvent, 0),
		SpendingId:         s.spending.SpendingId,
	}

	// The total contribution amount is the amount of money that needs to be
	// allocated to this spending item during the next funding event in order to
	// cover all the spending events that will happen between then and the
	// subsequent funding event.
	var totalContributionAmount int64

	// If we are full then assume we are empty for the next calculation. This
	// basically handles a spending object that is past its spending date but
	// hasn't been spent yet.
	if balance >= s.spending.TargetAmount && fundingFirst.Date.Before(nextRecurrence) {
		balance -= s.spending.TargetAmount
	}

	// If there are spending events in the next funding period then we need to
	// make sure that we calculate for those.
	if eventsBeforeSecond > 0 {
		// We need to subtract the spending that will happen before the next period
		// though. We have $5 allocated but between now and the next funding we need
		// to spend $5. So we cannot take the $5 we currently have into account when
		// we calculate how much will be needed for the next funding event.
		amountAfterCurrentSpending := myownsanity.Max(0, balance-(perSpendingAmount*eventsBeforeFirst))
		// The total amount we need is determined by how many times we will need the
		// target amount during the next period between funding events multiplied by
		// how much each spending event costs. If the current spending object is
		// over-allocated for this funding period and the next funding period then
		// this can result in a negative contribution amount. Because we would be
		// subtracting more than the calculated amount that we need.
		nextSpendingPeriodTotal := perSpendingAmount * eventsBeforeSecond
		// By taking the min of the amount we will have allocated and the amount
		// needed. We can safely arrive at a 0 contribution amount when we are
		// over-allocated.
		totalContributionAmount = nextSpendingPeriodTotal - myownsanity.Min(amountAfterCurrentSpending, nextSpendingPeriodTotal)
	} else {
		// Otherwise we can simply look at how much we need vs how much we already
		// have.
		amountNeeded := myownsanity.Max(0, perSpendingAmount-balance)
		// And how many times we will have a funding event before our due date. But
		// we add one second to the input time because `input` might be the exact
		// same timestamp of a contribution that is today per se. By adding one
		// second to the input and keeping nextRecurrence the same we basically make
		// this query end inclusive but start exclusive.
		// Essentially: events > start && events <= end.
		numberOfContributions := s.funding.GetNumberOfFundingEventsBetween(ctx, input.Add(1*time.Second), nextRecurrence, timezone)
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
		event.IsBehind = s.spending.TargetAmount > event.RollingAllocation
		event.RollingAllocation = event.RollingAllocation - s.spending.TargetAmount
	case nextRecurrence.Equal(fundingFirst.Date):
		// The next event will be both a contribution and a transaction.
		event.Date = nextRecurrence
		event.ContributionAmount = totalContributionAmount
		event.TransactionAmount = s.spending.TargetAmount
		adjustedRollingAllocation := (event.RollingAllocation + totalContributionAmount)
		event.IsBehind = s.spending.TargetAmount > adjustedRollingAllocation
		event.RollingAllocation = adjustedRollingAllocation - s.spending.TargetAmount
		event.Funding = []FundingEvent{
			fundingFirst,
		}
	}

	if event.RollingAllocation < 0 {
		// If the rolling allocation goes negative that means we will be spending
		// from our free-to-use instead. We want to represent this as an overspend
		// instead of a negative rolling allocation.
		event.OverspendAmount = event.RollingAllocation * -1
		event.RollingAllocation = 0
	}

	return &event
}
