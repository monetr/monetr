//go:build amd64 && !nosimd

package calc

//go:noescape
func __sparseDot32_AVX512(dense []float32, indices []int32, values []float32) float32

func init() {
	// There is only an AVX512 version of this. The whole thing is built around a
	// gather, and the AVX2 gather is slow enough that for the few indicies a
	// transaction actually has the plain Go loop wins anyway. AVX512 is where
	// the gather starts being worth it
	if HasAVX512() {
		sparseDotImplementation32 = __sparseDot32_AVX512
	}
}
