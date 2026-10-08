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

// func __sparseDot32_AVX_FMA(dense []float32, indices []int32, values []float32) float32
//
// Same idea as __sparseDot32_AVX512 but without the gather. VGATHERDPS is an
// AVX2 instruction and these versions only need AVX (and FMA for this one),
// same as the fourier and euclidean ones. So we load each dense[index]
// ourselves and build the vector up 1 lane at a time with VINSERTPS:
//
//   VMOVSS    dense[i0]          X2 = [d0,  0,  0,  0]
//   VINSERTPS dense[i1], lane 1  X2 = [d0, d1,  0,  0]
//   VINSERTPS dense[i2], lane 2  X2 = [d0, d1, d2,  0]
//   VINSERTPS dense[i3], lane 3  X2 = [d0, d1, d2, d3]
//
// After that it is a normal multiply and add with the 4 values that go with
// them. LOOP_AVXFMA does 2 of those (8 entries) each time around
//
// VINSERTPS can load straight from memory. On Intel that is the load plus 1
// uop on port 5, on Zen 4 it is just 1 uop. This way each entry costs 2 loads
// (the index and dense[index]) plus a quarter of a load for the values, where
// doing every entry with scalar loads would be 3. Ivy Bridge, Haswell and
// Skylake can only do 2 loads a cycle so the loads are the limit here, not the
// math
//   https://uops.info/html-instr/VINSERTPS_XMM_XMM_M32_I8.html
//
// On a 7950X this is actually faster than __sparseDot32_AVX512 (26ns vs 35ns
// for 128 entries) because Zen 4 can do 3 loads a cycle and its gather is slow.
// Skylake-SP is the other way around, its gather is fast and it only has 2
// load ports, see the notes above __sparseDot32_AVX512
//
// This only uses the 128 bit X registers. Going to 256 bits wouldn't help much
// since the loads are the limit, and every instruction in here being 128 bit
// VEX means the top halves of the registers never get dirty, so there is no
// VZEROUPPER at the end
TEXT ·__sparseDot32_AVX_FMA(SB), NOSPLIT, $0-76
  MOVQ dense_base+0(FP),    AX // Load the pointer of dense into AX
  MOVQ indices_base+24(FP), BX // Load the pointer of indices into BX
  MOVQ indices_len+32(FP),  DX // Load the length of indices into DX. This is how many entries we have to do
  MOVQ values_base+48(FP),  CX // Load the pointer of values into CX. It is the same length as indices

  // X0 and X1 are 2 separate sums, X0 for the first 4 entries of each loop and
  // X1 for the second 4. This way the 2 halves don't have to wait on each other
  VXORPS X0, X0, X0 // Zero out X0
  VXORPS X1, X1, X1 // Zero out X1

  CMPQ DX, $8      // Do we have at least 8 entries?
  JL   QUAD_AVXFMA // If we don't then jump straight to QUAD_AVXFMA

  LOOP_AVXFMA:
    MOVLQSX 0(BX),  R8  // Load the first index into R8. MOVLQSX sign extends the int32 so we can use it in an address
    MOVLQSX 4(BX),  R9  // Load the second index into R9
    MOVLQSX 8(BX),  R10 // Load the third index into R10
    MOVLQSX 12(BX), R11 // Load the fourth index into R11
    MOVLQSX 16(BX), R12 // Load the fifth index into R12
    MOVLQSX 20(BX), R13 // Load the sixth index into R13
    MOVLQSX 24(BX), SI  // Load the seventh index into SI
    MOVLQSX 28(BX), DI  // Load the eighth index into DI

    VMOVSS    (AX)(R8*4),           X2 // X2 = [dense[R8], 0, 0, 0], VMOVSS from memory zeroes the other 3 lanes
    VINSERTPS $0x10, (AX)(R9*4),  X2, X2 // Put dense[R9] in lane 1 of X2. The 1 in 0x10 is the lane
    VINSERTPS $0x20, (AX)(R10*4), X2, X2 // Put dense[R10] in lane 2 of X2
    VINSERTPS $0x30, (AX)(R11*4), X2, X2 // Put dense[R11] in lane 3 of X2, X2 now has the first 4
    VMOVSS    (AX)(R12*4),          X3 // X3 = [dense[R12], 0, 0, 0]
    VINSERTPS $0x10, (AX)(R13*4), X3, X3 // Put dense[R13] in lane 1 of X3
    VINSERTPS $0x20, (AX)(SI*4),  X3, X3 // Put dense[SI] in lane 2 of X3
    VINSERTPS $0x30, (AX)(DI*4),  X3, X3 // Put dense[DI] in lane 3 of X3, X3 now has the second 4

    VFMADD231PS 0(CX),  X2, X0 // X0 = (values[0:4] * X2) + X0
    VFMADD231PS 16(CX), X3, X1 // X1 = (values[4:8] * X3) + X1

    ADDQ $32, BX // Add 32 (8 * 4) to BX. This moves indices forward by 8 int32s
    ADDQ $32, CX // Add 32 (8 * 4) to CX. This moves values forward by 8 float32s
    SUBQ $8,  DX // Subtract 8 from DX since we just did 8 entries
    CMPQ DX,  $8 // Are there still at least 8 left?
    JGE  LOOP_AVXFMA // If there are then jump back to LOOP_AVXFMA, otherwise fall through to QUAD_AVXFMA

  // There are 0 to 7 entries left. If there are at least 4 then do 4 of them
  // the same way as LOOP_AVXFMA, then SINGLE_AVXFMA does the last 0 to 3
  QUAD_AVXFMA:
    VADDPS X1, X0, X0 // X0 = X0 + X1, we only need 1 sum from here on

    CMPQ DX, $4        // Do we have at least 4 entries left?
    JL   SINGLE_AVXFMA // If we don't then jump straight to SINGLE_AVXFMA

    MOVLQSX 0(BX),  R8  // Load the first index into R8
    MOVLQSX 4(BX),  R9  // Load the second index into R9
    MOVLQSX 8(BX),  R10 // Load the third index into R10
    MOVLQSX 12(BX), R11 // Load the fourth index into R11

    VMOVSS    (AX)(R8*4),           X2 // X2 = [dense[R8], 0, 0, 0]
    VINSERTPS $0x10, (AX)(R9*4),  X2, X2 // Put dense[R9] in lane 1 of X2
    VINSERTPS $0x20, (AX)(R10*4), X2, X2 // Put dense[R10] in lane 2 of X2
    VINSERTPS $0x30, (AX)(R11*4), X2, X2 // Put dense[R11] in lane 3 of X2

    VFMADD231PS 0(CX), X2, X0 // X0 = (values[0:4] * X2) + X0

    ADDQ $16, BX // Add 16 (4 * 4) to BX. This moves indices forward by 4 int32s
    ADDQ $16, CX // Add 16 (4 * 4) to CX. This moves values forward by 4 float32s
    SUBQ $4,  DX // Subtract 4 from DX since we just did 4 entries

  SINGLE_AVXFMA:
    TESTQ DX, DX        // Is DX zero?
    JZ    REDUCE_AVXFMA // If it is then there is nothing left, jump straight to REDUCE_AVXFMA

    SINGLELOOP_AVXFMA:
      MOVLQSX     0(BX),      R8 // Load the index into R8
      VMOVSS      (AX)(R8*4), X2 // X2 = [dense[R8], 0, 0, 0]
      VFMADD231SS 0(CX),  X2, X0 // Lane 0 of X0 = (values[0] * dense[R8]) + lane 0 of X0, the other 3 lanes are left alone

      ADDQ $4, BX            // Add 4 to BX. This moves indices forward by 1 int32
      ADDQ $4, CX            // Add 4 to CX. This moves values forward by 1 float32
      SUBQ $1, DX            // Subtract 1 from DX since we just did 1 entry
      JNZ  SINGLELOOP_AVXFMA // If DX is not zero then jump back to SINGLELOOP_AVXFMA

  // Add the 4 lanes of X0 up into 1, same as the end of the reduce in
  // __sparseDot32_AVX512
  REDUCE_AVXFMA:
    VMOVHLPS  X0, X0, X1 // X1 = lanes 2 and 3 of X0 moved down into lanes 0 and 1
    VADDPS    X1, X0, X0 // Lanes 0 and 1 of X0 are now (0 + 2) and (1 + 3)
    VMOVSHDUP X0, X1     // X1 = lane 1 of X0 copied down into lane 0
    VADDSS    X1, X0, X0 // X0 = lane 0 + lane 1, this is our dot product

    VMOVSS X0, ret+72(FP) // Store X0 as the return value
    RET                   // We are done, return

