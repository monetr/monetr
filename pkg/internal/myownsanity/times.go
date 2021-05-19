package myownsanity

import "time"

func TimeP(input time.Time) *time.Time {
	return &input
}

func TimesPEqual(a, b *time.Time) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	return a.Equal(*b)
}
