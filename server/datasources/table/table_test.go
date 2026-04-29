package table_test

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/monetr/monetr/server/datasources/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTable_Read(t *testing.T) {
	// Each case feeds a CSV string and a Mapping into NewTable, drains Read()
	// until an error, and asserts the produced rows + the terminal error. Full
	// table.Row equality is intentional: every field is part of expected so
	// unrelated wiring changes break these tests on purpose. Posted defaults to
	// true when the mapping omits a PostedSpec (every CSV monetr has seen in the
	// wild reports only posted rows), and Balance defaults to 0 when the
	// BalanceSpec is BalanceKindNone or BalanceKindSum. Cases that exercise the
	// wired paths set Posted/Balance explicitly.
	cases := []struct {
		name            string
		csv             string
		mapping         *table.Mapping
		firstRowHeaders bool
		want            []table.Row
		wantErr         string // substring; empty == expect io.EOF
	}{
		{
			name: "native id, sign amount, date, no headers",
			csv:  "1,12.34,2026-05-02,coffee\n2,-5.67,2026-05-02,refund\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Date", "Memo"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
				{
					RowNumber: 1,
					ID:        "2",
					Amount:    -567,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "native id with two fields joins on '::'",
			csv:  "1,checking,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
						{
							Name: "Account",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Account", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1::checking",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Hash precomputed: fnv.New64() over "1" then base32.StdEncoding. The
			// literal value is locked in so a regression that quietly changed the
			// hashing or encoding strategy would surface here.
			name: "hashed id over single field",
			csv:  "1,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindHashed,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "V5R32TEGAG364===",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "id derived from row number",
			csv:  "12.34,coffee,2026-05-02\n-5.67,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							DerivedKind: table.DerivedKindRowNumber,
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "0",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
				{
					RowNumber: 1,
					ID:        "1",
					Amount:    -567,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "sign amount with invert flips sign",
			csv:  "1,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind:   table.AmountKindSign,
					Invert: true,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    -1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "type amount, debit row keeps sign",
			csv:  "1,12.34,DEBIT,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindType,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
						{
							Name: "Type",
						},
					},
					Credit: "CREDIT",
					Debit:  "DEBIT",
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Type", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "type amount, credit row negates",
			csv:  "1,12.34,CREDIT,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindType,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
						{
							Name: "Type",
						},
					},
					Credit: "CREDIT",
					Debit:  "DEBIT",
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Type", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    -1234,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Invert composes after credit handling: CREDIT row gets *= -1, then
			// Invert flips it again. Net effect is the original positive value.
			name: "type amount, credit row with invert double-flips",
			csv:  "1,12.34,CREDIT,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind:   table.AmountKindType,
					Invert: true,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
						{
							Name: "Type",
						},
					},
					Credit: "CREDIT",
					Debit:  "DEBIT",
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Type", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Column kind: first field is the debit column, second is the credit
			// column. With debit populated and credit empty the debit value is
			// parsed as the amount.
			name: "column amount, debit column populated",
			csv:  "1,12.34,,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindColumn,
					Fields: []table.FieldRef{
						{
							Name: "Debit",
						},
						{
							Name: "Credit",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Debit", "Credit", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Mirror of the previous case with the credit cell populated instead.
			name: "column amount, credit column populated",
			csv:  "1,,12.34,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindColumn,
					Fields: []table.FieldRef{
						{
							Name: "Debit",
						},
						{
							Name: "Credit",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Debit", "Credit", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// The headers row consumes index 0 today: t.index increments after the
			// headers check passes, so the first body row is RowNumber 1.
			name: "first row headers consumed and matched",
			csv:  "ID,Amount,Memo,Date\n1,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: true,
			want: []table.Row{
				{
					RowNumber: 1,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "memo from derived row number",
			csv:  "1,12.34,2026-05-02\n2,-5.67,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{DerivedKind: table.DerivedKindRowNumber},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "0",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
				{
					RowNumber: 1,
					ID:        "2",
					Amount:    -567,
					Memo:      "1",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			name: "empty csv terminates with EOF",
			csv:  "",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "",
		},
		{
			name: "first row headers do not match mapping",
			csv:  "X,Y,Z\n1,12.34,coffee\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: true,
			want:            nil,
			wantErr:         "headers in file do not match",
		},
		{
			name: "column amount with both columns empty",
			csv:  "1,,,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindColumn,
					Fields: []table.FieldRef{
						{
							Name: "Debit",
						},
						{
							Name: "Credit",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Debit", "Credit", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "both debit and credit columns are empty",
		},
		{
			name: "column amount with both columns populated",
			csv:  "1,5.00,3.00,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindColumn,
					Fields: []table.FieldRef{
						{
							Name: "Debit",
						},
						{
							Name: "Credit",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Debit", "Credit", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "both debit and credit columns are populated",
		},
		{
			name: "amount parse failure surfaces as wrapped row error",
			csv:  "1,not-a-number,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "failed to derive amount from table for row",
		},
		{
			// Format variation: slashes with year last. [DateSpec.GetTimeFormat]
			// rewrites MM/DD/YYYY into Go's 01/02/2006 reference layout.
			name: "date format MM/DD/YYYY",
			csv:  "1,12.34,coffee,05/02/2026\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "MM/DD/YYYY",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Format variation: dashes with day first.
			name: "date format DD-MM-YYYY",
			csv:  "1,12.34,coffee,02-05-2026\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "DD-MM-YYYY",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Format variation: single-digit month/day plus two-digit year. Maps
			// to Go's 1/2/06 reference layout, which accepts both single- and
			// double-digit components when parsing.
			name: "date format M/D/YY",
			csv:  "1,12.34,coffee,5/2/26\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "M/D/YY",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Format variation: dotted European style.
			name: "date format DD.MM.YYYY",
			csv:  "1,12.34,coffee,02.05.2026\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "DD.MM.YYYY",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Date value does not match the declared format. [Table.getDate]
			// surfaces the time.Parse error wrapped with the row prefix.
			name: "date parse failure surfaces as wrapped row error",
			csv:  "1,12.34,coffee,not-a-date\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "failed to derive date from table for row",
		},
		{
			// [Table.Read] validates the mapping on the first call. An invalid
			// mapping (here, Date is omitted entirely) must surface that error
			// before any rows are produced.
			name: "invalid mapping is rejected on first read",
			csv:  "1,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo:    table.FieldRef{Name: "Memo"},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "failed to validate",
		},
		{
			// [Table.getPosted] returns true when the configured field matches
			// PostedSpec.Posted exactly. Any other status (including the empty
			// string) is treated as pending. Both rows here are posted.
			name: "posted spec, both rows match posted value",
			csv:  "1,12.34,POSTED,coffee,2026-05-02\n2,-5.67,POSTED,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Posted: &table.PostedSpec{
					Fields: []table.FieldRef{
						{
							Name: "Status",
						},
					},
					Posted: "POSTED",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Status", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
				{
					RowNumber: 1,
					ID:        "2",
					Amount:    -567,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// First row is PENDING which doesn't match POSTED, so it surfaces as
			// Posted=false. Second row matches and surfaces as true. The comparison
			// is exact, so case differences would also fall to pending.
			name: "posted spec, mixed posted and pending rows",
			csv:  "1,12.34,PENDING,coffee,2026-05-02\n2,-5.67,POSTED,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Posted: &table.PostedSpec{
					Fields: []table.FieldRef{
						{
							Name: "Status",
						},
					},
					Posted: "POSTED",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Status", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    false,
					Balance:   0,
				},
				{
					RowNumber: 1,
					ID:        "2",
					Amount:    -567,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Comparison is exact, not case-insensitive, so a "posted" row against a
			// "POSTED" needle is treated as pending. Worth pinning since the godoc on
			// [PostedSpec] still describes it as case- insensitive but the
			// implementation is not.
			name: "posted spec, case mismatch falls to pending",
			csv:  "1,12.34,posted,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Posted: &table.PostedSpec{
					Fields: []table.FieldRef{
						{
							Name: "Status",
						},
					},
					Posted: "POSTED",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindNone},
				Headers: []string{"ID", "Amount", "Status", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    false,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// Balance pulled from a column on each row. The first row is 100.00 USD =
			// 10000 minor units, the second is -25.50 = -2550. monetr does not
			// currently invert balances; the value passes through whatever sign the
			// source provided.
			name: "balance spec, field per row",
			csv:  "1,12.34,100.00,coffee,2026-05-02\n2,-5.67,-25.50,refund,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{
					Kind: table.BalanceKindField,
					Fields: []table.FieldRef{
						{
							Name: "Balance",
						},
					},
				},
				Headers: []string{"ID", "Amount", "Balance", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   10000,
				},
				{
					RowNumber: 1,
					ID:        "2",
					Amount:    -567,
					Memo:      "refund",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   -2550,
				},
			},
			wantErr: "",
		},
		{
			// BalanceKindSum is currently a documented no-op pending the
			// post-processing pass that would actually walk the rows. Pinning it as 0
			// here so the day that gets implemented this case fails loudly and
			// reminds us to update the test.
			name: "balance spec, sum kind is a no-op",
			csv:  "1,12.34,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{Kind: table.BalanceKindSum},
				Headers: []string{"ID", "Amount", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want: []table.Row{
				{
					RowNumber: 0,
					ID:        "1",
					Amount:    1234,
					Memo:      "coffee",
					Merchant:  nil,
					Date:      time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
					Posted:    true,
					Balance:   0,
				},
			},
			wantErr: "",
		},
		{
			// A non-numeric balance cell surfaces from [Table.getBalance] and is
			// wrapped with the row prefix the same way the amount path does.
			name: "balance parse failure surfaces as wrapped row error",
			csv:  "1,12.34,not-a-number,coffee,2026-05-02\n",
			mapping: &table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "ID",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{Name: "Memo"},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{
					Kind: table.BalanceKindField,
					Fields: []table.FieldRef{
						{
							Name: "Balance",
						},
					},
				},
				Headers: []string{"ID", "Amount", "Balance", "Memo", "Date"},
			},
			firstRowHeaders: false,
			want:            nil,
			wantErr:         "failed to derive balance from table for row",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			reader := csv.NewReader(strings.NewReader(tc.csv))
			tbl := table.NewTable(reader, tc.mapping, tc.firstRowHeaders)

			var got []table.Row
			var lastErr error
			for {
				row, err := tbl.Read()
				if err != nil {
					lastErr = err
					break
				}
				got = append(got, *row)
			}

			require.Len(t, got, len(tc.want), "row count must match expected")
			for i := range tc.want {
				assert.Equal(t, tc.want[i], got[i], "row %d", i)
			}
			if tc.wantErr == "" {
				assert.ErrorIs(t, lastErr, io.EOF, "last error should be io.EOF")
			} else {
				assert.ErrorContains(t, lastErr, tc.wantErr, "terminal error must match")
			}
		})
	}
}
