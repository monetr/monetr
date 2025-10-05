package models

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/util"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
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
func (o *FundingSchedule) GetNumberOfContributionsBetween(
	start, end time.Time,
	timezone *time.Location,
) int64 {
	rule := o.RuleSet.Set
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

func (FundingSchedule) CreateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		// validation.Key(
		// 	"bankAccountId",
		// 	validation.Required.Error("Must specify a bank account ID"),
		// 	ValidID[BankAccount]().Error("Bank account ID specified is not valid"),
		// ).Required(validators.Optional),
		validators.Name(validators.Require),
		validators.Description(),
		// TODO This is broken because we cannot take a string ruleset and MERGE it
		// into a ruleset struct. We need to implement a transformer here.
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
		validation.Key(
			"nextRecurrence",
			validation.Date(time.RFC3339).Min(time.Now()).Error("Next recurrence must be in the future"),
		).Required(validators.Optional),
	}
}

func (FundingSchedule) UpdateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validators.Name(validators.Optional),
		validators.Description(),
		validation.Key(
			"ruleset",
			validation.NewStringRule(func(input string) bool {
				_, err := NewRuleSet(input)
				return err == nil
			}, "Ruleset must be valid"),
		).Required(validators.Optional),
		validation.Key(
			"excludeWeekends",
			validation.In(true, false).Error("Exclude weekends must be a valid boolean"),
		).Required(validators.Optional),
		validation.Key(
			"estimatedDeposit",
			validation.Min(float64(0)).Error("Estimated deposit cannot be less than 0"),
		).Required(validators.Optional),
		validation.Key(
			"nextRecurrence",
			validation.Date(time.RFC3339).Min(time.Now()).Error("Next recurrence must be in the future"),
		).Required(validators.Optional),
	}
}

// UnmarshalRequest consumes a request body and an array of validation rules in
// order to create an object that can be persisted to the database. For updates,
// this function should be called on the existing object that is already stored
// in the database. The provided validators should prevent key or sensitive
// fields from being overwritten by the client's request body. For creates, the
// initial object can be left blank; or default values can be specified ahead of
// calling this function in case some fields are omitted in the intial request.
func (o *FundingSchedule) UnmarshalRequest(
	ctx context.Context,
	reader io.Reader,
	validators ...*validation.KeyRules,
) error {
	rawData := map[string]any{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return errors.WithStack(err)
	}

	if err := validation.ValidateWithContext(
		ctx,
		&rawData,
		validation.Map(
			validators...,
		),
	); err != nil {
		return err
	}

	if err := merge.Merge(
		o, rawData, merge.ErrorOnUnknownField,
	); err != nil {
		return errors.Wrap(err, "failed to merge patched data")
	}

	return nil
}
