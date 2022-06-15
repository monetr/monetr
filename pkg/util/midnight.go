package util

import (
	"time"

	"github.com/pkg/errors"
)

// MidnightInLocal will take the provided timestamp and return the midnight of that timestamp in the provided timezome.
// This can sometimes result in returning the day prior when the function is evaluated at such a time that the current
// input time is far enough ahead that the timezone is still the previous day.
func MidnightInLocal(input time.Time, timezone *time.Location) time.Time {
	clone := time.Date(
		input.Year(),
		input.Month(),
		input.Day(),
		input.Hour(),
		input.Minute(),
		input.Second(),
		input.Nanosecond(),
		timezone,
	)
	// We need to do this because we need to know the offset for a given timezone at the provided input's timestamp. This
	// way we can adjust the input to be in that timezone then truncate the time.
	_, offset := clone.Zone()
	inputTwo := input.Add(time.Second * time.Duration(offset))

	midnight := time.Date(
		inputTwo.Year(),  // Year
		inputTwo.Month(), // Month
		inputTwo.Day(),   // Day
		0,                // Hours
		0,                // Minutes
		0,                // Seconds
		0,                // Nano seconds
		timezone,         // The account's time zone.
	)

	return midnight
}

func InLocal(input time.Time, timezone *time.Location) time.Time {
	midnight := time.Date(
		input.Year(),       // Year
		input.Month(),      // Month
		input.Day(),        // Day
		input.Hour(),       // Hours
		input.Minute(),     // Minutes
		input.Second(),     // Seconds
		input.Nanosecond(), // Nano seconds
		timezone,           // The account's time zone.
	)

	return midnight
}

// ParseInLocal parses the time string provided into a time. But ignores any timezone on the time string itself. It
// assumes that the provided time string is always in the specified timezone. This is helpful when parsing things like
// dates that have no time information at all.
func ParseInLocal(format, input string, location *time.Location) (time.Time, error) {
	date, err := time.Parse(format, input)
	if err != nil {
		return date, errors.Wrap(err, "failed to parse time")
	}

	return InLocal(date, location), nil
}
