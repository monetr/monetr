package schema_test

import (
	"testing"
	"time"

	"github.com/Oudwins/zog"
	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schema"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestCreateSpending(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":              "Test",
			"description":       "Foobar",
			"fundingScheduleId": "fund_1234",
			"ruleset":           "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			"targetAmount":      int64(2000),
			"nextRecurrence":    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC().Format(time.RFC3339),
		}

		var result models.Spending

		issues := schema.CreateSpending.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		assert.EqualValues(t, "Test", result.Name)
		assert.EqualValues(t, "Foobar", result.Description)
		assert.EqualValues(t, models.SpendingTypeExpense, result.SpendingType)
		assert.EqualValues(t, int64(2000), result.TargetAmount)
	})

	t.Run("invalid funding schedule", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":              "Test",
			"description":       "Foobar",
			"fundingScheduleId": "bac_1234",
		}

		var result models.Spending

		issues := schema.CreateSpending.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.EqualValues(t, map[string][]string{
			"fundingScheduleId": {
				`expected id with prefix "fund"`,
			},
			"nextRecurrence": {
				"is required",
			},
			"ruleset": {
				"expenses must have a ruleset",
			},
			"targetAmount": {
				"is required",
			},
		}, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "bac_1234", result.FundingScheduleId)
	})
}

func TestPatchSpending(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name": "Test",
		}

		result := models.Spending{
			SpendingId:        "spnd_foo",
			AccountId:         "acct_1234",
			BankAccountId:     "bac_1234",
			FundingScheduleId: "fund_1234",
			Name:              "Original",
			Description:       "My expense",
			SpendingType:      models.SpendingTypeExpense,
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           testutils.Must(t, models.NewRuleSet, "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1"),
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		assert.Equal(t, "Test", result.Name)
	})

	t.Run("patch ruleset to goal", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":    "Test",
			"ruleset": "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
		}

		result := models.Spending{
			SpendingId:        "spnd_foo",
			AccountId:         "acct_1234",
			BankAccountId:     "bac_1234",
			FundingScheduleId: "fund_1234",
			Name:              "Original",
			Description:       "My expense",
			SpendingType:      models.SpendingTypeGoal,
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           nil,
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.EqualValues(t, map[string][]string{
			"ruleset": {
				"goals cannot have a ruleset",
			},
		}, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		assert.Equal(t, "Test", result.Name)
	})

	t.Run("wont overwrite with a nil", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":              "Test",
			"fundingScheduleId": nil,
		}

		result := models.Spending{
			SpendingId:        "spnd_foo",
			AccountId:         "acct_1234",
			BankAccountId:     "bac_1234",
			FundingScheduleId: "fund_1234",
			Name:              "Original",
			Description:       "My expense",
			SpendingType:      models.SpendingTypeGoal,
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           nil,
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		assert.Equal(t, "Test", result.Name)
	})
}
