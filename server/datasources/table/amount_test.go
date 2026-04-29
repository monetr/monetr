package table_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
)

func TestAmountSpec_Validate(t *testing.T) {
	// AmountSpec has three valid shapes, selected by Kind:
	//   - Sign:   exactly one Fields entry, no Credit/Debit strings
	//   - Type:   exactly two Fields entries, both Credit and Debit strings set
	//   - Column: exactly two Fields entries, no Credit/Debit strings
	// Invert may be either true or false regardless of Kind.
	ctx := table.WithColumns(
		t.Context(),
		[]string{"Date", "Amount", "DebitAmt", "CreditAmt", "TransType"},
	)
	cases := []struct {
		name    string
		spec    table.AmountSpec
		wantErr string
	}{
		{
			name: "Sign with one field, Invert false",
			spec: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Invert: false,
				Fields: []table.FieldRef{{Name: "Amount"}},
			},
			wantErr: "",
		},
		{
			name: "Sign with one field, Invert true",
			spec: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Invert: true,
				Fields: []table.FieldRef{{Name: "Amount"}},
			},
			wantErr: "",
		},
		{
			name: "Type with two fields and credit/debit, Invert false",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Invert: false,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CREDIT",
				Debit:  "DEBIT",
			},
			wantErr: "",
		},
		{
			name: "Type with two fields and credit/debit, Invert true",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Invert: true,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CREDIT",
				Debit:  "DEBIT",
			},
			wantErr: "",
		},
		{
			name: "Column with two fields, Invert false",
			spec: table.AmountSpec{
				Kind:   table.AmountKindColumn,
				Invert: false,
				Fields: []table.FieldRef{{Name: "DebitAmt"}, {Name: "CreditAmt"}},
			},
			wantErr: "",
		},
		{
			name: "Column with two fields, Invert true",
			spec: table.AmountSpec{
				Kind:   table.AmountKindColumn,
				Invert: true,
				Fields: []table.FieldRef{{Name: "DebitAmt"}, {Name: "CreditAmt"}},
			},
			wantErr: "",
		},
		{
			// Branch 3 (Column) now also reports "kind: cannot be blank" because
			// Required was added to its Kind rule. Prior to that fix, empty Kind
			// slipped through the Column branch.
			name:    "empty",
			spec:    table.AmountSpec{},
			wantErr: "input must be considered valid by: fields: cannot be blank; kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: cannot be blank; kind: must equal \"type\". or fields: cannot be blank; kind: must equal \"column\".",
		},
		{
			name:    "unknown kind with no fields",
			spec:    table.AmountSpec{Kind: table.AmountKind("bogus")},
			wantErr: "input must be considered valid by: fields: cannot be blank; kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: cannot be blank; kind: must equal \"type\". or fields: cannot be blank; kind: must equal \"column\".",
		},
		{
			name: "Sign with two fields violates length",
			spec: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
			},
			wantErr: "input must be considered valid by: fields: the length must be exactly 1. or credit: cannot be blank; debit: cannot be blank; kind: must equal \"type\". or kind: must equal \"column\".",
		},
		{
			name: "Sign with Credit set",
			spec: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Fields: []table.FieldRef{{Name: "Amount"}},
				Credit: "CREDIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified. or debit: cannot be blank; fields: the length must be exactly 2; kind: must equal \"type\". or credit: when kind is \"column\" credit cannot be specified; fields: the length must be exactly 2; kind: must equal \"column\".",
		},
		{
			name: "Sign with field name not in headers",
			spec: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Fields: []table.FieldRef{{Name: "NotPresent"}},
			},
			wantErr: "input must be considered valid by: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..). or credit: cannot be blank; debit: cannot be blank; fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..); kind: must equal \"type\". or fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..); kind: must equal \"column\".",
		},
		{
			name: "Type missing Credit",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or credit: cannot be blank. or debit: when kind is \"column\" debit cannot be specified; kind: must equal \"column\".",
		},
		{
			name: "Type with only one field violates length",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}},
				Credit: "CREDIT",
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; kind: must equal \"sign\". or fields: the length must be exactly 2. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; fields: the length must be exactly 2; kind: must equal \"column\".",
		},
		{
			name: "Column with Credit set",
			spec: table.AmountSpec{
				Kind:   table.AmountKindColumn,
				Fields: []table.FieldRef{{Name: "DebitAmt"}, {Name: "CreditAmt"}},
				Credit: "X",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or debit: cannot be blank; kind: must equal \"type\". or credit: when kind is \"column\" credit cannot be specified.",
		},
		{
			name: "Type with duplicate fields",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "Amount"}},
				Credit: "CREDIT",
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or fields: fields[1] is a duplicate of an earlier entry. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; fields: fields[1] is a duplicate of an earlier entry; kind: must equal \"column\".",
		},
		{
			name: "Column with duplicate fields",
			spec: table.AmountSpec{
				Kind:   table.AmountKindColumn,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "Amount"}},
			},
			wantErr: "input must be considered valid by: fields: the length must be exactly 1; kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: fields[1] is a duplicate of an earlier entry; kind: must equal \"type\". or fields: fields[1] is a duplicate of an earlier entry.",
		},
		{
			// Regression guard: an empty Kind used to pass the Column branch because
			// that branch only had validation.In(AmountKindColumn) with no Required;
			// In allows empty values. After Required was added to the Column branch's
			// Kind rule, all three branches now reject an empty Kind even when the
			// rest of the shape happens to look like a valid Column spec.
			name: "empty Kind with valid Column-shaped Fields",
			spec: table.AmountSpec{
				Fields: []table.FieldRef{{Name: "DebitAmt"}, {Name: "CreditAmt"}},
			},
			wantErr: "input must be considered valid by: fields: the length must be exactly 1; kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; kind: must equal \"type\". or kind: must equal \"column\".",
		},
		{
			// Credit at exactly 100 chars; upper boundary of Length(1, 100).
			name: "Type with Credit at max length boundary",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: strings.Repeat("A", 100),
				Debit:  "DEBIT",
			},
			wantErr: "",
		},
		{
			// Debit at exactly 100 chars; upper boundary of Length(1, 100).
			name: "Type with Debit at max length boundary",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CREDIT",
				Debit:  strings.Repeat("A", 100),
			},
			wantErr: "",
		},
		{
			name: "Type with Credit exceeds max length",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: strings.Repeat("A", 101),
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or credit: the length must be between 1 and 100. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; kind: must equal \"column\".",
		},
		{
			name: "Type with Debit exceeds max length",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CREDIT",
				Debit:  strings.Repeat("A", 101),
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or debit: the length must be between 1 and 100. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; kind: must equal \"column\".",
		},
		{
			// Tab is ASCII but not printable ASCII; the recent switch from is.ASCII
			// to is.PrintableASCII on Credit is what catches this.
			name: "Type with Credit containing tab",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CR\tEDIT",
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or credit: must contain printable ASCII characters only. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; kind: must equal \"column\".",
		},
		{
			name: "Type with Debit containing non-ASCII",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
				Credit: "CREDIT",
				Debit:  "Débit",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: the length must be exactly 1; kind: must equal \"sign\". or debit: must contain printable ASCII characters only. or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; kind: must equal \"column\".",
		},
		{
			// Type-branch Each coverage: second field is a blank FieldRef (first is
			// valid). Every other rule in the Type branch passes; Each surfaces the
			// invalid child via index 1.
			name: "Type with second child blank FieldRef",
			spec: table.AmountSpec{
				Kind:   table.AmountKindType,
				Fields: []table.FieldRef{{Name: "Amount"}, {}},
				Credit: "CREDIT",
				Debit:  "DEBIT",
			},
			wantErr: "input must be considered valid by: credit: when kind is \"sign\" credit cannot be specified; debit: when kind is \"sign\" debit cannot be specified; fields: (1: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..); kind: must equal \"sign\". or fields: (1: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..). or credit: when kind is \"column\" credit cannot be specified; debit: when kind is \"column\" debit cannot be specified; fields: (1: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..); kind: must equal \"column\".",
		},
		{
			// Column-branch Each coverage: first field references a column
			// that isn't in the context.
			name: "Column with first child name not in headers",
			spec: table.AmountSpec{
				Kind:   table.AmountKindColumn,
				Fields: []table.FieldRef{{Name: "NotPresent"}, {Name: "CreditAmt"}},
			},
			wantErr: "input must be considered valid by: fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..); kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..); kind: must equal \"type\". or fields: (0: input must be considered valid by: name: must be one of: [\"Date\", \"Amount\", \"DebitAmt\", \"CreditAmt\", \"TransType\"]. or derivedKind: cannot be blank; name: must be blank..).",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate(ctx)
			if tc.wantErr == "" {
				assert.NoError(t, err, "AmountSpec must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "AmountSpec must be rejected with the expected message")
			}
		})
	}
}
