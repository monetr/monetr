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

const validRuleset = "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1"

func TestCreateFundingSchedule(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":           "Test",
			"description":    "Foobar",
			"ruleset":        validRuleset,
			"nextRecurrence": util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC().Format(time.RFC3339),
		}

		var result models.FundingSchedule

		issues := schema.CreateFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.EqualValues(t, "Test", result.Name)
		assert.EqualValues(t, "Foobar", result.Description)
		assert.False(t, result.ExcludeWeekends)
		assert.False(t, result.AutoCreateTransaction)
		assert.Nil(t, result.EstimatedDeposit)
		if assert.NotNil(t, result.Ruleset) {
			assert.False(t, result.Ruleset.GetDTStart().IsZero())
		}
	})

	t.Run("excludeWeekends and autoCreateTransaction with estimatedDeposit", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":                  "Test",
			"description":           "Foobar",
			"ruleset":               validRuleset,
			"nextRecurrence":        util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC().Format(time.RFC3339),
			"excludeWeekends":       true,
			"autoCreateTransaction": true,
			"estimatedDeposit":      int64(12345),
		}

		var result models.FundingSchedule

		issues := schema.CreateFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.True(t, result.ExcludeWeekends)
		assert.True(t, result.AutoCreateTransaction)
		if assert.NotNil(t, result.EstimatedDeposit) {
			assert.EqualValues(t, int64(12345), *result.EstimatedDeposit)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"description": "Foobar",
		}

		var result models.FundingSchedule

		issues := schema.CreateFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.EqualValues(t, map[string][]string{
			"name": {
				"is required",
			},
			"ruleset": {
				"is required",
			},
			"nextRecurrence": {
				"is required",
			},
		}, zog.Issues.Flatten(issues))
	})
}

func TestFundingScheduleRuleset(t *testing.T) {
	cases := []struct {
		name    string
		ruleset string
		want    []string
	}{
		{
			name:    "malformed",
			ruleset: "not-a-ruleset",
			want:    []string{"invalid RRule"},
		},
		{
			name:    "missing DTSTART",
			ruleset: "RRULE:FREQ=DAILY",
			want:    []string{"DTSTART is required on rulesets"},
		},
		{
			name:    "unsupported FREQ HOURLY",
			ruleset: "DTSTART:20211231T060000Z\nRRULE:FREQ=HOURLY",
			want:    []string{"FREQ must be one of DAILY, WEEKLY, MONTHLY, YEARLY"},
		},
		{
			name:    "BYHOUR not supported",
			ruleset: "DTSTART:20211231T060000Z\nRRULE:FREQ=DAILY;BYHOUR=10",
			want:    []string{"BYHOUR is not supported"},
		},
		{
			name:    "BYMINUTE not supported",
			ruleset: "DTSTART:20211231T060000Z\nRRULE:FREQ=DAILY;BYMINUTE=30",
			want:    []string{"BYMINUTE is not supported"},
		},
		{
			name:    "BYSECOND not supported",
			ruleset: "DTSTART:20211231T060000Z\nRRULE:FREQ=DAILY;BYSECOND=15",
			want:    []string{"BYSECOND is not supported"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
			input := map[string]any{
				"name":           "Test",
				"description":    "Foobar",
				"ruleset":        tc.ruleset,
				"nextRecurrence": util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC().Format(time.RFC3339),
			}

			var result models.FundingSchedule

			issues := schema.CreateFundingSchedule.Parse(input, &result,
				zog.WithCtxValue("timezone", timezone),
				zog.WithCtxValue("clock", clock.New()),
			)
			assert.EqualValues(t, map[string][]string{
				"ruleset": tc.want,
			}, zog.Issues.Flatten(issues))
		})
	}

	t.Run("valid FREQ variants parse", func(t *testing.T) {
		valid := []string{
			"DTSTART:20211231T060000Z\nRRULE:FREQ=DAILY",
			"DTSTART:20211231T060000Z\nRRULE:FREQ=WEEKLY",
			"DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY",
			"DTSTART:20211231T060000Z\nRRULE:FREQ=YEARLY",
		}
		for _, r := range valid {
			t.Run(r, func(t *testing.T) {
				timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
				input := map[string]any{
					"name":           "Test",
					"description":    "Foobar",
					"ruleset":        r,
					"nextRecurrence": util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC().Format(time.RFC3339),
				}

				var result models.FundingSchedule

				issues := schema.CreateFundingSchedule.Parse(input, &result,
					zog.WithCtxValue("timezone", timezone),
					zog.WithCtxValue("clock", clock.New()),
				)
				assert.Empty(t, zog.Issues.Flatten(issues))
				assert.NotNil(t, result.Ruleset)
			})
		}
	})
}

func TestPatchFundingSchedule(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name": "Updated",
		}

		existing := models.FundingSchedule{
			FundingScheduleId:     "fund_foo",
			AccountId:             "acct_1234",
			BankAccountId:         "bac_1234",
			Name:                  "Original",
			Description:           "My funding schedule",
			Ruleset:               testutils.Must(t, models.NewRuleSet, validRuleset),
			ExcludeWeekends:       false,
			AutoCreateTransaction: false,
			NextRecurrence:        util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
		}
		result := existing

		issues := schema.PatchFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.Equal(t, "Updated", result.Name)
		assert.Equal(t, existing.Description, result.Description)
		assert.Equal(t, existing.ExcludeWeekends, result.ExcludeWeekends)
		assert.Equal(t, existing.AutoCreateTransaction, result.AutoCreateTransaction)
		assert.Equal(t, existing.NextRecurrence, result.NextRecurrence)
	})

	t.Run("wont overwrite with a nil", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":        nil,
			"description": nil,
		}

		existing := models.FundingSchedule{
			FundingScheduleId: "fund_foo",
			Name:              "Original",
			Description:       "My funding schedule",
			Ruleset:           testutils.Must(t, models.NewRuleSet, validRuleset),
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
		}
		result := existing

		issues := schema.PatchFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.Empty(t, zog.Issues.Flatten(issues))
		assert.Equal(t, "Original", result.Name)
		assert.Equal(t, "My funding schedule", result.Description)
	})

	t.Run("invalid ruleset patch", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"ruleset": "DTSTART:20211231T060000Z\nRRULE:FREQ=HOURLY",
		}

		existing := models.FundingSchedule{
			FundingScheduleId: "fund_foo",
			Name:              "Original",
			Description:       "My funding schedule",
			Ruleset:           testutils.Must(t, models.NewRuleSet, validRuleset),
			NextRecurrence:    util.Midnight(time.Now().Add(7*24*time.Hour), timezone).UTC(),
		}
		result := existing

		issues := schema.PatchFundingSchedule.Parse(input, &result,
			zog.WithCtxValue("timezone", timezone),
			zog.WithCtxValue("clock", clock.New()),
		)
		assert.EqualValues(t, map[string][]string{
			"ruleset": {
				"FREQ must be one of DAILY, WEEKLY, MONTHLY, YEARLY",
			},
		}, zog.Issues.Flatten(issues))
	})
}
