package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInLocal(t *testing.T) {
	t.Run("central time daylight savings", func(t *testing.T) {
		central, err := time.LoadLocation("America/Chicago")
		assert.NoError(t, err, "must be able to load the central time location")
		start := time.Date(2021, 11, 06, 0, 0, 0, 0, time.UTC)

		assert.False(t, start.IsDST(), "November 6th 2021 at midnight UTC should not be considered DST, UTC does not vibe with DST")

		inCentral := InLocal(start, central)
		assert.True(t, inCentral.IsDST(), "November 6th 2021 at midnight CDT should be considered DST")

		theNextDay := InLocal(start.Add(24*time.Hour), central)
		// This is because DST happens at 2AM.
		assert.True(t, theNextDay.IsDST(), "November 7th 2021 at midnight CDT should be considered DST")

		finallyDST := InLocal(start.Add(48*time.Hour), central)
		// This is because DST happens at 2AM.
		assert.False(t, finallyDST.IsDST(), "November 8th 2021 at midnight CST should not be considered DST")
	})
}

func TestMidnight(t *testing.T) {
	t.Run("panics for an empty time", func(t *testing.T) {
		sanLuis, err := time.LoadLocation("America/Argentina/San_Luis")
		assert.NoError(t, err, "must be able to load the sanLuis time location")

		assert.Panics(t, func() {
			_ = Midnight(time.Time{}, sanLuis)
		})
	})

	t.Run("central time", func(t *testing.T) {
		timezone, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must load central timezone")
		// 12 seconds after midnight already in the desired timezone.
		input := time.Date(2022, 9, 15, 0, 0, 12, 0, timezone)
		expected := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)

		midnight := Midnight(input, timezone)
		assert.Equal(t, expected, midnight, "should have truncated the time, but not the timezone")
	})

	t.Run("central but next day utc", func(t *testing.T) {
		timezone, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must load central timezone")
		now := time.Date(2023, 8, 14, 19, 30, 1, 0, timezone)
		nowUTC := now.UTC()
		expected := time.Date(2023, 8, 14, 0, 0, 0, 0, timezone)

		midnightNow := Midnight(now, timezone)
		assert.Equal(t, expected, midnightNow, "should not be in the future")

		midnightUTC := Midnight(nowUTC, timezone)
		assert.Equal(t, expected, midnightUTC, "should not be in the future")
	})
}
