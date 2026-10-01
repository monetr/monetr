package calc

import (
	"math"
	"math/bits"
	"math/cmplx"
	"sync"
)

//go:generate go run ./gen -size 4096 -output fourier_twiddles_amd64.s

const FourierSize = 4096

// The assembly reads tables generated for exactly 4096 points but takes the
// number of points from the length of the slice, so FourierSize has to match
// the tables. A negative constant does not convert to uint, so these pin
// FourierSize to 4096 at compile time. Changing it means changing the generate
// directive, and the symbol names and the 1/4096 constants in fourier_amd64.s
// too.
const (
	_ = uint(FourierSize - 4096)
	_ = uint(4096 - FourierSize)
)

// FastFourierTransformSlow is a recursive implementation of the fast Fourier
// transform.
// Deprecated: Use [FastFourierTransform] instead.
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

// InverseFastFourierTransformSlow is a recursive implementation of the inverse
// fast Fourier transform.
//
// Deprecated: Use [InverseFastFourierTransform] instead.
func InverseFastFourierTransformSlow(a []complex128) []complex128 {
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

// fastFourierTransform is swapped out for an assembly implementation by init()
// on platforms that have one.
var fastFourierTransform func(dst, src []complex128) = fastFourierTransformGo

// FastFourierTransform is a non-recursive forward fast Fourier transform for
// exactly FourierSize points. Fixing the size lets the twiddle factors and bit
// reversal offsets be computed ahead of time.
//
// On amd64 this uses AVX512, AVX with FMA, or plain AVX assembly depending on
// the CPU. Everything else falls back to fastFourierTransformGo.
func FastFourierTransform(a []complex128) []complex128 {
	if len(a) != FourierSize {
		panic("length of the input must be exactly FourierSize for the fixed size transform")
	}
	result := make([]complex128, FourierSize)
	fastFourierTransform(result, a)
	return result
}

// inverseFastFourierTransform is swapped out for an assembly implementation by
// init() on platforms that have one.
var inverseFastFourierTransform func(dst, src []complex128) = inverseFastFourierTransformGo

// InverseFastFourierTransform is a non-recursive inverse fast Fourier transform
// for exactly FourierSize points. It includes the 1/n scaling, so it gives back
// the series that was passed to FastFourierTransform.
//
// On amd64 this uses AVX512, AVX with FMA, or plain AVX assembly depending on
// the CPU. Everything else falls back to inverseFastFourierTransformGo.
func InverseFastFourierTransform(a []complex128) []complex128 {
	if len(a) != FourierSize {
		panic("length of the input must be exactly FourierSize for the fixed size transform")
	}
	result := make([]complex128, FourierSize)
	inverseFastFourierTransform(result, a)
	return result
}

// fourierTables holds everything about the fixed size transform that only
// depends on the size.
type fourierTables struct {
	// twiddles are the complex roots of unity for every radix-2 stage, packed end
	// to end smallest stage first. A stage whose half width is h reads h entries
	// starting at entry h-4, which works because the halves double every stage
	// and 4 + 8 + ... + h/2 comes out to exactly h-4.
	twiddles []complex128
	// scatter is the bit reversal permutation, stored as the index in the output
	// that each group of four results from the first pass belongs at.
	scatter []int
}

// fixedFourierTables are only built when the Go fallback runs. The assembly
// uses the generated tables in fourier_twiddles_amd64.s instead, which are laid
// out per radix-4 pass and store byte offsets, while these are laid out per
// radix-2 stage and store element indices.
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

// fastFourierTransformGo is the plain Go version of the fixed size transform,
// used on hosts without the right SIMD instructions. It stays radix-2 where the
// assembly folds pairs of stages into radix-4 passes, so the results differ by
// a few units in the last place.
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

	// Then every remaining radix-2 stage in place, from eight points wide all the
	// way up to the full transform.
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

// inverseFastFourierTransformGo is the plain Go version of the fixed size
// inverse transform, used on hosts without the right SIMD instructions. The
// inverse uses e^(+2*pi*i*j/N) where the forward transform uses
// e^(-2*pi*i*j/N), so this is fastFourierTransformGo with every twiddle factor
// conjugated and the input scaled by 1/n.
func inverseFastFourierTransformGo(dst, src []complex128) {
	tables := fixedFourierTables()

	// 1/n is a power of two, so scaling by it on the way in only changes the
	// exponent. As long as nothing in the transform gets small enough to go
	// subnormal, it comes out the same as scaling every output at the end without
	// needing another pass over dst.
	const scale = 1.0 / FourierSize
	const quarter = FourierSize / 4
	for q := range quarter {
		a0 := src[q] * scale
		a1 := src[q+FourierSize/2] * scale
		a2 := src[q+quarter] * scale
		a3 := src[q+quarter*3] * scale

		b0, b1 := a0+a1, a0-a1
		b2, b3 := a2+a3, a2-a3

		// The forward transform multiplies b3 by -i here and the inverse needs the
		// conjugate, +i. That is just -(-i * b3), so this makes -i * b3 the same
		// way and then swaps which of the outputs adds it and which subtracts it.
		b3 = complex(imag(b3), -real(b3))

		group := tables.scatter[q]
		dst[group+0] = b0 + b2
		dst[group+1] = b1 - b3
		dst[group+2] = b0 - b2
		dst[group+3] = b1 + b3
	}

	for half := 4; half <= FourierSize/2; half <<= 1 {
		twiddles := tables.twiddles[half-4 : half-4+half]
		for block := 0; block < FourierSize; block += half * 2 {
			for j := 0; j < half; j++ {
				u := dst[block+j]
				t := cmplx.Conj(twiddles[j]) * dst[block+j+half]
				dst[block+j] = u + t
				dst[block+j+half] = u - t
			}
		}
	}
}
