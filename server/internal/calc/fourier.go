package calc

import (
	"math"
	"math/cmplx"
)

const FourierSize = 4096

// FastFourierTransform is a recursive implementation of the fast Fourier
// transform.
func FastFourierTransform(a []complex128) []complex128 {
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

	fftEven := FastFourierTransform(even)
	fftOdd := FastFourierTransform(odd)

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
	fftConjugated := FastFourierTransform(conjugated)

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
