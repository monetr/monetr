package models

import (
	"context"
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/util"
	"github.com/uptrace/bun"
)

type SpendingType string

const (
	SpendingTypeExpense SpendingType = "expense" // 0
	SpendingTypeGoal    SpendingType = "goal"    // 1
)

type Spending struct {
	bun.BaseModel `bun:"table:spending,alias:spending"`

	SpendingId             ID[Spending]        `json:"spendingId" bun:"spending_id,notnull,pk"`
	AccountId              ID[Account]         `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account            `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId          ID[BankAccount]     `json:"bankAccountId" bun:"bank_account_id,notnull,pk,unique:per_bank"`
	BankAccount            *BankAccount        `json:"bankAccount,omitempty" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	FundingScheduleId      ID[FundingSchedule] `json:"fundingScheduleId" bun:"funding_schedule_id,notnull,nullzero"`
	FundingSchedule        *FundingSchedule    `json:"-" bun:"rel:belongs-to,join:funding_schedule_id=funding_schedule_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	SpendingType           SpendingType        `json:"spendingType" bun:"spending_type,notnull,unique:per_bank"`
	Name                   string              `json:"name" bun:"name,notnull,unique:per_bank,nullzero"`
	TargetAmount           int64               `json:"targetAmount" bun:"target_amount,notnull"`
	CurrentAmount          int64               `json:"currentAmount" bun:"current_amount,notnull"`
	UsedAmount             int64               `json:"usedAmount" bun:"used_amount,notnull"`
	RuleSet                *RuleSet            `json:"ruleset" bun:"ruleset,notnull,type:text"`
	LastSpentFrom          *time.Time          `json:"lastSpentFrom" bun:"last_spent_from"`
	LastRecurrence         *time.Time          `json:"lastRecurrence" bun:"last_recurrence"`
	NextRecurrence         time.Time           `json:"nextRecurrence" bun:"next_recurrence,notnull,nullzero"`
	NextContributionAmount int64               `json:"nextContributionAmount" bun:"next_contribution_amount,notnull"`
	IsBehind               bool                `json:"isBehind" bun:"is_behind,notnull"`
	IsPaused               bool                `json:"isPaused" bun:"is_paused,notnull"`
	AutoCreateTransaction  bool                `json:"autoCreateTransaction" bun:"auto_create_transaction,notnull"`
	CreatedAt              time.Time           `json:"createdAt" bun:"created_at,notnull,nullzero"`
}

func (Spending) IdentityPrefix() string {
	return "spnd"
}

var (
	_ bun.BeforeAppendModelHook = (*Spending)(nil)
)

