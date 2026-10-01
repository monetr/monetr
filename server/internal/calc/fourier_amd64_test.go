//go:build amd64 && !nosimd

package calc

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkFastFourierTransform_AVX512 preallocates the destination so only the
// transform is measured.
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
// fallback.
func BenchmarkFastFourierTransformFixed_Go(bench *testing.B) {
	series := randomSeries(1)
	result := make([]complex128, FourierSize)

	bench.ResetTimer()
	for bench.Loop() {
		fastFourierTransformGo(result, series)
	}
}

type fourierAssemblyVariant struct {
	Name      string
	Supported func() bool
	Transform func(dst, src []complex128)
}

// fourierAssemblyVariants is every forward assembly implementation in the
// package, in the same order init() prefers them.
var fourierAssemblyVariants = []fourierAssemblyVariant{
	{"AVX512", HasAVX512, __fastFourierTransform_AVX512},
	{"AVX_FMA", HasAVXFMA, __fastFourierTransform_AVX_FMA},
	{"AVX", HasAVX, __fastFourierTransform_AVX},
}

func TestFastFourierTransform_AVX512(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}
	testFastFourierTransformAssembly(t, __fastFourierTransform_AVX512)
}

func TestFastFourierTransform_AVX_FMA(t *testing.T) {
	if !HasAVXFMA() {
		t.Skip("host does not support AVX and FMA")
	}
	testFastFourierTransformAssembly(t, __fastFourierTransform_AVX_FMA)
}

func TestFastFourierTransform_AVX(t *testing.T) {
	if !HasAVX() {
		t.Skip("host does not support AVX")
	}
	testFastFourierTransformAssembly(t, __fastFourierTransform_AVX)
}

// testFastFourierTransformAssembly is every check the forward assembly
// transforms have to pass.
func testFastFourierTransformAssembly(t *testing.T, transform func(dst, src []complex128)) {
	t.Run("matches the recursive implementation", func(t *testing.T) {
		for _, seed := range []int64{1, 2, 3, 1024} {
			series := randomSeries(seed)
			expected := FastFourierTransformSlow(series)

			actual := make([]complex128, FourierSize)
			transform(actual, series)

			delta := worstDelta(t, expected, actual)
			t.Logf("seed %d worst relative delta %.3e", seed, delta)
			assert.Lessf(t, delta, 1e-13, "must agree with the recursive implementation for seed %d", seed)
		}
	})

	t.Run("matches the go implementation", func(t *testing.T) {
		series := randomSeries(7)

		expected := make([]complex128, FourierSize)
		fastFourierTransformGo(expected, series)

		actual := make([]complex128, FourierSize)
		transform(actual, series)

		// Radix-4 vs radix-2 and FMA rounding mean these won't match bit for bit,
		// just within a few units in the last place.
		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta against the go implementation %.3e", delta)
		assert.Less(t, delta, 1e-14, "assembly and go must agree to within rounding")
	})

	t.Run("impulse produces a flat spectrum", func(t *testing.T) {
		series := make([]complex128, FourierSize)
		series[0] = complex(1, 0)

		actual := make([]complex128, FourierSize)
		transform(actual, series)

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
		transform(actual, series)

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
		transform(actual, series)

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
}

func TestInverseFastFourierTransform_AVX512(t *testing.T) {
	if !HasAVX512() {
		t.Skip("host does not support AVX512")
	}
	testInverseFastFourierTransformAssembly(t, __fastFourierTransform_AVX512, __inverseFastFourierTransform_AVX512)
}

func TestInverseFastFourierTransform_AVX_FMA(t *testing.T) {
	if !HasAVXFMA() {
		t.Skip("host does not support AVX and FMA")
	}
	testInverseFastFourierTransformAssembly(t, __fastFourierTransform_AVX_FMA, __inverseFastFourierTransform_AVX_FMA)
}

func TestInverseFastFourierTransform_AVX(t *testing.T) {
	if !HasAVX() {
		t.Skip("host does not support AVX")
	}
	testInverseFastFourierTransformAssembly(t, __fastFourierTransform_AVX, __inverseFastFourierTransform_AVX)
}

// testInverseFastFourierTransformAssembly is every check the inverse assembly
// transforms have to pass. forward is the matching forward transform, which the
// round trip check runs first.
func testInverseFastFourierTransformAssembly(t *testing.T, forward, inverse func(dst, src []complex128)) {
	t.Run("matches the recursive implementation", func(t *testing.T) {
		for _, seed := range []int64{1, 2, 3, 1024} {
			spectrum := randomSeries(seed)
			expected := InverseFastFourierTransformSlow(spectrum)

			actual := make([]complex128, FourierSize)
			inverse(actual, spectrum)

			delta := worstDelta(t, expected, actual)
			t.Logf("seed %d worst relative delta %.3e", seed, delta)
			assert.Lessf(t, delta, 1e-13, "must agree with the recursive implementation for seed %d", seed)
		}
	})

	t.Run("matches the go implementation", func(t *testing.T) {
		spectrum := randomSeries(7)

		expected := make([]complex128, FourierSize)
		inverseFastFourierTransformGo(expected, spectrum)

		actual := make([]complex128, FourierSize)
		inverse(actual, spectrum)

		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta against the go implementation %.3e", delta)
		assert.Less(t, delta, 1e-14, "assembly and go must agree to within rounding")
	})

	t.Run("undoes the forward transform", func(t *testing.T) {
		series := randomSeries(8)

		spectrum := make([]complex128, FourierSize)
		forward(spectrum, series)

		actual := make([]complex128, FourierSize)
		inverse(actual, spectrum)

		delta := worstDelta(t, series, actual)
		t.Logf("worst relative delta after a round trip %.3e", delta)
		assert.Less(t, delta, 1e-14, "the inverse must give back the original series")
	})

	t.Run("flat spectrum produces an impulse", func(t *testing.T) {
		// Without the 1/n scaling the impulse would come out as n instead of 1.
		spectrum := make([]complex128, FourierSize)
		for i := range spectrum {
			spectrum[i] = complex(1, 0)
		}

		actual := make([]complex128, FourierSize)
		inverse(actual, spectrum)

		assert.InDelta(t, 1.0, real(actual[0]), 1e-12, "all of the energy belongs in the first point")
		assert.InDelta(t, 0.0, imag(actual[0]), 1e-12, "the first point must be purely real")
		for i := 1; i < FourierSize; i++ {
			require.InDeltaf(t, 0.0, cmplx.Abs(actual[i]), 1e-12, "point %d must be empty", i)
		}
	})

	t.Run("single bin produces a positive frequency", func(t *testing.T) {
		// The forward and inverse transforms only differ by the sign of the
		// exponent, and a flat spectrum looks the same either way. A single bin k
		// has to come back as e^(+2*pi*i*k*m/n), so this is what catches the
		// inverse using the forward twiddle factors.
		const bin = 3
		spectrum := make([]complex128, FourierSize)
		spectrum[bin] = complex(FourierSize, 0)

		actual := make([]complex128, FourierSize)
		inverse(actual, spectrum)

		for m := range actual {
			expected := complexExponential(2 * math.Pi * bin * float64(m) / FourierSize)
			require.InDeltaf(t, real(expected), real(actual[m]), 1e-12, "real part of point %d", m)
			require.InDeltaf(t, imag(expected), imag(actual[m]), 1e-12, "imaginary part of point %d", m)
		}
	})
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
