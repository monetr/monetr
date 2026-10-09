//go:build amd64 && !nosimd

package calc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildSparseVectors creates count normalized sparse vectors laid out end to
// end the way SparseNeighbors32 wants them. Each one has 1 to 7 entries, which
// is about what a transaction has.
func buildSparseVectors(
	rng *rand.Rand,
	count, size int,
) (signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32) {
	signatures = make([]uint64, count)
	norms = make([]float32, count)
	offsets = make([]int32, count+1)
	for i := range count {
		_, vectorIndices, vectorValues := buildSparseVector(rng, size, 1+rng.Intn(7))
		for x, index := range vectorIndices {
			norms[i] += vectorValues[x] * vectorValues[x]
			signatures[i] |= 1 << (uint64(index) % 64)
		}
		indices = append(indices, vectorIndices...)
		values = append(values, vectorValues...)
		offsets[i+1] = int32(len(indices))
	}

	return signatures, norms, offsets, indices, values
}

func TestSparseNeighbors32_AVX(t *testing.T) {
	if !HasAVX() {
		t.Skip("host does not support AVX")
	}

	t.Run("happy path", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 1000, 128)
		dense := make([]float32, 128)
		expected := make([]int32, 1000)
		result := make([]int32, 1000)
		var found int
		// Use every vector as the point once, the same way DBSCAN does
		for point := range 1000 {
			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = values[x]
			}

			expectedCount := sparseNeighbors32Go(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, result)
			require.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors of %d should match the go implementation", point)
			found += resultCount

			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = 0
			}
		}
		// Every point is its own neighbor, so make sure we found more than that
		assert.Greater(t, found, 1000, "should have found real neighbors")
	})

	t.Run("every length", func(t *testing.T) {
		// The signature pass and the dot product pass both have their own
		// leftovers, so walk the lengths around them
		for test := range 41 {
			rng := rand.New(rand.NewSource(int64(test)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, test, 48)
			dense, _, pointValues := buildSparseVector(rng, 48, 3)
			var signature uint64
			var norm2 float32
			for index, value := range dense {
				if value == 0 {
					continue
				}
				signature |= 1 << (uint64(index) % 64)
			}
			for _, value := range pointValues {
				norm2 += value * value
			}

			expected := make([]int32, test)
			result := make([]int32, test)
			expectedCount := sparseNeighbors32Go(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, result)
			assert.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors for %d vectors should match the go implementation", test)
		}
	})

	t.Run("nothing in common", func(t *testing.T) {
		rng := rand.New(rand.NewSource(2))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 100, 128)
		dense := make([]float32, 128)
		result := make([]int32, 100)
		// A signature of 0 can't have a bit in common with anything, even with an
		// epsilon that would take every vector
		resultCount := __sparseNeighbors32_AVX(dense, 0, 1, 100, signatures, norms, offsets, indices, values, result)
		assert.Zero(t, resultCount, "should not find any neighbors")
	})

	t.Run("doesnt write past output", func(t *testing.T) {
		// Every block stores all 4 lanes even when only some of them are
		// candidates. Every vector here is a candidate, so the last block is the
		// one that would go past the end if anything did
		rng := rand.New(rand.NewSource(3))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 24, 48)
		for i := range signatures {
			signatures[i] = 1
		}
		dense := make([]float32, 48)
		buffer := make([]int32, 28)
		for i := range buffer {
			buffer[i] = -1
		}
		__sparseNeighbors32_AVX(dense, 1, 1, 100, signatures, norms, offsets, indices, values, buffer[:24:24])
		assert.Equal(t, []int32{-1, -1, -1, -1}, buffer[24:], "nothing past the end of output should change")
	})
}

