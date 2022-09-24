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

func TestMidnightInLocal(t *testing.T) {
	t.Run("weird timezone", func(t *testing.T) {
		sanLuis, err := time.LoadLocation("America/Argentina/San_Luis")
		assert.NoError(t, err, "must be able to load the sanLuis time location")
		start := time.Date(2022, 06, 15, 2, 24, 38, 0, time.UTC)

		midnight := MidnightInLocal(start, sanLuis)
		assert.Equal(t, time.Date(2022, 06, 15, 0, 0, 0, 0, sanLuis), midnight, "midnight in a different timezone is a different day")
	})

	t.Run("panics for an empty time", func(t *testing.T) {
		sanLuis, err := time.LoadLocation("America/Argentina/San_Luis")
		assert.NoError(t, err, "must be able to load the sanLuis time location")

		assert.Panics(t, func() {
			_ = MidnightInLocal(time.Time{}, sanLuis)
		})
	})

	t.Run("central time", func(t *testing.T) {
		timezone, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must load central timezone")
		// 12 seconds after midnight already in the desired timezone.
		input := time.Date(2022, 9, 15, 0, 0, 12, 0, timezone)
		expected := time.Date(2022, 9, 15, 0, 0, 0, 0, timezone)

		midnight := MidnightInLocal(input, timezone)
		assert.Equal(t, expected, midnight, "should have truncated the time, but not the timezone")
	})
}
