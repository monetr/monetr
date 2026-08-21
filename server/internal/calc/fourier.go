package calc

import (
	"math"
	"math/bits"
	"math/cmplx"
	"sync"
)

//go:generate go run ./gen -size 4096 -output fourier_twiddles_amd64.s

const FourierSize = 4096

// The assembly implementations look up their twiddle factors and their bit
// reversal offsets in tables that are generated for one specific size and named
// after it, and they take the number of points to transform from the length of
// the slice they are handed. Those two facts have to agree. If FourierSize ever
// changes without the go:generate line above changing to match and the assembly
// being pointed at the new symbols, the transform would walk off the end of the
// tables and start writing at whatever offsets it found there.
//
// This makes that a compile error instead. Changing the size means changing
// this line, the generate directive, and the symbol names in fourier_amd64.s
// together, which is exactly the set of things that have to move at once.
// A negative constant does not convert to uint, so these two together pin
// FourierSize to exactly 4096, one catching each direction.
const (
	_ = uint(FourierSize - 4096)
	_ = uint(4096 - FourierSize)
)

// FastFourierTransformSlow is a recursive implementation of the fast Fourier
// transform.
func FastFourierTransformSlow(a []complex128) []complex128 {
	n := len(a)
	if n <= 1 {
		return a
	}

	// Ensure n is a power of two
	if (n & (n - 1)) != 0 {
		panic("Length of input must be a power of two")
	}

	// Is the same as n/2, SHRQ shifts the bits left by $1
	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = a[i*2]
		odd[i] = a[i*2+1]
	}

	fftEven := FastFourierTransformSlow(even)
	fftOdd := FastFourierTransformSlow(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := complexExponential(-2*math.Pi*float64(k)/float64(n)) * fftOdd[k]
		result[k] = fftEven[k] + t
		result[k+n/2] = fftEven[k] - t
	}
	return result
}

// InverseFastFourierTransform is a recursive implementation of the inverse fast
// Fourier transform.
func InverseFastFourierTransform(a []complex128) []complex128 {
	n := len(a)

	// Conjugate the input
	conjugated := make([]complex128, n)
	for i := range a {
		conjugated[i] = cmplx.Conj(a[i])
	}

	// Apply FFT to the conjugated input
	fftConjugated := FastFourierTransformSlow(conjugated)

	// Conjugate the result and scale by 1/n
	for i := range fftConjugated {
		fftConjugated[i] = cmplx.Conj(fftConjugated[i]) / complex(float64(n), 0)
	}

	return fftConjugated
}

// Compute complex exponential (Euler's formula)
func complexExponential(theta float64) complex128 {
	return complex(math.Cos(theta), math.Sin(theta))
}

// fastFourierTransform is whichever implementation of the fixed size
// transform the host CPU can actually run. Platforms that have a hand written
// one swap this out from their own init function, the same way the euclidean
// distance and vector normalization implementations do.
var fastFourierTransform func(dst, src []complex128) = fastFourierTransformGo

// FastFourierTransform is a non-recursive forward fast Fourier transform for
// exactly FourierSize points. It does the same thing as
// FastFourierTransformSlow but it can only ever do it at the one size, which is
// what lets every root of unity and every bit reversal offset be worked out
// ahead of time instead of during the transform.
//
// On amd64 this runs in hand written SIMD assembly against tables that were
// computed at build time and baked into the binary as read only data. There are
// three of those, picked by CPU feature in init(): AVX512, AVX with fused
// multiply-add, and plain AVX. Anything else, including every other
// architecture, falls back to the equivalent Go below.
func FastFourierTransform(a []complex128) []complex128 {
	if len(a) != FourierSize {
		panic("length of the input must be exactly FourierSize for the fixed size transform")
	}
	result := make([]complex128, FourierSize)
	fastFourierTransform(result, a)
	return result
}

