package myownsanity

import "math"

// Deprecated: Use Pointer instead.
func Float32P(value float32) *float32 {
	return &value
}

// Deprecated: Use Pointer instead.
func Float64P(value float64) *float64 {
	return &value
}

// Deprecated: Use Pointer instead.
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

func Abs(input int64) int64 {
	mask := input >> 63
	return (input ^ mask) - mask
}

// AbsFloat32 returns the absolute value of a provided float. This is a
// non-branching implementation.
func AbsFloat32(input float32) float32 {
	// This seems silly to use instead of just doing math.Abs but math.Abs
	// requires casting to a float64 no matter what! This preserves the 32 bit
	// space and is doing the same bit trickery under the hood as math.Abs but
	// without an unnecessary cast.
	bits := math.Float32bits(input)
	bits &= 0x7FFFFFFF
	return math.Float32frombits(bits)
}

// AbsFloat64 returns the absolute value of a provided float using a
// non-branching implementation.
func AbsFloat64(input float64) float64 {
	bits := math.Float64bits(input)
	bits &= 0x7FFFFFFFFFFFFFFF
	return math.Float64frombits(bits)
}
