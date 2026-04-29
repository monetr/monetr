package table_test

import (
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/monetr/monetr/server/validators"
	"github.com/stretchr/testify/assert"
)

func TestFieldRef_Validate(t *testing.T) {
	// A FieldRef must reference exactly one of: a named column (via Name, and
	// the name must appear in the columns attached to the context), or a derived
	// value (via DerivedKind restricted to the known constants). Setting both or
	// neither is rejected, and an unknown DerivedKind value is rejected.
	ctx := table.WithColumns(
		t.Context(),
		[]string{"Date", "Description", "Amount", "Id"},
	)
	cases := []struct {
		name    string
		ref     table.FieldRef
		wantErr string
	}{
		{
			name:    "name only",
			ref:     table.FieldRef{Name: "Date"},
			wantErr: "",
		},
		{
			name:    "derived row number",
			ref:     table.FieldRef{DerivedKind: table.DerivedKindRowNumber},
			wantErr: "",
		},
		{
			name:    "name not in headers",
			ref:     table.FieldRef{Name: "NotPresent"},
			wantErr: "input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\"]. or derivedKind: cannot be blank; name: must be blank.",
		},
		{
			name:    "empty",
			ref:     table.FieldRef{},
			wantErr: "input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank.",
		},
		{
			name:    "both name and derived",
			ref:     table.FieldRef{Name: "Date", DerivedKind: table.DerivedKindRowNumber},
			wantErr: "input must be considered valid by: derivedKind: must be blank. or name: must be blank.",
		},
		{
			name:    "unknown derived kind",
			ref:     table.FieldRef{DerivedKind: table.DerivedKind("bogus")},
			wantErr: "input must be considered valid by: derivedKind: must be blank; name: cannot be blank. or derivedKind: must be one of: [\"rowNumber\"].",
		},
		{
			name:    "name with unknown derived",
			ref:     table.FieldRef{Name: "Date", DerivedKind: table.DerivedKind("bogus")},
			wantErr: "input must be considered valid by: derivedKind: must be blank. or derivedKind: must be one of: [\"rowNumber\"]; name: must be blank.",
		},
		{
			// A Name with a tab in it can't ever match a real column because the
			// Headers themselves are validated through [validators.PrintableUnicode]
			// upstream in [Mapping]. So the In check is what surfaces here, the print
			// rule on Name is more or less defense-in-depth.
			name:    "name with tab",
			ref:     table.FieldRef{Name: "Dat\te"},
			wantErr: "input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\"]. or derivedKind: cannot be blank; name: must be blank.",
		},
		{
			name:    "name with newline",
			ref:     table.FieldRef{Name: "Dat\ne"},
			wantErr: "input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\"]. or derivedKind: cannot be blank; name: must be blank.",
		},
		{
			// "Café" used to fail the print rule, but [validators.PrintableUnicode]
			// is fine with it. The case still fails here because "Café" isn't in this
			// context's column list, which is the more meaningful thing to test now.
			name:    "name not in columns, with non-ASCII",
			ref:     table.FieldRef{Name: "Café"},
			wantErr: "input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\"]. or derivedKind: cannot be blank; name: must be blank.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validators.FlattenValidationError(tc.ref.Validate(ctx))

			if tc.wantErr == "" {
				assert.NoError(t, err, "FieldRef must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "FieldRef must be rejected with the expected message")
			}
		})
	}
}
