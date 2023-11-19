package recurring

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWindowGetDeviation(t *testing.T) {
	t.Run("monthly", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		date := time.Date(2023, 11, 15, 0, 0, 0, 0, timezone)
		monthly := windowMonthly(date)

		{ // Test the start date
			delta, ok := monthly.GetDeviation(date)
			assert.EqualValues(t, 0, delta, "input date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date
			next := date.AddDate(0, 1, 0)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 0, delta, "next date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date with one day after
			next := date.AddDate(0, 1, 1)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 1, delta, "one day after the next should have a delta of 1")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date with one day before
			next := date.AddDate(0, 1, -1)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 1, delta, "one day before the next should have a delta of 1")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test before the start day
			next := date.AddDate(0, -1, 0)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, -1, delta, "invalid date should have a delta of -1")
			assert.False(t, ok, "ok should be false if the provided date comes before the start")
		}

		{ // Test outside the window
			next := date.AddDate(0, 0, 13)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, -1, delta, "should have a delta of -1 for an invalid day")
			assert.False(t, ok, "ok should be false if the provided date is outside the window")
		}
	})
}
