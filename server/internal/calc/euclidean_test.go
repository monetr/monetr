package calc

import (
	"math/rand"
	"testing"
)

func BenchmarkEuclideanDistance(bench *testing.B) {
	bench.StopTimer()
	a := make([]float64, 512)
	b := make([]float64, 512)
	for i := range a {
		a[i] = rand.Float64()
		b[i] = rand.Float64()
	}

	bench.StartTimer()
	for i := 0; i < bench.N; i++ {
		euclideanDistanceGo(a, b)
	}
}
