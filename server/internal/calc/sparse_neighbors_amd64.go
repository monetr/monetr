//go:build amd64 && !nosimd

package calc

import "golang.org/x/sys/cpu"

//go:noescape
func __sparseNeighbors32_AVX(
	dense []float32,
	signature uint64,
	norm2, epsilon float32,
	signatures []uint64,
	norms []float32,
	offsets []int32,
	indices []int32,
	values []float32,
	output []int32,
) int

//go:noescape
func __sparseNeighbors32_AVX512VL(
	dense []float32,
	signature uint64,
	norm2, epsilon float32,
	signatures []uint64,
	norms []float32,
	offsets []int32,
	indices []int32,
	values []float32,
	output []int32,
) int

//go:noescape
func __sparseNeighbors32_AVX512(
	dense []float32,
	signature uint64,
	norm2, epsilon float32,
	signatures []uint64,
	norms []float32,
	offsets []int32,
	indices []int32,
	values []float32,
	output []int32,
) int

func init() {
	// There isn't a feature flag for "512 bit registers don't slow the core
	// down", but AVX512VBMI2 is close. Skylake and Cascade Lake Xeons don't have
	// it and they are the ones that slow down. Ice Lake and newer, and Zen 4 and
	// newer all have it and slow down a lot less or not at all
	if HasAVX512() && cpu.X86.HasAVX512VBMI2 {
		sparseNeighborsImplementation32 = __sparseNeighbors32_AVX512
	} else if HasAVX512VL() {
		sparseNeighborsImplementation32 = __sparseNeighbors32_AVX512VL
	} else if HasAVX() {
		sparseNeighborsImplementation32 = __sparseNeighbors32_AVX
	}
}
