package table_test

import (
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

// makeNamedFields builds a slice of n unique FieldRefs whose Name values are
// formatted from pattern (e.g. "F%02d"). Used to drive length-cap and bulk
// validation cases.
func makeNamedFields(pattern string, n int) []table.FieldRef {
	fields := make([]table.FieldRef, n)
	for i := range fields {
		fields[i] = table.FieldRef{Name: fmt.Sprintf(pattern, i)}
	}
	return fields
}

func TestIDSpec_Validate(t *testing.T) {
	// An IDSpec must have a recognized Kind (native or hashed) and at least one
	// FieldRef. Each FieldRef inside Fields is validated against the same column
	// set attached to the context; invalid children surface via validation.Each
	// wrapped as "fields: (<index>: ...)". The column list is padded to 21 unique
	// names so the 20-field length cap and duplicate-field cases below have
	// enough columns to reference.
	columns := []string{"Date", "Description", "Amount", "Id"}
	for i := 0; i < 21; i++ {
		columns = append(columns, fmt.Sprintf("F%02d", i))
	}
	ctx := table.WithColumns(t.Context(), columns)
	cases := []struct {
		name    string
		spec    table.IDSpec
		wantErr string
	}{
		{
			name:    "native with named field",
			spec:    table.IDSpec{Kind: table.IDSpecKindNative, Fields: []table.FieldRef{{Name: "Id"}}},
			wantErr: "",
		},
		{
			name:    "hashed with multiple named",
			spec:    table.IDSpec{Kind: table.IDSpecKindHashed, Fields: []table.FieldRef{{Name: "Date"}, {Name: "Amount"}}},
			wantErr: "",
		},
		{
			name:    "hashed with named and derived",
			spec:    table.IDSpec{Kind: table.IDSpecKindHashed, Fields: []table.FieldRef{{Name: "Date"}, {DerivedKind: table.DerivedKindRowNumber}}},
			wantErr: "",
		},
		{
			name:    "empty",
			spec:    table.IDSpec{},
			wantErr: "failed to validate *table.IDSpec: fields: cannot be blank; kind: cannot be blank.",
		},
		{
			name:    "unknown kind",
			spec:    table.IDSpec{Kind: table.IDSpecKind("bogus"), Fields: []table.FieldRef{{Name: "Id"}}},
			wantErr: "failed to validate *table.IDSpec: kind: must be one of: [\"native\", \"hashed\"].",
		},
		{
			name:    "valid kind, no fields",
			spec:    table.IDSpec{Kind: table.IDSpecKindNative, Fields: []table.FieldRef{}},
			wantErr: "failed to validate *table.IDSpec: fields: cannot be blank.",
		},
		{
			name:    "valid kind, blank child",
			spec:    table.IDSpec{Kind: table.IDSpecKindNative, Fields: []table.FieldRef{{}}},
			wantErr: "failed to validate *table.IDSpec: fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..).",
		},
		{
			name:    "valid kind, child with name and derived",
			spec:    table.IDSpec{Kind: table.IDSpecKindNative, Fields: []table.FieldRef{{Name: "Id", DerivedKind: table.DerivedKindRowNumber}}},
			wantErr: "failed to validate *table.IDSpec: fields: (0: input must be considered valid by: derivedKind: must be blank. or name: must be blank..).",
		},
		{
			name:    "valid kind, child not in headers",
			spec:    table.IDSpec{Kind: table.IDSpecKindNative, Fields: []table.FieldRef{{Name: "NotPresent"}}},
			wantErr: "failed to validate *table.IDSpec: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\", \"F00\", \"F01\", \"F02\", \"F03\", \"F04\", \"F05\", \"F06\", \"F07\", \"F08\", \"F09\", \"F10\", \"F11\", \"F12\", \"F13\", \"F14\", \"F15\", \"F16\", \"F17\", \"F18\", \"F19\", \"F20\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
		{
			// 20 unique fields; exactly at the Length(1, 20) upper bound.
			name: "hashed at max field count",
			spec: table.IDSpec{
				Kind:   table.IDSpecKindHashed,
				Fields: makeNamedFields("F%02d", 20),
			},
			wantErr: "",
		},
		{
			// 21 unique fields; exceeds the cap.
			name: "hashed over max field count",
			spec: table.IDSpec{
				Kind:   table.IDSpecKindHashed,
				Fields: makeNamedFields("F%02d", 21),
			},
			wantErr: "failed to validate *table.IDSpec: fields: the length must be between 1 and 20.",
		},
		{
			// Two identical FieldRefs pass Length (2 is in [1, 20]) so Unique is what
			// rejects them. This path isn't reachable from any other IDSpec case
			// today because length was previously unbounded.
			name: "hashed with duplicate fields",
			spec: table.IDSpec{
				Kind:   table.IDSpecKindHashed,
				Fields: []table.FieldRef{{Name: "Id"}, {Name: "Id"}},
			},
			wantErr: "failed to validate *table.IDSpec: fields: fields[1] is a duplicate of an earlier entry.",
		},
		{
			// Two identical derived FieldRefs fail Unique the same way.
			name: "hashed with duplicate derived",
			spec: table.IDSpec{
				Kind:   table.IDSpecKindHashed,
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumber}, {DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "failed to validate *table.IDSpec: fields: fields[1] is a duplicate of an earlier entry.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate(ctx)
			if tc.wantErr == "" {
				assert.NoError(t, err, "IDSpec must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "IDSpec must be rejected with the expected message")
			}
		})
	}
}
