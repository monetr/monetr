package table_test

import (
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

func TestBalanceSpec_Validate(t *testing.T) {
	// BalanceSpec uses a two-branch JoinErrorMaybe structure:
	//   Branch 1 (Kind is none or sum): Fields must be empty.
	//   Branch 2 (Kind is field):       Fields must contain exactly one valid
	//                                   FieldRef.
	// At least one branch must succeed for the value to validate. When both
	// branches fail, their errors are joined with "\n" and the whole thing is
	// wrapped with "input must be considered valid by: ". Sub-errors within
	// each branch are sorted alphabetically by JSON tag and separated by "; ".
	ctx := table.WithColumns(
		t.Context(),
		[]string{"Date", "Description", "Amount", "RunningBalance"},
	)
	cases := []struct {
		name    string
		spec    table.BalanceSpec
		wantErr string
	}{
		// --- Happy paths (at least one branch passes) ---
		{
			name: "none with nil fields",
			spec: table.BalanceSpec{
				Kind: table.BalanceKindNone,
			},
			wantErr: "",
		},
		{
			name: "none with empty fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindNone,
				Fields: []table.FieldRef{},
			},
			wantErr: "",
		},
		{
			name: "sum with nil fields",
			spec: table.BalanceSpec{
				Kind: table.BalanceKindSum,
			},
			wantErr: "",
		},
		{
			name: "sum with empty fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindSum,
				Fields: []table.FieldRef{},
			},
			wantErr: "",
		},
		{
			name: "field with named field",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "RunningBalance"}},
			},
			wantErr: "",
		},
		{
			name: "field with row number",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "",
		},
		{
			name: "field with row per day",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumberPerDay}},
			},
			wantErr: "",
		},
		{
			name: "field with row per day per amount",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKindRowNumberPerDayPerAmount}},
			},
			wantErr: "",
		},

		// --- Cross-validation: Kind and Fields must agree ---
		{
			// Rejected because none forbids fields (branch 1) and the kind isn't
			// "field" (branch 2). Previously accepted under the single-branch schema;
			// its rejection is the primary reason for this rewrite.
			name: "none with a field",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindNone,
				Fields: []table.FieldRef{{Name: "RunningBalance"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank. or kind: must equal \"field\".",
		},
		{
			name: "sum with a field",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindSum,
				Fields: []table.FieldRef{{Name: "RunningBalance"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank. or kind: must equal \"field\".",
		},
		{
			// Two-fields cases expose a different branch-2 error (length) than the
			// single-field cross-validation above, so both are worth distinguishing.
			name: "none with multiple fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindNone,
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "Amount"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank. or fields: the length must be exactly 1; kind: must equal \"field\".",
		},
		{
			name: "sum with multiple fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindSum,
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "Amount"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank. or fields: the length must be exactly 1; kind: must equal \"field\".",
		},
		{
			name: "field with no fields",
			spec: table.BalanceSpec{
				Kind: table.BalanceKindField,
			},
			wantErr: "input must be considered valid by: kind: must be one of: [\"none\", \"sum\"]. or fields: cannot be blank.",
		},
		{
			// An explicitly empty slice is treated the same as nil for Required
			// because IsEmpty([]) is true.
			name: "field with empty fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{},
			},
			wantErr: "input must be considered valid by: kind: must be one of: [\"none\", \"sum\"]. or fields: cannot be blank.",
		},

		// --- Kind-layer invalid ---
		{
			name:    "empty",
			spec:    table.BalanceSpec{},
			wantErr: "input must be considered valid by: kind: cannot be blank. or fields: cannot be blank; kind: must equal \"field\".",
		},
		{
			name: "unknown kind, no fields",
			spec: table.BalanceSpec{
				Kind: table.BalanceKind("bogus"),
			},
			wantErr: "input must be considered valid by: kind: must be one of: [\"none\", \"sum\"]. or fields: cannot be blank; kind: must equal \"field\".",
		},
		{
			// Branch 2 accepts the field but rejects the kind, so only the kind error
			// appears in branch 2. Branch 1 rejects both.
			name: "unknown kind, valid field",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKind("bogus"),
				Fields: []table.FieldRef{{Name: "RunningBalance"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or kind: must equal \"field\".",
		},

		// --- Length violation in branch 2 (kind=field with multiple fields) ---
		{
			name: "field with too many fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "Amount"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: the length must be exactly 1.",
		},
		{
			name: "field with three fields",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "Amount"}, {Name: "Date"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: the length must be exactly 1.",
		},
		{
			// Length(1,1) fails before Unique runs, so the duplicate case reports the
			// length error rather than the uniqueness error.
			name: "duplicates caught by length check",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "RunningBalance"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: the length must be exactly 1.",
		},

		// --- Child FieldRef invalid (exercises Each inside branch 2) ---
		{
			name: "field with blank child",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..).",
		},
		{
			name: "field child not in headers",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "NotPresent"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"RunningBalance\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
		{
			name: "field child with name and derived",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{Name: "RunningBalance", DerivedKind: table.DerivedKindRowNumber}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: (0: input must be considered valid by: derivedKind: must be blank. or name: must be blank..).",
		},
		{
			name: "field child with unknown derived",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKindField,
				Fields: []table.FieldRef{{DerivedKind: table.DerivedKind("bogus")}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: (0: input must be considered valid by: derivedKind: must be blank; name: cannot be blank. or derivedKind: must be one of: [\"rowNumber\", \"rowNumberPerDay\", \"rowNumberPerDayPerAmount\"]..).",
		},

		// --- Combined (both kind and fields misbehaving) ---
		{
			// Empty kind fails Required in both branches; adding two fields surfaces
			// the Empty rule in branch 1 and Length in branch 2, so both branches
			// produce two keys sorted alphabetically.
			name: "empty kind with two fields",
			spec: table.BalanceSpec{
				Fields: []table.FieldRef{{Name: "RunningBalance"}, {Name: "Amount"}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: cannot be blank. or fields: the length must be exactly 1; kind: must equal \"field\".",
		},
		{
			name: "unknown kind, invalid child",
			spec: table.BalanceSpec{
				Kind:   table.BalanceKind("bogus"),
				Fields: []table.FieldRef{{}},
			},
			wantErr: "input must be considered valid by: fields: must be blank; kind: must be one of: [\"none\", \"sum\"]. or fields: (0: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..); kind: must equal \"field\".",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate(ctx)
			if tc.wantErr == "" {
				assert.NoError(t, err, "BalanceSpec must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "BalanceSpec must be rejected with the expected message")
			}
		})
	}
}
