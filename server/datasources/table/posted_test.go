package table_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

func TestPostedSpec_Validate(t *testing.T) {
	// PostedSpec requires exactly one FieldRef (valid against the column set in
	// context). Posted is optional; when set it must contain ASCII characters
	// only and be no longer than 50 characters. An empty Posted string is
	// accepted because Length(0, 50) treats empty as valid.
	ctx := table.WithColumns(
		t.Context(),
		[]string{"Date", "Status", "Description", "Amount"},
	)
	cases := []struct {
		name    string
		spec    table.PostedSpec
		wantErr string
	}{
		{
			name: "named field with empty Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
			},
			wantErr: "",
		},
		{
			name: "named field with short ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "Posted",
			},
			wantErr: "",
		},
		{
			name: "named field with mid-range ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 50),
			},
			wantErr: "",
		},
		{
			name: "named field with Posted at min non-empty length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "A",
			},
			wantErr: "",
		},
		{
			name: "named field with ASCII punctuation Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "posted! @#$%^&*()_+-=",
			},
			wantErr: "",
		},
		{
			name: "derived row number field with empty Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "",
		},
		{
			name: "derived row number per day field with ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumberPerDay}},
				Posted: "POSTED",
			},
			wantErr: "",
		},
		{
			name: "derived row number per day per amount field with ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumberPerDayPerAmount}},
				Posted: "Y",
			},
			wantErr: "",
		},
		{
			name:    "empty",
			spec:    table.PostedSpec{},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "nil Fields with valid Posted",
			spec: table.PostedSpec{
				Posted: "POSTED",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "empty slice Fields with valid Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{},
				Posted: "POSTED",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "two named fields violates length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Date"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			name: "three fields violates length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Date"}, {Name: "Amount"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			// Length(1,1) fails before the Unique rule runs, so the message reports
			// the length error rather than "duplicate".
			name: "duplicate fields are caught by length check first",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Status"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			name: "child field is empty",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..).",
		},
		{
			name: "child field name not in headers",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "NotPresent"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Status\", \"Description\", \"Amount\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
		{
			name: "child field has both name and derived kind",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status", DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: derivedKind: must be blank. or name: must be blank..).",
		},
		{
			name: "child field has unknown derived kind",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKind("bogus")}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: derivedKind: must be blank; name: cannot be blank. or derivedKind: must be one of: [\"rowNumber\", \"rowNumberPerDay\", \"rowNumberPerDayPerAmount\"]..).",
		},
		{
			name: "non-ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "Pösted",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable ASCII characters only.",
		},
		{
			// Tab is ASCII (0x09) but not *printable* ASCII (0x20-0x7E). The switch
			// from is.ASCII to is.PrintableASCII is what rejects it.
			name: "Posted with tab character",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POS\tTED",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable ASCII characters only.",
		},
		{
			name: "Posted with newline character",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POS\nTED",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable ASCII characters only.",
		},
		{
			// DEL (0x7F) is still ASCII but sits outside the printable range.
			name: "Posted with DEL character",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POSTED\x7f",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable ASCII characters only.",
		},
		{
			name: "Posted at 100-char max length boundary",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 100),
			},
			wantErr: "",
		},
		{
			name: "Posted exceeds max length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 101),
			},
			wantErr: "failed to validate *table.PostedSpec: posted: the length must be no more than 100.",
		},
		{
			name: "empty Fields with non-ASCII Posted",
			spec: table.PostedSpec{
				Posted: "Pösted",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank; posted: must contain printable ASCII characters only.",
		},
		{
			name: "empty Fields with over-length Posted",
			spec: table.PostedSpec{
				Posted: strings.Repeat("A", 101),
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank; posted: the length must be no more than 100.",
		},
		{
			name: "invalid child field with non-ASCII Posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{}},
				Posted: "Pösted",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..); posted: must contain printable ASCII characters only.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate(ctx)
			if tc.wantErr == "" {
				assert.NoError(t, err, "PostedSpec must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "PostedSpec must be rejected with the expected message")
			}
		})
	}
}
