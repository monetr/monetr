//go:build amd64 && !nosimd

package calc

import "golang.org/x/sys/cpu"

//go:noescape
func __normalizeVector64_AVX(input []float64)

//go:noescape
func __normalizeVector64_AVX512(input []float64)

//go:noescape
func __normalizeVector32_AVX(input []float32)

//go:noescape
func __normalizeVector32_AVX512(input []float32)

func init() {
	if cpu.X86.HasAVX512F {
		normalizeVectorImplementation64 = __normalizeVector64_AVX512
		normalizeVectorImplementation32 = __normalizeVector32_AVX512
	} else if cpu.X86.HasAVX {
		normalizeVectorImplementation64 = __normalizeVector64_AVX
		normalizeVectorImplementation32 = __normalizeVector32_AVX
	}
}
