//go:build amd64 && !nosimd

package calc

//go:noescape
func __fastFourierTransform_AVX512(dst, src []complex128)

//go:noescape
func __fastFourierTransform_AVX_FMA(dst, src []complex128)

//go:noescape
func __fastFourierTransform_AVX(dst, src []complex128)

func init() {
	// Prefer the widest instruction set the host can actually run, the same way
	// the euclidean distance and vector normalization implementations do.
	//
	// Everything the AVX512 version uses is part of the AVX512 foundation
	// instruction set, so plain AVX512F support is enough for it. The two 256
	// bit versions need nothing newer than AVX, which is why the sign flips in
	// them go through VXORPD rather than the AVX2 only VPXOR.
	switch {
	case HasAVX512():
		fastFourierTransform = __fastFourierTransform_AVX512
	case HasAVXFMA():
		fastFourierTransform = __fastFourierTransform_AVX_FMA
	case HasAVX():
		fastFourierTransform = __fastFourierTransform_AVX
	}
}
