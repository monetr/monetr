package models

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/pkg/errors"
)

type SpendingType uint8

const (
	SpendingTypeExpense SpendingType = iota
	SpendingTypeGoal
	SpendingTypeOverflow
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

func (e *Spending) CalculateNextContribution(
	ctx context.Context,
	accountTimezone string,
	fundingSchedule *FundingSchedule,
	now time.Time,
) error {
	span := sentry.StartSpan(ctx, "CalculateNextContribution")
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

	nextContributionDate := fundingSchedule.GetNextContributionDateAfter(now, timezone)

	nextRecurrence := e.NextRecurrence
	if e.RecurrenceRule != nil {
		// Same thing as the contribution rule, make sure that we are incrementing with the existing dates as the base
		// rather than the current timestamp (which is what RRule defaults to).
		e.RecurrenceRule.DTStart(e.NextRecurrence)

		// If the next recurrence of the spending is in the past, then bump it as well.
		if nextRecurrence.Before(now) {
			nextRecurrence = e.RecurrenceRule.After(now, false)
		}
	}

	targetAmount := e.TargetAmount
	currentAmount := e.GetProgressAmount()
	var nextContribution int64
	var isBehind bool
	switch {
	case nextContributionDate.After(nextRecurrence) && e.SpendingType == SpendingTypeExpense:
		// If the next contribution date happens after the next time the expense recurs, then we need to see if the
		// expense recurs at all again before that contribution. Then make sure that we set the contribution amount to
		// represent the total needed for every recurrence.
		numberOfRecurrences := int64(len(e.RecurrenceRule.Between(now, nextContributionDate, false)))
		// Multiple how much we need by how many times this expense will be needed before we will get funding again.
		targetAmount *= numberOfRecurrences
		// Calculate how much we need to contribute next time based on that need.
		nextContribution = targetAmount - currentAmount
		// We are behind if there is anything left to contribute that we cannot contribute before the next recurrence.
		isBehind = targetAmount-currentAmount > 0
	case nextContributionDate.After(nextRecurrence) && e.SpendingType == SpendingTypeGoal:
		// If we won't receive another contribution before the goal is due then just set the next contribution amount to
		// be the amount we need.
		nextContribution = int64(math.Max(float64(targetAmount-currentAmount), 0))
		// We are behind if there is anything left to contribute that we cannot contribute before the next recurrence.
		isBehind = targetAmount-currentAmount > 0
	case nextRecurrence.After(nextContributionDate) && e.SpendingType == SpendingTypeExpense:
		// Check to see how many times this expense will be needed between the next contribution date and the
		// contribution date succeeding it. If the expense is needed multiple times then we need to allocate more to the
		// expense with each contribution.
		frequent := e.RecurrenceRule.Between(
			nextContributionDate,
			fundingSchedule.GetNextContributionDateAfter(nextContributionDate, timezone),
			false,
		)
		if len(frequent) > 1 {
			// If the expense is needed more than one time, then multiply the target amount so we can allocate enough
			// with the next contribution to cover us.
			targetAmount *= int64(len(frequent))
		}
		fallthrough
	case nextRecurrence.After(nextContributionDate):
		// If the next recurrence happens after a contribution, then calculate how many contributions will occur before
		// that recurrence and set that to be the next contribution amount.
		// Technically this works a bit oddly with expenses that recur more frequently than they can be funded, but
		// because we have adjusted the target amount above (if this is the case) then the calculation will still be
		// correct.
		numberOfContributions := fundingSchedule.GetNumberOfContributionsBetween(now, nextRecurrence)

		// If for some reason there are no contributions to be made, then prevent us from trying to divide by zero.
		if numberOfContributions == 0 {
			nextContribution = 0
		} else {
			nextContribution = (targetAmount - currentAmount) / numberOfContributions
		}
	case nextRecurrence.Equal(nextContributionDate) && e.SpendingType == SpendingTypeExpense:
		// Check to see how many times this expense will recur between the next contribution date and the one that
		// succeeds it. But this time make it an inclusive search (the true at the end). Because the next contribution
		// date is also the next recurrence we need to allocate an additional targetAmount towards the expense with the
		// next contribution to cover everything.
		frequent := e.RecurrenceRule.Between(
			nextContributionDate,
			fundingSchedule.GetNextContributionDateAfter(nextContributionDate, timezone),
			true,
		)
		if len(frequent) > 1 {
			targetAmount *= int64(len(frequent))
		}
		fallthrough
	case nextRecurrence.Equal(nextContributionDate):
		// Super simple, we know how much we need (targetAmount) and we know how much we have (currentAmount), and the
		// next time this expense will be needed is on the same day that the expense will receive funding. So just
		// subtract our current from what we need to determine how much we want to allocate.
		nextContribution = int64(math.Max(float64(targetAmount-currentAmount), 0))
	}

	// If the next contribution would be less than zero (likely because the user manually transferred extra funds to this)
	// spending object. Then make sure we don't actually attempt to allocate negative funds.
	if nextContribution < 0 {
		nextContribution = 0
	}

	e.NextContributionAmount = nextContribution
	e.IsBehind = isBehind

	// If the current nextRecurrence on the object is in the past, then bump it to our new next recurrence.
	if e.NextRecurrence.Before(now) {
		e.LastRecurrence = &e.NextRecurrence
		e.NextRecurrence = nextRecurrence
	}

	return nil
}
