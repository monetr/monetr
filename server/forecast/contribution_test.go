package forecast

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateNextContribution(t *testing.T) {
	t.Run("simple monthly expense", func(t *testing.T) {
		t.Skip("not relevant yet")
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		fundingRule := testutils.NewRuleSet(t, 2022, 1, 15, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		spendingRule := testutils.NewRuleSet(t, 2022, 10, 8, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8")
		now := time.Date(2024, 4, 9, 0, 0, 1, 0, timezone).UTC()
		log := testutils.GetLog(t)

		funding := models.FundingSchedule{
			RuleSet:         fundingRule,
			ExcludeWeekends: false,
			NextOccurrence:  time.Date(2024, 4, 15, 0, 0, 0, 0, timezone),
		}
		spending := models.Spending{
			SpendingType:   models.SpendingTypeExpense,
			TargetAmount:   5000,
			CurrentAmount:  0,
			NextRecurrence: time.Date(2024, 5, 8, 0, 0, 0, 0, timezone),
			RuleSet:        spendingRule,
		}

		contribution, err := CalculateNextContribution(
			context.Background(),
			spending,
			funding,
			timezone,
			now,
			log,
		)
		assert.NoError(t, err, "should be able to calculate next contribution")
		assert.Equal(t, Contribution{
			IsBehind:           false,
			ContributionAmount: 2500,
			NextRecurrence:     time.Date(2024, 5, 8, 0, 0, 0, 0, timezone),
		}, contribution)
	})
}