// func __sparseDot32Scalar_AVX_FMA(dense *float32, indices *int32, values *float32, count int) float32
//
// A scalar version for really short vectors, 2 to 8 entries is most of what
// DBSCAN sees. At that size the work is tiny and most of the time is just the
// call, so this is built to cost as little as possible to get in and out of:
//
//   It takes pointers and a count instead of 3 slices. Go calls assembly
//   through a wrapper that copies every argument onto the stack first, so 4
//   words instead of 9 makes the call cheaper. On a 7950X an empty function
//   took 2.41ns with slices and 2.24ns with pointers
//
//   It goes 2 entries at a time into 2 separate sums, X0 and X1, so the 2 FMAs
//   in each loop don't wait on each other
//
//   Both indicies for a pair come from 1 64 bit load and get split apart in
//   registers. That is 2.5 loads per entry instead of 3, and Zen 4 can only do
//   3 loads a cycle so the loads are the limit here
//
// On a 7950X this beat sparseDot32Go at every size (2.0ns vs 2.2ns at 2 and
// 3.3ns vs 4.1ns at 8). Against the 1 sum loop that is written right into
// SparseDot32 it ties at 2 and 3 entries, where the call is basically the
// whole cost, and wins from 4 up
TEXT ·__sparseDot32Scalar_AVX_FMA(SB), NOSPLIT, $0-36
  MOVQ dense+0(FP),   AX // Load the pointer of dense into AX
  MOVQ indices+8(FP), BX // Load the pointer of indices into BX
  MOVQ values+16(FP), CX // Load the pointer of values into CX
  MOVQ count+24(FP),  DX // Load the number of entries into DX

  VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
  VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

  CMPQ DX, $2       // Do we have at least 2 entries?
  JB   SINGLE_SCALAR // If we don't then jump straight to SINGLE_SCALAR

  LOOP_SCALAR:
    MOVQ    0(BX), R8 // Load the next 2 indicies into R8 at once, the first one is the low 32 bits and the second is the high 32 bits
    MOVLQSX R8,    R9 // R9 = the low 32 bits of R8 sign extended, this is the first index
    SARQ    $32,   R8 // Shift R8 right by 32 keeping the sign, now R8 is the second index

    VMOVSS (AX)(R9*4), X2 // X2 = dense[R9]
    VMOVSS (AX)(R8*4), X3 // X3 = dense[R8]

    VFMADD231SS 0(CX), X2, X0 // X0 = (values[0] * X2) + X0
    VFMADD231SS 4(CX), X3, X1 // X1 = (values[1] * X3) + X1

    ADDQ $8, BX      // Add 8 (2 * 4) to BX. This moves indices forward by 2 int32s
    ADDQ $8, CX      // Add 8 (2 * 4) to CX. This moves values forward by 2 float32s
    SUBQ $2, DX      // Subtract 2 from DX since we just did 2 entries
    CMPQ DX, $2      // Are there still at least 2 left?
    JAE  LOOP_SCALAR // If there are then jump back to LOOP_SCALAR, otherwise fall through to SINGLE_SCALAR

  // If the count was odd there is 1 entry left over
  SINGLE_SCALAR:
    TESTQ DX, DX      // Is DX zero?
    JZ    DONE_SCALAR // If it is then there is nothing left, jump straight to DONE_SCALAR

    MOVLQSX     0(BX),      R8 // Load the last index into R8
    VMOVSS      (AX)(R8*4), X2 // X2 = dense[R8]
    VFMADD231SS 0(CX),  X2, X0 // X0 = (values[0] * X2) + X0

  DONE_SCALAR:
    VADDSS X1, X0, X0     // X0 = X0 + X1, this is our dot product
    VMOVSS X0, ret+32(FP) // Store X0 as the return value
    RET                   // We are done, return

