package ofx

import (
	"bytes"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	t.Run("standard format", func(t *testing.T) {
		ofxDate := "20240104164454.232"
		result, err := ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse a known good OFX timestamp")
		assert.EqualValues(t, time.Date(2024, 1, 4, 16, 44, 54, 232000000, time.UTC), result)
	})

	t.Run("normal alternative format", func(t *testing.T) {
		// See: https://github.com/monetr/monetr/issues/2362
		ofxDate := "20250124120000"
		result, err := ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse the alternative OFX timestamp")
		assert.EqualValues(t, time.Date(2025, 1, 24, 12, 0, 0, 0, time.UTC), result)
	})

	t.Run("weirder alternative format", func(t *testing.T) {
		// See: https://github.com/monetr/monetr/issues/2380
		ofxDate := "20240101000000[-6:CST]"
		result, err := ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse the alternative OFX timestamp")
		assert.EqualValues(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), result)
	})

	t.Run("date format without timestamp", func(t *testing.T) {
		// See: https://github.com/monetr/monetr/issues/2575
		ofxDate := "20250101"
		result, err := ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse the alternative OFX timestamp")
		assert.EqualValues(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), result)
	})

	t.Run("invalid input", func(t *testing.T) {
		ofxDate := "January 01, 2025"
		result, err := ParseDate(ofxDate, time.UTC)
		assert.EqualError(t, err, "failed to parse OFX timestamp [January 01, 2025], found 0 matching patterns")
		assert.True(t, result.IsZero(), "date returned must be zero")
	})
}

func TestParse(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-nfcu.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("nfcu wrapped", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-nfcu-wrapped.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("nfcu 2", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-nfcu-2.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("us bank", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-usbank.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.CREDITCARDMSGSRSV1, "credit card message response must not be nil")
	})

	t.Run("no curdef MXN", func(t *testing.T) {
		reader := bytes.NewReader(fixtures.LoadFile(t, "no-curdef-mxn.ofx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("non-utf8 OFX", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-non-utf8.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("non-utf8 XML", func(t *testing.T) {
		reader := bytes.NewReader(GetFixtures(t, "sample-non-utf8-xml.qfx"))

		result, err := Parse(reader)
		assert.NoError(t, err, "must not return an error parsing known valid file")
		assert.NotNil(t, result, "resulting OFX object should not be nil")
		assert.NotNil(t, result.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, result.BANKMSGSRSV1, "bank message response must not be nil")
	})
}
