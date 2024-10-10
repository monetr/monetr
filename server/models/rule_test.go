package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRuleSet_MarshalJSON(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		set, err := NewRuleSet("DTSTART:20230831T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		assert.NoError(t, err, "should not have an error parsing the rule set string")
		assert.NotNil(t, set, "returned rule set should not be nil")

		after := set.After(time.Date(2023, 9, 28, 0, 0, 0, 0, time.UTC), false)
		assert.Equal(t, time.Date(2023, 9, 30, 5, 0, 0, 0, time.UTC), after, "should end up being the 30th with a 5am offset")

		var data struct {
			RuleSet       *RuleSet `json:"ruleSet"`
			NewLineSample string   `json:"newlineSample"`
		}
		data.RuleSet = set
		data.NewLineSample = "foo\nbar"

		j, err := json.Marshal(data)
		assert.NoError(t, err, "should not return an error when converting to json")
		assert.JSONEq(t, `{
			"ruleSet": "DTSTART:20230831T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			"newlineSample": "foo\nbar"
		}`, string(j))
	})
}

func TestRuleSet_UnmarshalJSON(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var data struct {
			RuleSet       *RuleSet `json:"ruleSet"`
			NewLineSample string   `json:"newlineSample"`
		}

		jsonBlob := []byte(`{
			"ruleSet": "DTSTART:20230831T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			"newlineSample": "foo\nbar"
		}`)

		err := json.Unmarshal(jsonBlob, &data)
		assert.NoError(t, err, "should not return an error when parsing json")

		assert.Equal(
			t,
			"DTSTART:20230831T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			data.RuleSet.String(),
			"should have parsed the ruleset properly",
		)

		// Make sure we get the same result as we did when when we built it normally.
		after := data.RuleSet.After(time.Date(2023, 9, 28, 0, 0, 0, 0, time.UTC), false)
		assert.Equal(t, time.Date(2023, 9, 30, 5, 0, 0, 0, time.UTC), after, "should end up being the 30th with a 5am offset")
	})
}

func TestRuleSet_DaylightSavingsTime(t *testing.T) {
	t.Run("2024 daylight savings time transition", func(t *testing.T) {
		ruleString := "DTSTART:20230401T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1"
		rule, err := NewRuleSet(ruleString)
		assert.NoError(t, err, "must be able to parse the rule")

		timezone, err := time.LoadLocation("America/Chicago")
		assert.NoError(t, err, "must be able to load the central time timezone")

		expected := time.Date(2025, 01, 01, 0, 0, 0, 0, timezone)
		now, err := time.Parse(time.RFC3339, "2024-10-08T22:15:04.541Z")
		assert.NoError(t, err, "must be able to get now")

		// Without the fix
		assert.NotEqual(t, expected.UTC(), rule.After(now, false).UTC(), "does not handle the DST transition properly")

		// Fix for bug
		rule.DTStart(rule.GetDTStart().In(timezone))

		// With the fix
		assert.Equal(t, expected.UTC(), rule.After(now, false).UTC(), "DST transition is now correct")
	})
}
