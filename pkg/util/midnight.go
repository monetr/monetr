package util

import (
	"time"

	"github.com/pkg/errors"
)

// Midnight will take the provider timestamp and a desired timezone. It does not assume that the provided timestamp is
// in the desired timezone, or that it is in UTC. It calculates the difference in timezone offsets and adjusts the
// timestamp accordingly. Then truncates the timestamp to produce a midnight or "start of day" timestamp which it
// returns.
// As a result, this function should **NEVER** return a timestamp that comes after the provided input. The output may be
// a later date, but will still be before the provided input once timezone offsets have been accounted for.
func Midnight(input time.Time, timezone *time.Location) time.Time {
	if input.IsZero() {
		panic("cannot calculate the midnight in local of an empty time")
	}
	// Get the timezone offsets from the input and from the timezone specified.
	_, inputZoneOffset := input.Zone()
	_, tzOffset := input.In(timezone).Zone()

	// Calculate the difference betwen them.
	// TODO: This might need to be math.Abs instead. If the timezone is _ahead_ of the input location I think there would
	// be a bug here. Not quite sure yet.
	delta := inputZoneOffset - tzOffset

	// Subtract the delta from the input time, this accounts for the timezone difference potential between the input
	// timezone and the provided timezone.
	// The input time should be treated as the absolute time of the moment. But might be in ANY timezone. It could already
	// be in the specified timezone, or it could be in UTC or some other timezone for example. But we must make sure that
	// our adjustment takes that into account. So this delta will do just that.
	adjusted := input.Add(time.Duration(-delta) * time.Second)
	// Create a timestamp in the specified timezone that is midnight.
	midnight := time.Date(
		adjusted.Year(),  // Year
		adjusted.Month(), // Month
		adjusted.Day(),   // Day
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
