package calc

import (
	"math"
)

var (
	euclideanImplementation64 func(a, b []float64) float64 = euclideanDistance64Go
	euclideanImplementation32 func(a, b []float32) float32 = euclideanDistance32Go
)

func euclideanDistance64Go(a, b []float64) float64 {
	var distance float64
	for i := range a {
		x := a[i] - b[i]
		distance += x * x
	}
	return distance
}

func euclideanDistanceSlow(a, b []float64) float64 {
	var distance float64
	for i := range a {
		distance += math.Pow(a[i]-b[i], 2)
	}
	return distance
}

func EuclideanDistance64(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("invalid euclidean vectors provided, length must match!")
	}
	if len(a)%8 != 0 {
		panic("length of a and b must be divisible by 8 for compatability reasons")
	}
	return euclideanImplementation64(a, b)
}

func euclideanDistance32Go(a, b []float32) float32 {
	var distance float32
	for i := range a {
		x := a[i] - b[i]
		distance += x * x
	}
	return distance
}

func EuclideanDistance32(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("invalid euclidean vectors provided, length must match!")
	}
	if len(a)%16 != 0 {
		panic("length of a and b must be divisible by 16 for compatability reasons")
	}
	return euclideanImplementation32(a, b)
}
