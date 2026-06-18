package myownsanity_test

import (
	"encoding/json"
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		assert.Empty(t, set, "a set with no initial values should be empty")
	})

	t.Run("with initial values", func(t *testing.T) {
		set := myownsanity.NewSet("a", "b", "c")
		assert.Len(t, set, 3, "should have all three of the initial values")
		assert.True(t, set.Has("a"), "should contain a")
		assert.True(t, set.Has("b"), "should contain b")
		assert.True(t, set.Has("c"), "should contain c")
	})

	t.Run("deduplicates initial values", func(t *testing.T) {
		// The whole point of a set is that duplicate values collapse into a
		// single entry, so passing the same value twice should only result in
		// one item.
		set := myownsanity.NewSet("a", "a", "b")
		assert.Len(t, set, 2, "duplicate initial values should be collapsed into one")
		assert.True(t, set.Has("a"), "should contain a")
		assert.True(t, set.Has("b"), "should contain b")
	})

	t.Run("works with non-string types", func(t *testing.T) {
		set := myownsanity.NewSet(1, 2, 3, 3)
		assert.Len(t, set, 3, "should dedupe the integer set as well")
		assert.True(t, set.Has(1), "should contain 1")
		assert.False(t, set.Has(4), "should not contain a value that was never added")
	})
}

func TestSet_Has(t *testing.T) {
	t.Run("present and absent", func(t *testing.T) {
		set := myownsanity.NewSet("present")
		assert.True(t, set.Has("present"), "should find a value that was added")
		assert.False(t, set.Has("absent"), "should not find a value that was never added")
	})

	t.Run("empty set has nothing", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		assert.False(t, set.Has("anything"), "an empty set should not contain anything")
	})
}

func TestSet_Add(t *testing.T) {
	t.Run("adds a value", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		set.Add("a")
		assert.True(t, set.Has("a"), "should contain the value after adding it")
		assert.Len(t, set, 1, "should only have the one value")
	})

	t.Run("adding the same value twice is idempotent", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		set.Add("a")
		set.Add("a")
		assert.Len(t, set, 1, "adding the same value twice should not change the size")
	})

	t.Run("returns the set so it can be chained", func(t *testing.T) {
		// Add returns the set itself, this lets us chain calls together which is
		// handy when we want to build a set up inline.
		set := myownsanity.NewSet[string]().Add("a").Add("b")
		assert.Len(t, set, 2, "chained adds should both land in the set")
		assert.True(t, set.Has("a"), "should contain a")
		assert.True(t, set.Has("b"), "should contain b")
	})
}

func TestSet_Remove(t *testing.T) {
	t.Run("removes a value", func(t *testing.T) {
		set := myownsanity.NewSet("a", "b")
		set.Remove("a")
		assert.False(t, set.Has("a"), "should no longer contain the removed value")
		assert.True(t, set.Has("b"), "but should still contain the value we left alone")
		assert.Len(t, set, 1, "should only have the one remaining value")
	})

	t.Run("removing a value that is not present is a no-op", func(t *testing.T) {
		// Deleting a missing key from a map is safe in Go, so removing something
		// that was never in the set should not blow up or change anything.
		set := myownsanity.NewSet("a")
		set.Remove("does not exist")
		assert.Len(t, set, 1, "removing a missing value should not change the set")
		assert.True(t, set.Has("a"), "should still contain the original value")
	})

	t.Run("returns the set so it can be chained", func(t *testing.T) {
		set := myownsanity.NewSet("a", "b", "c").Remove("a").Remove("b")
		assert.Len(t, set, 1, "chained removes should both take effect")
		assert.True(t, set.Has("c"), "should still contain the value we did not remove")
	})
}

func TestSet_MarshalJSON(t *testing.T) {
	t.Run("empty set marshals to an empty array", func(t *testing.T) {
		// We want an empty set to come out as [] and not null, otherwise things
		// consuming the json have to special case the null.
		set := myownsanity.NewSet[string]()
		result, err := json.Marshal(set)
		assert.NoError(t, err, "must be able to marshal an empty set")
		assert.JSONEq(t, `[]`, string(result), "an empty set should marshal to an empty array")
	})

	t.Run("marshals values to an array", func(t *testing.T) {
		set := myownsanity.NewSet("a", "b", "c")
		result, err := json.Marshal(set)
		require.NoError(t, err, "must be able to marshal the set")

		// The order of a map is not deterministic in Go so we cannot just
		// compare the json string directly, instead pull it back out into a
		// slice and make sure all of the values are present.
		var items []string
		require.NoError(t, json.Unmarshal(result, &items), "must be able to read the marshalled set back out")
		assert.ElementsMatch(t, []string{"a", "b", "c"}, items, "all of the values should be present in the json")
	})
}

func TestSet_UnmarshalJSON(t *testing.T) {
	t.Run("reads values from an array", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		err := json.Unmarshal([]byte(`["a","b","c"]`), &set)
		assert.NoError(t, err, "must be able to unmarshal into the set")
		assert.Len(t, set, 3, "should have read all three values")
		assert.True(t, set.Has("a"), "should contain a")
		assert.True(t, set.Has("b"), "should contain b")
		assert.True(t, set.Has("c"), "should contain c")
	})

	t.Run("deduplicates values from the array", func(t *testing.T) {
		set := myownsanity.NewSet[string]()
		err := json.Unmarshal([]byte(`["a","a","b"]`), &set)
		assert.NoError(t, err, "must be able to unmarshal into the set")
		assert.Len(t, set, 2, "duplicate values in the json should be collapsed")
		assert.True(t, set.Has("a"), "should contain a")
		assert.True(t, set.Has("b"), "should contain b")
	})

	t.Run("round trip", func(t *testing.T) {
		// Marshal a set and then read it straight back into a new set, the two
		// should end up identical.
		original := myownsanity.NewSet("a", "b", "c")
		data, err := json.Marshal(original)
		require.NoError(t, err, "must be able to marshal the original set")

		result := myownsanity.NewSet[string]()
		require.NoError(t, json.Unmarshal(data, &result), "must be able to unmarshal back into a set")
		assert.Equal(t, original, result, "the round tripped set should match the original")
	})
}
