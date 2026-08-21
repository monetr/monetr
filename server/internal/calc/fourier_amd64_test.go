//go:build amd64 && !nosimd

package calc

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// fourierAssemblyVariant is one of the hand written transforms along with the
// CPU feature test that says whether this host can run it.
type fourierAssemblyVariant struct {
	Name      string
	Supported func() bool
	Transform func(dst, src []complex128)
}

// fourierAssemblyVariants is every assembly implementation in the package, in
// the same order init() prefers them.
var fourierAssemblyVariants = []fourierAssemblyVariant{
	{"AVX512", HasAVX512, __fastFourierTransform_AVX512},
	{"AVX_FMA", HasAVXFMA, __fastFourierTransform_AVX_FMA},
	{"AVX", HasAVX, __fastFourierTransform_AVX},
}

// TestFastFourierTransformVariants holds every implementation to the same
// standard rather than trusting that the narrower ones came out right just
// because the widest one did. The 256 bit versions are not simply the AVX512
// one with smaller registers: they transpose two lanes instead of four, they
// use VXORPD where the AVX512 version uses VPXORQ, and the non-FMA one does its
// complex multiply with VADDSUBPD, an instruction AVX512 does not even have.
func TestFastFourierTransformVariants(t *testing.T) {
	for _, variant := range fourierAssemblyVariants {
		t.Run(variant.Name, func(t *testing.T) {
			if !variant.Supported() {
				t.Skipf("host does not support %s", variant.Name)
			}

			t.Run("matches the recursive implementation", func(t *testing.T) {
				for _, seed := range []int64{1, 2, 3, 1024} {
					series := randomSeries(seed)
					expected := FastFourierTransformSlow(series)

					actual := make([]complex128, FourierSize)
					variant.Transform(actual, series)

					delta := worstDelta(t, expected, actual)
					t.Logf("seed %d worst relative delta %.3e", seed, delta)
					assert.Lessf(t, delta, 1e-13, "%s must agree with the recursive implementation for seed %d", variant.Name, seed)
				}
			})

			t.Run("matches the go implementation", func(t *testing.T) {
				// Worth checking each variant against the Go separately rather than
				// only chaining through the recursive comparison, because the three
				// assembly versions do their arithmetic differently enough from each
				// other that a shared reference is the only fair judge.
				series := randomSeries(7)

				expected := make([]complex128, FourierSize)
				fastFourierTransformGo(expected, series)

				actual := make([]complex128, FourierSize)
				variant.Transform(actual, series)

				// These will not agree bit for bit. The assembly folds the stages
				// into radix-4 passes while the Go stays radix-2, so the additions
				// associate differently, and two of the three variants use a fused
				// multiply-add that rounds once where Go rounds twice. Neither is an
				// error, so all that is worth asserting is that they land within a
				// few units in the last place.
				delta := worstDelta(t, expected, actual)
				t.Logf("worst relative delta against the go implementation %.3e", delta)
				assert.Less(t, delta, 1e-14, "assembly and go must agree to within rounding")
			})

			t.Run("impulse produces a flat spectrum", func(t *testing.T) {
				// A single impulse at the front has to transform to a flat spectrum
				// of ones. This is a spectrum shape check, not a bounds check;
				// TestFastFourierTransformStaysInBounds is what covers the buffers.
				series := make([]complex128, FourierSize)
				series[0] = complex(1, 0)

				actual := make([]complex128, FourierSize)
				variant.Transform(actual, series)

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
				variant.Transform(actual, series)

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
				variant.Transform(actual, series)

				var timeDomain, frequencyDomain float64
				for i := range series {
					timeDomain += real(series[i]) * real(series[i])
				}
				for i := range actual {
					magnitude := cmplx.Abs(actual[i])
					frequencyDomain += magnitude * magnitude
				}
				frequencyDomain /= float64(FourierSize)

				assert.InDelta(t, timeDomain, frequencyDomain, 1e-6, "must validate Parseval's theorem")
			})
		})
	}
}

func BenchmarkFastFourierTransformVariants(bench *testing.B) {
	for _, variant := range fourierAssemblyVariants {
		bench.Run(variant.Name, func(bench *testing.B) {
			if !variant.Supported() {
				bench.Skipf("host does not support %s", variant.Name)
			}

			series := randomSeries(1)
			result := make([]complex128, FourierSize)

			bench.ResetTimer()
			for bench.Loop() {
				variant.Transform(result, series)
			}
		})
	}
}
