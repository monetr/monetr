package models

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/util"
)

type FundingSchedule struct {
	tableName string `pg:"funding_schedules"`

	FundingScheduleId uint64       `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64       `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account           *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId     uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,pk,on_delete:CASCADE,unique:per_bank,type:'bigint'"`
	BankAccount       *BankAccount `json:"bankAccount,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	Name              string       `json:"name" pg:"name,notnull,unique:per_bank"`
	Description       string       `json:"description,omitempty" pg:"description"`
	Rule              *Rule        `json:"rule" pg:"rule,notnull,type:'text'" swaggertype:"string" example:"FREQ=MONTHLY;BYMONTHDAY=15,-1"`
	ExcludeWeekends   bool         `json:"excludeWeekends" pg:"exclude_weekends,notnull"`
	LastOccurrence    *time.Time   `json:"lastOccurrence" pg:"last_occurrence"`
	NextOccurrence    time.Time    `json:"nextOccurrence" pg:"next_occurrence,notnull"`
}

func (f *FundingSchedule) GetNumberOfContributionsBetween(start, end time.Time) int64 {
	return int64(len(f.Rule.Between(start, end, false)))
}

func (f *FundingSchedule) GetNextContributionDateAfter(now time.Time, timezone *time.Location) time.Time {
	nextContributionDate := util.MidnightInLocal(f.NextOccurrence, timezone)
	if now.Before(nextContributionDate) {
		// If now is before the already established next occurrence, then just return that.
		// This might be goofy if we want to test stuff in the distant past?
		return nextContributionDate
	}

	nextContributionRule := f.Rule.RRule

	// Force the start of the rule to be the next contribution date. This fixes a bug where the rule would increment
	// properly, but would include the current timestamp in that increment causing incorrect comparisons below. This
	// makes sure that the rule will increment in the user's timezone as intended.
	nextContributionRule.DTStart(f.NextOccurrence)

	// Keep track of an un-adjusted next contribution date. Because we might subtract days to account for early
	// funding, we need to make sure we are still incrementing relative to the _real_ contribution dates. Not the
	// adjusted ones.
	actualNextContributionDate := nextContributionDate
	for !nextContributionDate.After(now) {
		// If the next contribution date is not after now, then increment it.
		nextContributionDate = nextContributionRule.After(actualNextContributionDate, false)
		// Store the real contribution date for later use.
		actualNextContributionDate = nextContributionDate

		// If we are excluding weekends, and the next contribution date falls on a weekend; then we need to adjust the
		// date to the previous business day.
		if f.ExcludeWeekends {
			switch nextContributionDate.Weekday() {
			case time.Sunday:
				// If it lands on a sunday then subtract 2 days to put the contribution date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -2)
			case time.Saturday:
				// If it lands on a sunday then subtract 1 day to put the contribution date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -1)
			}
		}

		nextContributionDate = util.MidnightInLocal(nextContributionDate, timezone)
	}

	return nextContributionDate
}

func (f *FundingSchedule) CalculateNextOccurrence(ctx context.Context, timezone *time.Location) bool {
	span := sentry.StartSpan(ctx, "CalculateNextOccurrence")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"fundingScheduleId": f.FundingScheduleId,
		"timezone":          timezone.String(),
	}

	now := time.Now()

	if now.Before(f.NextOccurrence) {
		crumbs.Debug(span.Context(), "Skipping processing funding schedule, it does not occur yet", map[string]interface{}{
			"fundingScheduleId": f.FundingScheduleId,
			"nextOccurrence":    f.NextOccurrence,
		})
		return false
	}

	nextFundingOccurrence := f.GetNextContributionDateAfter(now, timezone)

	current := f.NextOccurrence
	f.LastOccurrence = &current
	f.NextOccurrence = nextFundingOccurrence

	return true
}