func TestSparseNeighbors32_AVX512VL(t *testing.T) {
	if !HasAVX512VL() {
		t.Skip("host does not support AVX512VL")
	}

	t.Run("happy path", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 1000, 128)
		dense := make([]float32, 128)
		expected := make([]int32, 1000)
		result := make([]int32, 1000)
		var found int
		// Use every vector as the point once, the same way DBSCAN does
		for point := range 1000 {
			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = values[x]
			}

			expectedCount := sparseNeighbors32Go(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX512VL(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, result)
			require.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors of %d should match the go implementation", point)
			found += resultCount

			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = 0
			}
		}
		// Every point is its own neighbor, so make sure we found more than that
		assert.Greater(t, found, 1000, "should have found real neighbors")
	})

	t.Run("every length", func(t *testing.T) {
		// The signature pass does blocks of 8 and then 1 at a time, so walk the
		// lengths around a few blocks
		for test := range 41 {
			rng := rand.New(rand.NewSource(int64(test)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, test, 48)
			dense, _, pointValues := buildSparseVector(rng, 48, 3)
			var signature uint64
			var norm2 float32
			for index, value := range dense {
				if value == 0 {
					continue
				}
				signature |= 1 << (uint64(index) % 64)
			}
			for _, value := range pointValues {
				norm2 += value * value
			}

			expected := make([]int32, test)
			result := make([]int32, test)
			expectedCount := sparseNeighbors32Go(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX512VL(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, result)
			assert.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors for %d vectors should match the go implementation", test)
		}
	})

	t.Run("nothing in common", func(t *testing.T) {
		rng := rand.New(rand.NewSource(2))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 100, 128)
		dense := make([]float32, 128)
		result := make([]int32, 100)
		// A signature of 0 can't have a bit in common with anything, even with an
		// epsilon that would take every vector
		resultCount := __sparseNeighbors32_AVX512VL(dense, 0, 1, 100, signatures, norms, offsets, indices, values, result)
		assert.Zero(t, resultCount, "should not find any neighbors")
	})

	t.Run("doesnt write past output", func(t *testing.T) {
		// Every block stores all 8 lanes even when only some of them are
		// candidates. Every vector here is a candidate, so the last block is the
		// one that would go past the end if anything did
		rng := rand.New(rand.NewSource(3))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 24, 48)
		for i := range signatures {
			signatures[i] = 1
		}
		dense := make([]float32, 48)
		buffer := make([]int32, 32)
		for i := range buffer {
			buffer[i] = -1
		}
		__sparseNeighbors32_AVX512VL(dense, 1, 1, 100, signatures, norms, offsets, indices, values, buffer[:24:24])
		assert.Equal(t, []int32{-1, -1, -1, -1, -1, -1, -1, -1}, buffer[24:], "nothing past the end of output should change")
	})
}

func TestSparseNeighbors32_AVX512(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}

	t.Run("happy path", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 1000, 128)
		dense := make([]float32, 128)
		expected := make([]int32, 1000)
		result := make([]int32, 1000)
		var found int
		// Use every vector as the point once, the same way DBSCAN does
		for point := range 1000 {
			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = values[x]
			}

			expectedCount := sparseNeighbors32Go(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX512(dense, signatures[point], norms[point], 0.5, signatures, norms, offsets, indices, values, result)
			require.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors of %d should match the go implementation", point)
			found += resultCount

			for x := offsets[point]; x < offsets[point+1]; x++ {
				dense[indices[x]] = 0
			}
		}
		// Every point is its own neighbor, so make sure we found more than that
		assert.Greater(t, found, 1000, "should have found real neighbors")
	})

	t.Run("every length", func(t *testing.T) {
		// The signature pass does blocks of 16 and then 1 at a time, so walk the
		// lengths around a few blocks
		for test := range 41 {
			rng := rand.New(rand.NewSource(int64(test)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, test, 48)
			dense, _, pointValues := buildSparseVector(rng, 48, 3)
			var signature uint64
			var norm2 float32
			for index, value := range dense {
				if value == 0 {
					continue
				}
				signature |= 1 << (uint64(index) % 64)
			}
			for _, value := range pointValues {
				norm2 += value * value
			}

			expected := make([]int32, test)
			result := make([]int32, test)
			expectedCount := sparseNeighbors32Go(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, expected)
			resultCount := __sparseNeighbors32_AVX512(dense, signature, norm2, 0.98, signatures, norms, offsets, indices, values, result)
			assert.Equal(t, expected[:expectedCount], result[:resultCount], "neighbors for %d vectors should match the go implementation", test)
		}
	})

	t.Run("nothing in common", func(t *testing.T) {
		rng := rand.New(rand.NewSource(2))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 100, 128)
		dense := make([]float32, 128)
		result := make([]int32, 100)
		// A signature of 0 can't have a bit in common with anything, even with an
		// epsilon that would take every vector
		resultCount := __sparseNeighbors32_AVX512(dense, 0, 1, 100, signatures, norms, offsets, indices, values, result)
		assert.Zero(t, resultCount, "should not find any neighbors")
	})

	t.Run("doesnt write past output", func(t *testing.T) {
		// Every block stores all 16 lanes even when only some of them are
		// candidates. Every vector here is a candidate, so the last block is the
		// one that would go past the end if anything did
		rng := rand.New(rand.NewSource(3))
		signatures, norms, offsets, indices, values := buildSparseVectors(rng, 32, 48)
		for i := range signatures {
			signatures[i] = 1
		}
		dense := make([]float32, 48)
		buffer := make([]int32, 48)
		for i := range buffer {
			buffer[i] = -1
		}
		__sparseNeighbors32_AVX512(dense, 1, 1, 100, signatures, norms, offsets, indices, values, buffer[:32:32])
		assert.Equal(t, []int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, buffer[32:], "nothing past the end of output should change")
	})
}

