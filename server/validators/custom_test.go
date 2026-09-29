package validators_test

import (
	"testing"

	"github.com/monetr/monetr/server/validators"
	"github.com/stretchr/testify/assert"
)

func TestUnique_Strings(t *testing.T) {
	rule := validators.Unique[string]()
	cases := []struct {
		name    string
		input   []string
		wantErr string
	}{
		{
			name:    "empty",
			input:   []string{},
			wantErr: "",
		},
		{
			name:    "single element",
			input:   []string{"a"},
			wantErr: "",
		},
		{
			name:    "all unique",
			input:   []string{"a", "b", "c"},
			wantErr: "",
		},
		{
			name:    "duplicate at index 1",
			input:   []string{"a", "a"},
			wantErr: "fields[1] is a duplicate of an earlier entry",
		},
		{
			name:    "non-adjacent duplicate surfaces at the later index",
			input:   []string{"a", "b", "a"},
			wantErr: "fields[2] is a duplicate of an earlier entry",
		},
		{
			name:    "three identical reports the first repeat",
			input:   []string{"x", "x", "x"},
			wantErr: "fields[1] is a duplicate of an earlier entry",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Validate(tc.input)
			if tc.wantErr == "" {
				assert.NoError(t, err, "slice must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "slice must be rejected with the expected message")
			}
		})
	}
}

func TestUnique_Ints(t *testing.T) {
	rule := validators.Unique[int]()

	assert.NoError(t, rule.Validate([]int{1, 2, 3}), "unique ints accepted")
	assert.EqualError(t, rule.Validate([]int{1, 2, 1}), "fields[2] is a duplicate of an earlier entry", "repeated int flagged")
}

func TestUnique_ComparableStruct(t *testing.T) {
	// Unique works on any `comparable` type, including structs whose fields are
	// all comparable. This mirrors how it's used against FieldRef in the csv
	// package: two entries are duplicates only if every field matches.
	type pair struct {
		Name string
		Kind string
	}

	rule := validators.Unique[pair]()

	assert.NoError(
		t,
		rule.Validate([]pair{{Name: "Date"}, {Name: "Amount"}}),
		"structurally distinct structs accepted",
	)
	assert.NoError(
		t,
		rule.Validate([]pair{{Name: "Amount"}, {Kind: "rowNumber"}}),
		"same zero field is fine when the other differs",
	)
	assert.EqualError(
		t,
		rule.Validate([]pair{{Name: "Date"}, {Name: "Date"}}),
		"fields[1] is a duplicate of an earlier entry",
		"fully equal structs flagged",
	)
}

func TestUnique_PointerToSlice(t *testing.T) {
	// Struct field validation can hand the rule a pointer to the slice rather
	// than the slice itself.
	rule := validators.Unique[string]()

	unique := []string{"a", "b"}
	duplicate := []string{"a", "b", "a"}
	var missing *[]string

	assert.NoError(t, rule.Validate(&unique), "pointer to unique slice accepted")
	assert.EqualError(t, rule.Validate(&duplicate), "fields[2] is a duplicate of an earlier entry", "pointer to slice with repeat flagged")
	assert.NoError(t, rule.Validate(missing), "nil pointer to slice is skipped")
}

func TestUnique_DecodedJSONArray(t *testing.T) {
	// Schemas validate the raw decoded request body, so arrays arrive as []any
	// rather than a typed slice.
	rule := validators.Unique[string]()

	assert.NoError(t, rule.Validate([]any{}), "empty decoded array accepted")
	assert.NoError(t, rule.Validate([]any{"a", "b", "c"}), "unique decoded array accepted")
	assert.EqualError(t, rule.Validate([]any{"a", "b", "a"}), "fields[2] is a duplicate of an earlier entry", "decoded array with repeat flagged")
	assert.EqualError(t, rule.Validate([]any{"a", 1}), "fields[1] is not a valid value", "element of the wrong type rejected")
}

func TestUnique_NotASlice(t *testing.T) {
	// Values that are not a slice of T are left for other rules to reject.
	rule := validators.Unique[string]()

	assert.NoError(t, rule.Validate(nil), "nil is skipped")
	assert.NoError(t, rule.Validate("a"), "non-slice value is skipped")
	assert.NoError(t, rule.Validate([]int{1, 1}), "slice of another type is skipped")
}