// func __sparseDot32Scalar_AVX(dense *float32, indices *int32, values *float32, count int) float32
//
// This is __sparseDot32Scalar_AVX_FMA for CPUs that have AVX but not FMA, like
// Ivy Bridge. Every FMA becomes a VMULSS and then a VADDSS, see the notes above
// __sparseDot32Scalar_AVX_FMA for how the rest of it works
//
// The 2 separate sums matter even more here. An add on Ivy Bridge takes 3
// cycles before the next add into the same sum can start, so with 1 sum (like
// the Go loop) we can only do 1 entry every 3 cycles. With 2 sums it is 2
// entries every 3 cycles, which is about as fast as Ivy Bridge's 2 load ports
// can feed it anyway
//   https://uops.info/html-instr/ADDSS_XMM_XMM.html
TEXT ·__sparseDot32Scalar_AVX(SB), NOSPLIT, $0-36
  MOVQ dense+0(FP),   AX // Load the pointer of dense into AX
  MOVQ indices+8(FP), BX // Load the pointer of indices into BX
  MOVQ values+16(FP), CX // Load the pointer of values into CX
  MOVQ count+24(FP),  DX // Load the number of entries into DX

  VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
  VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

  CMPQ DX, $2          // Do we have at least 2 entries?
  JB   SINGLE_SCALARAVX // If we don't then jump straight to SINGLE_SCALARAVX

  LOOP_SCALARAVX:
    MOVQ    0(BX), R8 // Load the next 2 indicies into R8 at once, the first one is the low 32 bits and the second is the high 32 bits
    MOVLQSX R8,    R9 // R9 = the low 32 bits of R8 sign extended, this is the first index
    SARQ    $32,   R8 // Shift R8 right by 32 keeping the sign, now R8 is the second index

    VMOVSS (AX)(R9*4), X2 // X2 = dense[R9]
    VMOVSS (AX)(R8*4), X3 // X3 = dense[R8]

    VMULSS 0(CX), X2, X2 // X2 = values[0] * X2
    VMULSS 4(CX), X3, X3 // X3 = values[1] * X3
    VADDSS X2,    X0, X0 // X0 = X0 + X2
    VADDSS X3,    X1, X1 // X1 = X1 + X3

    ADDQ $8, BX         // Add 8 (2 * 4) to BX. This moves indices forward by 2 int32s
    ADDQ $8, CX         // Add 8 (2 * 4) to CX. This moves values forward by 2 float32s
    SUBQ $2, DX         // Subtract 2 from DX since we just did 2 entries
    CMPQ DX, $2         // Are there still at least 2 left?
    JAE  LOOP_SCALARAVX // If there are then jump back to LOOP_SCALARAVX, otherwise fall through to SINGLE_SCALARAVX

  // If the count was odd there is 1 entry left over
  SINGLE_SCALARAVX:
    TESTQ DX, DX         // Is DX zero?
    JZ    DONE_SCALARAVX // If it is then there is nothing left, jump straight to DONE_SCALARAVX

    MOVLQSX 0(BX),      R8 // Load the last index into R8
    VMOVSS  (AX)(R8*4), X2 // X2 = dense[R8]
    VMULSS  0(CX),  X2, X2 // X2 = values[0] * X2
    VADDSS  X2,     X0, X0 // X0 = X0 + X2

  DONE_SCALARAVX:
    VADDSS X1, X0, X0     // X0 = X0 + X1, this is our dot product
    VMOVSS X0, ret+32(FP) // Store X0 as the return value
    RET                   // We are done, return

