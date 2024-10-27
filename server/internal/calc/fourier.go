package calc

import (
	"math"
	"math/cmplx"
)

// FFT function (recursive implementation)
func fft(a []complex128) []complex128 {
	n := len(a)
	if n <= 1 {
		return a
	}

	// Evil LEAQ hacks to do math on values in a single instruction lmao. BX is
	// `n`. So doing `-1(BX)` gives us n-1 in DX. Then doing TESTQ gives us a
	// bitwise logical AND of the value of BX and DX. Then do a JNE so if TESTQ is
	// not 0 then we panic.
	// LEAQ    -1(BX), DX
	// TESTQ   BX, DX
	// JNE     panic.boi
	// Ensure n is a power of two
	if (n & (n - 1)) != 0 {
		panic("Length of input must be a power of two")
	}

	// If n is in the CX register then
	// SHRQ    $1, CX
	// Is the same as n/2, SHRQ shifts the bits left by $1
	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		// 		 MOVQ    main..autotmp_37+72(SP), DX
		//     MOVQ    main.a+96(SP), SI
		//     MOVQ    main.a+104(SP), CX
		//     MOVQ    main..autotmp_36+48(SP), DI
		//     XORL    BX, BX
		//     JMP     main_fft_pc167
		// main_fft_pc167:
		//     CMPQ    BX, DI
		//     JGE     main_fft_pc245
		//     MOVQ    BX, R8
		//     SHLQ    $1, BX
		//     CMPQ    CX, BX
		//     JLS     main_fft_pc637
		//     MOVQ    R8, R9
		//     SHLQ    $5, R8 // x32 because 16*2 = 32 **MATH** :crying:
		//     MOVSD   (SI)(R8*1), X0
		//     MOVSD   8(SI)(R8*1), X1
		even[i] = a[i*2]
		odd[i] = a[i*2+1]
	}

	fftEven := fft(even)
	fftOdd := fft(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		// -2 * math.Pi comes out to:
		// MOVSD   $f64.c01921fb54442d18(SB), X1 = -6.283185307179586
		t := cmplxExp(-2*math.Pi*float64(k)/float64(n)) * fftOdd[k]
		result[k] = fftEven[k] + t
		result[k+n/2] = fftEven[k] - t
	}
	return result
}

// IFFT function (recursive implementation)
func ifft(a []complex128) []complex128 {
	n := len(a)

	// Conjugate the input
	conjugated := make([]complex128, n)
	for i := range a {
		conjugated[i] = cmplx.Conj(a[i])
	}

	// Apply FFT to the conjugated input
	fftConjugated := fft(conjugated)

	// Conjugate the result and scale by 1/n
	for i := range fftConjugated {
		fftConjugated[i] = cmplx.Conj(fftConjugated[i]) / complex(float64(n), 0)
	}

	return fftConjugated
}

// Compute complex exponential (Euler's formula)
func cmplxExp(theta float64) complex128 {
	return complex(math.Cos(theta), math.Sin(theta))
}

// Magnitude helper function
func cmplxAbs(c complex128) float64 {
	return cmplx.Abs(c)
}
