package forecast

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/sirupsen/logrus"
)

type Contribution struct {
	IsBehind           bool
	ContributionAmount int64
	NextRecurrence     time.Time
}

// Do not use this yet!
func CalculateNextContribution(
	ctx context.Context,
	spending models.Spending,
	fundingSchedule models.FundingSchedule,
	timezone *time.Location,
	now time.Time,
	log *logrus.Entry,
) (Contribution, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	crumbs.Debug(span.Context(), "Calculating next contribution for spending", map[string]interface{}{
		"spending": map[string]interface{}{
			"spendingId":    spending.SpendingId,
			"spendingType":  spending.SpendingType,
			"ruleset":       spending.RuleSet,
			"targetAmount":  spending.TargetAmount,
			"usedAmount":    spending.UsedAmount,
			"currentAmount": spending.CurrentAmount,
			"isPaused":      spending.IsPaused,
		},
		"funding": map[string]interface{}{
			"fundingScheduleId": fundingSchedule.FundingScheduleId,
			"ruleset":           fundingSchedule.RuleSet,
			"excludeWeekends":   fundingSchedule.ExcludeWeekends,
		},
	})

	// At the time of writing this overflow spending isn't implemented really
	// anywhere. But its pretty simple. We don't make pre-calculated contributions
	// to them. They are meant to be used as a way to just be an earmark outside
	// of the forecast.
	if spending.SpendingType == models.SpendingTypeOverflow || spending.GetIsPaused() {
		return Contribution{
			IsBehind:           false,
			ContributionAmount: 0,
			NextRecurrence:     spending.NextRecurrence,
		}, nil
	}

	// Don't change the time by converting it to the account timezone. This will
	// make debugging easier if there is a problem. It's possible that the time
	// was already in the account's timezone, but this still is good to have
	// because it makes this function consistent. If the time was already in the
	// account's timezone then this won't do anything.
	now = now.In(timezone)

	fundingInstructions := NewFundingScheduleFundingInstructions(
		log,
		fundingSchedule,
	)

	spendingInstructions := NewSpendingInstructions(
		log,
		spending,
		fundingInstructions,
	)

	spendingEvent, err := spendingInstructions.GetNextNSpendingEventsAfter(
		span.Context(),
		1,
		now,
		timezone,
	)
	if err != nil {
		return Contribution{}, nil
	}

	nextRecurrence := spending.NextRecurrence
	isBehind := false
	if len(spendingEvent) == 1 {
		isBehind = spendingEvent[0].IsBehind
		nextRecurrence = spendingEvent[0].Date
	}

	contributionEvent, err := spendingInstructions.GetNextContributionEvent(
		span.Context(),
		now,
		timezone,
	)
	if err != nil {
		return Contribution{}, nil
	}

	return Contribution{
		IsBehind:           isBehind,
		ContributionAmount: contributionEvent.ContributionAmount,
		NextRecurrence:     nextRecurrence,
	}, nil
}
