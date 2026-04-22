package myownsanity

import "time"

// Deprecated: Use Pointer instead.
func TimeP(input time.Time) *time.Time {
	return &input
}

func TimesPEqual(a, b *time.Time) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	// Just to silence editor warning. Neither should be nil at this point.
	if a == nil || b == nil {
		return false
	}

	return a.Equal(*b)
}

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
