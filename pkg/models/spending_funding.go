package models

import (
	"time"

	"github.com/monetr/monetr/pkg/internal/myownsanity"
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

type SpendingFundingDay struct {
	Date    time.Time
	Funding []SpendingFunding
}

type SpendingFundingHelper []SpendingFunding

func NewSpendingFundingHelper(funding []SpendingFunding) SpendingFundingHelper {
	if len(funding) == 0 {
		panic("must provide at least a single funding schedule")
	}
	return funding
}

func (s SpendingFundingHelper) GetNextContributionDateAfter(input time.Time, timezone *time.Location) SpendingFundingDay {
	var earliest time.Time
	result := make([]SpendingFunding, 0, len(s))
	for _, schedule := range s {
		next := schedule.FundingSchedule.GetNextContributionDateAfter(input, timezone)

		switch {
		case earliest.IsZero():
			earliest = next
			fallthrough
		case next.Equal(earliest):
			result = append(result, schedule)
		case next.Before(earliest):
			earliest = next
			result = []SpendingFunding{
				schedule,
			}
		}
	}

	myownsanity.Assert(!earliest.IsZero(), "The earliest next contribution cannot be zero, something is wrong with the provided funding instructions.")

	return SpendingFundingDay{
		Date:    earliest,
		Funding: result,
	}
}

func (s SpendingFundingHelper) GetNextTwoContributionDatesAfter(start time.Time, timezone *time.Location) [2]SpendingFundingDay {
	var result [2]SpendingFundingDay
	for i := 0; i < len(result); i ++ {
		if i == 0 {
			result[i] = s.GetNextContributionDateAfter(start, timezone)
			continue
		}

		result[i] = s.GetNextContributionDateAfter(result[i - 1].Date, timezone)
	}

	return result
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
