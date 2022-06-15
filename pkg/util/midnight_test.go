package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	t.Run("late in the day", func(t *testing.T) {
		central, err := time.LoadLocation("America/Chicago")
		assert.NoError(t, err, "must be able to load the central time location")
		start := time.Date(2022, 06, 15, 2, 24, 38, 0, time.UTC)

		midnight := MidnightInLocal(start, central)
		assert.Equal(t, time.Date(2022, 06, 14, 0, 0, 0, 0, central), midnight, "midnight in a different timezone is a different day")
	})
}
