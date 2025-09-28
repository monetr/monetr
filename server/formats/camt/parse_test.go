package camt_test

import (
	"bytes"
	"testing"

	"github.com/monetr/monetr/server/formats/camt"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("goldman sachs US v2 sample file", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "goldman-us-camt053-v2.xml"))
		result, err := camt.Parse(reader)
		assert.NoError(t, err, "must succeed in parsing valid sample file")
		assert.NotEmpty(t, result)
	})

	t.Run("goldman sachs US v2 with wires sample file", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "goldman-us-camt053-wire-v2.xml"))
		result, err := camt.Parse(reader)
		assert.NoError(t, err, "must succeed in parsing valid sample file")
		assert.NotEmpty(t, result)
	})

	t.Run("goldman sachs UK v2 sample file", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "goldman-uk-camt053-v2.xml"))
		result, err := camt.Parse(reader)
		assert.NoError(t, err, "must succeed in parsing valid sample file")
		assert.NotEmpty(t, result)
	})

	t.Run("goldman sachs EU v8 sample file", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "goldman-eu-camt053-v8.xml"))
		result, err := camt.Parse(reader)
		assert.NoError(t, err, "must succeed in parsing valid sample file")
		assert.NotEmpty(t, result)
	})
}

func TestParseTransactionAmount(t *testing.T) {
	t.Run("us example debit", func(t *testing.T) {
		input := camt.ReportEntry15{
			Amt: &camt.ActiveOrHistoricCurrencyAndAmount{
				CcyAttr: "USD",
				Value:   100.00,
			},
			CdtDbtInd: "DBIT",
		}
		result, err := camt.ParseTransactionAmount(input)
		assert.NoError(t, err, "should not return an error parsing known good values")
		assert.EqualValues(t, 10000, result, "should match the expected amount")
	})

	t.Run("us example credit", func(t *testing.T) {
		input := camt.ReportEntry15{
			Amt: &camt.ActiveOrHistoricCurrencyAndAmount{
				CcyAttr: "USD",
				Value:   100.00,
			},
			CdtDbtInd: "CRDT",
		}
		result, err := camt.ParseTransactionAmount(input)
		assert.NoError(t, err, "should not return an error parsing known good values")
		assert.EqualValues(t, -10000, result, "should match the expected amount")
	})
}
