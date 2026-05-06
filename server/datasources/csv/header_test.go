package csv_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/datasources/csv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeekHeader(t *testing.T) {
	t.Run("comma-delimited bank export", func(t *testing.T) {
		input := "Date,Description,Amount,Balance\n" +
			"2026-01-15,COFFEE SHOP,-4.50,1234.56\n" +
			"2026-01-16,GAS STATION,-45.00,1189.56\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		require.NoError(t, err, "must parse a comma-delimited CSV")
		assert.Equal(t, ',', delimeter)
		assert.Equal(t, []string{"Date", "Description", "Amount", "Balance"}, headers, "must extract all four headers")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read the returned reader to completion")
		assert.Equal(t, input, string(contents), "returned reader must preserve original content")
	})

	t.Run("pipe-delimited export", func(t *testing.T) {
		input := "Date|Description|Amount\n" +
			"2026-01-15|GROCERY|-42.10\n" +
			"2026-01-16|RESTAURANT|-23.45\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		require.NoError(t, err, "must parse a pipe-delimited CSV")
		assert.Equal(t, '|', delimeter)
		assert.Equal(t, []string{"Date", "Description", "Amount"}, headers, "must extract pipe-delimited headers")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader")
		assert.Equal(t, input, string(contents), "returned reader must preserve pipe content")
	})

	t.Run("tab-delimited export", func(t *testing.T) {
		input := "Date\tDescription\tAmount\tCategory\n" +
			"2026-01-15\tUTIL\t-99.00\tBills\n" +
			"2026-01-16\tPAYCHECK\t2000.00\tIncome\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		assert.Equal(t, '\t', delimeter)
		require.NoError(t, err, "must parse a tab-delimited CSV")
		assert.Equal(t, []string{"Date", "Description", "Amount", "Category"}, headers, "must extract tab-delimited headers")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader")
		assert.Equal(t, input, string(contents), "returned reader must preserve tab content")
	})

	t.Run("headers with leading whitespace are trimmed", func(t *testing.T) {
		input := "Date, Description, Amount\n" +
			"2026-01-15, COFFEE, -4.50\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		assert.Equal(t, ',', delimeter)
		require.NoError(t, err, "must parse CSV with leading spaces")
		assert.Equal(t, []string{"Date", "Description", "Amount"}, headers, "leading spaces must be trimmed from headers")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader")
		assert.Equal(t, input, string(contents), "returned reader must preserve original bytes including interior spaces")
	})

	t.Run("quoted fields containing commas", func(t *testing.T) {
		input := "Date,Description,Amount,Balance\n" +
			"2026-01-15,\"PAYROLL, INC\",\"1,234.56\",\"10,000.00\"\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		assert.Equal(t, ',', delimeter)
		require.NoError(t, err, "must parse CSV with quoted comma fields")
		assert.Equal(t, []string{"Date", "Description", "Amount", "Balance"}, headers, "headers must be extracted despite quoted commas in body rows")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader")
		assert.Equal(t, input, string(contents), "returned reader must preserve quoted content verbatim")
	})

	t.Run("input smaller than peek size", func(t *testing.T) {
		// bufio.Reader.Peek returns io.EOF for inputs shorter than the 1000-byte
		// peek window. The function tolerates EOF as long as the available preview
		// parses into a valid header row.
		input := "Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n"

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		assert.Equal(t, ',', delimeter)
		require.NoError(t, err, "short CSV must still parse")
		assert.Equal(t, []string{"Date", "Description", "Amount"}, headers, "short-input headers must be extracted")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader")
		assert.Equal(t, input, string(contents), "short-input reader must preserve content")
	})

	t.Run("input larger than peek size", func(t *testing.T) {
		// Build a CSV whose body exceeds the 1000-byte peek window so we exercise
		// the non-EOF peek path and prove the returned reader yields the full
		// original payload, not just the peeked prefix. This is the load-bearing
		// contract for FileStorage.Store.
		var b strings.Builder
		b.WriteString("Date,Description,Amount\n")
		for i := 0; i < 100; i++ {
			b.WriteString("2026-01-15,SOME TRANSACTION DESCRIPTION STRING,-12.34\n")
		}
		input := b.String()
		require.Greater(t, len(input), 1000, "test setup: input must exceed the 1000-byte peek window")

		delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(input)))
		assert.Equal(t, ',', delimeter)
		require.NoError(t, err, "large CSV must parse")
		assert.Equal(t, []string{"Date", "Description", "Amount"}, headers, "headers must be read from the first peek window")

		contents, err := io.ReadAll(reader)
		require.NoError(t, err, "must read returned reader to completion")
		assert.Equal(t, input, string(contents), "returned reader must yield the entire original payload, not just the peeked window")
	})

	t.Run("invalid cases", func(t *testing.T) {
		// Invalid inputs: fewer than 3 columns under every candidate delimiter, or
		// data that does not parse as CSV at all. Each case must produce an error and
		// nil outputs.
		invalidCases := []struct {
			name  string
			input string
		}{
			{"single column", "Date\n2026-01-15\n2026-01-16\n"},
			{"two columns", "Date,Amount\n2026-01-15,-4.50\n"},
			{"empty input", ""},
			{"binary noise", "\x00\x01\x02\xff\xfe garbage \x00"},
			{"prose without delimiters", "this is just a sentence with no structure whatsoever"},
			{"html document", "<html><body>Not a CSV</body></html>"},
		}
		for _, tc := range invalidCases {
			tc := tc
			t.Run("invalid/"+tc.name, func(t *testing.T) {
				delimeter, headers, reader, err := csv.PeekHeader(bytes.NewReader([]byte(tc.input)))
				assert.Equal(t, ',', delimeter)
				assert.Error(t, err, "must reject %q as non-CSV", tc.name)
				assert.Nil(t, headers, "headers must be nil on failure")
				assert.Nil(t, reader, "reader must be nil on failure")
			})
		}
	})
}
