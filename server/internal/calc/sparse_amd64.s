//go:build amd64 && !nosimd

#include "textflag.h"

// func __sparseDot32Scalar_AVX(dense *float32, denseLength int, indices *int32, values *float32, count int) float32
//
// This is the dot product of dense and a sparse vector, but only at the spots
// in indices. The plain Go version is sparseDot32Go in sparse.go:
//
//   for i, index := range indices {
//     dot += dense[index] * values[i]
//   }
//
// Most of what DBSCAN sees is only 2 to 8 entries. At that size the work is
// tiny and a lot of the time is just the call, so this is built to be cheap to
// get in and out of:
//
//   It takes pointers and counts instead of 3 slices. Go calls assembly
//   through a wrapper that copies every argument onto the stack first, so 5
//   words instead of 9 makes the call cheaper. On a 7950X an empty function
//   took 2.41ns with 3 slices and 2.24ns with pointers. denseLength is only
//   there for sparseDot32Go, this ignores it
//
//   It goes 2 entries at a time into 2 separate sums, X0 and X1. An add takes
//   3 cycles before the next add into the same sum can start (on both Ivy
//   Bridge and Zen 4), so with 1 sum like the plain Go loop we can only do 1
//   entry every 3 cycles. With 2 sums it is 2 every 3 cycles
//     https://uops.info/html-instr/ADDSS_XMM_XMM.html
//
//   Both indicies for a pair come from 1 64 bit load and get split apart in
//   registers. That is 2.5 loads per entry instead of 3. Ivy Bridge can only
//   do 2 loads a cycle and Zen 4 can do 3, so the loads are the other limit
//   here
//
// This uses a VMULSS and then a VADDSS instead of an FMA. Ivy Bridge doesn't
// have FMA, and on Zen 4 an FMA version was slower anyway (41ns vs 34ns for
// 128 entries). An FMA takes 4 cycles before the next one into the same sum
// can start, the VMULSS here isn't part of that wait at all
//
// This does not check that the indicies actually fit in dense, a bad index
// just reads whatever memory is there instead of panicking. The caller has to
// make sure of that. In DBSCAN the indicies come from the TFIDF vocabulary,
// which is the same width as dense, so they always fit
//
// Everything in here is 128 bit VEX, so the top halves of the registers never
// get dirty and there is no VZEROUPPER at the end
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VMULSS 0(CX), X2, X2 is X2 = X2 * values[0]
TEXT ·__sparseDot32Scalar_AVX(SB), NOSPLIT, $0-44
  MOVQ dense+0(FP),    AX // Load the pointer of dense into AX
  MOVQ indices+16(FP), BX // Load the pointer of indices into BX
  MOVQ values+24(FP),  CX // Load the pointer of values into CX
  MOVQ count+32(FP),   DX // Load the number of entries into DX

  VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
  VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

  CMPQ DX, $2  // Do we have at least 2 entries?
  JB   SINGLE  // If we don't then jump straight to SINGLE

  LOOP:
    MOVQ    0(BX), R8 // Load the next 2 indicies into R8 at once, the first one is the low 32 bits and the second is the high 32 bits
    MOVLQSX R8,    R9 // R9 = the low 32 bits of R8 sign extended, this is the first index
    SARQ    $32,   R8 // Shift R8 right by 32 keeping the sign, now R8 is the second index

    VMOVSS (AX)(R9*4), X2 // X2 = dense[R9]
    VMOVSS (AX)(R8*4), X3 // X3 = dense[R8]

    VMULSS 0(CX), X2, X2 // X2 = values[0] * X2
    VMULSS 4(CX), X3, X3 // X3 = values[1] * X3
    VADDSS X2,    X0, X0 // X0 = X0 + X2
    VADDSS X3,    X1, X1 // X1 = X1 + X3

    ADDQ $8, BX // Add 8 (2 * 4) to BX. This moves indices forward by 2 int32s
    ADDQ $8, CX // Add 8 (2 * 4) to CX. This moves values forward by 2 float32s
    SUBQ $2, DX // Subtract 2 from DX since we just did 2 entries
    CMPQ DX, $2 // Are there still at least 2 left?
    JAE  LOOP   // If there are then jump back to LOOP, otherwise fall through to SINGLE

  // If the count was odd there is 1 entry left over
  SINGLE:
    TESTQ DX, DX // Is DX zero?
    JZ    DONE   // If it is then there is nothing left, jump straight to DONE

    MOVLQSX 0(BX),      R8 // Load the last index into R8
    VMOVSS  (AX)(R8*4), X2 // X2 = dense[R8]
    VMULSS  0(CX),  X2, X2 // X2 = values[0] * X2
    VADDSS  X2,     X0, X0 // X0 = X0 + X2

  DONE:
    VADDSS X1, X0, X0     // X0 = X0 + X1, this is our dot product
    VMOVSS X0, ret+40(FP) // Store X0 as the return value
    RET                   // We are done, return