func BenchmarkSparseNeighbors32_Go(bench *testing.B) {
	for _, count := range []int{256, 1024, 4096} {
		bench.Run(fmt.Sprint(count), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(count)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, count, count/3)
			dense := make([]float32, count/3)
			for x := offsets[0]; x < offsets[1]; x++ {
				dense[indices[x]] = values[x]
			}
			output := make([]int32, count)
			for bench.Loop() {
				sparseNeighbors32Go(dense, signatures[0], norms[0], 0.5, signatures, norms, offsets, indices, values, output)
			}
		})
	}
}

func BenchmarkSparseNeighbors32_AVX(bench *testing.B) {
	if !HasAVX() {
		bench.Skip("host does not support AVX")
	}

	for _, count := range []int{256, 1024, 4096} {
		bench.Run(fmt.Sprint(count), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(count)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, count, count/3)
			dense := make([]float32, count/3)
			for x := offsets[0]; x < offsets[1]; x++ {
				dense[indices[x]] = values[x]
			}
			output := make([]int32, count)
			for bench.Loop() {
				__sparseNeighbors32_AVX(dense, signatures[0], norms[0], 0.5, signatures, norms, offsets, indices, values, output)
			}
		})
	}
}

func BenchmarkSparseNeighbors32_AVX512VL(bench *testing.B) {
	if !HasAVX512VL() {
		bench.Skip("host does not support AVX512VL")
	}

	for _, count := range []int{256, 1024, 4096} {
		bench.Run(fmt.Sprint(count), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(count)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, count, count/3)
			dense := make([]float32, count/3)
			for x := offsets[0]; x < offsets[1]; x++ {
				dense[indices[x]] = values[x]
			}
			output := make([]int32, count)
			for bench.Loop() {
				__sparseNeighbors32_AVX512VL(dense, signatures[0], norms[0], 0.5, signatures, norms, offsets, indices, values, output)
			}
		})
	}
}

func BenchmarkSparseNeighbors32_AVX512(bench *testing.B) {
	if !HasAVX512() {
		bench.Skip("host does not support AVX512")
	}

	for _, count := range []int{256, 1024, 4096} {
		bench.Run(fmt.Sprint(count), func(bench *testing.B) {
			rng := rand.New(rand.NewSource(int64(count)))
			signatures, norms, offsets, indices, values := buildSparseVectors(rng, count, count/3)
			dense := make([]float32, count/3)
			for x := offsets[0]; x < offsets[1]; x++ {
				dense[indices[x]] = values[x]
			}
			output := make([]int32, count)
			for bench.Loop() {
				__sparseNeighbors32_AVX512(dense, signatures[0], norms[0], 0.5, signatures, norms, offsets, indices, values, output)
			}
		})
	}
}
