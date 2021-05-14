package util

import (
	"github.com/pkg/errors"
	"time"
)

func MidnightInLocal(input time.Time, timezone *time.Location) time.Time {
	midnight := time.Date(
		input.Year(),  // Year
		input.Month(), // Month
		input.Day(),   // Day
		0,             // Hours
		0,             // Minutes
		0,             // Seconds
		0,             // Nano seconds
		timezone,      // The account's time zone.
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
