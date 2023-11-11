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
