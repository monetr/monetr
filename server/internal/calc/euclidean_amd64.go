//go:build amd64 && !nosimd

package calc

//go:noescape
func __euclideanDistance64_AVX(a, b []float64) float64

//go:noescape
func __euclideanDistance64_AVX_FMA(a, b []float64) float64

//go:noescape
func __euclideanDistance32_AVX(a, b []float32) float32

//go:noescape
func __euclideanDistance32_AVX_FMA(a, b []float32) float32

//go:noescape
func __euclideanDistance64_AVX512(a, b []float64) float64

//go:noescape
func __euclideanDistance64_AVX512_FMA(a, b []float64) float64

//go:noescape
func __euclideanDistance32_AVX512(a, b []float32) float32

//go:noescape
func __euclideanDistance32_AVX512_FMA(a, b []float32) float32

func init() {
	// If the CPU is capable of AVX512 floating point instructions AND
	// fused-multiply-add instructions then prefer those. Same for the other
	// branches. Prefer fused-multiply-add first then fallback to regular
	// instructions.
	if HasAVX512FMA() {
		euclideanImplementation64 = __euclideanDistance64_AVX512_FMA
		euclideanImplementation32 = __euclideanDistance32_AVX512_FMA
	} else if HasAVX512() {
		euclideanImplementation64 = __euclideanDistance64_AVX512
		euclideanImplementation32 = __euclideanDistance32_AVX512
	} else if HasAVXFMA() {
		euclideanImplementation64 = __euclideanDistance64_AVX_FMA
		euclideanImplementation32 = __euclideanDistance32_AVX_FMA
	} else if HasAVX() {
		euclideanImplementation64 = __euclideanDistance64_AVX
		euclideanImplementation32 = __euclideanDistance32_AVX
	}
}
