package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

func NewRuleSet(t *testing.T, year, month, day int, timezone *time.Location, rule string) *models.RuleSet {
	ruleString := fmt.Sprintf(
		"DTSTART:%s\nRRULE:%s",
		time.Date(year, time.Month(month), day, 0, 0, 0, 0, timezone).UTC().Format("20060102T150405Z"),
		rule,
	)

	set, err := models.NewRuleSet(ruleString)
	require.NoError(t, err, "must be able to parse rule and start into ruleset: %s", ruleString)

	return set
}

func RuleToSet(t *testing.T, timezone *time.Location, ruleString string, now time.Time) *models.RuleSet {
	rule, err := rrule.StrToRRule(ruleString)
	require.NoError(t, err, "must be able to parse rule string")
	rule.DTStart(now)

	after := rule.After(now, false)
	dtstart := util.Midnight(after, timezone)

	ruleSetString := fmt.Sprintf(
		"DTSTART:%s\nRRULE:%s",
		dtstart.UTC().Format("20060102T150405Z"),
		ruleString,
	)

	set, err := models.NewRuleSet(ruleSetString)
	require.NoError(t, err, "must be able to parse rule and start into ruleset: %s", ruleSetString)

	return set
}

// RuleSetInTimezone is similar to [RuleToSet] except this one does not take
// only a rule string, instead taking a ruleset string complete with a dtstart
// parameter. It converts the dtstart parameter into a midnight timestamp in the
// specified time zone.
func RuleSetInTimezone(t *testing.T, timezone *time.Location, ruleset string) *models.RuleSet {
	set, err := models.NewRuleSet(ruleset)
	require.NoError(t, err, "Must not encounter an error building the rule set from a string")

	midnight := util.Midnight(set.GetDTStart(), timezone)
	set.DTStart(midnight)

	return set
}
