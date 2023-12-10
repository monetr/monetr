package calc

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkNormalizeVector64_Go(bench *testing.B) {
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
				normalizeVector64Go(input)
			}
		})
	}
}

func BenchmarkNormalizeVector32_Go(bench *testing.B) {
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
			input := make([]float32, size)
			for i := range input {
				input[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				normalizeVector32Go(input)
			}
		})
	}
}