func (o *Spending) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.SpendingId.IsZero() {
			o.SpendingId = NewID[Spending]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
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

// GetRecurrencesBefore will return an array of times that this spending item
// will be used (based on the recurrence rule) between the provided now and
// before in the specified time zone. The window includes now but stops just
// short of before, so an event that lands right on before belongs to the next
// window instead of this one. Occurrences that were already spent from early
// are left out, since that money is already gone. Goals will at most return a
// single time if the goal is due within that window.
func (e *Spending) GetRecurrencesBefore(
	now, before time.Time,
	timezone *time.Location,
) []time.Time {
	switch e.SpendingType {
	case SpendingTypeExpense:
		// Make sure that the rule increments relative to midnight in the user's
		// timezone. The stored DTSTART is the same instant but is represented in
		// UTC, which can shift occurrences onto the wrong day for timezones that
		// are ahead of UTC.
		rule := e.RuleSet.InTimezone(timezone)
		// We want every occurrence from now up to but not including before. That
		// way an event that lands right on now still counts here, but one that
		// lands on before rolls into the next window. rrule's Between treats both
		// ends the same, so we ask for both included and then drop the last one if
		// it happens to fall exactly on before.
		occurrences := rule.Between(now, before, true)
		if n := len(occurrences); n > 0 && occurrences[n-1].Equal(before) {
			occurrences = occurrences[:n-1]
		}
		return RecurrencesAfterSpentFrom(
			occurrences,
			e.GetRecurrenceSpentFrom(timezone),
		)
	case SpendingTypeGoal:
		if !e.NextRecurrence.Before(now) && e.NextRecurrence.Before(before) {
			return []time.Time{e.NextRecurrence}
		}
		fallthrough
	default:
		return nil
	}
}

// GetRecurrenceSpentFrom returns the occurrence of the expense that was last
// spent from, based on LastSpentFrom. Bills don't always come in on the exact
// day they are due, so a spend counts towards whichever occurrence it is
// closest to. That way paying a few days early or a few days late still lines
// up with the occurrence it was meant for. If we don't know, or this is not an
// expense, then this returns a zero time.
func (e *Spending) GetRecurrenceSpentFrom(timezone *time.Location) time.Time {
	if e.SpendingType != SpendingTypeExpense ||
		e.RuleSet == nil ||
		e.LastSpentFrom == nil ||
		e.LastSpentFrom.IsZero() {
		return time.Time{}
	}

	return e.getRecurrenceSpentFrom(e.RuleSet.InTimezone(timezone), timezone)
}

// getRecurrenceSpentFrom is the guts of [Spending.GetRecurrenceSpentFrom], but
// takes the rule already in the user's timezone so we don't have to clone it
// more than once when we need the rule for other things too.
func (e *Spending) getRecurrenceSpentFrom(
	rule *RuleSet,
	timezone *time.Location,
) time.Time {
	spentOn := util.Midnight(*e.LastSpentFrom, timezone)
	previous := rule.Before(spentOn, true)
	next := rule.After(spentOn, false)
	switch {
	case next.IsZero():
		return previous
	case previous.IsZero():
		return next
	case next.Sub(spentOn) < spentOn.Sub(previous):
		return next
	default:
		return previous
	}
}

// GetIsLate returns true if the most recent occurrence of the expense before
// now has not been spent from yet, but the occurrence before that one has. So
// the bill was due but the transaction just hasn't come in yet. This only holds
// while now is closer to that occurrence than the next one, past that point a
// spend would count towards the next occurrence anyway. If the expense was not
// spent from recently enough then we don't know, and this returns false.
func (e *Spending) GetIsLate(timezone *time.Location, now time.Time) bool {
	if e.SpendingType != SpendingTypeExpense ||
		e.RuleSet == nil ||
		e.LastSpentFrom == nil ||
		e.LastSpentFrom.IsZero() {
		return false
	}

	rule := e.RuleSet.InTimezone(timezone)
	spentFrom := e.getRecurrenceSpentFrom(rule, timezone)
	if spentFrom.IsZero() {
		return false
	}

	previous := rule.Before(now, false)
	if previous.IsZero() || !rule.Before(previous, false).Equal(spentFrom) {
		return false
	}

	next := rule.After(previous, false)
	if next.IsZero() {
		return false
	}

	return now.Sub(previous) < next.Sub(now)
}

// RecurrencesAfterSpentFrom takes a sorted set of occurrences and drops the
// ones that have already been spent from. If spentFrom is zero then nothing is
// dropped.
func RecurrencesAfterSpentFrom(
	occurrences []time.Time,
	spentFrom time.Time,
) []time.Time {
	if spentFrom.IsZero() {
		return occurrences
	}
	for len(occurrences) > 0 && !occurrences[0].After(spentFrom) {
		occurrences = occurrences[1:]
	}
	return occurrences
}

func (e *Spending) CalculateNextContribution(
	ctx context.Context,
	timezone *time.Location,
	fundingSchedule *FundingSchedule,
	now time.Time,
	log *slog.Logger,
) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	spndLog := log.With(
		"spendingId", e.SpendingId,
		"timezone", timezone.String(),
		slog.Group("spending",
			"ruleset", e.RuleSet,
			"targetAmount", e.TargetAmount,
			"currentAmount", e.CurrentAmount,
		),
		slog.Group("funding",
			"ruleset", fundingSchedule.RuleSet,
			"nextRecurrence", fundingSchedule.NextRecurrence,
			"excludeWeekends", fundingSchedule.ExcludeWeekends,
		),
		slog.Group("before",
			"isBehind", e.IsBehind,
			"nextContributionAmount", e.NextContributionAmount,
			"lastRecurrence", e.LastRecurrence,
			"nextRecurrence", e.NextRecurrence,
		),
	)

	result := calculateNextContribution(
		span.Context(),
		*e,
		*fundingSchedule,
		timezone,
		now,
	)
	e.IsBehind = result.IsBehind
	e.NextContributionAmount = result.NextContributionAmount
	e.LastRecurrence = result.LastRecurrence
	e.NextRecurrence = result.NextRecurrence

	spndLog.DebugContext(span.Context(), "calculated next spending contribution",
		slog.Group("after",
			"isBehind", e.IsBehind,
			"nextContributionAmount", e.NextContributionAmount,
			"lastRecurrence", e.LastRecurrence,
			"nextRecurrence", e.NextRecurrence,
		),
	)
}

