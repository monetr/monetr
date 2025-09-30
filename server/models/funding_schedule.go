package models

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/util"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

type FundingSchedule struct {
	tableName string `pg:"funding_schedules"`

	FundingScheduleId      ID[FundingSchedule] `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,pk"`
	AccountId              ID[Account]         `json:"-" pg:"account_id,notnull,pk"`
	Account                *Account            `json:"-" pg:"rel:has-one"`
	BankAccountId          ID[BankAccount]     `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank"`
	BankAccount            *BankAccount        `json:"bankAccount,omitempty" pg:"rel:has-one"`
	Name                   string              `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string              `json:"description,omitempty" pg:"description"`
	RuleSet                *RuleSet            `json:"ruleset" pg:"ruleset,notnull,type:'text'"`
	ExcludeWeekends        bool                `json:"excludeWeekends" pg:"exclude_weekends,notnull,use_zero"`
	WaitForDeposit         bool                `json:"waitForDeposit" pg:"wait_for_deposit,notnull,use_zero"`
	EstimatedDeposit       *int64              `json:"estimatedDeposit" pg:"estimated_deposit"`
	LastRecurrence         *time.Time          `json:"lastRecurrence" pg:"last_recurrence"`
	NextRecurrence         time.Time           `json:"nextRecurrence" pg:"next_recurrence,notnull"`
	NextRecurrenceOriginal time.Time           `json:"nextRecurrenceOriginal" pg:"next_recurrence_original,notnull"`
}

func (FundingSchedule) IdentityPrefix() string {
	return "fund"
}

var (
	_ pg.BeforeInsertHook = (*FundingSchedule)(nil)
)

func (o *FundingSchedule) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.FundingScheduleId.IsZero() {
		o.FundingScheduleId = NewID(o)
	}

	return ctx, nil
}

// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (f *FundingSchedule) GetNumberOfContributionsBetween(start, end time.Time, timezone *time.Location) int64 {
	rule := f.RuleSet.Set
	// Make sure that the rule is using the timezone of the dates provided. This
	// is an easy way to force that. We also need to truncate the hours on the
	// start time. To make sure that we are operating relative to midnight.
	items := rule.Between(start, end, true)
	return int64(len(items))
}

// GetNextTwoContributionDatesAfter returns the next two contribution dates
// relative to the timestamp provided. This is used to better calculate
// contributions to funds that recur more frequently than they can be funded.
// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (f *FundingSchedule) GetNextTwoContributionDatesAfter(now time.Time, timezone *time.Location) (time.Time, time.Time) {
	nextOne, _ := f.GetNextContributionDateAfter(now, timezone)
	subsequent, _ := f.GetNextContributionDateAfter(nextOne, timezone)

	return nextOne, subsequent
}

// Deprecated: Use the forecasting package funding instructions interface
// instead.
func (f *FundingSchedule) GetNextContributionDateAfter(now time.Time, timezone *time.Location) (actual, original time.Time) {
	// Make debugging easier.
	now = now.In(timezone)
	nextContributionRule := f.RuleSet.Clone()
	// Force the start of the rule to be the next contribution date. This fixes a
	// bug where the rule would increment properly, but would include the current
	// timestamp in that increment causing incorrect comparisons below. This makes
	// sure that the rule will increment in the user's timezone as intended.
	nextContributionRule.DTStart(nextContributionRule.GetDTStart().In(timezone))
	var nextContributionDate time.Time
	if !f.NextRecurrence.IsZero() {
		nextContributionDate = util.Midnight(f.NextRecurrence, timezone)
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
		if f.ExcludeWeekends {
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
func (f *FundingSchedule) CalculateNextOccurrence(ctx context.Context, now time.Time, timezone *time.Location) bool {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "CalculateNextOccurrence"

	span.Data = map[string]interface{}{
		"fundingScheduleId": f.FundingScheduleId,
		"timezone":          timezone.String(),
	}

	if now.Before(f.NextRecurrence) {
		crumbs.Debug(span.Context(), "Skipping processing funding schedule, it does not occur yet", map[string]interface{}{
			"fundingScheduleId": f.FundingScheduleId,
			"now":               now,
			"nextOccurrence":    f.NextRecurrence,
		})
		return false
	}

	nextFundingOccurrence, originalNextFundingOccurrence := f.GetNextContributionDateAfter(now, timezone)

	crumbs.Debug(span.Context(), "Calculated next recurrence for funding schedule", map[string]interface{}{
		"fundingScheduleId": f.FundingScheduleId,
		"excludeWeekends":   f.ExcludeWeekends,
		"ruleset":           f.RuleSet,
		"before": map[string]interface{}{
			"lastRecurrence":         f.LastRecurrence,
			"nextRecurrence":         f.NextRecurrence,
			"nextRecurrenceOriginal": f.NextRecurrenceOriginal,
		},
		"after": map[string]interface{}{
			"lastRecurrence":         f.NextRecurrence,
			"nextRecurrence":         nextFundingOccurrence,
			"nextRecurrenceOriginal": originalNextFundingOccurrence,
		},
	})

	current := f.NextRecurrence
	f.LastRecurrence = &current
	f.NextRecurrence = nextFundingOccurrence
	f.NextRecurrenceOriginal = originalNextFundingOccurrence

	return true
}

func (FundingSchedule) CreateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validation.Key(
			"bankAccountId",
			validation.Required.Error("Must specify a bank account ID"),
			ValidID[BankAccount]().Error("Bank account ID specified is not valid"),
		).Required(validators.Require),
		validators.Name(validators.Require),
		validation.Key(
			"description",
			validation.Length(1, 300).Error("Description must be between 1 and 300 characters"),
		).Required(validators.Optional),
		validation.Key(
			"ruleset",
			validation.Required.Error("Ruleset must be specified for funding schedules"),
			validation.NewStringRule(func(input string) bool {
				_, err := NewRuleSet(input)
				return err == nil
			}, "Ruleset must be valid"),
		).Required(validators.Require),
		validation.Key(
			"excludeWeekends",
			validation.In(true, false).Error("Exclude weekends must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key(
			"estimatedDeposit",
			validation.Min(float64(0)).Error("Estimated deposit cannot be less than 0"),
		).Required(validators.Optional),
		// TODO Validate next recurrence based on the current timestamp
	}
}
