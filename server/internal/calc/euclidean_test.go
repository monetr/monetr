package calc

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkEuclideanDistance64_Go(bench *testing.B) {
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
				euclideanDistance64Go(a, b)
			}
		})
	}
}

func BenchmarkEuclideanDistance64_GoSlow(bench *testing.B) {
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
				euclideanDistanceSlow(a, b)
			}
		})
	}
}

func BenchmarkEuclideanDistance32_Go(bench *testing.B) {
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
			a := make([]float32, size)
			b := make([]float32, size)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				euclideanDistance32Go(a, b)
			}
		})
	}
}
