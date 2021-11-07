package myownsanity

func Float32P(value float32) *float32 {
	return &value
}

func Int32P(value int32) *int32 {
	return &value
}

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
