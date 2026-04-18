package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Oudwins/zog"
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
			"nextRecurrence":    util.Midnight(time.Now(), timezone).UTC().Add(1 * time.Minute).Format(time.RFC3339),
		}

		var result models.Spending

		issues := schema.CreateSpending.Parse(input, &result, zog.WithCtxValue("timezone", timezone))
		assert.Empty(t, zog.Issues.Prettify(issues))
		j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
		fmt.Println(string(j))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		j, _ = json.MarshalIndent(input, "", "  ")
		fmt.Println(string(j))
		j, _ = json.MarshalIndent(result, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("invalid funding schedule", func(t *testing.T) {
		input := map[string]any{
			"name":              "Test",
			"description":       "Foobar",
			"fundingScheduleId": "bac_1234",
		}

		var result models.Spending

		issues := schema.CreateSpending.Parse(input, &result)
		assert.EqualValues(t, map[string][]string{
			"fundingScheduleId": []string{
				`expected id with prefix "fund"`,
			},
			"nextRecurrence": []string{
				"is required",
			},
			"ruleset": []string{
				"expenses must have a ruleset",
			},
			"targetAmount": []string{
				"is required",
			},
		}, zog.Issues.Flatten(issues))
		// TODO Make test precise
		assert.NotEmpty(t, zog.Issues.Prettify(issues))
		j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
		fmt.Println(string(j))
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
			SpendingType:      "expense",
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           testutils.Must(t, models.NewRuleSet, "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1"),
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result, zog.WithCtxValue("timezone", timezone))
		assert.Empty(t, zog.Issues.Prettify(issues))
		j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
		fmt.Println(string(j))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		j, _ = json.MarshalIndent(input, "", "  ")
		fmt.Println(string(j))
		j, _ = json.MarshalIndent(result, "", "  ")
		fmt.Println(string(j))
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
			SpendingType:      "goal",
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           nil,
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result, zog.WithCtxValue("timezone", timezone))
		assert.EqualValues(t, map[string][]string{
			"ruleset": []string{
				"goals cannot have a ruleset",
			},
		}, zog.Issues.Flatten(issues))
		j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
		fmt.Println(string(j))
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
		j, _ = json.MarshalIndent(input, "", "  ")
		fmt.Println(string(j))
		j, _ = json.MarshalIndent(result, "", "  ")
		fmt.Println(string(j))
		assert.Equal(t, "Test", result.Name)
	})

	t.Run("wont overwrite with a nil", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":              "Test",
			"fudningScheduleId": nil,
		}

		result := models.Spending{
			SpendingId:        "spnd_foo",
			AccountId:         "acct_1234",
			BankAccountId:     "bac_1234",
			FundingScheduleId: "fund_1234",
			Name:              "Original",
			Description:       "My expense",
			SpendingType:      "goal",
			TargetAmount:      2000,
			CurrentAmount:     1000,
			UsedAmount:        0,
			Ruleset:           nil,
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
			IsPaused:          false,
			CreatedAt:         time.Now(),
		}

		issues := schema.PatchSpending.Parse(input, &result, zog.WithCtxValue("timezone", timezone))
		assert.Empty(t, zog.Issues.Flatten(issues))
		{ // Issues
			fmt.Println("=== Issues ===")
			j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
			fmt.Println(string(j))
		}
		{ // Input
			fmt.Println("=== Input ===")
			j, _ := json.MarshalIndent(input, "", "  ")
			fmt.Println(string(j))
		}
		{ // Result
			fmt.Println("=== Result ===")
			j, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(j))
		}
		assert.EqualValues(t, "fund_1234", result.FundingScheduleId)
	})
}
