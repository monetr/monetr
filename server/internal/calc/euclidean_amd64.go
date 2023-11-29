//go:build amd64 && !nosimd

package calc

//go:noescape
func __euclideanDistance(a, b []float64) float64

func init() {
	euclideanImplementation = __euclideanDistance
}
