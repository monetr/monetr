//go:build amd64 && !nosimd

#include "textflag.h"

// sparseTailMask is indexed by the number of entries left over after the main
// loop has consumed every full group of 16. Entry N has its lowest N bits set,
// which is exactly the AVX512 mask needed to process N lanes and ignore the
// rest. Masked loads suppress faults on the lanes that are masked off, so this
// is also what keeps the kernel from reading past the end of the slices.
DATA  sparseTailMask+0(SB)/2,  $0x0000
DATA  sparseTailMask+2(SB)/2,  $0x0001
DATA  sparseTailMask+4(SB)/2,  $0x0003
DATA  sparseTailMask+6(SB)/2,  $0x0007
DATA  sparseTailMask+8(SB)/2,  $0x000F
DATA  sparseTailMask+10(SB)/2, $0x001F
DATA  sparseTailMask+12(SB)/2, $0x003F
DATA  sparseTailMask+14(SB)/2, $0x007F
DATA  sparseTailMask+16(SB)/2, $0x00FF
DATA  sparseTailMask+18(SB)/2, $0x01FF
DATA  sparseTailMask+20(SB)/2, $0x03FF
DATA  sparseTailMask+22(SB)/2, $0x07FF
DATA  sparseTailMask+24(SB)/2, $0x0FFF
DATA  sparseTailMask+26(SB)/2, $0x1FFF
DATA  sparseTailMask+28(SB)/2, $0x3FFF
DATA  sparseTailMask+30(SB)/2, $0x7FFF
DATA  sparseTailMask+32(SB)/2, $0xFFFF
GLOBL sparseTailMask(SB), RODATA|NOPTR, $34

// func __sparseDot32_AVX512(dense []float32, indices []int32, values []float32) float32
TEXT ·__sparseDot32_AVX512(SB), NOSPLIT, $0-76
  MOVQ dense_base+0(FP), AX     // Load the pointer of the dense vector we index into.
  MOVQ indices_base+24(FP), BX  // Load the pointer of the sparse index array.
  MOVQ indices_len+32(FP), DX   // Load the number of non-zero entries into DX.
  MOVQ values_base+48(FP), CX   // Load the pointer of the sparse value array.

  VXORPS Z0, Z0, Z0             // Clear the accumulator that holds the running dot product.

  CMPQ DX, $16                  // A transaction usually occupies fewer than 16 indicies,
  JL   TAIL                     // so the common case skips the main loop entirely.

  LOOP:
    VMOVDQU32   0(BX), Z1          // Load the next 16 indicies into ZMM1.
    KXNORW      K1, K1, K1         // Set every bit of K1 so the gather reads all 16 lanes.
    VXORPS      Z2, Z2, Z2         // A gather leaves masked off lanes untouched, so start from zero.
    VGATHERDPS  0(AX)(Z1*4), K1, Z2 // Pull the dense value sitting at each of those 16 indicies.
    VMOVUPS     0(CX), Z3          // Load the 16 sparse values that pair with those indicies.
    VFMADD231PS Z2, Z3, Z0         // Multiply the two sets together and add them into the accumulator.

    ADDQ $64, BX                   // Move the index pointer forward by 16 int32s.
    ADDQ $64, CX                   // Move the value pointer forward by 16 float32s.
    SUBQ $16, DX                   // We just consumed 16 of the entries.
    CMPQ DX, $16                   // If there are still at least 16 left then go
    JGE  LOOP                      // around again, otherwise fall into the tail.

  TAIL:
    TESTQ DX, DX                   // If nothing is left over then there is no tail
    JZ    REDUCE                   // to process and we can go straight to the reduce.

    LEAQ  sparseTailMask(SB), SI   // Look up the mask for however many entries remain.
    MOVWLZX (SI)(DX*2), R8         // Entry DX has its lowest DX bits set.
    KMOVW R8, K1                   // K1 masks the two loads.
    KMOVW R8, K2                   // K2 masks the gather. It is consumed by the instruction.

    VMOVDQU32.Z 0(BX), K1, Z1      // Load only the remaining indicies, zeroing the lanes past the end.
    VXORPS      Z2, Z2, Z2         // Again, the gather only writes the lanes it is told to.
    VGATHERDPS  0(AX)(Z1*4), K2, Z2 // Gather only the lanes that are really there.
    VMOVUPS.Z   0(CX), K1, Z3      // Load only the remaining values, zeroing the rest.
    VFMADD231PS Z2, Z3, Z0         // The lanes past the end are 0 * 0, so they add nothing.

  REDUCE:
    // Collapse the 16 lanes of ZMM0 down into a single float32. This is the same
    // sequence the euclidean distance kernels use, see README.md for the tables
    // that walk through what each step is doing.
    VEXTRACTF32X8 $1, Z0, Y1       // Extract the high 256-bits of ZMM0 into YMM1.
    VHADDPS       Y0, Y0, Y0       // Horizontal add on YMM0 (the low 256-bits of ZMM0).
    VPERM2F128    $1, Y0, Y0, Y2   // Extract the high 128 bits of YMM0 into YMM2.
    VHADDPS       Y1, Y1, Y1       // Horizontal add on YMM1 (from the first extract).
    VPERM2F128    $1, Y1, Y1, Y3   // Extract the high 128 bits of YMM1 into YMM3.
    VADDPS        X0, X1, X0       // XMM0 += XMM1, only the low 32 bits matter from here.
    VADDPS        X0, X2, X0       // XMM0 += XMM2
    VADDPS        X0, X3, X0       // XMM0 += XMM3
    VHADDPS       X0, X0, X0       // Add the 32-bit pairs so every lane holds the total.
    MOVL          X0, ret+72(FP)   // Move the low 32 bits out into the return address space.
    // Clear the upper bits of the vector registers. The go compiler only emits
    // legacy SSE, and returning with those bits still dirty makes that SSE code
    // pay a transition penalty on intel hardware.
    VZEROUPPER
    RET                            // Return
