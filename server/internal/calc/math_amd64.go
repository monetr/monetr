//go:build amd64 && !nosimd

package calc

//go:noescape
func __cosine64(input float64) float64

//go:noescape
func __absFloat64(input float64) float64

func init() {
	cosineImplementation = __cosine64
}

