//go:build amd64 && !nosimd

package calc

//go:noescape
func __sparseDot32_AVX512(dense []float32, indices []int32, values []float32) float32

//go:noescape
func __sparseDot32_AVX_FMA(dense []float32, indices []int32, values []float32) float32

//go:noescape
func __sparseDot32_AVX(dense []float32, indices []int32, values []float32) float32

// TODO SparseDot32 doesn't use this yet. Calling it from inside SparseDot32 is
// a second function call and that costs more than this saves (about 0.7ns on a
// 7950X). It only wins when the caller calls it directly, which means
// SparseDot32 needs to be small enough for Go to inline
//
//go:noescape
func __sparseDot32Scalar_AVX_FMA(dense *float32, indices *int32, values *float32, count int) float32

func init() {
	// The AVX512 version is built around VGATHERDPS. The AVX_FMA and AVX
	// versions only need AVX (and FMA), same as the fourier and euclidean ones,
	// so they can't use the gather since that is AVX2. They load each value
	// themselves instead, see the notes in sparse_amd64.s
	//
	// On a 7950X the AVX_FMA version is actually faster than the AVX512 one
	// because Zen 4's gather is slow. Skylake-SP's gather is a lot faster than
	// doing the loads ourselves though, so AVX512 still goes first
	switch {
	case HasAVX512():
		sparseDotImplementation32 = __sparseDot32_AVX512
	case HasAVXFMA():
		sparseDotImplementation32 = __sparseDot32_AVX_FMA
	case HasAVX():
		sparseDotImplementation32 = __sparseDot32_AVX
	}
}
