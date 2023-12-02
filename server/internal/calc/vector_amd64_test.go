//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/klauspost/cpuid/v2"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeVector64_AVX(t *testing.T) {
	if !cpuid.CPU.Has(cpuid.AVX) {
		t.Skip("host does not support AVX")
	}

	t.Run("simple-8", func(t *testing.T) {
		base := []float64{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
		}

		goData := make([]float64, len(base))
		copy(goData, base)
		normalizeVector64Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float64, len(base))
		copy(avxData, base)

		__normalizeVector64_AVX(avxData)
		fmt.Println("AVX:", avxData)

		assert.EqualValues(t, goData, avxData, "both resulting vectors should be equal")
	})

	t.Run("random-128", func(t *testing.T) {
		base := make([]float64, 128)
		for i := range base {
			base[i] = rand.Float64()
		}

		goData := make([]float64, len(base))
		copy(goData, base)
		normalizeVector64Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float64, len(base))
		copy(avxData, base)

		__normalizeVector64_AVX(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal")
	})
}

func BenchmarkNormalizeVector64_AVX(bench *testing.B) {
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
			input := make([]float64, size)
			for i := range input {
				input[i] = rand.Float64()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__normalizeVector64_AVX(input)
			}
		})
	}
}
