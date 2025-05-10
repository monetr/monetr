//go:build amd64 && !nosimd

package calc

import (
	"golang.org/x/sys/cpu"
)

//go:noescape
func __euclideanDistance64_AVX(a, b []float64) float64

//go:noescape
func __euclideanDistance32_AVX(a, b []float32) float32

//go:noescape
func __euclideanDistance64_AVX512(a, b []float64) float64

//go:noescape
func __euclideanDistance32_AVX512(a, b []float32) float32

func init() {
	if cpu.X86.HasAVX512F {
		euclideanImplementation64 = __euclideanDistance64_AVX512
		euclideanImplementation32 = __euclideanDistance32_AVX512
	} else if cpu.X86.HasAVX {
		// TODO Technically these rely on both AVX and FMA instructions so there
		// might need to be additional checks here?
		euclideanImplementation64 = __euclideanDistance64_AVX
		euclideanImplementation32 = __euclideanDistance32_AVX
	}
}
