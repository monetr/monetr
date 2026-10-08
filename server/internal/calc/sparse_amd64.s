//go:build amd64 && !nosimd

#include "textflag.h"

// func __sparseDot32_AVX512(dense []float32, indices []int32, values []float32) float32
//
// This is the dot product of dense and a sparse vector, but only at the spots
// in indices. The plain Go version is sparseDot32Go in sparse.go:
//
//   for i, index := range indices {
//     dot += dense[index] * values[i]
//   }
//
// We do it in 2 parts:
//
//   LOOP:   Does 16 entries at a time with VGATHERDPS, which takes 16 indicies
//           and loads dense[index] for each of them into one ZMM register.
//           After that it is just a multiply and add with 16 values, same as
//           any other dot product
//
//   QUAD:   Does whatever is left (0 to 15 entries) with plain scalar loads, 4
//           at a time and then SINGLE does the last 0 to 3 one at a time
//
// The gather is the expensive part. On Zen 4 one 16 lane gather is 81 uops and
// a new one can only start every 16.57 cycles, so about 1 cycle per lane. That
// is about the same as doing the loads ourselves, Zen 4 can do 3 loads a cycle
// and each entry needs 3 (the index, dense[index] and the value). Skylake-SP
// does the whole gather in 9.25 cycles, but it can only do 2 loads a cycle so
// doing the loads ourselves would take 24 cycles for the same 16 entries. So
// the gather is a tie on Zen 4 and a big win on Skylake-SP
//   https://uops.info/html-instr/VGATHERDPS_ZMM_K_VSIB_ZMM.html
//
// The left overs don't use a masked gather. On Zen 4 a masked gather costs the
// same as a full one no matter how many lanes are masked off, when this used a
// masked gather for the left overs 2 entries took the same 6.7ns as 8 did. So
// 17 entries was almost as slow as 32
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VFMADD231PS Z2, Z3, Z0 is Z0 = (Z3 * Z2) + Z0
TEXT ·__sparseDot32_AVX512(SB), NOSPLIT, $0-76
  MOVQ dense_base+0(FP),    AX // Load the pointer of dense into AX
  MOVQ indices_base+24(FP), BX // Load the pointer of indices into BX
  MOVQ indices_len+32(FP),  DX // Load the length of indices into DX. This is how many entries we have to do
  MOVQ values_base+48(FP),  CX // Load the pointer of values into CX. It is the same length as indices

  // Z0 is the accumulator for LOOP. X4 through X7 are the accumulators for QUAD
  // and SINGLE, see the notes above QUAD for why there are 4 of them
  VXORPS Z0, Z0, Z0 // Zero out Z0
  VXORPS X4, X4, X4 // Zero out X4
  VXORPS X5, X5, X5 // Zero out X5
  VXORPS X6, X6, X6 // Zero out X6
  VXORPS X7, X7, X7 // Zero out X7

  CMPQ DX, $16 // Do we have at least 16 entries?
  JL   QUAD    // If we don't then there isn't a full group of 16, jump straight to QUAD

  LOOP:
    VMOVDQU32   0(BX),           Z1 // Load the next 16 indicies into Z1
    KXNORW      K1,          K1, K1 // XNOR K1 with itself, this sets all 16 bits. The gather clears K1 when it is done so we have to do this every loop
    VXORPS      Z2,          Z2, Z2 // Zero out Z2 so the gather doesn't have to wait on whatever the last loop left in there
    VGATHERDPS  0(AX)(Z1*4), K1, Z2 // For each lane load the float32 at AX + index*4, so Z2 is now dense[index] for all 16
    VFMADD231PS 0(CX),       Z2, Z0 // Z0 = (values * Z2) + Z0. The 16 values get loaded straight from CX as part of the FMA

    ADDQ $64, BX   // Add 64 (16 * 4) to BX. This moves indices forward by 16 int32s
    ADDQ $64, CX   // Add 64 (16 * 4) to CX. This moves values forward by 16 float32s
    SUBQ $16, DX   // Subtract 16 from DX since we just did 16 entries
    CMPQ DX,  $16  // Are there still at least 16 left?
    JGE  LOOP      // If there are then jump back to LOOP, otherwise fall through to QUAD

  // QUAD does 4 entries at a time with plain scalar loads. LOOP already took
  // every full group of 16 so this only ever goes around 3 times at most
  //
  // Each of the 4 gets its own accumulator. A scalar FMA takes 4 cycles before
  // its result can be used by the next one (on both Zen 4 and Skylake-SP), so
  // if all 4 went into the same register then each one would sit and wait on
  // the one before it. With 4 registers they can all go at the same time
  //   https://uops.info/html-instr/VFMADD231SS_XMM_XMM_M32.html
  QUAD:
    CMPQ DX, $4 // Do we have at least 4 entries left?
    JL   SINGLE // If we don't then jump straight to SINGLE

    QUADLOOP:
      MOVLQSX 0(BX),  R8  // Load the first index into R8. MOVLQSX sign extends the int32, same as the gather does
      MOVLQSX 4(BX),  R9  // Load the second index into R9
      MOVLQSX 8(BX),  R10 // Load the third index into R10
      MOVLQSX 12(BX), R11 // Load the fourth index into R11

      VMOVSS (AX)(R8*4),  X8  // X8 = dense[R8]
      VMOVSS (AX)(R9*4),  X9  // X9 = dense[R9]
      VMOVSS (AX)(R10*4), X10 // X10 = dense[R10]
      VMOVSS (AX)(R11*4), X11 // X11 = dense[R11]

      VFMADD231SS 0(CX),  X8,  X4 // X4 = (values[0] * X8) + X4
      VFMADD231SS 4(CX),  X9,  X5 // X5 = (values[1] * X9) + X5
      VFMADD231SS 8(CX),  X10, X6 // X6 = (values[2] * X10) + X6
      VFMADD231SS 12(CX), X11, X7 // X7 = (values[3] * X11) + X7

      ADDQ $16, BX  // Add 16 (4 * 4) to BX. This moves indices forward by 4 int32s
      ADDQ $16, CX  // Add 16 (4 * 4) to CX. This moves values forward by 4 float32s
      SUBQ $4,  DX  // Subtract 4 from DX since we just did 4 entries
      CMPQ DX,  $4  // Are there still at least 4 left?
      JGE  QUADLOOP // If there are then jump back to QUADLOOP, otherwise fall through to SINGLE

  // SINGLE does the last 0 to 3 entries one at a time. There are so few of
  // these that they can all just go into X4
  SINGLE:
    TESTQ DX, DX // Is DX zero?
    JZ    REDUCE // If it is then there is nothing left, jump straight to REDUCE

    SINGLELOOP:
      MOVLQSX     0(BX),      R8 // Load the index into R8
      VMOVSS      (AX)(R8*4), X8 // X8 = dense[R8]
      VFMADD231SS 0(CX),  X8, X4 // X4 = (values[0] * X8) + X4

      ADDQ $4, BX     // Add 4 to BX. This moves indices forward by 1 int32
      ADDQ $4, CX     // Add 4 to CX. This moves values forward by 1 float32
      SUBQ $1, DX     // Subtract 1 from DX since we just did 1 entry
      JNZ  SINGLELOOP // If DX is not zero then jump back to SINGLELOOP

  // Now we need to add everything up into 1 number. Call the 16 lanes of Z0 s0
  // through s15. Each step adds the top half of what is left onto the bottom
  // half, so 16 lanes, then 8, then 4, then 2, then 1
  //
  // We don't use VHADDPS for this. On Zen 4 it is 3 uops and a new one can only
  // start every 2 cycles, while the extracts and shuffles here are 1 uop each
  // and a few of them can start every cycle
  //   https://uops.info/html-instr/VHADDPS_YMM_YMM_YMM.html
  //   https://uops.info/html-instr/VEXTRACTF128_XMM_YMM_I8.html
  //   https://uops.info/html-instr/VMOVHLPS_XMM_XMM_XMM.html
  //
  // The extract has to happen first. Any VEX instruction that writes Y0 zeroes
  // the top half of Z0, so s8 through s15 would be gone
  REDUCE:
    VEXTRACTF32X8 $1, Z0, Y1 // Y1 = [s8 ... s15], the high 256 bits of Z0
    VADDPS        Y1, Y0, Y0 // Y0 = [s0+s8, s1+s9 ... s7+s15], 8 lanes left
    VEXTRACTF128  $1, Y0, X1 // X1 = the high 4 lanes of Y0
    VADDPS        X1, X0, X0 // X0 = the low 4 lanes + the high 4 lanes, 4 lanes left
    VMOVHLPS      X0, X0, X1 // X1 = lanes 2 and 3 of X0 moved down into lanes 0 and 1
    VADDPS        X1, X0, X0 // Lanes 0 and 1 of X0 are now (0 + 2) and (1 + 3), 2 lanes left
    VMOVSHDUP     X0, X1     // X1 = lane 1 of X0 copied down into lane 0
    VADDSS        X1, X0, X0 // X0 = lane 0 + lane 1, everything from LOOP is now in lane 0 of X0

    VADDSS X5, X4, X4     // X4 = X4 + X5
    VADDSS X7, X6, X6     // X6 = X6 + X7
    VADDSS X6, X4, X4     // X4 = X4 + X6, everything from QUAD and SINGLE is now in X4
    VADDSS X4, X0, X0     // X0 = X0 + X4, this is our dot product
    VMOVSS X0, ret+72(FP) // Store X0 as the return value

    // We have to clear the top halves of the vector registers before going
    // back to Go. Right after this returns Go zeroes X15 with XORPS, which is
    // an old SSE instruction, and on Intel running SSE instructions while the
    // top halves are dirty is slow
    VZEROUPPER // Zero the upper bits of the vector registers
    RET        // We are done, return
