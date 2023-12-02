package calc

import (
	"math"
)

var (
	normalizeVectorImplementation64 func(input []float64) = normalizeVector64Go
)

func normalizeVector64Go(input []float64) {
	var norm float64
	for _, value := range input {
		norm += value * value
	}
	norm = math.Sqrt(norm)
	for i, value := range input {
		input[i] = value / norm
	}
}

func NormalizeVector64(input []float64) {
	if len(input)%8 != 0 {
		panic("length of the input vector must be divisible by 8 for compatability reasons")
	}
	normalizeVectorImplementation64(input)
}
