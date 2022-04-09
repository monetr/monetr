package models

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
)

type SpendingType uint8

const (
	SpendingTypeExpense SpendingType = iota
	SpendingTypeGoal
)

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

// CalculateNextContribution will take the provided details about the next contribution's date and frequency and
// determine how much will need to be allocated on that day to this spending object. This is to the best of my knowledge
// a "pure" function, it should produce the same results given the same inputs every single time. As such it must be
// provided things like "now".
func (e *Spending) CalculateNextContribution(
	ctx context.Context,
	accountTimezone string,
	nextContributionDate time.Time,
	nextContributionRule *Rule,
	now time.Time,
) error {
	span := sentry.StartSpan(ctx, "CalculateNextContribution")
	defer span.Finish()

	span.SetTag("spendingId", strconv.FormatUint(e.SpendingId, 10))

	timezone, err := time.LoadLocation(accountTimezone)
	if err != nil {
		return errors.Wrap(err, "failed to parse account's timezone")
	}

	// The total needed needs to be calculated differently for goals and expenses. How much expenses need is always a
	// representation of the target amount minus the current amount allocated to the expense. But goals work a bit
	// differently because the allocated amount can fluctuate throughout the life of the goal. When a transaction is
	// spent from a goal it deducts from the current amount, but adds to the used amount. This is to keep track of how
	// much the goal has actually progressed while maintaining existing patterns for calculating allocations. As a
	// result for us to know how much a goal needs, we need to subtract the current amount plus the used amount from the
	// target for goals.
	progressAmount := e.GetProgressAmount()

	nextContributionDate = util.MidnightInLocal(nextContributionDate, timezone)

	// If we have achieved our expense then we don't need to do anything.
	if e.TargetAmount <= progressAmount {
		e.IsBehind = false
		e.NextContributionAmount = 0
	}

	// Always cast to the timezone, this way if the timezone changes (like crossing DST) we make sure we have the
	// correct midnight value.
	nextDueDate := util.MidnightInLocal(e.NextRecurrence, timezone)

	if e.RecurrenceRule != nil {
		// This will trick RRule into calculating the "after" based on the current next due date, so the next one will
		// be relative to the current one.
		e.RecurrenceRule.DTStart(nextDueDate)
	}

	if now.After(nextDueDate) && e.RecurrenceRule != nil {
		// Bump the last time this spending object recurred to the "nextDueDate" which should now be in the past.
		e.LastRecurrence = &nextDueDate
		// Calculate the next time this spending object will need to have a full balance.
		e.NextRecurrence = util.MidnightInLocal(
			e.RecurrenceRule.After(nextDueDate, false),
			timezone, // Make sure we calculate this in the account's timezone.
		)
		nextDueDate = e.NextRecurrence
	}

	// This is just to make absolutely sure that we are working in the user's timezone and not something else.
	nowInTimezone := now.In(timezone)

	// Keep track of how much we need for a single recurrence of the expense.
	targetAmount := e.TargetAmount

	// Make sure we calculate our next contribution relative to the next contribution date. This will make sure that the
	// RRule library does not do anything with the hours, minutes or seconds that can throw of calculations.
	nextContributionRule.DTStart(nextContributionDate)

	// If we are a recurring expense then check to see if it recurs more frequently than we get funding for it.
	if e.RecurrenceRule != nil {
		// Start with the next contribution date.
		subsequentContributionDate := nextContributionDate

		// If the next time we need this expense ready is not before the next contribution date (could be equal to the
		// next contribution date) then look ahead one more contribution.
		if !e.NextRecurrence.Before(nextContributionDate) {
			subsequentContributionDate = nextContributionRule.After(nextContributionDate, false)
		}
		//subsequentContributionDate := nextContributionRule.After(nextContributionDate, false)
		// Then see how many times this expense is due between now and then.
		numberOfTimesNeededBeforeNextContribution := len(e.RecurrenceRule.Between(nowInTimezone, subsequentContributionDate, false))
		// If it is due at least once then that means we need to modify our target amount for the next contribution, so
		// we can over-allocate.
		if numberOfTimesNeededBeforeNextContribution > 0 {
			targetAmount *= int64(numberOfTimesNeededBeforeNextContribution)
		}
	}

	// This is just how much we need to meet our target.
	needed := int64(math.Max(float64(targetAmount-progressAmount), 0))

	if nextContributionDate.After(nextDueDate) {
		// If the next time we would contribute to this expense is after the next time the expense is due, then the
		// expense has fallen behind. Mark it as behind and set the contribution to be the difference.
		// This might over allocate towards the expense in some scenarios where it believes it is behind when it is not.
		e.NextContributionAmount = needed
		e.IsBehind = progressAmount < targetAmount
		return nil
	} else if nextContributionDate.Equal(nextDueDate) {
		// If the next time we would contribute is the same day it's due, this is okay. The user could change the due
		// date if they want a bit of a buffer, and we would plan it differently. But we don't want to consider this
		// "behind".
		e.IsBehind = false
		e.NextContributionAmount = needed
		return nil
	} else if progressAmount >= targetAmount {
		e.IsBehind = false
	} else {
		// Fix weird edge case where this isn't being unset.
		e.IsBehind = false
	}

	// Count the number of times the contribution rule will happen before the next time we are due. We can then divide
	// the amount needed by the number of times we will have a chance to contribute.
	numberOfContributions := len(nextContributionRule.Between(nowInTimezone, nextDueDate, false))

	if numberOfContributions == 0 {
		// This is a bit weird, I'm not sure what causes this yet off the top of my head. I ran into while testing when
		// I made the due date the 29th, and the next payday the 28th. (Note: it was the 28th at the time). And it
		// caused this to break. Pretty sure this is just a bug with the funding schedule next contribution date being
		// in the past, but this is a short term fix for now. This also acts as a slight safety net for a divide by 0
		// error.
		e.NextContributionAmount = needed
	} else {
		perContribution := needed / int64(numberOfContributions)
		e.NextContributionAmount = perContribution
	}

	return nil
}
