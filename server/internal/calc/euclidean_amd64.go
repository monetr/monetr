//go:build amd64 && !nosimd

package calc

//go:noescape
func __euclideanDistance64(a, b []float64) float64

//go:noescape
func __euclideanDistance64_AVX512(a, b []float64) float64

func init() {
	euclideanImplementation64 = __euclideanDistance64
}
