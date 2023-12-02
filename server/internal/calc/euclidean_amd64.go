//go:build amd64 && !nosimd

package calc

import "github.com/klauspost/cpuid/v2"

//go:noescape
func __euclideanDistance64(a, b []float64) float64

//go:noescape
func __euclideanDistance64_AVX512(a, b []float64) float64

func init() {
	if cpuid.CPU.Supports(cpuid.AVX512F) {
		euclideanImplementation64 = __euclideanDistance64_AVX512
	} else if cpuid.CPU.Supports(cpuid.AVX) {
		euclideanImplementation64 = __euclideanDistance64
	}
}
