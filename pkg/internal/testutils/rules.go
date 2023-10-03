package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/require"
)

func NewRuleSet(t *testing.T, year, month, day int, timezone *time.Location, rule string) *models.RuleSet {
	ruleString := fmt.Sprintf(
		"DTSTART:%s%sRRULE:%s",
		time.Date(year, time.Month(month), day, 0, 0, 0, 0, timezone).UTC().Format("20060102T150405Z"),
		`\n`,
		rule,
	)

	set, err := models.NewRuleSet(ruleString)
	require.NoError(t, err, "must be able to parse rule and start into ruleset: %s", ruleString)

	return set
}
