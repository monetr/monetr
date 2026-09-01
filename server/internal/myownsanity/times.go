package myownsanity

import "time"

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	} else {
		return b
	}
}

func MaxNonNilTime(times ...*time.Time) *time.Time {
	var maxTime *time.Time
	for _, t := range times {
		if t == nil {
			continue
		}
		if maxTime == nil {
			maxTime = t
			continue
		}

		if t.After(*maxTime) {
			maxTime = t
		}
	}

	return maxTime
}
