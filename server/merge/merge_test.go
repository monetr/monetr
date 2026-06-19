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

type Embedded struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

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

	t.Run("null clears a pointer field", func(t *testing.T) {
		// An explicit null in the source body should actually unset a nullable
		// destination field. This is the bug that meant a client could never clear
		// a bank account mask, the null was just silently ignored and the existing
		// value was left in place.
		type Foo struct {
			Bar *string `json:"bar"`
		}

		existing := "this is a string"
		dst := Foo{
			Bar: &existing,
		}
		src := map[string]any{
			"bar": nil,
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "must be able to merge an explicit null")
		assert.Nil(t, dst.Bar, "the null should have cleared the pointer field")
	})

	t.Run("null does not affect a non nullable field", func(t *testing.T) {
		// A null cannot be represented by a non pointer field, so just like
		// encoding/json we leave the destination alone rather than zeroing it out.
		// This matters because the validation layer does NOT reject a null on a non
		// nullable field, so without this a client could accidentally wipe a name or
		// zero out a balance just by sending null.
		type Foo struct {
			Name    string `json:"name"`
			Balance int64  `json:"balance"`
		}

		dst := Foo{
			Name:    "do not clobber me",
			Balance: 1234,
		}
		src := map[string]any{
			"name":    nil,
			"balance": nil,
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "a null on a non nullable field should not error")
		assert.Equal(t, "do not clobber me", dst.Name, "the null should not have changed the string field")
		assert.EqualValues(t, 1234, dst.Balance, "the null should not have changed the integer field")
	})

	t.Run("null skips when skipping zero values", func(t *testing.T) {
		// When the caller asked us to skip zero values a null is just another zero
		// so we leave the nullable field exactly as it was. This also proves we do
		// not panic trying to call IsZero on an invalid value.
		type Foo struct {
			Bar *string `json:"bar"`
		}

		existing := "this is a string"
		dst := Foo{
			Bar: &existing,
		}
		src := map[string]any{
			"bar": nil,
		}

		err := merge.Merge(&dst, src, merge.SkipZeroValues)
		assert.NoError(t, err, "must be able to merge an explicit null while skipping zero values")
		assert.NotNil(t, dst.Bar, "the pointer should be left alone when skipping zero values")
		assert.Equal(t, existing, *dst.Bar, "the existing value should be untouched")
	})

	t.Run("merge embedded struct fields", func(t *testing.T) {
		type Foo struct {
			Embedded

			AdjustsBalance bool `json:"adjustsBalance"`
		}

		dst := Foo{}
		src := map[string]any{
			"name":           "this is a string",
			"amount":         json.Number("12345"),
			"adjustsBalance": true,
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge into the promoted fields of an embedded struct")
		assert.Equal(t, src["name"], dst.Name, "the promoted name field should be merged")
		assert.EqualValues(t, 12345, dst.Amount, "the promoted amount field should be merged")
		assert.True(t, dst.AdjustsBalance, "the wrapper's own field should be merged")
	})

	t.Run("merge embedded struct fields handle duplicate", func(t *testing.T) {
		type Foo struct {
			Embedded

			// Duplicate of field on embedded
			Name           string
			AdjustsBalance bool `json:"adjustsBalance"`
		}

		dst := Foo{}
		src := map[string]any{
			"name":           "this is a string",
			"amount":         json.Number("12345"),
			"adjustsBalance": true,
		}

		err := merge.Merge(&dst, src)
		assert.EqualError(t, err, "duplicate field in destination struct: Name")
		assert.Empty(t, dst)
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

	t.Run("handle weird nullable custom types", func(t *testing.T) {
		type Foo struct {
			SpendingId *models.ID[models.Spending] `json:"spendingId"`
		}

		dst := Foo{}
		src := map[string]any{
			"spendingId": "spnd_12345",
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge with odd destination type aliases")
		// Make sure that we parsed it properly!
		assert.EqualValues(t, src["spendingId"], *dst.SpendingId)
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

	t.Run("handle unsigned json numbers", func(t *testing.T) {
		type Foo struct {
			Amount uint64 `json:"amount"`
		}

		dst := Foo{}
		src := map[string]any{
			"amount": json.Number("12345"),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge a json number into an unsigned field")
		assert.EqualValues(t, 12345, dst.Amount, "amount should be merged properly!")
	})

	t.Run("handle unsigned json pointer numbers", func(t *testing.T) {
		type Foo struct {
			Amount *uint64 `json:"amount"`
		}

		dst := Foo{}
		src := map[string]any{
			"amount": json.Number("12345"),
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge a json number into an unsigned pointer field")
		assert.EqualValues(t, 12345, *dst.Amount, "amount should be merged properly!")
	})

	t.Run("handle unsigned json numbers larger than the max int64", func(t *testing.T) {
		// This is the whole reason we parse the string as a uint instead of going
		// through Int64. A value like this fits just fine in a uint64 but Int64
		// would choke on it, so make sure the full range actually works.
		type Foo struct {
			Amount uint64 `json:"amount"`
		}

		dst := Foo{}
		src := map[string]any{
			"amount": json.Number("18446744073709551615"), // math.MaxUint64
		}

		err := merge.Merge(&dst, src)
		assert.NoError(t, err, "Must be able to merge a uint64 value that does not fit in an int64")
		assert.EqualValues(t, uint64(18446744073709551615), dst.Amount, "the max uint64 should merge properly")
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
