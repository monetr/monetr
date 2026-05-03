package table_test

import (
	"encoding/json"
	"testing"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/monetr/monetr/server/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapping_Validate(t *testing.T) {
	// Mapping.Validate composes the per-sub-spec validators and injects the
	// mapping's own Headers into the context. Each sub-spec's error is nested
	// under its JSON tag (id, amount, memo, merchant, date, posted, balance)
	// and keys are emitted alphabetically. Merchant and Posted short-circuit
	// to no error when nil; non-nil values must satisfy their respective
	// validators.

	// Build a fresh fully-valid Mapping per case so a single test's mutation
	// cannot leak into other cases through shared slices or pointers.
	validMapping := func() table.Mapping {
		return table.Mapping{
			ID: table.IDSpec{
				Kind:   table.IDSpecKindNative,
				Fields: []table.FieldRef{{Name: "Id"}},
			},
			Amount: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Fields: []table.FieldRef{{Name: "Amount"}},
			},
			Memo: table.FieldRef{Name: "Description"},
			Date: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DD",
			},
			Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
			Headers: []string{
				"Date",
				"Description",
				"Amount",
				"Id",
				"TransType",
				"DebitAmt",
				"CreditAmt",
				"Merchant",
				"Status",
				"RunningBalance",
			},
		}
	}

	cases := []struct {
		name    string
		mutate  func(m *table.Mapping)
		wantErr string
	}{
		{
			name:   "minimal valid, no optional fields",
			mutate: func(*table.Mapping) {},
		},
		{
			name: "valid with merchant",
			mutate: func(m *table.Mapping) {
				m.Merchant = &table.FieldRef{Name: "Merchant"}
			},
		},
		{
			name: "valid with posted",
			mutate: func(m *table.Mapping) {
				m.Posted = &table.PostedSpec{
					Fields: []table.FieldRef{{Name: "Status"}},
					Posted: "POSTED",
				}
			},
		},
		{
			name: "valid with merchant and posted",
			mutate: func(m *table.Mapping) {
				m.Merchant = &table.FieldRef{Name: "Merchant"}
				m.Posted = &table.PostedSpec{
					Fields: []table.FieldRef{{Name: "Status"}},
					Posted: "POSTED",
				}
			},
		},
		{
			name: "valid with derived id",
			mutate: func(m *table.Mapping) {
				m.ID = table.IDSpec{
					Kind:   table.IDSpecKindHashed,
					Fields: []table.FieldRef{{Name: "Date"}, {DerivedKind: table.DerivedKindRowNumberPerDay}},
				}
			},
		},
		{
			name: "valid with amount type",
			mutate: func(m *table.Mapping) {
				m.Amount = table.AmountSpec{
					Kind:   table.AmountKindType,
					Fields: []table.FieldRef{{Name: "Amount"}, {Name: "TransType"}},
					Credit: "CREDIT",
					Debit:  "DEBIT",
				}
			},
		},
		{
			name: "valid with amount column",
			mutate: func(m *table.Mapping) {
				m.Amount = table.AmountSpec{
					Kind:   table.AmountKindColumn,
					Fields: []table.FieldRef{{Name: "DebitAmt"}, {Name: "CreditAmt"}},
				}
			},
		},
		{
			name: "valid with balance field",
			mutate: func(m *table.Mapping) {
				m.Balance = table.BalanceSpec{
					Kind:   table.BalanceKindField,
					Fields: []table.FieldRef{{Name: "RunningBalance"}},
				}
			},
		},
		{
			name:    "empty mapping, all sub-spec errors",
			mutate:  func(m *table.Mapping) { *m = table.Mapping{} },
			wantErr: "failed to validate *table.Mapping: amount: input must be considered valid by: fields: cannot be blank; kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: cannot be blank; kind: must equal \"type\". or fields: cannot be blank; kind: must equal \"column\".; balance: input must be considered valid by: kind: cannot be blank. or fields: cannot be blank; kind: must equal \"field\".; date: failed to validate *table.DateSpec: fields: cannot be blank; format: cannot be blank.; headers: cannot be blank; id: failed to validate *table.IDSpec: fields: cannot be blank; kind: cannot be blank.; memo: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..",
		},
		{
			name: "invalid id kind",
			mutate: func(m *table.Mapping) {
				m.ID = table.IDSpec{
					Kind:   table.IDSpecKind("bogus"),
					Fields: []table.FieldRef{{Name: "Id"}},
				}
			},
			wantErr: "failed to validate *table.Mapping: id: failed to validate *table.IDSpec: kind: must be one of: [\"native\", \"hashed\"]..",
		},
		{
			name: "invalid amount kind",
			mutate: func(m *table.Mapping) {
				m.Amount = table.AmountSpec{
					Kind:   table.AmountKind("bogus"),
					Fields: []table.FieldRef{{Name: "Amount"}},
				}
			},
			wantErr: "failed to validate *table.Mapping: amount: input must be considered valid by: kind: must equal \"sign\". or credit: cannot be blank; debit: cannot be blank; fields: the length must be exactly 2; kind: must equal \"type\". or fields: the length must be exactly 2; kind: must equal \"column\"..",
		},
		{
			name: "memo not in columns",
			mutate: func(m *table.Mapping) {
				m.Memo = table.FieldRef{Name: "NotPresent"}
			},
			wantErr: "failed to validate *table.Mapping: memo: input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\", \"TransType\", \"DebitAmt\", \"CreditAmt\", \"Merchant\", \"Status\", \"RunningBalance\"]. or derivedKind: cannot be blank; name: must be blank..",
		},
		{
			name: "merchant not in columns",
			mutate: func(m *table.Mapping) {
				m.Merchant = &table.FieldRef{Name: "NotPresent"}
			},
			wantErr: "failed to validate *table.Mapping: merchant: input must be considered valid by: name: must be one of: [\"Date\", \"Description\", \"Amount\", \"Id\", \"TransType\", \"DebitAmt\", \"CreditAmt\", \"Merchant\", \"Status\", \"RunningBalance\"]. or derivedKind: cannot be blank; name: must be blank..",
		},
		{
			name: "invalid date format",
			mutate: func(m *table.Mapping) {
				m.Date = table.DateSpec{
					Fields: []table.FieldRef{{Name: "Date"}},
					Format: "nope",
				}
			},
			wantErr: "failed to validate *table.Mapping: date: failed to validate *table.DateSpec: format: Date format does not include the year..",
		},
		{
			name: "invalid posted, empty spec",
			mutate: func(m *table.Mapping) {
				m.Posted = &table.PostedSpec{}
			},
			wantErr: "failed to validate *table.Mapping: posted: failed to validate *table.PostedSpec: fields: cannot be blank..",
		},
		{
			name: "invalid balance kind",
			mutate: func(m *table.Mapping) {
				m.Balance = table.BalanceSpec{Kind: table.BalanceKind("bogus")}
			},
			wantErr: "failed to validate *table.Mapping: balance: input must be considered valid by: kind: must be one of: [\"none\", \"sum\"]. or fields: cannot be blank; kind: must equal \"field\"..",
		},
		{
			name: "invalid id and balance",
			mutate: func(m *table.Mapping) {
				m.ID = table.IDSpec{
					Kind:   table.IDSpecKind("bogus"),
					Fields: []table.FieldRef{{Name: "Id"}},
				}
				m.Balance = table.BalanceSpec{Kind: table.BalanceKind("bogus")}
			},
			wantErr: "failed to validate *table.Mapping: balance: input must be considered valid by: kind: must be one of: [\"none\", \"sum\"]. or fields: cannot be blank; kind: must equal \"field\".; id: failed to validate *table.IDSpec: kind: must be one of: [\"native\", \"hashed\"]..",
		},
		{
			name: "invalid merchant, posted nil ok",
			mutate: func(m *table.Mapping) {
				m.Merchant = &table.FieldRef{}
			},
			wantErr: "failed to validate *table.Mapping: merchant: input must be considered valid by: name: cannot be blank. or derivedKind: cannot be blank..",
		},
		{
			name: "invalid posted, merchant nil ok",
			mutate: func(m *table.Mapping) {
				m.Posted = &table.PostedSpec{
					Fields: []table.FieldRef{{Name: "Status"}},
					Posted: "Pösted",
				}
			},
			wantErr: "failed to validate *table.Mapping: posted: failed to validate *table.PostedSpec: posted: must contain printable ASCII characters only..",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := validMapping()
			tc.mutate(&m)
			err := m.Validate(t.Context())
			if tc.wantErr == "" {
				assert.NoError(t, err, "Mapping must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "Mapping must be rejected with the expected message")
			}
		})
	}
}

