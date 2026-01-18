package calc

import "golang.org/x/sys/cpu"

// HasAVX512FMA returns true if the current CPU on the host system has both the
// AVX512F and the AVX512VL instruction sets available. To debug this you can
// run the following command on unix systems: `lscpu | grep avx512`.
// This means that the fused-multply-add instructions are available for the
// operations we want to perform. Other CPUs may designate the fused-multply-add
// instructions differently, these functions should be used to reduce any
// confusion around what is possible on the host system.
func HasAVX512FMA() bool {
	return cpu.X86.HasAVX512F && cpu.X86.HasAVX512VL
}

// HasAVX512 returns true of the CPU supports the most bassic AVX512 instruction
// sets that monetr can take advantage of. This function can return true even
// when HasAVX512FMA is also true as this function indicates a subset of
// supported instructions.
func HasAVX512() bool {
	return cpu.X86.HasAVX512F
}

func HasAVXFMA() bool {
	return cpu.X86.HasAVX && cpu.X86.HasFMA
}

func HasAVX() bool {
	return cpu.X86.HasAVX
}
