package myownsanity

import "strings"

func CoalesceStrings(str ...string) string {
	for _, value := range str {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func CoalesceInts(i ...int) int {
	for _, value := range i {
		if value > 0 {
			return value
		}
	}

	return 0
}