func TestMapping_Validate_JSONShape(t *testing.T) {
	// MarshalErrorTree plus the OneOfError / validation.Errors MarshalJSON
	// methods should produce a nested JSON tree that mirrors the structure of
	// the failures: per-field maps for non-union sub-specs, a {"oneOf": [...]}
	// envelope for union sub-specs, and recursion through nested unions.
	validMapping := func() table.Mapping {
		return table.Mapping{
			ID: table.IDSpec{
				Kind:   table.IDSpecKindNative,
				Fields: []table.FieldRef{{Name: "Id"}},
			},
			Amount: table.AmountSpec{
				Kind:   table.AmountKindSign,
				Fields: []table.FieldRef{{Name: "Amount"}},
			},
			Memo: table.FieldRef{Name: "Description"},
			Date: table.DateSpec{
				Fields: []table.FieldRef{{Name: "Date"}},
				Format: "YYYY-MM-DD",
			},
			Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
			Headers: []string{"Date", "Description", "Amount", "Id"},
		}
	}

	t.Run("empty mapping json shape", func(t *testing.T) {
		m := table.Mapping{}
		err := m.Validate(t.Context())
		require.Error(t, err)

		raw, jerr := json.Marshal(validators.MarshalErrorTree(err))
		require.NoError(t, jerr)

		var decoded struct {
			Amount struct {
				OneOf []map[string]string `json:"oneOf"`
			} `json:"amount"`
			Balance struct {
				OneOf []map[string]string `json:"oneOf"`
			} `json:"balance"`
			Date map[string]string `json:"date"`
			ID   map[string]string `json:"id"`
			Memo struct {
				OneOf []map[string]string `json:"oneOf"`
			} `json:"memo"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))

		assert.Len(t, decoded.Amount.OneOf, 3, "amount has three union variants")
		assert.Len(t, decoded.Balance.OneOf, 2, "balance has two union variants")
		assert.Len(t, decoded.Memo.OneOf, 2, "memo (FieldRef) has two union variants")
		assert.Equal(t, map[string]string{
			"fields": "cannot be blank",
			"format": "cannot be blank",
		}, decoded.Date, "date is not a union, serializes as a flat object")
		assert.Equal(t, map[string]string{
			"fields": "cannot be blank",
			"kind":   "cannot be blank",
		}, decoded.ID, "id is not a union, serializes as a flat object")
	})

	t.Run("partial amount union shape", func(t *testing.T) {
		// User submits kind=sign with two fields. Variant 1 (sign) only fails
		// the length-1 check; variants 2 and 3 fail their kind constraint. The
		// per-variant errors implicitly identify which one was being attempted.
		m := validMapping()
		m.Amount = table.AmountSpec{
			Kind:   table.AmountKindSign,
			Fields: []table.FieldRef{{Name: "Date"}, {Name: "Amount"}},
		}
		err := m.Validate(t.Context())
		require.Error(t, err)

		raw, jerr := json.Marshal(validators.MarshalErrorTree(err))
		require.NoError(t, jerr)

		var decoded struct {
			Amount struct {
				OneOf []map[string]string `json:"oneOf"`
			} `json:"amount"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		require.Len(t, decoded.Amount.OneOf, 3)

		assert.NotContains(t, decoded.Amount.OneOf[0], "kind", "sign variant: kind passes")
		assert.Equal(t, "the length must be exactly 1", decoded.Amount.OneOf[0]["fields"])
		assert.Equal(t, `must equal "type"`, decoded.Amount.OneOf[1]["kind"], "type variant: kind fails")
		assert.Equal(t, `must equal "column"`, decoded.Amount.OneOf[2]["kind"], "column variant: kind fails")
	})

	t.Run("nested union recurses", func(t *testing.T) {
		// A FieldRef with neither name nor derivedKind sits inside the sign
		// variant's fields[0]. The inner FieldRef.Validate returns its own
		// OneOfError, which must serialize as a nested {"oneOf": [...]} under
		// the outer amount.oneOf[0].fields path.
		m := validMapping()
		m.Amount = table.AmountSpec{
			Kind:   table.AmountKindSign,
			Fields: []table.FieldRef{{}},
		}
		err := m.Validate(t.Context())
		require.Error(t, err)

		raw, jerr := json.Marshal(validators.MarshalErrorTree(err))
		require.NoError(t, jerr)

		// fields key here can be either a string (top-level error) or a nested
		// object describing the per-element FieldRef failure. Decode loosely.
		var decoded struct {
			Amount struct {
				OneOf []map[string]json.RawMessage `json:"oneOf"`
			} `json:"amount"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		require.Len(t, decoded.Amount.OneOf, 3)

		signFields, ok := decoded.Amount.OneOf[0]["fields"]
		require.True(t, ok, "sign variant should report a fields error")

		var indexed map[string]struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(signFields, &indexed))
		require.Contains(t, indexed, "0", "fields[0] must surface as the '0' key")
		assert.Len(t, indexed["0"].OneOf, 2, "inner FieldRef has two union variants")
	})
}
