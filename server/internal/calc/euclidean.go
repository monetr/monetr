package calc

import (
	"math"
)

var (
	euclideanImplementation64 func(a, b []float64) float64 = euclideanDistanceGo
)

func euclideanDistanceGo(a, b []float64) float64 {
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
	if len(a)%4 != 0 {
		panic("length of a and b must be divisible by 4 for compatability reasons")
	}
	return euclideanImplementation64(a, b)
}
