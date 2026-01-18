//go:build amd64 && !nosimd

package calc

//go:noescape
func __normalizeVector64_AVX(input []float64)

//go:noescape
func __normalizeVector64_AVX_FMA(input []float64)

//go:noescape
func __normalizeVector64_AVX512(input []float64)

//go:noescape
func __normalizeVector64_AVX512_FMA(input []float64)

//go:noescape
func __normalizeVector32_AVX(input []float32)

//go:noescape
func __normalizeVector32_AVX_FMA(input []float32)

//go:noescape
func __normalizeVector32_AVX512(input []float32)

//go:noescape
func __normalizeVector32_AVX512_FMA(input []float32)

func init() {
	if HasAVX512FMA() {
		normalizeVectorImplementation64 = __normalizeVector64_AVX512_FMA
		normalizeVectorImplementation32 = __normalizeVector32_AVX512_FMA
	} else if HasAVX512() {
		normalizeVectorImplementation64 = __normalizeVector64_AVX512
		normalizeVectorImplementation32 = __normalizeVector32_AVX512
	} else if HasAVXFMA() {
		normalizeVectorImplementation64 = __normalizeVector64_AVX_FMA
		normalizeVectorImplementation32 = __normalizeVector32_AVX_FMA
	} else if HasAVX() {
		normalizeVectorImplementation64 = __normalizeVector64_AVX
		normalizeVectorImplementation32 = __normalizeVector32_AVX
	}
}
