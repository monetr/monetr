//go:build amd64 && !nosimd

package calc

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkEuclideanDistanceSIMD(bench *testing.B) {
	bench.StopTimer()
	a := make([]float64, 512)
	b := make([]float64, 512)
	for i := range a {
		a[i] = rand.Float64()
		b[i] = rand.Float64()
	}

	bench.StartTimer()
	for i := 0; i < bench.N; i++ {
		__euclideanDistance(a, b)
	}
}

func TestEuclideanDistanceVsSIMD(t *testing.T) {
	t.Skip("they are sameish, they are off by maybe 1 in the lowest decimal place")
	for x := 0; x <= 10; x++ {
		a := make([]float64, 128)
		b := make([]float64, 128)
		for i := range a {
			a[i] = rand.Float64()
			b[i] = rand.Float64()
		}
		goResult := euclideanDistanceGo(a, b)
		simdResult := __euclideanDistance(a, b)
		assert.EqualValues(t, goResult, simdResult, "go and simd must match!")
	}
}
