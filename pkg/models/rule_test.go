package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRule(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		input := "FREQ=MONTHLY;BYMONTHDAY=15,-1"
		rule, err := NewRule(input)
		assert.NoError(t, err, "must be able to parse semi-monthly rule")
		nextRecurrence := rule.After(time.Date(2022, 4, 5, 0, 0, 0, 0, time.UTC), false).Truncate(24 * time.Hour)
		assert.Equal(t, time.Date(2022, 4, 15, 0, 0, 0, 0, time.UTC), nextRecurrence, "next recurrence should be equal")
	})
}

func TestRuleString(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		input := "FREQ=MONTHLY;BYMONTHDAY=15,-1"
		rule, err := NewRule(input)
		assert.NoError(t, err, "must be able to parse semi-monthly rule")
		assert.NotEmpty(t, rule.String(), "should produce a string")
		assert.Equal(t, input, rule.RRule.OrigOptions.RRuleString(), "should produce the expect string rule")
	})
}

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
