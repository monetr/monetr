package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SpendingEvent struct {
	Date               time.Time      `json:"date"`
	TransactionAmount  int64          `json:"transactionAmount"`
	ContributionAmount int64          `json:"contributionAmount"`
	RollingAllocation  int64          `json:"rollingAllocation"`
	Funding            []FundingEvent `json:"funding"`
	SpendingId         ID[Spending]   `json:"spendingId"`
}

var (
	_ SpendingInstructions = &spendingInstructionBase{}
)

type SpendingInstructions interface {
	GetNextNSpendingEventsAfter(ctx context.Context, n int, input time.Time, timezone *time.Location) ([]SpendingEvent, error)
	GetSpendingEventsBetween(ctx context.Context, start, end time.Time, timezone *time.Location) ([]SpendingEvent, error)
	GetNextInflowEventAfter(ctx context.Context, input time.Time, timezone *time.Location) (*SpendingEvent, error)
}

type spendingInstructionBase struct {
	log      *logrus.Entry
	ruleset  *RuleSet
	spending Spending
	funding  FundingInstructions
}

func NewSpendingInstructions(log *logrus.Entry, spending Spending, fundingInstructions FundingInstructions) SpendingInstructions {
	instructions := &spendingInstructionBase{
		log:      log,
		ruleset:  nil,
		spending: spending,
		funding:  fundingInstructions,
	}
	if spending.RuleSet != nil {
		instructions.ruleset = spending.RuleSet.Clone()
	}

	return instructions
}

func (s *spendingInstructionBase) GetSpendingEventsBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) ([]SpendingEvent, error) {
	events := make([]SpendingEvent, 0)

	var err error
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
				return nil, errors.WithStack(err)
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
		event, err = s.getNextSpendingEventAfter(ctx, afterDate, timezone, allocation)
		if err != nil {
			return nil, err
		}

		// No event returned means there are no more.
		if event == nil {
			ilog.Trace("no more spending events to calculate")
			break
		}

		if event.Date.After(end) {
			ilog.Trace("calculated next spending event, but it happens after the end window, discarding and exiting calculation")
			break
		}

		// This should not happen, and to some degree there are now tests to prove
		// this. But if it does happen that means there has been a regression. Send
		// something to sentry with some contextual data so it can be diagnosted.
		if !event.Date.After(afterDate) {
			// Don't log the name of the spending object, its nobodies business.
			s.spending.Name = "[REDACTED]"
			ilog.WithFields(logrus.Fields{
				"bug": true,
				"debug": logrus.Fields{
					"badEvent":      event,
					"previousEvent": events[i-1],
					"spending":      s.spending,
					"afterDate":     afterDate,
					"timezone":      timezone,
					"allocation":    allocation,
					"i":             i,
					"count":         len(events),
					"start":         start,
					"end":           end,
				},
			}).Error("calculated a spending event that does not come after the after date specified! there is a bug somewhere!!!")

			// This might not make it into sentry because of sampling :(
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
	var err error
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return nil, errors.WithStack(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return events, nil
		default:
			// Do nothing
		}

		var event *SpendingEvent
		if i == 0 {
			event, err = s.getNextSpendingEventAfter(ctx, input, timezone, s.spending.CurrentAmount)
		} else {
			event, err = s.getNextSpendingEventAfter(ctx, events[i-1].Date, timezone, events[i-1].RollingAllocation)
		}

		// If there was a problem or a timeout, just return
		if err != nil {
			return nil, err
		}

		// No event returned means there are no more.
		if event == nil {
			break
		}

		events = append(events, *event)
	}

	return events, nil
}

func (s *spendingInstructionBase) GetRecurrencesBetween(
	ctx context.Context,
	start, end time.Time,
	timezone *time.Location,
) ([]time.Time, error) {
	switch s.spending.SpendingType {
	case SpendingTypeExpense:
		rule := s.ruleset
		rule.DTStart(rule.GetDTStart().In(timezone))

		// This little bit is really confusing. Basically we want to know how many times this spending boi happens
		// before the specified end date. This can include the start date, but we want to exclude the end date. This is
		// because this function is **INTENDED** to be called with the start being now or the next funding event, and
		// end being the next funding event immediately after that. We can't control what happens after the later
		// funding event, so we need to know how much will be spent before then, so we know how much to allocate.
		items := rule.Between(start, end.Add(-1*time.Second), true)
		return items, nil
	case SpendingTypeGoal:
		if s.spending.NextRecurrence.After(start) && s.spending.NextRecurrence.Before(end) {
			return []time.Time{s.spending.NextRecurrence}, nil
		}
		fallthrough
	default:
		return nil, nil
	}
}

