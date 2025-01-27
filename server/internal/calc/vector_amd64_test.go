//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/cpu"
)

func TestNormalizeVector64_AVX(t *testing.T) {
	if !cpu.X86.HasAVX {
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

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
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

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
	})

	t.Run("random-1024", func(t *testing.T) {
		base := make([]float64, 1024)
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

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
	})
}

func TestNormalizeVector64_AVX512(t *testing.T) {
	if !cpu.X86.HasAVX512F {
		t.Skip("host does not support AVX512")
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

		__normalizeVector64_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
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

		__normalizeVector64_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
	})

	t.Run("random-1024", func(t *testing.T) {
		base := make([]float64, 1024)
		for i := range base {
			base[i] = rand.Float64()
		}

		goData := make([]float64, len(base))
		copy(goData, base)
		normalizeVector64Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float64, len(base))
		copy(avxData, base)

		__normalizeVector64_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-16, "both resulting vectors should be equal-ish")
	})
}

func TestNormalizeVector32_AVX(t *testing.T) {
	if !cpu.X86.HasAVX {
		t.Skip("host does not support AVX")
	}

	t.Run("simple-8", func(t *testing.T) {
		base := []float32{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal-ish")
	})

	t.Run("random-128", func(t *testing.T) {
		base := make([]float32, 128)
		for i := range base {
			base[i] = rand.Float32()
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal-ish")
	})

	t.Run("random-1024", func(t *testing.T) {
		base := make([]float32, 1024)
		for i := range base {
			base[i] = rand.Float32()
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal")
	})
}

func TestNormalizeVector32_AVX512(t *testing.T) {
	if !cpu.X86.HasAVX512F {
		t.Skip("host does not support AVX512")
	}

	t.Run("simple-8", func(t *testing.T) {
		base := []float32{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			10,
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal-ish")
	})

	t.Run("random-128", func(t *testing.T) {
		base := make([]float32, 128)
		for i := range base {
			base[i] = rand.Float32()
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal-ish")
	})

	t.Run("random-1024", func(t *testing.T) {
		base := make([]float32, 1024)
		for i := range base {
			base[i] = rand.Float32()
		}

		goData := make([]float32, len(base))
		copy(goData, base)
		normalizeVector32Go(goData)
		fmt.Println("Go: ", goData)

		avxData := make([]float32, len(base))
		copy(avxData, base)

		__normalizeVector32_AVX512(avxData)
		fmt.Println("AVX:", avxData)

		assert.InDeltaSlice(t, goData, avxData, 1e-7, "both resulting vectors should be equal-ish")
	})
}

func BenchmarkNormalizeVector64_AVX(bench *testing.B) {
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

func BenchmarkNormalizeVector64_AVX512(bench *testing.B) {
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
			input := make([]float64, size)
			for i := range input {
				input[i] = rand.Float64()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__normalizeVector64_AVX512(input)
			}
		})
	}
}

func BenchmarkNormalizeVector32_AVX(bench *testing.B) {
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
			input := make([]float32, size)
			for i := range input {
				input[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__normalizeVector32_AVX(input)
			}
		})
	}
}

func BenchmarkNormalizeVector32_AVX512(bench *testing.B) {
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
			input := make([]float32, size)
			for i := range input {
				input[i] = rand.Float32()
			}

			bench.StartTimer()
			for i := 0; i < bench.N; i++ {
				__normalizeVector32_AVX512(input)
			}
		})
	}
}
