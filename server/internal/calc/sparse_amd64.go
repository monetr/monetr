//go:build amd64 && !nosimd

package calc

//go:noescape
func __sparseDot32_AVX512(dense []float32, indices []int32, values []float32) float32

func init() {
	// There is deliberately no AVX or AVX2 tier for the sparse dot product. The
	// kernel is built around a gather, and the AVX2 gather instructions are slow
	// enough that for the handful of indicies a transaction vector actually
	// occupies the pure go loop is the better option. AVX512F is the first
	// instruction set where the gather is worth reaching for.
	if HasAVX512() {
		sparseDotImplementation32 = __sparseDot32_AVX512
	}
}