func (s *spendingInstructionBase) getNextSpendingEventAfter(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
	balance int64,
) (*SpendingEvent, error) {
	// If the spending object is paused then there wont be any events for it at
	// all.
	if s.spending.IsPaused {
		return nil, nil
	}

	input = util.Midnight(input, timezone)

	var rule *RuleSet
	if s.spending.RuleSet != nil {
		rule = s.ruleset
	}

	nextRecurrence := util.Midnight(s.spending.NextRecurrence, timezone)
	switch s.spending.SpendingType {
	case SpendingTypeOverflow:
		return nil, nil
	case SpendingTypeGoal:
		// If we are working with a goal and it has already "completed" then there
		// is nothing more to do, no more events will come up for this spending
		// object.
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			return nil, nil
		}
	case SpendingTypeExpense:
		myownsanity.ASSERT_NOTNIL(rule, "expense spending type must have a recurrence rule!")
		rule.DTStart(rule.GetDTStart().In(timezone))
		if !nextRecurrence.After(input) || nextRecurrence.Equal(input) {
			nextRecurrence = rule.After(input, false)
		}
	}

	var fundingFirst, fundingSecond FundingEvent
	{ // Get our next two funding events
		fundingEvents, err := s.funding.GetNFundingEventsAfter(ctx, 2, input, timezone)
		if err != nil {
			return nil, err
		}
		if len(fundingEvents) != 2 {
			// TODO, if there are multiple funding schedules and they land on the same
			// day, this will happen.
			panic("invalid number of funding events returned;")
		}

		fundingFirst, fundingSecond = fundingEvents[0], fundingEvents[1]
	}

	var eventsBeforeFirst, eventsBeforeSecond int64
	{
		beforeFirst, err := s.GetRecurrencesBetween(ctx, input, fundingFirst.Date, timezone)
		if err != nil {
			return nil, err
		}

		// The number of times this item will be spent before it receives funding
		// again. This is considered the current funding period. This is used to
		// determine if the spending is currently behind. As the total amount that
		// will be spent must be <= the amount currently allocated to this spending
		// item. If it is not then there will not be enough funds to cover each
		// spending event between now and the next funding event.
		eventsBeforeFirst = int64(len(beforeFirst))

		beforeSecond, err := s.GetRecurrencesBetween(ctx, fundingFirst.Date, fundingSecond.Date, timezone)
		if err != nil {
			return nil, err
		}
		// The number of times this item will be spent in the subsequent funding
		// period. This is used to determine how much needs to be allocated at the
		// beginning of the next funding period.
		eventsBeforeSecond = int64(len(beforeSecond))
	}

	// The amount of funds needed for each individual spending event.
	var perSpendingAmount int64
	switch s.spending.SpendingType {
	case SpendingTypeExpense:
		perSpendingAmount = s.spending.TargetAmount
	case SpendingTypeGoal:
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
		numberOfContributions, err := s.funding.GetNumberOfFundingEventsBetween(
			ctx,
			input.Add(1*time.Second),
			nextRecurrence,
			timezone,
		)
		if err != nil {
			return nil, err
		}
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
		// NOTE At the time of writing this, event.RollingAllocation is not being
		// defined anywhere. But this is ultimately what the math will end up being
		// once it is defined, and we calculate the effects of a transaction.
		event.RollingAllocation = event.RollingAllocation - s.spending.TargetAmount
	case nextRecurrence.Equal(fundingFirst.Date):
		// The next event will be both a contribution and a transaction.
		event.Date = nextRecurrence
		event.ContributionAmount = totalContributionAmount
		event.TransactionAmount = s.spending.TargetAmount
		// NOTE At the time of writing this, event.RollingAllocation is not being
		// defined anywhere. But this is ultimately what the math will end up being
		// once it is defined, and we calculate the effects of a transaction.
		event.RollingAllocation = (event.RollingAllocation + totalContributionAmount) - s.spending.TargetAmount
		event.Funding = []FundingEvent{
			fundingFirst,
		}
	}

	return &event, nil
}

func (s *spendingInstructionBase) GetNextInflowEventAfter(
	ctx context.Context,
	input time.Time,
	timezone *time.Location,
) (*SpendingEvent, error) {
	log := s.log.
		WithContext(ctx).
		WithFields(logrus.Fields{
			"input":    input,
			"timezone": timezone.String(),
		})

	afterDate := input
	allocation := s.spending.CurrentAmount
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				log.
					WithError(err).
					Error("timed out while trying to determine next inflow event")
				return nil, errors.WithStack(err)
			}
			crumbs.Warn(ctx, "Received done context signal with no error", "spending", nil)
			return nil, nil
		default:
			// Do nothing
		}

		event, err := s.getNextSpendingEventAfter(ctx, afterDate, timezone, allocation)
		if err != nil {
			return nil, err
		}

		if event == nil {
			return nil, nil
		}

		// If the event we found is an actual contribution then return it.
		if event.ContributionAmount > 0 {
			return event, nil
		}

		// Otherwise we have to go around again, update our baseline and continue.
		afterDate = event.Date
		allocation = event.RollingAllocation
	}
}

type SpendingContribution struct {
	Amount int64
}

func CalculateSpendingContributionAfter(
	ctx context.Context,
	log *logrus.Entry,
	spending Spending,
	funding FundingSchedule,
	input time.Time,
	timezone *time.Location,
) (SpendingContribution, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	fundingInstructions := NewFundingScheduleFundingInstructions(log, funding)
	spendingInstructions := NewSpendingInstructions(log, spending, fundingInstructions)

	result, err := spendingInstructions.GetNextInflowEventAfter(ctx, input, timezone)
	if err != nil {
		return SpendingContribution{
			Amount: 0,
		}, err
	}

	if result == nil {
		return SpendingContribution{
			Amount: 0,
		}, nil
	}

	return SpendingContribution{
		Amount: result.ContributionAmount,
	}, nil
}
