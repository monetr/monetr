package models

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/util"
	"github.com/uptrace/bun"
)

type FundingSchedule struct {
	bun.BaseModel `bun:"table:funding_schedules,alias:funding_schedule"`

	FundingScheduleId      ID[FundingSchedule] `json:"fundingScheduleId" bun:"funding_schedule_id,notnull,pk"`
	AccountId              ID[Account]         `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account            `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId          ID[BankAccount]     `json:"bankAccountId" bun:"bank_account_id,notnull,pk,unique:per_bank"`
	BankAccount            *BankAccount        `json:"bankAccount,omitempty" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	Name                   string              `json:"name" bun:"name,notnull,unique:per_bank,nullzero"`
	Description            string              `json:"description,omitempty" bun:"description,nullzero"`
	RuleSet                *RuleSet            `json:"ruleset" bun:"ruleset,notnull,type:text"`
	ExcludeWeekends        bool                `json:"excludeWeekends" bun:"exclude_weekends,notnull"`
	WaitForDeposit         bool                `json:"waitForDeposit" bun:"wait_for_deposit,notnull"`
	AutoCreateTransaction  bool                `json:"autoCreateTransaction" bun:"auto_create_transaction,notnull"`
	EstimatedDeposit       *int64              `json:"estimatedDeposit" bun:"estimated_deposit"`
	LastRecurrence         *time.Time          `json:"lastRecurrence" bun:"last_recurrence"`
	NextRecurrence         time.Time           `json:"nextRecurrence" bun:"next_recurrence,notnull,nullzero"`
	NextRecurrenceOriginal time.Time           `json:"nextRecurrenceOriginal" bun:"next_recurrence_original,notnull,nullzero"`
}

func (FundingSchedule) IdentityPrefix() string {
	return "fund"
}

var (
	_ bun.BeforeAppendModelHook = (*FundingSchedule)(nil)
)

func (o *FundingSchedule) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.FundingScheduleId.IsZero() {
			o.FundingScheduleId = NewID[FundingSchedule]()
		}
	}

	return nil
}

// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (o *FundingSchedule) GetNumberOfContributionsBetween(
	start, end time.Time,
	timezone *time.Location,
) int64 {
	// Make sure that the rule is using the timezone of the dates provided. This
	// is an easy way to force that. We also need to truncate the hours on the
	// start time. To make sure that we are operating relative to midnight.
	rule := o.RuleSet.Clone()
	rule.DTStart(rule.GetDTStart().In(timezone))
	items := rule.Between(start, end, true)
	return int64(len(items))
}

// GetNextTwoContributionDatesAfter returns the next two contribution dates
// relative to the timestamp provided. This is used to better calculate
// contributions to funds that recur more frequently than they can be funded.
// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (o *FundingSchedule) GetNextTwoContributionDatesAfter(
	now time.Time,
	timezone *time.Location,
) (time.Time, time.Time) {
	nextOne, _ := o.GetNextContributionDateAfter(now, timezone)
	subsequent, _ := o.GetNextContributionDateAfter(nextOne, timezone)

	return nextOne, subsequent
}

// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (o *FundingSchedule) GetNextContributionDateAfter(
	now time.Time,
	timezone *time.Location,
) (actual, original time.Time) {
	// Make debugging easier.
	now = now.In(timezone)
	nextContributionRule := o.RuleSet.Clone()
	// Force the start of the rule to be the next contribution date. This fixes a
	// bug where the rule would increment properly, but would include the current
	// timestamp in that increment causing incorrect comparisons below. This makes
	// sure that the rule will increment in the user's timezone as intended.
	nextContributionRule.DTStart(nextContributionRule.GetDTStart().In(timezone))
	var nextContributionDate time.Time
	if !o.NextRecurrence.IsZero() {
		nextContributionDate = util.Midnight(o.NextRecurrence, timezone)
	} else {
		nextContributionDate = util.Midnight(nextContributionRule.Before(now, false), timezone)
	}
	if now.Before(nextContributionDate) {
		// If now is before the already established next occurrence, then just
		// return that. This might be goofy if we want to test stuff in the distant
		// past?
		return nextContributionDate, nextContributionDate
	}

	// Keep track of an un-adjusted next contribution date. Because we might
	// subtract days to account for early funding, we need to make sure we are
	// still incrementing relative to the _real_ contribution dates. Not the
	// adjusted ones.
	actualNextContributionDate := nextContributionDate
	for !nextContributionDate.After(now) {
		// If the next contribution date is not after now, then increment it.
		nextContributionDate = nextContributionRule.After(actualNextContributionDate, false)
		// Store the real contribution date for later use.
		actualNextContributionDate = nextContributionDate

		// If we are excluding weekends, and the next contribution date falls on a
		// weekend; then we need to adjust the date to the previous business day.
		if o.ExcludeWeekends {
			switch nextContributionDate.Weekday() {
			case time.Sunday:
				// If it lands on a sunday then subtract 2 days to put the contribution
				// date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -2)
			case time.Saturday:
				// If it lands on a sunday then subtract 1 day to put the contribution
				// date on a Friday.
				nextContributionDate = nextContributionDate.AddDate(0, 0, -1)
			}
		}

		nextContributionDate = util.Midnight(nextContributionDate, timezone)
	}

	return nextContributionDate, actualNextContributionDate
}

// Deprecated: This function should no longer be used, use the forecasting code
// instead.
func (o *FundingSchedule) CalculateNextOccurrence(
	ctx context.Context,
	now time.Time,
	timezone *time.Location,
) bool {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "CalculateNextOccurrence"

	span.Data = map[string]any{
		"fundingScheduleId": o.FundingScheduleId,
		"timezone":          timezone.String(),
	}

	if now.Before(o.NextRecurrence) {
		crumbs.Debug(span.Context(), "Skipping processing funding schedule, it does not occur yet", map[string]any{
			"fundingScheduleId": o.FundingScheduleId,
			"now":               now,
			"nextOccurrence":    o.NextRecurrence,
		})
		return false
	}

	nextFundingOccurrence, originalNextFundingOccurrence := o.GetNextContributionDateAfter(now, timezone)

	crumbs.Debug(span.Context(), "Calculated next recurrence for funding schedule", map[string]any{
		"fundingScheduleId": o.FundingScheduleId,
		"excludeWeekends":   o.ExcludeWeekends,
		"ruleset":           o.RuleSet,
		"before": map[string]any{
			"lastRecurrence":         o.LastRecurrence,
			"nextRecurrence":         o.NextRecurrence,
			"nextRecurrenceOriginal": o.NextRecurrenceOriginal,
		},
		"after": map[string]any{
			"lastRecurrence":         o.NextRecurrence,
			"nextRecurrence":         nextFundingOccurrence,
			"nextRecurrenceOriginal": originalNextFundingOccurrence,
		},
	})

	current := o.NextRecurrence
	o.LastRecurrence = &current
	o.NextRecurrence = nextFundingOccurrence
	o.NextRecurrenceOriginal = originalNextFundingOccurrence

	return true
}
