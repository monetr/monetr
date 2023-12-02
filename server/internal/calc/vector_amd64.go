//go:build amd64 && !nosimd

package calc

import "github.com/klauspost/cpuid/v2"

//go:noescape
func __normalizeVector64_AVX(input []float64)

func init() {
	if cpuid.CPU.Supports(cpuid.AVX) {
		normalizeVectorImplementation64 = __normalizeVector64_AVX
	}
}
