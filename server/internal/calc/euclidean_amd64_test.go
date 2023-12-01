//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkEuclideanDistanceAVX(bench *testing.B) {
	sizes := []int{
		16,
		32,
		64,
		128,
		256,
		512,
		1024,
		2048,
		4096,
		8192,
	}
	for _, size := range sizes {
		bench.Run(fmt.Sprint(size), func(bench *testing.B) {
			bench.StopTimer()
			a := make([]float64, size)
			b := make([]float64, size)
			for i := range a {
				a[i] = rand.Float64()
				b[i] = rand.Float64()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__euclideanDistance64(a, b)
			}
		})
	}
}

func BenchmarkEuclideanDistanceAVX512(bench *testing.B) {
	sizes := []int{
		16,
		32,
		64,
		128,
		256,
		512,
		1024,
		2048,
		4096,
		8192,
	}
	for _, size := range sizes {
		bench.Run(fmt.Sprint(size), func(bench *testing.B) {
			bench.StopTimer()
			a := make([]float64, size)
			b := make([]float64, size)
			for i := range a {
				a[i] = rand.Float64()
				b[i] = rand.Float64()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__euclideanDistance64_AVX512(a, b)
			}
		})
	}
}

func TestEuclideanDistance64AVX(t *testing.T) {
	for x := 0; x <= 10; x++ {
		a := make([]float64, 128)
		b := make([]float64, 128)
		for i := range a {
			a[i] = rand.Float64()
			b[i] = rand.Float64()
		}
		goResult := euclideanDistanceGo(a, b)
		simdResult := __euclideanDistance64(a, b)
		assert.InDelta(t, goResult, simdResult, 1e-13, "must be within delta")
	}
}