// fourierTables holds everything about a fixed size transform that only depends
// on the size, which means all of it can be worked out once and then reused
// forever.
type fourierTables struct {
	// twiddles are the complex roots of unity for every radix-2 stage, packed
	// end to end smallest stage first. A stage whose half width is h reads h
	// entries starting at entry h-4, which works because the halves double
	// every stage and 4 + 8 + ... + h/2 comes out to exactly h-4.
	twiddles []complex128
	// scatter is the bit reversal permutation, stored as the index in the
	// output that each group of four results from the first pass belongs at.
	scatter []int
}

// fixedFourierTables are only ever built if something actually asks for them.
// A host running one of the assembly implementations gets its tables out of
// read only data in a generated .s file instead, so nothing is computed at
// runtime there and this never gets called.
//
// Those generated tables are not laid out the same way as these. The assembly
// runs the stages in radix-4 pairs, so its table is grouped per pass as h
// copies of W(2h)^j followed by h copies of W(4h)^j, and its permutation is
// stored as byte offsets. These are grouped per radix-2 stage and the
// permutation is stored as element indices. The two describe the same
// transform, but do not expect one to be checkable against the other.
var fixedFourierTables = sync.OnceValue(func() fourierTables {
	tables := fourierTables{
		twiddles: make([]complex128, 0, FourierSize),
		scatter:  make([]int, FourierSize/4),
	}

	for half := 4; half <= FourierSize/2; half <<= 1 {
		length := half * 2
		for j := 0; j < half; j++ {
			sin, cos := math.Sincos(-2 * math.Pi * float64(j) / float64(length))
			tables.twiddles = append(tables.twiddles, complex(cos, sin))
		}
	}

	// The first pass produces four outputs at a time and the group they land in
	// is the bit reversal of the group they came from.
	width := bits.Len(uint(FourierSize/4)) - 1
	for q := range tables.scatter {
		tables.scatter[q] = int(bits.Reverse64(uint64(q))>>(64-width)) * 4
	}

	return tables
})

// fastFourierTransformGo is the plain Go version of the fixed size transform.
// It is the fallback for hosts without the right SIMD instructions, and it is
// the readable statement of what the assembly is doing.
//
// It is kept in the radix-2 form on purpose. The assembly folds every pair of
// stages into a radix-4 pass, which is worth a good deal of speed but makes the
// arithmetic much harder to follow, and none of that helps a reader trying to
// understand the algorithm or a host that is running this path because it has
// no AVX512. The two agree to within a few units in the last place, which is
// all a different association order costs, and there is a test that says so.
func fastFourierTransformGo(dst, src []complex128) {
	tables := fixedFourierTables()

	// The first pass does the bit reversal permutation and the first two radix-2
	// stages in one go, as a single radix-4 butterfly. Those two stages only ever
	// multiply by 1 or by -i so there is nothing to look up, and permuting on the
	// way out of src means no scratch buffer is needed.
	const quarter = FourierSize / 4
	for q := range quarter {
		a0 := src[q]
		a1 := src[q+FourierSize/2]
		a2 := src[q+quarter]
		a3 := src[q+quarter*3]

		b0, b1 := a0+a1, a0-a1
		b2, b3 := a2+a3, a2-a3

		// Multiplying a complex number by -i is (x + iy) * -i = y - ix, so it is
		// a swap of the two halves and a sign flip rather than a multiply.
		b3 = complex(imag(b3), -real(b3))

		group := tables.scatter[q]
		dst[group+0] = b0 + b2
		dst[group+1] = b1 + b3
		dst[group+2] = b0 - b2
		dst[group+3] = b1 - b3
	}

	// Then every remaining radix-2 stage in place, from eight points wide all
	// the way up to the full transform.
	for half := 4; half <= FourierSize/2; half <<= 1 {
		twiddles := tables.twiddles[half-4 : half-4+half]
		for block := 0; block < FourierSize; block += half * 2 {
			for j := 0; j < half; j++ {
				u := dst[block+j]
				t := twiddles[j] * dst[block+j+half]
				dst[block+j] = u + t
				dst[block+j+half] = u - t
			}
		}
	}
}
