package table_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

func TestPostedSpec_Validate(t *testing.T) {
	// PostedSpec requires exactly one FieldRef (valid against the column set in
	// context). Posted is optional; when set it must contain printable
	// characters only (per [validators.PrintableUnicode], so umlauts and the
	// like are fine but tabs and zero-width joiners are not) and be 1 to 100
	// characters long. An empty Posted string is still accepted because
	// Length(1, 100) treats empty as valid (the rule only fires once the value
	// is non-empty).
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
			name: "named field, empty posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
			},
			wantErr: "",
		},
		{
			name: "named field, short ascii",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "Posted",
			},
			wantErr: "",
		},
		{
			name: "named field, mid-range ascii",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 50),
			},
			wantErr: "",
		},
		{
			name: "named field, at min length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "A",
			},
			wantErr: "",
		},
		{
			name: "named field, ascii punctuation",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "posted! @#$%^&*()_+-=",
			},
			wantErr: "",
		},
		{
			name: "row number field, empty posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "",
		},
		{
			name:    "empty",
			spec:    table.PostedSpec{},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "nil fields, valid posted",
			spec: table.PostedSpec{
				Posted: "POSTED",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "empty fields, valid posted",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{},
				Posted: "POSTED",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "two named fields",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Date"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			name: "three fields",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Date"}, {Name: "Amount"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			// Length(1,1) fails before the Unique rule runs, so the message reports
			// the length error rather than "duplicate".
			name: "duplicates caught by length check",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}, {Name: "Status"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: the length must be exactly 1.",
		},
		{
			name: "child field empty",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..).",
		},
		{
			name: "child field not in headers",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "NotPresent"}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Status\", \"Description\", \"Amount\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
		{
			name: "child field with name and derived",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status", DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: derivedKind: must be blank. or name: must be blank..).",
		},
		{
			name: "child field with unknown derived",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKind("bogus")}},
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: derivedKind: must be blank; name: cannot be blank. or derivedKind: must be one of: [\"rowNumber\"]..).",
		},
		{
			// "Pösted" is the canonical reason monetr swapped from
			// is.PrintableASCII to [validators.PrintableUnicode]. It used to
			// fail here, now it's a perfectly fine Posted needle.
			name: "posted with umlauts",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "Pösted",
			},
			wantErr: "",
		},
		{
			// Tab still gets caught because [unicode.IsPrint] only treats
			// the regular ASCII space as printable, not the C0 controls.
			name: "posted with tab",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POS\tTED",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable characters only.",
		},
		{
			name: "posted with newline",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POS\nTED",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable characters only.",
		},
		{
			// DEL (0x7F) is still ASCII but it's outside the printable range
			// so the new rule rejects it the same way the old one did.
			name: "posted with DEL",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: "POSTED\x7f",
			},
			wantErr: "failed to validate *table.PostedSpec: posted: must contain printable characters only.",
		},
		{
			name: "posted at max length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 100),
			},
			wantErr: "",
		},
		{
			name: "posted over max length",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{Name: "Status"}},
				Posted: strings.Repeat("A", 101),
			},
			wantErr: "failed to validate *table.PostedSpec: posted: the length must be between 1 and 100.",
		},
		{
			// Pösted is now fine on its own, so the only problem left here is
			// that fields is blank. The case stays around to make sure the
			// fields error still surfaces when the Posted needle happens to
			// contain non-ASCII content.
			name: "empty fields, posted with umlauts",
			spec: table.PostedSpec{
				Posted: "Pösted",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank.",
		},
		{
			name: "empty fields, over-length posted",
			spec: table.PostedSpec{
				Posted: strings.Repeat("A", 101),
			},
			wantErr: "failed to validate *table.PostedSpec: fields: cannot be blank; posted: the length must be between 1 and 100.",
		},
		{
			// Same idea as the previous case but with the child FieldRef
			// also broken. Since Pösted itself is fine now, only the fields
			// error makes it through.
			name: "invalid child, posted with umlauts",
			spec: table.PostedSpec{
				Fields: []table.FieldRef{{}},
				Posted: "Pösted",
			},
			wantErr: "failed to validate *table.PostedSpec: fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..).",
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
