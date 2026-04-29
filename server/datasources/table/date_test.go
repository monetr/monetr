package table_test

import (
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

func TestDateSpec_Validate(t *testing.T) {
	// DateSpec must have exactly one FieldRef (valid against the column set in
	// context) and a non-empty human-readable Format like YYYY-MM-DD.
	ctx := table.WithColumns(
		t.Context(),
		[]string{"Date", "PostDate", "Amount"},
	)
	cases := []struct {
		name    string
		spec    table.DateSpec
		wantErr string
	}{
		{
			name: "YYYY-MM-DD",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DD",
			},
			wantErr: "",
		},
		{
			name: "MM/DD/YYYY",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "MM/DD/YYYY",
			},
			wantErr: "",
		},
		{
			name: "M/D/YYYY single-digit month and day",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "M/D/YYYY",
			},
			wantErr: "",
		},
		{
			name: "DD-MM-YY two-digit year",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "DD-MM-YY",
			},
			wantErr: "",
		},
		{
			name:    "empty",
			spec:    table.DateSpec{},
			wantErr: "failed to validate *table.DateSpec: fields: cannot be blank; format: cannot be blank.",
		},
		{
			name:    "missing fields",
			spec:    table.DateSpec{Format: "YYYY-MM-DD"},
			wantErr: "failed to validate *table.DateSpec: fields: cannot be blank.",
		},
		{
			name: "missing format",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
			},
			wantErr: "failed to validate *table.DateSpec: format: cannot be blank.",
		},
		{
			name: "two fields violates length",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}, {Name: "PostDate"}},
				Format: "YYYY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: fields: the length must be exactly 1.",
		},
		{
			name: "field name not in headers",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "NotPresent"}},
				Format: "YYYY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"PostDate\", \"Amount\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
		{
			name: "field with both name and derived kind set",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date", DerivedKind: table.DerivedKindRowNumber}},
				Format: "YYYY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: fields: (0: input must be considered valid by: derivedKind: must be blank. or name: must be blank..).",
		},
		{
			name: "format is purely junk letters",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "hello",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			name: "format is only separators",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "---",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			name: "format missing day component",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the day of the month.",
		},
		{
			name: "format missing year component",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			name: "format missing month component",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the month.",
		},
		{
			name: "format has two year tokens",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-YY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			// YYYY-MM-MM is exactly 10 chars, so Length passes and the input drives
			// the month regex straight into its "exactly one" failure.
			name: "format has two month tokens",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-MM",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the month.",
		},
		{
			name: "format has two day tokens",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DD-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the day of the month.",
		},
		{
			name: "format has three Y run (neither YY nor YYYY)",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			name: "format has five Y run",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYYY-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the year.",
		},
		{
			name: "format has three M run",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MMM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the month.",
		},
		{
			name: "format has three D run",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DDD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the day of the month.",
		},
		{
			// 13 chars but the chars whitelist trips before the length rule, so the
			// foreign letter surfaces as the explicit character-class error.
			name: "format contains foreign uppercase letter",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-HH-MM-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format contains invalid characters.",
		},
		{
			// Lowercase m is not in the whitelist, but the month regex (which runs
			// before the char whitelist) fires first because [^M]* matches lowercase
			// runs and never finds an uppercase M.
			name: "format contains lowercase m instead of uppercase M",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-mm-DD",
			},
			wantErr: "failed to validate *table.DateSpec: format: Date format does not include the month.",
		},
		{
			// Five characters: every Match rule passes but Length (6, 10) fails.
			name: "format shorter than minimum length",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYMMD",
			},
			wantErr: "failed to validate *table.DateSpec: format: the length must be between 6 and 10.",
		},
		{
			// Eleven characters: every Match rule passes but Length (6, 10) fails.
			// Also covers the "trailing separator" footgun from earlier probing.
			name: "format longer than maximum length",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DD.",
			},
			wantErr: "failed to validate *table.DateSpec: format: the length must be between 6 and 10.",
		},
		{
			name: "format shortest valid",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YY-M-D",
			},
			wantErr: "",
		},
		{
			// All three separator classes (period, slash, dash) mixed in a single
			// format; accepted because the char whitelist allows any combination of
			// them.
			name: "format with mixed separators",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YY.MM/DD",
			},
			wantErr: "",
		},
		{
			// Space is in the separator whitelist, so space-separated formats
			// validate even though they're unusual.
			name: "format with space separator",
			spec: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YY MM DD",
			},
			wantErr: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate(ctx)
			if tc.wantErr == "" {
				assert.NoError(t, err, "DateSpec must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "DateSpec must be rejected with the expected message")
			}
		})
	}
}
