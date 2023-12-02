//go:build amd64 && !nosimd

package calc

//go:noescape
func __normalizeVector64_AVX(input []float64)
