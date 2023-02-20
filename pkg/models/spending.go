package models

import (
	"context"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
)

type SpendingType uint8

const (
	SpendingTypeExpense SpendingType = iota
	SpendingTypeGoal
	SpendingTypeOverflow
)

var _ pg.BeforeInsertHook = (*Spending)(nil)

type Spending struct {
	tableName string `pg:"spending"`

	SpendingId             uint64           `json:"spendingId" pg:"spending_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                *Account         `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE,type:'bigint'"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	FundingScheduleId      uint64           `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,on_delete:RESTRICT"`
	FundingSchedule        *FundingSchedule `json:"-" pg:"rel:has-one" swaggerignore:"true"`
	SpendingType           SpendingType     `json:"spendingType" pg:"spending_type,notnull,use_zero,unique:per_bank"`
	Name                   string           `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" pg:"description"`
	TargetAmount           int64            `json:"targetAmount" pg:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" pg:"current_amount,notnull,use_zero"`
	UsedAmount             int64            `json:"usedAmount" pg:"used_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" pg:"recurrence_rule,type:'text'" swaggertype:"string"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" pg:"last_recurrence"`
	NextRecurrence         time.Time        `json:"nextRecurrence" pg:"next_recurrence,notnull"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" pg:"is_behind,notnull,use_zero"`
	IsPaused               bool             `json:"isPaused" pg:"is_paused,notnull,use_zero"`
	DateCreated            time.Time        `json:"dateCreated" pg:"date_created,notnull"`
	DateStarted            time.Time        `json:"dateStarted" pg:"date_started,notnull"`
}

func (e Spending) GetIsStale(now time.Time) bool {
	return e.NextRecurrence.Before(now)
}

func (e Spending) GetIsPaused() bool {
	return e.IsPaused
}

func (e Spending) GetProgressAmount() int64 {
	switch e.SpendingType {
	case SpendingTypeGoal:
		return e.CurrentAmount + e.UsedAmount
	case SpendingTypeExpense:
		fallthrough
	default:
		return e.CurrentAmount
	}
}

// GetRecurrencesBefore will return an array of times that this spending item will be used (based on the recurrence
// rule) between the provided now and before in the specified time zone. Goals will at most return a single time if the
// goal is due within that window.
func (e *Spending) GetRecurrencesBefore(now, before time.Time, timezone *time.Location) []time.Time {
	switch e.SpendingType {
	case SpendingTypeExpense:
		dtMidnight := util.MidnightInLocal(now, timezone)
		e.RecurrenceRule.DTStart(dtMidnight)
		return e.RecurrenceRule.Between(now, before, false)
	case SpendingTypeGoal:
		if e.NextRecurrence.After(now) && e.NextRecurrence.Before(before) {
			return []time.Time{e.NextRecurrence}
		}
		fallthrough
	default:
		return nil
	}
}

func (e *Spending) CalculateNextContribution(
	ctx context.Context,
	accountTimezone string,
	fundingSchedule *FundingSchedule,
	now time.Time,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("spendingId", strconv.FormatUint(e.SpendingId, 10))

	if e.SpendingType == SpendingTypeOverflow {
		crumbs.Debug(ctx, "No need to calculate contribution for overflow spending", nil)
		return nil
	}

	timezone, err := time.LoadLocation(accountTimezone)
	if err != nil {
		return errors.Wrap(err, "failed to parse account's timezone")
	}
	// Don't change the time by convert it to the account timezone. This will make debugging easier if there is a
	// problem.
	now = now.In(timezone)

	// Get the timestamps for the next two funding events, this is so we can determine how many spending events will
	// happen during these two funding windows.
	fundingFirst, fundingSecond := fundingSchedule.GetNextTwoContributionDatesAfter(now, timezone)
	nextRecurrence := util.MidnightInLocal(e.NextRecurrence, timezone)
	if e.RecurrenceRule != nil {
		// Same thing as the contribution rule, make sure that we are incrementing with the existing dates as the base
		// rather than the current timestamp (which is what RRule defaults to).
		e.RecurrenceRule.DTStart(nextRecurrence)

		// If the next recurrence of the spending is in the past, then bump it as well.
		if nextRecurrence.Before(now) {
			nextRecurrence = e.RecurrenceRule.After(now, false)
		}
	}

	// The number of times this item will be spent before it receives funding again. This is considered the current
	// funding period. This is used to determine if the spending is currently behind. As the total amount that will be
	// spent must be <= the amount currently allocated to this spending item. If it is not then there will not be enough
	// funds to cover each spending event between now and the next funding event.
	eventsBeforeFirst := int64(len(e.GetRecurrencesBefore(now, fundingFirst, timezone)))
	// The number of times this item will be spent in the subsequent funding period. This is used to determine how much
	// needs to be allocated at the beginning of the next funding period.
	eventsBeforeSecond := int64(len(e.GetRecurrencesBefore(fundingFirst, fundingSecond, timezone)))

	// The amount of funds needed for each individual spending event.
	perSpendingAmount := e.TargetAmount
	// The amount of funds currently allocated towards this spending item. This is not increased until the next funding
	// event, or the user transfers funds to this spending item.
	currentAmount := e.GetProgressAmount()

	// We are behind if we do not currently have enough funds for all the spending events between now and the next time
	// this spending object will receive funding.
	e.IsBehind = eventsBeforeFirst > 0 && (perSpendingAmount*eventsBeforeFirst) > currentAmount

	// The total contribution amount is the amount of money that needs to be allocated to this spending item during the
	// next funding event in order to cover all the spending events that will happen between then and the subsequent
	// funding event.
	var totalContributionAmount int64
	// If there are spending events in the next funding period then we need to make sure that we calculate for those.
	if eventsBeforeSecond > 0 {
		// We need to subtract the spending that will happen before the next period though.
		// We have $5 allocated but between now and the next funding we need to spend $5. So we cannot take the $5 we
		// currently have into account when we calculate how much will be needed for the next funding event.
		amountAfterCurrentSpending := myownsanity.Max(0, currentAmount-(perSpendingAmount*eventsBeforeFirst))
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
		amountNeeded := myownsanity.Max(0, perSpendingAmount-currentAmount)
		// And how many times we will have a funding event before our due date.
		numberOfContributions := fundingSchedule.GetNumberOfContributionsBetween(now, nextRecurrence, timezone)
		// Then determine how much we would need at each of those funding events.
		totalContributionAmount = amountNeeded / myownsanity.Max(1, numberOfContributions)
	}

	// Update the spending item with our calculated contribution amount.
	e.NextContributionAmount = totalContributionAmount

	// If the current nextRecurrence on the object is in the past, then bump it to our new next recurrence.
	if e.NextRecurrence.Before(now) {
		e.LastRecurrence = &e.NextRecurrence
		e.NextRecurrence = nextRecurrence
	}

	return nil
}

func (s *Spending) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Make sure when we are creating a funding schedule that we set the date started field for the first instance. This
	// way subsequent rule evaluations can use this date started as a reference point.
	if s.DateStarted.IsZero() {
		s.DateStarted = s.NextRecurrence
	}

	return ctx, nil
}
