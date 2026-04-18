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
		// TODO Make test precise
		assert.NotEmpty(t, zog.Issues.Prettify(issues))
		j, _ := json.MarshalIndent(zog.Issues.Flatten(issues), "", "  ")
		fmt.Println(string(j))
		assert.EqualValues(t, "bac_1234", result.FundingScheduleId)
	})
}
