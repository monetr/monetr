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
