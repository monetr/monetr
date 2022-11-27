package models

import (
	"time"

	"github.com/monetr/monetr/pkg/util"
)

type SpendingFunding struct {
	tableName string `pg:"spending_funding"`

	SpendingFundingId      uint64           `json:"spendingFundingId" pg:"spending_funding_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                *Account         `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE,type:'bigint'"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	SpendingId             uint64           `json:"spendingId" pg:"spending_id,notnull,on_delete:CASCADE,type:'bigint'"`
	Spending               *Spending        `json:"-" pg:"rel:has-one"`
	FundingScheduleId      uint64           `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,on_delete:CASCADE,type:'bigint'"`
	FundingSchedule        *FundingSchedule `json:"-" pg:"rel:has-one"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
}

type SpendingFundingHelper []SpendingFunding

func NewSpendingFundingHelper(funding []SpendingFunding) SpendingFundingHelper {
	if len(funding) == 0 {
		panic("must provide at least a single funding schedule")
	}
	return funding
}

func (s SpendingFundingHelper) GetNextTwoContributionDatesAfter(start time.Time, timezone *time.Location) (time.Time, time.Time) {
	var first, second time.Time
	for _, schedule := range s {
		next := schedule.FundingSchedule.GetNextContributionDateAfter(start, timezone)
		if next.Before(first) || first.IsZero() {
			first = next
		}
	}

	for _, schedule := range s {
		next := schedule.FundingSchedule.GetNextContributionDateAfter(first, timezone)
		if next.Before(second) || second.IsZero() {
			second = next
		}
	}

	return first, second
}

func (s SpendingFundingHelper) GetNumberOfContributionsBetween(start, end time.Time, timezone *time.Location) int64 {
	// We need to get the unique days that we will be contributing to something. So key the map by the unix timestamp.
	contributions := map[int64]struct{}{}
	for _, schedule := range s {
		rule := schedule.FundingSchedule.Rule.RRule
		// Make sure that the rule is using the timezone of the dates provided. This is an easy way to force that.
		// We also need to truncate the hours on the start time. To make sure that we are operating relative to
		// midnight.
		dtStart := util.MidnightInLocal(start, timezone)
		rule.DTStart(dtStart)
		items := rule.Between(start, end, true)
		for _, item := range items {
			contributions[util.MidnightInLocal(item, timezone).Unix()] = struct{}{}
		}
	}

	return int64(len(contributions))
}

func (s SpendingFundingHelper) GetNextFunding(start time.Time, timezone *time.Location) SpendingFunding {
	var first time.Time
	var funding SpendingFunding
	for _, schedule := range s {
		next := schedule.FundingSchedule.GetNextContributionDateAfter(start, timezone)

		if next.Before(first) || first.IsZero() {
			first = next
			funding = schedule
		}
	}

	return funding
}
