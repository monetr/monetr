package calc

import (
	"math/cmplx"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// No build constraints here so the Go fallback is tested on non amd64 builds
// and with the nosimd tag.

// worstDelta reports the largest absolute difference between two transforms
// relative to the largest magnitude in the expected result, since the
// coefficients span several orders of magnitude.
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
	require.NotZero(t, peak, "the expected transform must not be all zeroes")
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

func TestFastFourierTransformFixedGo(t *testing.T) {
	series := randomSeries(99)
	expected := FastFourierTransformSlow(series)

	actual := make([]complex128, FourierSize)
	fastFourierTransformGo(actual, series)

	delta := worstDelta(t, expected, actual)
	t.Logf("worst relative delta %.3e", delta)
	assert.Less(t, delta, 1e-13, "go implementation must agree with the recursive implementation")
}

// TestFastFourierTransformExported covers the dispatch in init() and the length
// guard, which the other tests skip by calling the implementations directly.
func TestFastFourierTransformExported(t *testing.T) {
	t.Run("transforms through whichever implementation init picked", func(t *testing.T) {
		series := randomSeries(5)
		expected := FastFourierTransformSlow(series)

		actual := FastFourierTransform(series)

		require.Len(t, actual, FourierSize, "must return a full length transform")
		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta %.3e", delta)
		assert.Less(t, delta, 1e-13, "the dispatched implementation must be correct")
	})

	t.Run("does not modify the input", func(t *testing.T) {
		series := randomSeries(6)
		original := make([]complex128, FourierSize)
		copy(original, series)

		FastFourierTransform(series)

		assert.Equal(t, original, series, "the input must be left alone")
	})

	t.Run("panics on the wrong length", func(t *testing.T) {
		// The assembly trusts the slice length, so a bad length here would mean
		// out of bounds writes.
		for _, length := range []int{0, 1, FourierSize - 1, FourierSize + 1, FourierSize * 2} {
			assert.Panicsf(t, func() {
				FastFourierTransform(make([]complex128, length))
			}, "must panic for a series of %d points", length)
		}
		assert.Panics(t, func() {
			FastFourierTransform(nil)
		}, "must panic for a nil series")
	})
}

func TestInverseFastFourierTransformFixedGo(t *testing.T) {
	t.Run("matches the recursive implementation", func(t *testing.T) {
		spectrum := randomSeries(99)
		expected := InverseFastFourierTransformSlow(spectrum)

		actual := make([]complex128, FourierSize)
		inverseFastFourierTransformGo(actual, spectrum)

		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta %.3e", delta)
		assert.Less(t, delta, 1e-13, "go implementation must agree with the recursive implementation")
	})

	t.Run("undoes the forward transform", func(t *testing.T) {
		series := randomSeries(98)

		spectrum := make([]complex128, FourierSize)
		fastFourierTransformGo(spectrum, series)

		actual := make([]complex128, FourierSize)
		inverseFastFourierTransformGo(actual, spectrum)

		delta := worstDelta(t, series, actual)
		t.Logf("worst relative delta after a round trip %.3e", delta)
		assert.Less(t, delta, 1e-14, "the inverse must give back the original series")
	})
}

// TestInverseFastFourierTransformExported is TestFastFourierTransformExported
// for the inverse transform.
func TestInverseFastFourierTransformExported(t *testing.T) {
	t.Run("transforms through whichever implementation init picked", func(t *testing.T) {
		spectrum := randomSeries(5)
		expected := InverseFastFourierTransformSlow(spectrum)

		actual := InverseFastFourierTransform(spectrum)

		require.Len(t, actual, FourierSize, "must return a full length transform")
		delta := worstDelta(t, expected, actual)
		t.Logf("worst relative delta %.3e", delta)
		assert.Less(t, delta, 1e-13, "the dispatched implementation must be correct")
	})

	t.Run("undoes FastFourierTransform", func(t *testing.T) {
		series := randomSeries(4)

		actual := InverseFastFourierTransform(FastFourierTransform(series))

		delta := worstDelta(t, series, actual)
		t.Logf("worst relative delta after a round trip %.3e", delta)
		assert.Less(t, delta, 1e-14, "the inverse must give back the original series")
	})

	t.Run("does not modify the input", func(t *testing.T) {
		spectrum := randomSeries(6)
		original := make([]complex128, FourierSize)
		copy(original, spectrum)

		InverseFastFourierTransform(spectrum)

		assert.Equal(t, original, spectrum, "the input must be left alone")
	})

	t.Run("panics on the wrong length", func(t *testing.T) {
		for _, length := range []int{0, 1, FourierSize - 1, FourierSize + 1, FourierSize * 2} {
			assert.Panicsf(t, func() {
				InverseFastFourierTransform(make([]complex128, length))
			}, "must panic for a series of %d points", length)
		}
		assert.Panics(t, func() {
			InverseFastFourierTransform(nil)
		}, "must panic for a nil series")
	})
}

// TestFastFourierTransformStaysInBounds pads both buffers with a sentinel value
// and makes sure the transform never writes outside of them.
func TestFastFourierTransformStaysInBounds(t *testing.T) {
	assertTransformStaysInBounds(t, fastFourierTransform)
}

// TestInverseFastFourierTransformStaysInBounds is the same check for the
// inverse transform.
func TestInverseFastFourierTransformStaysInBounds(t *testing.T) {
	assertTransformStaysInBounds(t, inverseFastFourierTransform)
}

func assertTransformStaysInBounds(t *testing.T, transform func(dst, src []complex128)) {
	t.Helper()
	const padding = 512

	sentinel := complex(-98765.4321, 12345.6789)

	newPadded := func() (whole, middle []complex128) {
		whole = make([]complex128, FourierSize+2*padding)
		for i := range whole {
			whole[i] = sentinel
		}
		// The subslice has the right length but a much larger capacity, so an
		// implementation that walked past its length would corrupt the padding
		// instead of tripping over Go's own bounds checks.
		return whole, whole[padding : padding+FourierSize : padding+FourierSize]
	}

	assertPaddingIntact := func(t *testing.T, whole []complex128, what string) {
		t.Helper()
		for i := 0; i < padding; i++ {
			require.Equalf(t, sentinel, whole[i], "%s was written %d elements before its start", what, padding-i)
		}
		for i := padding + FourierSize; i < len(whole); i++ {
			require.Equalf(t, sentinel, whole[i], "%s was written %d elements past its end", what, i-(padding+FourierSize)+1)
		}
	}

	sourceWhole, source := newPadded()
	for i := range source {
		source[i] = complex(float64(i%37)-18, float64(i%23)-11)
	}
	sourceCopy := make([]complex128, FourierSize)
	copy(sourceCopy, source)

	destinationWhole, destination := newPadded()

	transform(destination, source)

	assertPaddingIntact(t, destinationWhole, "the destination")
	assertPaddingIntact(t, sourceWhole, "the source")
	assert.Equal(t, sourceCopy, source, "the source must not be written at all")

	// Make sure every bin was actually written too.
	for i := range destination {
		require.NotEqualf(t, sentinel, destination[i], "bin %d was never written", i)
	}
}
