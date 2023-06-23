package myownsanity

func Float32P(value float32) *float32 {
	return &value
}

func Float64P(value float64) *float64 {
	return &value
}

func Int32P(value int32) *int32 {
	return &value
}

type Number interface {
	int | int32 | int64
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}

	return b
}