// calculateNextContribution takes a spending object and its funding schedule, a
// timezone and a point in time. It then calculates what the next contribution
// to that spending object should be based on that data. It will then return an
// updated spending object with whatever has changed. This can include:
//   - IsBehind
//   - ContributionAmount
//   - LastRecurrence
//   - NextRecurrence
//
// The provided objects are unmodified.
func calculateNextContribution(
	ctx context.Context,
	spending Spending,
	fundingSchedule FundingSchedule,
	timezone *time.Location,
	now time.Time,
) Spending {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("spendingId", spending.SpendingId.String())

	// Don't change the time by convert it to the account timezone. This will make
	// debugging easier if there is a problem.
	// It's possible that the time was already in the account's timezone, but this
	// still is good to have because it makes this function consistent.
	now = now.In(timezone)

	var rule *RuleSet
	if spending.RuleSet != nil {
		rule = spending.RuleSet.InTimezone(timezone)
	}

	fundingFirst, fundingSecond := fundingSchedule.GetNextTwoContributionDatesAfter(now, timezone)
	nextRecurrence := util.Midnight(spending.NextRecurrence, timezone)
	if rule != nil {
		// If the next recurrence of the spending is in the past, then bump it as well.
		if nextRecurrence.Before(now) {
			nextRecurrence = rule.After(now, false)
		}
	}

	// The number of times this item will be spent before it receives funding
	// again. This is considered the current funding period. This is used to
	// determine if the spending is currently behind. As the total amount that
	// will be spent must be <= the amount currently allocated to this spending
	// item. If it is not then there will not be enough funds to cover each
	// spending event between now and the next funding event.
	eventsBeforeFirst := int64(len(spending.GetRecurrencesBefore(now, fundingFirst, timezone)))
	// The number of times this item will be spent in the subsequent funding
	// period. This is used to determine how much needs to be allocated at the
	// beginning of the next funding period.
	eventsBeforeSecond := int64(len(spending.GetRecurrencesBefore(fundingFirst, fundingSecond, timezone)))

	// If the last occurrence was due but hasn't been spent from yet, then it is
	// still going to be spent before the next funding. So count it with the rest
	// of the events in the current funding period.
	var lateEvents int64
	if spending.GetIsLate(timezone, now) {
		lateEvents = 1
		eventsBeforeFirst += lateEvents
	}

	// The amount of funds needed for each individual spending event.
	perSpendingAmount := spending.TargetAmount
	// The amount of funds currently allocated towards this spending item. This is
	// not increased until the next funding event, or the user transfers funds to
	// this spending item.
	currentAmount := spending.GetProgressAmount()

	// We are behind if we do not currently have enough funds for all the spending
	// events between now and the next time this spending object will receive
	// funding.
	spending.IsBehind = eventsBeforeFirst > 0 && (perSpendingAmount*eventsBeforeFirst) > currentAmount

	// The total contribution amount is the amount of money that needs to be
	// allocated to this spending item during the next funding event in order to
	// cover all the spending events that will happen between then and the
	// subsequent funding event.
	var totalContributionAmount int64
	// If there are spending events in the next funding period then we need to
	// make sure that we calculate for those.
	if eventsBeforeSecond > 0 {
		// We need to subtract the spending that will happen before the next period
		// though. We have $5 allocated but between now and the next funding we need
		// to spend $5. So we cannot take the $5 we currently have into account when
		// we calculate how much will be needed for the next funding event.
		amountAfterCurrentSpending := max(
			0,
			currentAmount-(perSpendingAmount*eventsBeforeFirst),
		)
		// The total amount we need is determined by how many times we will need the
		// target amount during the next period between funding events multiplied by
		// how much each spending event costs. If the current spending object is
		// over-allocated for this funding period and the next funding period then
		// this can result in a negative contribution amount. Because we would be
		// subtracting more than the calculated amount that we need.
		nextSpendingPeriodTotal := perSpendingAmount * eventsBeforeSecond
		// By taking the min of the amount we will have allocated and the amount
		// needed. We can safely arrive at a 0 contribution amount when we are
		// over-allocated.
		totalContributionAmount = nextSpendingPeriodTotal - min(
			amountAfterCurrentSpending,
			nextSpendingPeriodTotal,
		)
	} else {
		// Otherwise we can simply look at how much we need vs how much we already
		// have. If the last occurrence is late then whatever it needs is still
		// sitting here, so it can't count towards the next one.
		amountNeeded := max(
			0,
			perSpendingAmount-max(0, currentAmount-(perSpendingAmount*lateEvents)),
		)
		// If the upcoming occurrence was already spent from early then that money
		// is already gone, so we want to be saving for the one after it instead.
		dueDate := nextRecurrence
		if spentFrom := spending.GetRecurrenceSpentFrom(timezone); !spentFrom.IsZero() && !dueDate.After(spentFrom) {
			if after := rule.After(spentFrom, false); !after.IsZero() {
				dueDate = after
			}
		}
		// And how many times we will have a funding event before our due date.
		numberOfContributions := fundingSchedule.GetNumberOfContributionsBetween(
			now,
			dueDate,
			timezone,
		)
		// Then determine how much we would need at each of those funding events.
		totalContributionAmount = amountNeeded / max(1, numberOfContributions)
	}

	// Update the spending item with our calculated contribution amount.
	spending.NextContributionAmount = totalContributionAmount

	// If the current nextRecurrence on the object is in the past, then bump it to
	// our new next recurrence.
	if spending.NextRecurrence.Before(now) {
		lastRecurrence := spending.NextRecurrence
		spending.LastRecurrence = &lastRecurrence
		spending.NextRecurrence = nextRecurrence
	}

	return spending
}
