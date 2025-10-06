package merge_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/teambition/rrule-go"
)

func TestMerge(t *testing.T) {
	t.Run("cannot merge into a non struct", func(t *testing.T) {
		var dst int
		src := map[string]any{
			"bar": "this is a string",
		}
		err := merge.Merge(&dst, src)
		assert.EqualError(t, err, "cannot build field map for destination of type: int")
	})

	t.Run("cannot merge into nil", func(t *testing.T) {
		type Foo struct {
			Bar *string `json:"bar"`
		}
		var dst *Foo = nil
		src := map[string]any{
			"bar": "this is a string",
		}

		err := merge.Merge(dst, src)
		assert.EqualError(t, err, "cannot merge into a nil destination")
	})

	t.Run("merge simple", func(t *testing.T) {
		type Foo struct {
			Bar string `json:"bar"`
		}

		dst := Foo{}
		src := map[string]any{
			"bar": "this is a string",
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge simple structures")
		assert.Equal(t, src["bar"], dst.Bar, "the source and destination fields should now match!")
	})

	t.Run("merge pointer", func(t *testing.T) {
		type Foo struct {
			Bar *string `json:"bar"`
		}

		dst := Foo{}
		src := map[string]any{
			"bar": "this is a string",
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with pointers in the struct")
		assert.Equal(t, src["bar"], *dst.Bar, "the source and destination fields should now match!")
	})

	t.Run("merge using json.Unmarshaller", func(t *testing.T) {
		type Foo struct {
			RuleSet *models.RuleSet `json:"ruleset"`
		}

		dst := Foo{}
		src := map[string]any{
			"ruleset": testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", time.Now()).String(),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with pointers in the struct")
		// Make sure that we parsed it properly!
		assert.EqualValues(t, rrule.MONTHLY, dst.RuleSet.GetRRule().Options.Freq)
		assert.EqualValues(t, 1, dst.RuleSet.GetRRule().Options.Interval)
	})

	t.Run("handle weird custom types", func(t *testing.T) {
		type Foo struct {
			SpendingId models.ID[models.Spending] `json:"spendingId"`
			RuleSet    *models.RuleSet            `json:"ruleset"`
		}

		dst := Foo{}
		src := map[string]any{
			"spendingId": "spnd_12345",
			"ruleset":    testutils.RuleToSet(t, time.UTC, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", time.Now()).String(),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with odd destination type aliases")
		// Make sure that we parsed it properly!
		assert.EqualValues(t, src["spendingId"], dst.SpendingId)
		assert.EqualValues(t, rrule.MONTHLY, dst.RuleSet.GetRRule().Options.Freq)
		assert.EqualValues(t, 1, dst.RuleSet.GetRRule().Options.Interval)
	})

	t.Run("handle json numbers", func(t *testing.T) {
		type Foo struct {
			Amount int64 `json:"amount"`
		}

		dst := Foo{}
		src := map[string]any{
			"amount": json.Number("12345"),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with json numbers in the source")
		assert.EqualValues(t, 12345, dst.Amount, "amount should be merged properly!")
	})

	t.Run("handle json pointer numbers", func(t *testing.T) {
		type Foo struct {
			Amount *int64 `json:"amount"`
		}

		dst := Foo{}
		src := map[string]any{
			"amount": json.Number("12345"),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with json numbers in the source")
		assert.EqualValues(t, 12345, *dst.Amount, "amount should be merged properly!")
	})

	t.Run("handle timestamps", func(t *testing.T) {
		type Foo struct {
			Timestamp time.Time `json:"timestamp"`
		}

		now := time.Now()
		dst := Foo{}
		src := map[string]any{
			"timestamp": now.Format(time.RFC3339Nano),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with dates and strings")
		assert.EqualValues(t, now.Unix(), dst.Timestamp.Unix(), "timestamp should merge properly")
	})
}
