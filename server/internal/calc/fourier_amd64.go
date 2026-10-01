//go:build amd64 && !nosimd

package calc

//go:noescape
func __fastFourierTransform_AVX512(dst, src []complex128)

//go:noescape
func __fastFourierTransform_AVX_FMA(dst, src []complex128)

//go:noescape
func __fastFourierTransform_AVX(dst, src []complex128)

//go:noescape
func __inverseFastFourierTransform_AVX512(dst, src []complex128)

//go:noescape
func __inverseFastFourierTransform_AVX_FMA(dst, src []complex128)

//go:noescape
func __inverseFastFourierTransform_AVX(dst, src []complex128)

func init() {
	// The AVX512 versions only use AVX512F instructions, and the 256 bit
	// versions only need AVX, which is why they use VXORPD rather than the AVX2
	// VPXOR.
	switch {
	case HasAVX512():
		fastFourierTransform = __fastFourierTransform_AVX512
		inverseFastFourierTransform = __inverseFastFourierTransform_AVX512
	case HasAVXFMA():
		fastFourierTransform = __fastFourierTransform_AVX_FMA
		inverseFastFourierTransform = __inverseFastFourierTransform_AVX_FMA
	case HasAVX():
		fastFourierTransform = __fastFourierTransform_AVX
		inverseFastFourierTransform = __inverseFastFourierTransform_AVX
	}
}
