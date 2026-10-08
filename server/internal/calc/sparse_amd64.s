//go:build amd64 && !nosimd

#include "textflag.h"

// sparseTailMask is for whatever is left over after LOOP has done every full
// group of 16. Entry N has its lowest N bits set, so if there are 5 left we
// load entry 5, which is 0x001F (0000000000011111), and only lanes 0 through 4
// get touched. Masked off lanes don't fault either, so this is also what keeps
// us from reading past the end of indices and values
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
//
// This is the dot product of dense and a sparse vector, but only at the spots
// in indices. The plain Go version is sparseDot32Go in sparse.go:
//
//   for i, index := range indices {
//     dot += dense[index] * values[i]
//   }
//
// We do 16 entries at a time. The trick is VGATHERDPS, it takes 16 indicies
// and loads dense[index] for each of them into one ZMM register. After that it
// is just a multiply and add with 16 values, same as any other dot product
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VFMADD231PS Z2, Z3, Z0 is Z0 = (Z3 * Z2) + Z0
TEXT ·__sparseDot32_AVX512(SB), NOSPLIT, $0-76
  MOVQ dense_base+0(FP),    AX // Load the pointer of dense into AX
  MOVQ indices_base+24(FP), BX // Load the pointer of indices into BX
  MOVQ indices_len+32(FP),  DX // Load the length of indices into DX. This is how many entries we have to do
  MOVQ values_base+48(FP),  CX // Load the pointer of values into CX. It is the same length as indices

  VXORPS Z0, Z0, Z0 // Zero out Z0, this is our accumulator for the dot product

  CMPQ DX, $16 // Do we have at least 16 entries?
  JL   TAIL    // If we don't then there isn't a full group of 16, jump straight to TAIL

  LOOP:
    VMOVDQU32   0(BX),           Z1 // Load the next 16 indicies into Z1
    KXNORW      K1,          K1, K1 // XNOR K1 with itself, this sets all 16 bits. The gather clears K1 when it is done so we have to do this every loop
    VXORPS      Z2,          Z2, Z2 // Zero out Z2 so the gather doesn't have to wait on whatever the last loop left in there
    VGATHERDPS  0(AX)(Z1*4), K1, Z2 // For each lane load the float32 at AX + index*4, so Z2 is now dense[index] for all 16
    VMOVUPS     0(CX),           Z3 // Load the 16 values that go with those indicies into Z3
    VFMADD231PS Z2,          Z3, Z0 // Z0 = (Z3 * Z2) + Z0, multiply each dense value by its value and add it into the accumulator

    ADDQ $64, BX   // Add 64 (16 * 4) to BX. This moves indices forward by 16 int32s
    ADDQ $64, CX   // Add 64 (16 * 4) to CX. This moves values forward by 16 float32s
    SUBQ $16, DX   // Subtract 16 from DX since we just did 16 entries
    CMPQ DX,  $16  // Are there still at least 16 left?
    JGE  LOOP      // If there are then jump back to LOOP, otherwise fall through to TAIL

  // TAIL does whatever is left, which is somewhere from 0 to 15 entries. It is
  // the same as LOOP but with a mask from sparseTailMask so we only touch the
  // lanes that are really there. Everything past the end gets loaded as 0
  TAIL:
    TESTQ DX, DX // Is DX zero?
    JZ    REDUCE // If it is then there is nothing left, jump straight to REDUCE

    LEAQ    sparseTailMask(SB), SI // Point SI at the start of the tail mask table
    MOVWLZX (SI)(DX*2),         R8 // Load entry DX into R8. Each entry is 2 bytes so it is at SI + DX*2
    KMOVW   R8,                 K1 // Copy the mask into K1, this is for the two loads
    KMOVW   R8,                 K2 // Copy the mask into K2 for the gather. The gather clears its mask so it can't share K1

    VMOVDQU32.Z 0(BX),       K1, Z1 // Load only the indicies that are left into Z1, the rest of the lanes get zeroed
    VXORPS      Z2,          Z2, Z2 // Zero out Z2. The gather won't write the lanes past the end, and if one had a NaN in it then NaN * 0 is still NaN
    VGATHERDPS  0(AX)(Z1*4), K2, Z2 // Same gather as LOOP, but only for the lanes set in K2
    VMOVUPS.Z   0(CX),       K1, Z3 // Load only the values that are left into Z3, the rest of the lanes get zeroed
    VFMADD231PS Z2,          Z3, Z0 // Z0 = (Z3 * Z2) + Z0, the lanes past the end are 0 * 0 so they don't change anything

  // Z0 has 16 partial sums now and we need to add them all up into 1. Call the
  // lanes of Z0 s0 through s15. This is the same reduce the euclidean distance
  // code uses, README.md has tables that walk through it
  //
  // The extract has to happen first. Any VEX instruction that writes Y0 (like
  // VHADDPS) zeroes the top half of Z0, so s8 through s15 would be gone
  REDUCE:
    VEXTRACTF32X8 $1, Z0, Y1     // Y1 = [s8 ... s15], the high 256 bits of Z0
    VHADDPS       Y0, Y0, Y0     // Add neighboring pairs, the low 128 bits of Y0 start with s0+s1, s2+s3 and the high 128 bits start with s4+s5, s6+s7
    VPERM2F128    $1, Y0, Y0, Y2 // Swap the 128 bit halves of Y0 into Y2, so X2 starts with s4+s5, s6+s7
    VHADDPS       Y1, Y1, Y1     // Same thing for Y1, so X1 starts with s8+s9, s10+s11
    VPERM2F128    $1, Y1, Y1, Y3 // Swap the halves of Y1 into Y3, so X3 starts with s12+s13, s14+s15
    VADDPS        X0, X1, X0     // X0 = X1 + X0
    VADDPS        X0, X2, X0     // X0 = X2 + X0
    VADDPS        X0, X3, X0     // X0 = X3 + X0. Now lanes 0 and 1 of X0 each have half of the 16 added up
    VHADDPS       X0, X0, X0     // Add lanes 0 and 1 together, lane 0 of X0 is now the whole dot product
    MOVL          X0, ret+72(FP) // Store lane 0 of X0 as the return value

    // We have to clear the top halves of the vector registers before going
    // back to Go. Right after this returns Go zeroes X15 with XORPS, which is
    // an old SSE instruction, and on Intel running SSE instructions while the
    // top halves are dirty is slow
    VZEROUPPER // Zero the upper bits of the vector registers
    RET        // We are done, return
