//go:build amd64 && !nosimd

package calc

//go:noescape
func __fastFourierTransform_AVX512(dst, src []complex128)

func init() {
	// Everything the transform uses (VFMADDSUB213PD, VSHUFF64X2, VPERMILPD,
	// VMOVDDUP and VPXORQ on ZMM registers) is part of the AVX512 foundation
	// instruction set, so plain AVX512F support is all that is needed here.
	if HasAVX512() {
		fastFourierTransform = __fastFourierTransform_AVX512
	}
}
