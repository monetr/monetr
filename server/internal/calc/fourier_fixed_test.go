package calc

import (
	"math/cmplx"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Everything in this file is deliberately free of build constraints. The
// architecture specific tests live in fourier_amd64_test.go and only ever run
// on a host with the right instructions, which means that without this file the
// pure Go implementation would go completely untested on exactly the builds
// where it is the only implementation there is: every non amd64 architecture,
// and any build with the nosimd tag.

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
	// An all zero expectation would make the ratio meaningless, and no test here
	// is supposed to produce one, so say so rather than quietly returning a
	// delta of zero that passes every threshold.
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
	// The pure Go path is what runs anywhere the assembly cannot, so it needs
	// checking on its own rather than only through the assembly comparison.
	series := randomSeries(99)
	expected := FastFourierTransformSlow(series)

	actual := make([]complex128, FourierSize)
	fastFourierTransformGo(actual, series)

	delta := worstDelta(t, expected, actual)
	t.Logf("worst relative delta %.3e", delta)
	assert.Less(t, delta, 1e-13, "go implementation must agree with the recursive implementation")
}

// TestFastFourierTransformExported goes through the exported entry point rather
// than reaching past it to the unexported implementations. Every other test in
// the package calls the implementations directly, which means nothing else
// would notice if init() picked the wrong one or the length guard stopped
// working.
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
		// The assembly derives its addressing from the length it is handed while
		// indexing tables that are fixed at FourierSize, so this guard is the
		// thing standing between a bad caller and out of bounds writes.
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

// TestFastFourierTransformStaysInBounds is the test that actually earns the
// name. Checking the values of a transform proves nothing about whether it
// wrote outside its output buffer, because the assertions never look there. So
// put a known pattern on both sides of both buffers and check afterwards that
// none of it moved.
func TestFastFourierTransformStaysInBounds(t *testing.T) {
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

	fastFourierTransform(destination, source)

	assertPaddingIntact(t, destinationWhole, "the destination")
	assertPaddingIntact(t, sourceWhole, "the source")
	assert.Equal(t, sourceCopy, source, "the source must not be written at all")

	// And every slot of the destination has to actually be written, otherwise
	// the padding check above would pass for an implementation that simply did
	// nothing.
	for i := range destination {
		require.NotEqualf(t, sentinel, destination[i], "bin %d was never written", i)
	}
}
