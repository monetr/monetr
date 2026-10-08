//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildSparseVector creates a normalized vector with the requested number of
// non-zero entries, returned in both the dense and the sparse representation.
func buildSparseVector(rng *rand.Rand, size, nnz int) (dense []float32, indices []int32, values []float32) {
	dense = make([]float32, size)
	taken := make(map[int32]struct{}, nnz)
	for len(taken) < nnz {
		taken[int32(rng.Intn(size))] = struct{}{}
	}
	var norm float64
	for index := int32(0); index < int32(size); index++ {
		if _, ok := taken[index]; !ok {
			continue
		}
		value := rng.Float32() + 0.01
		dense[index] = value
		norm += float64(value) * float64(value)
	}
	norm = math.Sqrt(norm)
	for index := int32(0); index < int32(size); index++ {
		if dense[index] == 0 {
			continue
		}
		dense[index] = float32(float64(dense[index]) / norm)
		indices = append(indices, index)
		values = append(values, dense[index])
	}
	return dense, indices, values
}

func TestSparseDot32_AVX512(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}

	// The tail handling is the interesting part of this kernel, so walk every
	// length either side of the 16 lane boundary rather than just the sizes the
	// real data tends to produce.
	for _, nnz := range []int{0, 1, 2, 3, 7, 8, 15, 16, 17, 31, 32, 33, 64, 100} {
		t.Run(fmt.Sprint(nnz), func(t *testing.T) {
			rng := rand.New(rand.NewSource(int64(nnz)))
			dense, indices, values := buildSparseVector(rng, 2048, nnz)
			require.Len(t, indices, nnz, "must have built the requested number of entries")

			assert.InDelta(t,
				sparseDot32Go(dense, indices, values),
				__sparseDot32_AVX512(dense, indices, values),
				1e-6,
				"the assembly implementation must agree with the go implementation",
			)
		})
	}
}

// TestSparseDot32_MatchesEuclideanDistance32 proves the identity the sparse
// kernel relies on. The squared euclidean distance between two vectors is
// ||a||^2 + ||b||^2 - 2(a . b), so the dot product of one dense vector and one
// sparse vector is enough to recover the distance that EuclideanDistance32
// would have calculated by walking every index of both.
func TestSparseDot32_MatchesEuclideanDistance32(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}

	rng := rand.New(rand.NewSource(9))
	for range 500 {
		aDense, _, aValues := buildSparseVector(rng, 2048, 1+rng.Intn(12))
		bDense, bIndices, bValues := buildSparseVector(rng, 2048, 1+rng.Intn(12))

		var aNorm2, bNorm2 float32
		for _, value := range aValues {
			aNorm2 += value * value
		}
		for _, value := range bValues {
			bNorm2 += value * value
		}

		dot := SparseDot32(aDense, bIndices, bValues)
		assert.InDelta(t,
			EuclideanDistance32(aDense, bDense),
			aNorm2+bNorm2-2*dot,
			1e-5,
			"the sparse identity must agree with the dense euclidean distance",
		)
	}
}

// TestSparseDot32_ZeroIntersection covers the assumption the DBSCAN signature
// prefilter is built on. Two normalized vectors that share no indicies have a
// dot product of exactly zero, which puts them at a squared distance of ~2.0.
func TestSparseDot32_ZeroIntersection(t *testing.T) {
	dense := make([]float32, 64)
	dense[0], dense[1] = 0.6, 0.8

	assert.Zero(t,
		SparseDot32(dense, []int32{2, 3}, []float32{0.6, 0.8}),
		"vectors that share no indicies must have a dot product of zero",
	)
}

func BenchmarkSparseDot32_Go(bench *testing.B) {
	for _, nnz := range []int{2, 4, 8, 16, 32} {
		bench.Run(fmt.Sprint(nnz), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(nnz)))
			dense, indices, values := buildSparseVector(rng, 2048, nnz)
			bench.ResetTimer()
			for bench.Loop() {
				sparseDot32Go(dense, indices, values)
			}
		})
	}
}

func BenchmarkSparseDot32_AVX512(bench *testing.B) {
	if !HasAVX512() {
		bench.Skip("host does not support AVX512")
	}

	for _, nnz := range []int{2, 4, 8, 16, 32} {
		bench.Run(fmt.Sprint(nnz), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(nnz)))
			dense, indices, values := buildSparseVector(rng, 2048, nnz)
			bench.ResetTimer()
			for bench.Loop() {
				__sparseDot32_AVX512(dense, indices, values)
			}
		})
	}
}
