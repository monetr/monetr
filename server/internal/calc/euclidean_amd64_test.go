//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/cpu"
)

func BenchmarkEuclideanDistance64_AVX(bench *testing.B) {
	if !cpu.X86.HasAVX {
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
				__euclideanDistance64_AVX(a, b)
			}
		})
	}
}

func BenchmarkEuclideanDistance32_AVX(bench *testing.B) {
	if !cpu.X86.HasAVX {
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
			a := make([]float32, size)
			b := make([]float32, size)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__euclideanDistance32_AVX(a, b)
			}
		})
	}
}

func BenchmarkEuclideanDistance64_AVX512(bench *testing.B) {
	if !cpu.X86.HasAVX512F {
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

func BenchmarkEuclideanDistance32_AVX512(bench *testing.B) {
	if !cpu.X86.HasAVX512F {
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
			a := make([]float32, size)
			b := make([]float32, size)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__euclideanDistance32_AVX512(a, b)
			}
		})
	}
}

func TestEuclideanDistance64_AVX(t *testing.T) {
	if !cpu.X86.HasAVX {
		t.Skip("host does not support AVX")
	}

	for x := 0; x <= 10; x++ {
		a := make([]float64, 128)
		b := make([]float64, 128)
		for i := range a {
			a[i] = rand.Float64()
			b[i] = rand.Float64()
		}
		goResult := euclideanDistance64Go(a, b)
		simdResult := __euclideanDistance64_AVX(a, b)
		assert.InDelta(t, goResult, simdResult, 1e-13, "must be within delta")
	}
}

func TestEuclideanDistance32_AVX(t *testing.T) {
	if !cpu.X86.HasAVX {
		t.Skip("host does not support AVX")
	}

	t.Run("simple-8", func(t *testing.T) {
		a := []float32{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
		}
		b := []float32{
			8,
			7,
			6,
			5,
			4,
			3,
			2,
			1,
		}
		goResult := euclideanDistance32Go(a, b)
		simdResult := __euclideanDistance32_AVX(a, b)
		assert.InDelta(t, goResult, simdResult, 1e-4, "must be within delta")
	})

	t.Run("complex-128", func(t *testing.T) {
		for x := 0; x <= 10; x++ {
			a := make([]float32, 128)
			b := make([]float32, 128)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}
			goResult := euclideanDistance32Go(a, b)
			simdResult := __euclideanDistance32_AVX(a, b)
			fmt.Println("go:  ", goResult)
			fmt.Println("simd:", simdResult)
			assert.InDelta(t, goResult, simdResult, 1e-3, "must be within delta")
		}
	})

	t.Run("complex-1024", func(t *testing.T) {
		for x := 0; x <= 10; x++ {
			a := make([]float32, 1024)
			b := make([]float32, 1024)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}
			goResult := euclideanDistance32Go(a, b)
			simdResult := __euclideanDistance32_AVX(a, b)
			fmt.Println("go:  ", goResult)
			fmt.Println("simd:", simdResult)
			fmt.Println()
			assert.InDelta(t, goResult, simdResult, 1e-3, "must be within delta")
		}
	})
}

func TestEuclideanDistance64_AVX512(t *testing.T) {
	if !cpu.X86.HasAVX512F {
		t.Skip("host does not support AVX512")
	}

	for x := 0; x <= 10; x++ {
		a := make([]float64, 128)
		b := make([]float64, 128)
		for i := range a {
			a[i] = rand.Float64()
			b[i] = rand.Float64()
		}
		goResult := euclideanDistance64Go(a, b)
		simdResult := __euclideanDistance64_AVX512(a, b)
		assert.InDelta(t, goResult, simdResult, 1e-13, "must be within delta")
	}
}

func TestEuclideanDistance32_AVX512(t *testing.T) {
	if !cpu.X86.HasAVX512F {
		t.Skip("host does not support AVX")
	}

	t.Run("simple-8", func(t *testing.T) {
		a := []float32{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
			13,
			14,
			15,
			16,
		}
		b := []float32{
			16,
			15,
			14,
			13,
			12,
			11,
			10,
			9,
			8,
			7,
			6,
			5,
			4,
			3,
			2,
			1,
		}
		goResult := euclideanDistance32Go(a, b)
		simdResult := __euclideanDistance32_AVX512(a, b)
		fmt.Println("Go:     ", goResult)
		fmt.Println("avx512: ", simdResult)
		assert.InDelta(t, goResult, simdResult, 1e-4, "must be within delta")
	})

	t.Run("complex-128", func(t *testing.T) {
		for x := 0; x <= 10; x++ {
			a := make([]float32, 128)
			b := make([]float32, 128)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}
			goResult := euclideanDistance32Go(a, b)
			simdResult := __euclideanDistance32_AVX512(a, b)
			fmt.Println("go:  ", goResult)
			fmt.Println("simd:", simdResult)
			assert.InDelta(t, goResult, simdResult, 1e-3, "must be within delta")
		}
	})

	t.Run("complex-1024", func(t *testing.T) {
		for x := 0; x <= 10; x++ {
			a := make([]float32, 1024)
			b := make([]float32, 1024)
			for i := range a {
				a[i] = rand.Float32()
				b[i] = rand.Float32()
			}
			goResult := euclideanDistance32Go(a, b)
			simdResult := __euclideanDistance32_AVX512(a, b)
			fmt.Println("go:  ", goResult)
			fmt.Println("simd:", simdResult)
			fmt.Println()
			assert.InDelta(t, goResult, simdResult, 1e-3, "must be within delta")
		}
	})
}
