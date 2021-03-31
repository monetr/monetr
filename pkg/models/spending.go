package models

import (
	"github.com/pkg/errors"
	"time"
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
	FundingSchedule        *FundingSchedule `json:"fundingSchedule,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	SpendingType           SpendingType     `json:"spendingType" pg:"spending_type,notnull,use_zero"`
	Name                   string           `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" pg:"description"`
	TargetAmount           int64            `json:"targetAmount" pg:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" pg:"current_amount,notnull,use_zero"`
	UsedAmount             int64            `json:"userAmount" pg:"used_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" pg:"recurrence_rule,notnull,type:'text'" swaggertype:"string"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" pg:"last_recurrence"`
	NextRecurrence         time.Time        `json:"nextRecurrence" pg:"next_recurrence,notnull"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" pg:"is_behind,notnull,use_zero"`
}

func midnightInLocal(input time.Time, timezone *time.Location) time.Time {
	midnight := time.Date(
		input.Year(),  // Year
		input.Month(), // Month
		input.Day(),   // Day
		0,             // Hours
		0,             // Minutes
		0,             // Seconds
		0,             // Nano seconds
		timezone,      // The account's time zone.
	)

	return midnight
}

func (e *Spending) CalculateNextContribution(
	accountTimezone string,
	nextContributionDate time.Time,
	nextContributionRule *Rule,
) error {
	timezone, err := time.LoadLocation(accountTimezone)
	if err != nil {
		return errors.Wrap(err, "failed to parse account's timezone")
	}

	nextContributionDate = midnightInLocal(nextContributionDate, timezone)

	// If we have achieved our expense then we don't need to do anything.
	if e.TargetAmount <= e.CurrentAmount {
		e.IsBehind = false
		e.NextContributionAmount = 0
	}

	nextDueDate := midnightInLocal(e.NextRecurrence, timezone)
	if time.Now().After(nextDueDate) {
		e.LastRecurrence = &nextDueDate
		e.NextRecurrence = e.RecurrenceRule.After(nextDueDate, false)
		nextDueDate = midnightInLocal(e.NextRecurrence, timezone)
	}

	// If the next time we would contribute to this expense is after the next time the expense is due, then the expense
	// has fallen behind. Mark it as behind and set the contribution to be the difference.
	if nextContributionDate.After(nextDueDate) {
		e.IsBehind = true
		e.NextContributionAmount = e.TargetAmount - e.CurrentAmount
		return nil
	} else if nextContributionDate.Equal(nextDueDate) {
		// If the next time we would contribute is the same day it's due, this is okay. The user could change the due
		// date if they want a bit of a buffer and we would plan it differently. But we don't want to consider this
		// "behind".
		e.IsBehind = false
		e.NextContributionAmount = e.TargetAmount - e.CurrentAmount
		return nil
	}

	// If the next time we would contribute to this expense is not behind and has more than one contribution to meet its
	// target then we need to calculate a partial contribution.
	numberOfContributions := 0
	if nextContributionDate.Before(nextDueDate) {
		numberOfContributions++
	}

	// TODO Handle expenses that recur more frequently than they are funded.
	midnightToday := midnightInLocal(time.Now(), timezone)
	nextContributionRule.DTStart(midnightToday)
	contributionDateX := nextContributionDate
	for {
		contributionDateX = nextContributionRule.After(contributionDateX, false)
		if nextDueDate.Before(contributionDateX) {
			break
		}

		numberOfContributions++
	}

	totalNeeded := e.TargetAmount - e.CurrentAmount
	perContribution := totalNeeded / int64(numberOfContributions)

	e.NextContributionAmount = perContribution
	return nil
}