// func __sparseDot32_AVX(dense []float32, indices []int32, values []float32) float32
//
// This is __sparseDot32_AVX_FMA for CPUs that have AVX but not FMA, like Ivy
// Bridge. The only difference is every FMA becomes a VMULPS and then a VADDPS,
// see the notes above __sparseDot32_AVX_FMA for how the rest of it works
TEXT ·__sparseDot32_AVX(SB), NOSPLIT, $0-76
  MOVQ dense_base+0(FP),    AX // Load the pointer of dense into AX
  MOVQ indices_base+24(FP), BX // Load the pointer of indices into BX
  MOVQ indices_len+32(FP),  DX // Load the length of indices into DX. This is how many entries we have to do
  MOVQ values_base+48(FP),  CX // Load the pointer of values into CX. It is the same length as indices

  VXORPS X0, X0, X0 // Zero out X0, the sum for the first 4 entries of each loop
  VXORPS X1, X1, X1 // Zero out X1, the sum for the second 4 entries of each loop

  CMPQ DX, $8   // Do we have at least 8 entries?
  JL   QUAD_AVX // If we don't then jump straight to QUAD_AVX

  LOOP_AVX:
    MOVLQSX 0(BX),  R8  // Load the first index into R8
    MOVLQSX 4(BX),  R9  // Load the second index into R9
    MOVLQSX 8(BX),  R10 // Load the third index into R10
    MOVLQSX 12(BX), R11 // Load the fourth index into R11
    MOVLQSX 16(BX), R12 // Load the fifth index into R12
    MOVLQSX 20(BX), R13 // Load the sixth index into R13
    MOVLQSX 24(BX), SI  // Load the seventh index into SI
    MOVLQSX 28(BX), DI  // Load the eighth index into DI

    VMOVSS    (AX)(R8*4),           X2 // X2 = [dense[R8], 0, 0, 0]
    VINSERTPS $0x10, (AX)(R9*4),  X2, X2 // Put dense[R9] in lane 1 of X2
    VINSERTPS $0x20, (AX)(R10*4), X2, X2 // Put dense[R10] in lane 2 of X2
    VINSERTPS $0x30, (AX)(R11*4), X2, X2 // Put dense[R11] in lane 3 of X2, X2 now has the first 4
    VMOVSS    (AX)(R12*4),          X3 // X3 = [dense[R12], 0, 0, 0]
    VINSERTPS $0x10, (AX)(R13*4), X3, X3 // Put dense[R13] in lane 1 of X3
    VINSERTPS $0x20, (AX)(SI*4),  X3, X3 // Put dense[SI] in lane 2 of X3
    VINSERTPS $0x30, (AX)(DI*4),  X3, X3 // Put dense[DI] in lane 3 of X3, X3 now has the second 4

    VMULPS 0(CX),  X2, X2 // X2 = values[0:4] * X2
    VMULPS 16(CX), X3, X3 // X3 = values[4:8] * X3
    VADDPS X2,     X0, X0 // X0 = X0 + X2
    VADDPS X3,     X1, X1 // X1 = X1 + X3

    ADDQ $32, BX  // Add 32 (8 * 4) to BX. This moves indices forward by 8 int32s
    ADDQ $32, CX  // Add 32 (8 * 4) to CX. This moves values forward by 8 float32s
    SUBQ $8,  DX  // Subtract 8 from DX since we just did 8 entries
    CMPQ DX,  $8  // Are there still at least 8 left?
    JGE  LOOP_AVX // If there are then jump back to LOOP_AVX, otherwise fall through to QUAD_AVX

  QUAD_AVX:
    VADDPS X1, X0, X0 // X0 = X0 + X1, we only need 1 sum from here on

    CMPQ DX, $4     // Do we have at least 4 entries left?
    JL   SINGLE_AVX // If we don't then jump straight to SINGLE_AVX

    MOVLQSX 0(BX),  R8  // Load the first index into R8
    MOVLQSX 4(BX),  R9  // Load the second index into R9
    MOVLQSX 8(BX),  R10 // Load the third index into R10
    MOVLQSX 12(BX), R11 // Load the fourth index into R11

    VMOVSS    (AX)(R8*4),           X2 // X2 = [dense[R8], 0, 0, 0]
    VINSERTPS $0x10, (AX)(R9*4),  X2, X2 // Put dense[R9] in lane 1 of X2
    VINSERTPS $0x20, (AX)(R10*4), X2, X2 // Put dense[R10] in lane 2 of X2
    VINSERTPS $0x30, (AX)(R11*4), X2, X2 // Put dense[R11] in lane 3 of X2

    VMULPS 0(CX), X2, X2 // X2 = values[0:4] * X2
    VADDPS X2,    X0, X0 // X0 = X0 + X2

    ADDQ $16, BX // Add 16 (4 * 4) to BX. This moves indices forward by 4 int32s
    ADDQ $16, CX // Add 16 (4 * 4) to CX. This moves values forward by 4 float32s
    SUBQ $4,  DX // Subtract 4 from DX since we just did 4 entries

  SINGLE_AVX:
    TESTQ DX, DX     // Is DX zero?
    JZ    REDUCE_AVX // If it is then there is nothing left, jump straight to REDUCE_AVX

    SINGLELOOP_AVX:
      MOVLQSX 0(BX),      R8 // Load the index into R8
      VMOVSS  (AX)(R8*4), X2 // X2 = [dense[R8], 0, 0, 0]
      VMULSS  0(CX),  X2, X2 // Lane 0 of X2 = values[0] * dense[R8]
      VADDSS  X2,     X0, X0 // Lane 0 of X0 = lane 0 of X0 + lane 0 of X2, the other 3 lanes are left alone

      ADDQ $4, BX         // Add 4 to BX. This moves indices forward by 1 int32
      ADDQ $4, CX         // Add 4 to CX. This moves values forward by 1 float32
      SUBQ $1, DX         // Subtract 1 from DX since we just did 1 entry
      JNZ  SINGLELOOP_AVX // If DX is not zero then jump back to SINGLELOOP_AVX

  REDUCE_AVX:
    VMOVHLPS  X0, X0, X1 // X1 = lanes 2 and 3 of X0 moved down into lanes 0 and 1
    VADDPS    X1, X0, X0 // Lanes 0 and 1 of X0 are now (0 + 2) and (1 + 3)
    VMOVSHDUP X0, X1     // X1 = lane 1 of X0 copied down into lane 0
    VADDSS    X1, X0, X0 // X0 = lane 0 + lane 1, this is our dot product

    VMOVSS X0, ret+72(FP) // Store X0 as the return value
    RET                   // We are done, return
