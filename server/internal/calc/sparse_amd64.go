//go:build amd64 && !nosimd

package calc

//go:noescape
func __sparseDot32Scalar_AVX(dense *float32, denseLength int, indices *int32, values *float32, count int) float32

func init() {
	// Any CPU with AVX gets the assembly version. It only needs AVX so it runs on
	// Ivy Bridge too, anything older stays on sparseDot32Go
	if HasAVX() {
		sparseDotImplementation32 = __sparseDot32Scalar_AVX
	}
}
