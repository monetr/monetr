package calc

import (
	"math"
)

func euclideanDistanceVanilla(a, b []float64) float64 {
	var distance float64
	for i, value := range a {
		distance += math.Pow(value-b[i], 2)
	}
	return distance
}
