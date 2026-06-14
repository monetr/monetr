package schemas_test

import (
	"testing"

	"github.com/monetr/monetr/server/schemas"
	"github.com/stretchr/testify/assert"
)

func TestRuleset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// A plain monthly rule with a dtstart like we use for funding schedules
		// should be totally fine.
		err := schemas.Ruleset().Validate("DTSTART:20230831T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		assert.NoError(t, err, "a basic monthly ruleset with a dtstart should be valid")
	})

	t.Run("empty string is allowed", func(t *testing.T) {
		// The ruleset rule does not enforce presence, thats what the Required rule
		// on the schema is for. An empty value short circuits inside the validation
		// library before our function even runs so it comes back clean here.
		err := schemas.Ruleset().Validate("")
		assert.NoError(t, err, "an empty ruleset is left for the Required rule to catch, not this one")
	})

	t.Run("garbage is not a valid ruleset", func(t *testing.T) {
		// This one cant even be parsed into a rule so NewRuleSet fails and we
		// reject it.
		err := schemas.Ruleset().Validate("this is definitely not a ruleset")
		assert.EqualError(t, err, "Ruleset must be valid", "an unparseable ruleset should be rejected")
	})

	t.Run("requires a dtstart", func(t *testing.T) {
		// A bare RRULE without a DTSTART will actually parse, the rrule library
		// just gives it a zero value start time. monetr does not want that so we
		// reject any ruleset that did not explicitly say when it starts.
		err := schemas.Ruleset().Validate("RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		assert.EqualError(t, err, "Ruleset must be valid", "a ruleset without a dtstart should be rejected")
	})

	t.Run("rejects a ruleset with byhour", func(t *testing.T) {
		// monetr only cares about day level precision for these rules. A BYHOUR
		// parses just fine as an rrule but we dont want it because it implies a
		// time of day we are not going to honor, so we reject it on purpose.
		err := schemas.Ruleset().Validate("DTSTART:20230831T050000Z\nRRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=10")
		assert.EqualError(t, err, "Ruleset must be valid", "a ruleset with a by hour component should be rejected")
	})

	t.Run("rejects a ruleset with byminute", func(t *testing.T) {
		// Same idea as byhour, a minute level rule is more precision than monetr
		// wants to deal with.
		err := schemas.Ruleset().Validate("DTSTART:20230831T050000Z\nRRULE:FREQ=DAILY;INTERVAL=1;BYMINUTE=30")
		assert.EqualError(t, err, "Ruleset must be valid", "a ruleset with a by minute component should be rejected")
	})

	t.Run("rejects a ruleset with bysecond", func(t *testing.T) {
		// And the same for seconds, theres no world where we want a funding
		// schedule firing on a specific second.
		err := schemas.Ruleset().Validate("DTSTART:20230831T050000Z\nRRULE:FREQ=DAILY;INTERVAL=1;BYSECOND=15")
		assert.EqualError(t, err, "Ruleset must be valid", "a ruleset with a by second component should be rejected")
	})
}
