package calc

import "math"

var (
	cosineImplementation func(input float64) float64 = math.Cos
)

func Cosine(input float64) float64 {
	return cosineImplementation(input)
}
