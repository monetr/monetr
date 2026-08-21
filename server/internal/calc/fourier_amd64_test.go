//go:build amd64 && !nosimd

package calc

import (
	"math"
	"math/cmplx"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// worstDelta reports the largest absolute difference between two transforms
// relative to the largest magnitude in the expected result. Comparing relative
// to the peak is the honest way to do this, an absolute tolerance would be
// meaningless when the coefficients themselves span several orders of
// magnitude.
func worstDelta(t testing.TB, expected, actual []complex128) float64 {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "both transforms must be the same length")

	var worst, peak float64
	for i := range expected {
		if delta := cmplx.Abs(expected[i] - actual[i]); delta > worst {
			worst = delta
		}
		if magnitude := cmplx.Abs(expected[i]); magnitude > peak {
			peak = magnitude
		}
	}
	if peak == 0 {
		return worst
	}
	return worst / peak
}

func randomSeries(seed int64) []complex128 {
	random := rand.New(rand.NewSource(seed))
	series := make([]complex128, FourierSize)
	for i := range series {
		series[i] = complex(random.NormFloat64(), random.NormFloat64())
	}
	return series
}

func TestFastFourierTransform_AVX512(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}

	t.Run("matches the recursive implementation", func(t *testing.T) {
		for _, seed := range []int64{1, 2, 3, 1024} {
			series := randomSeries(seed)
			expected := FastFourierTransformSlow(series)

			actual := make([]complex128, FourierSize)
			__fastFourierTransform_AVX512(actual, series)

			delta := worstDelta(t, expected, actual)
			t.Logf("seed %d worst relative delta %.3e", seed, delta)
			assert.Lessf(t, delta, 1e-13, "assembly must agree with the recursive implementation for seed %d", seed)
		}
	})

	t.Run("matches the go implementation", func(t *testing.T) {
		series := randomSeries(7)

		expected := make([]complex128, FourierSize)
		fastFourierTransformGo(expected, series)

		actual := make([]complex128, FourierSize)
		__fastFourierTransform_AVX512(actual, series)

		// These two run the same algorithm in the same order off the same
		// twiddle factors, but they still will not agree bit for bit. The
		// assembly does its complex multiply with a fused multiply-add, which
		// rounds once, where Go rounds the multiply and the add separately. That
		// leaves the assembly a little more accurate than the Go, not less, so
		// the only thing worth asserting is that they agree to within a couple
		// of units in the last place.
		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta %.3e", delta)
		assert.Less(t, delta, 1e-15, "assembly and go must agree to within rounding")
	})

	t.Run("does not read past the input", func(t *testing.T) {
		// A single impulse at the front transforms into a flat spectrum of ones.
		// Anything reading or writing outside the series would show up loudly.
		series := make([]complex128, FourierSize)
		series[0] = complex(1, 0)

		actual := make([]complex128, FourierSize)
		__fastFourierTransform_AVX512(actual, series)

		for i := range actual {
			require.InDeltaf(t, 1.0, real(actual[i]), 1e-12, "real part of bin %d", i)
			require.InDeltaf(t, 0.0, imag(actual[i]), 1e-12, "imaginary part of bin %d", i)
		}
	})

	t.Run("constant input only has a dc component", func(t *testing.T) {
		series := make([]complex128, FourierSize)
		for i := range series {
			series[i] = complex(1, 0)
		}

		actual := make([]complex128, FourierSize)
		__fastFourierTransform_AVX512(actual, series)

		assert.InDelta(t, float64(FourierSize), real(actual[0]), 1e-9, "all of the energy belongs in bin zero")
		assert.InDelta(t, 0.0, imag(actual[0]), 1e-9, "bin zero must be purely real")
		for i := 1; i < FourierSize; i++ {
			require.InDeltaf(t, 0.0, cmplx.Abs(actual[i]), 1e-9, "bin %d must be empty", i)
		}
	})

	t.Run("parsevals theorem", func(t *testing.T) {
		series := make([]complex128, FourierSize)
		for i := range series {
			series[i] = complex(math.Sin(2*math.Pi*5*float64(i)/128.0), 0)
		}

		actual := make([]complex128, FourierSize)
		__fastFourierTransform_AVX512(actual, series)

		var timeDomain, frequencyDomain float64
		for i := range series {
			timeDomain += real(series[i]) * real(series[i])
		}
		for i := range actual {
			magnitude := cmplx.Abs(actual[i])
			frequencyDomain += magnitude * magnitude
		}
		frequencyDomain /= float64(FourierSize)

		t.Logf("energy in the time domain %.9f, in the frequency domain %.9f", timeDomain, frequencyDomain)
		assert.InDelta(t, timeDomain, frequencyDomain, 1e-6, "must validate Parseval's theorem")
	})
}

// BenchmarkFastFourierTransform_AVX512 times just the transform itself with the
// destination buffer already allocated, so the number is the cost of the
// assembly on its own rather than the cost of the assembly plus a 64KB
// allocation.
func BenchmarkFastFourierTransform_AVX512(bench *testing.B) {
	if !HasAVX512() {
		bench.Skip("host does not support AVX512")
	}

	series := randomSeries(1)
	result := make([]complex128, FourierSize)

	bench.ResetTimer()
	for bench.Loop() {
		__fastFourierTransform_AVX512(result, series)
	}
}

// BenchmarkFastFourierTransformFixed_Go is the same measurement for the pure Go
// fallback, which is what hosts without AVX512 end up running.
func BenchmarkFastFourierTransformFixed_Go(bench *testing.B) {
	series := randomSeries(1)
	result := make([]complex128, FourierSize)

	bench.ResetTimer()
	for bench.Loop() {
		fastFourierTransformGo(result, series)
	}
}

func TestFastFourierTransformFixedGo(t *testing.T) {
	// The pure Go path is what runs on hosts without AVX512, so it needs to be
	// checked on its own rather than only through the assembly comparison.
	series := randomSeries(99)
	expected := FastFourierTransformSlow(series)

	actual := make([]complex128, FourierSize)
	fastFourierTransformGo(actual, series)

	delta := worstDelta(t, expected, actual)
	t.Logf("worst relative delta %.3e", delta)
	assert.Less(t, delta, 1e-13, "go implementation must agree with the recursive implementation")
}
