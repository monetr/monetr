package util

import "time"

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
