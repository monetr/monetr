//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/klauspost/cpuid/v2"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEuclideanDistanceAVX(bench *testing.B) {
	if !cpuid.CPU.Has(cpuid.AVX) {
		bench.Skip("host does not support AVX")
	}

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
	if !cpuid.CPU.Has(cpuid.AVX512F) {
		bench.Skip("host does not support AVX512")
	}

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
	if !cpuid.CPU.Has(cpuid.AVX) {
		t.Skip("host does not support AVX")
	}

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

func TestEuclideanDistance64AVX512(t *testing.T) {
	if !cpuid.CPU.Has(cpuid.AVX512F) {
		t.Skip("host does not support AVX512")
	}

	for x := 0; x <= 10; x++ {
		a := make([]float64, 128)
		b := make([]float64, 128)
		for i := range a {
			a[i] = rand.Float64()
			b[i] = rand.Float64()
		}
		goResult := euclideanDistanceGo(a, b)
		simdResult := __euclideanDistance64_AVX512(a, b)
		assert.InDelta(t, goResult, simdResult, 1e-13, "must be within delta")
	}
}
