package ofx_test

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/formats/ofx"
	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	t.Run("standard format", func(t *testing.T) {
		ofxDate := "20240104164454.232"
		result, err := ofx.ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse a known good OFX timestamp")
		assert.EqualValues(t, time.Date(2024, 1, 4, 16, 44, 54, 232000000, time.UTC), result)
	})

	t.Run("alternative format", func(t *testing.T) {
		// See: https://github.com/monetr/monetr/issues/2362
		ofxDate := "20250124120000"
		result, err := ofx.ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse the alternative OFX timestamp")
		assert.EqualValues(t, time.Date(2025, 1, 24, 12, 0, 0, 0, time.UTC), result)
	})

	t.Run("weirder alternative format", func(t *testing.T) {
		// See: https://github.com/monetr/monetr/issues/2380
		ofxDate := "20240101000000[-6:CST]"
		result, err := ofx.ParseDate(ofxDate, time.UTC)
		assert.NoError(t, err, "must be able to parse the alternative OFX timestamp")
		assert.EqualValues(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), result)
	})

	t.Run("invalid input", func(t *testing.T) {
		ofxDate := "20240101"
		result, err := ofx.ParseDate(ofxDate, time.UTC)
		assert.EqualError(t, err, "failed to parse OFX timestamp [20240101], found 0 matching patterns")
		assert.True(t, result.IsZero(), "date returned must be zero")
	})
}
