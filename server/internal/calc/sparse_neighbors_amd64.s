//go:build amd64 && !nosimd

// vim: set expandtab tabstop=2 shiftwidth=2 textwidth=80 comments=\://:

#include "textflag.h"

// Some notes before reading this file
//
// The plain Go version of this is sparseNeighbors32Go in sparse_neighbors.go,
// and the dot product in the middle of it is sparseDot32Go in sparse.go. The
// notes above each part below have the Go that part lines up with. The assembly
// doesn't go in the same order as the Go though. It checks the signature of
// every vector first and then does the dot product for just the ones that made
// it. So where the order is different the notes have the Go from the original
// loop, and then that same Go written out the way the assembly does it
//
// There are 3 versions of this and they only differ in how they check the
// signatures. Everything from FILTERED down is the exact same code in all 3:
//
//   __sparseNeighbors32_AVX:      1 signature at a time, for Ivy Bridge (the
//                                 E5-2667 v2) and anything else without
//                                 AVX-512
//
//   __sparseNeighbors32_AVX512VL: 8 at a time with 256 bit registers, for
//                                 Skylake and Cascade Lake Xeons
//
//   __sparseNeighbors32_AVX512:   16 at a time with 512 bit registers, for
//                                 Zen 4 and Ice Lake or newer
//
// The dot product adds up in the exact same order as sparseDot32Go, X0 gets the
// even entries and X1 gets the odd ones. So the distance is the same down to
// the last bit as the Go version, and a pair that lands right on epsilon
// doesn't change sides depending on which version ran
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VSUBSS X0, X3, X3 is X3 = X3 - X0

// const_sparse_neighbors_lanes is 0 through 15 as int32s. BLOCK adds i to these
// to get the position of each vector it is looking at, the 256 bit version only
// uses the first 8
DATA const_sparse_neighbors_lanes<>+0(SB)/4,  $0
DATA const_sparse_neighbors_lanes<>+4(SB)/4,  $1
DATA const_sparse_neighbors_lanes<>+8(SB)/4,  $2
DATA const_sparse_neighbors_lanes<>+12(SB)/4, $3
DATA const_sparse_neighbors_lanes<>+16(SB)/4, $4
DATA const_sparse_neighbors_lanes<>+20(SB)/4, $5
DATA const_sparse_neighbors_lanes<>+24(SB)/4, $6
DATA const_sparse_neighbors_lanes<>+28(SB)/4, $7
DATA const_sparse_neighbors_lanes<>+32(SB)/4, $8
DATA const_sparse_neighbors_lanes<>+36(SB)/4, $9
DATA const_sparse_neighbors_lanes<>+40(SB)/4, $10
DATA const_sparse_neighbors_lanes<>+44(SB)/4, $11
DATA const_sparse_neighbors_lanes<>+48(SB)/4, $12
DATA const_sparse_neighbors_lanes<>+52(SB)/4, $13
DATA const_sparse_neighbors_lanes<>+56(SB)/4, $14
DATA const_sparse_neighbors_lanes<>+60(SB)/4, $15
GLOBL const_sparse_neighbors_lanes<>(SB), RODATA|NOPTR, $64

// func __sparseNeighbors32_AVX(dense []float32, signature uint64, norm2, epsilon float32, signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32, output []int32) int
//
// This is the whole inner loop of DBSCAN's getNeighbors in one call. It writes
// the position of every vector within epsilon of dense into output and returns
// how many there were. We go over the vectors in 2 passes:
//
//   FILTER:    Checks every signature against ours and writes the position of
//              every vector with at least 1 bit in common into output. This is
//              where most vectors get thrown out
//
//   CANDIDATE: Goes over just the vectors FILTER kept. PAIR and SINGLE do the
//              dot product, then DISTANCE works out the distance and keeps the
//              ones that are close enough. Those get written back into output
//              over the top of the candidates
//
// The arguments are slices instead of pointers like __sparseDot32Scalar_AVX.
// This gets called once for each point instead of once for each pair, so the
// cost of copying the arguments doesn't matter anymore
//
// The only thing this needs is AVX for the VEX floating point instructions,
// FILTER is plain integer code. That way this runs on Ivy Bridge (E5-2667 v2)
// which doesn't have AVX2 or BMI, and everything newer too
TEXT ·__sparseNeighbors32_AVX(SB), NOSPLIT, $0-192
  MOVQ signature+24(FP),       R8 // Load our signature into R8
  MOVQ signatures_base+40(FP), SI // Load the pointer of signatures into SI
  MOVQ signatures_len+48(FP),  CX // Load the number of vectors into CX
  MOVQ output_base+160(FP),    DI // Load the pointer of output into DI

  XORQ AX, AX // AX = i, the vector we are looking at
  XORQ DX, DX // DX = how many candidates we have written into output so far

  TESTQ CX, CX   // Do we have any vectors at all?
  JZ    FILTERED // If we don't then there aren't any candidates, jump straight to FILTERED

  // FILTER is this part of the loop in sparseNeighbors32Go:
  //
  //   for i := range signatures {
  //     if signature&signatures[i] == 0 {
  //       continue
  //     }
  //     ...
  //   }
  //
  // But it doesn't do the dot product right away, it keeps the candidates in
  // output for CANDIDATE. And it doesn't branch on whether a vector made it.
  // Whether a pair has a bit in common is pretty much random, so a branch would
  // guess wrong a lot and every wrong guess costs about 15 to 20 cycles.
  // Instead we always write i into output and then only move forward if i made
  // it, otherwise the next one just writes over it. Written out in Go that is
  // this, with AX as i and DX as candidates:
  //
  //   candidates := 0
  //   for i := range signatures {
  //     output[candidates] = int32(i)
  //     if signature&signatures[i] != 0 {
  //       candidates++
  //     }
  //   }
  //
  // The if is a SETNE and an ADDQ though, not a jump
  FILTER:
    MOVQ  (SI)(AX*8), R9 // R9 = signatures[i]
    MOVL  AX, (DI)(DX*4) // output[DX] = i, this only sticks if i is a candidate
    XORL  R10, R10       // Zero R10 so that SETNE below only has to set the low byte
    ANDQ  R8,  R9        // R9 = R9 & our signature, this sets ZF if they have no bits in common
    SETNE R10B           // R10 = 1 if they have a bit in common, 0 if they don't
    ADDQ  R10, DX        // Move DX forward by 1 only if i was a candidate
    INCQ  AX             // i++
    CMPQ  AX,  CX        // Have we checked every vector?
    JB    FILTER         // If we haven't then jump back to FILTER

  // Now output[0:DX] is every candidate. CANDIDATE is the rest of the loop in
  // sparseNeighbors32Go, but only for the candidates:
  //
  //   start, end := offsets[i], offsets[i+1]
  //   dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //   distance := norm2 + norms[i] - 2*dot
  //   if distance <= epsilon {
  //     output[count] = int32(i)
  //     count++
  //   }
  //
  // The ones that are close enough get written back into output over the top of
  // the candidates, with the same trick as FILTER so there is no branch on the
  // distance either. We never write past the candidate we are reading, so doing
  // this in place is fine. Written out in Go that is this, with DI walking over
  // the candidates, R8 as &output[count] and R9 as where the candidates end:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     output[count] = c
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if distance <= epsilon {
  //       count++
  //     }
  //   }
  FILTERED:
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], where the first neighbor goes

  MOVQ   dense_base+0(FP),     AX  // Load the pointer of dense into AX
  MOVQ   indices_base+112(FP), BX  // Load the pointer of indices into BX
  MOVQ   values_base+136(FP),  CX  // Load the pointer of values into CX
  MOVQ   offsets_base+88(FP),  DX  // Load the pointer of offsets into DX
  MOVQ   norms_base+64(FP),    SI  // Load the pointer of norms into SI
  VMOVSS norm2+32(FP),         X14 // X14 = our squared norm
  VMOVSS epsilon+36(FP),       X15 // X15 = epsilon

  CMPQ DI, R9 // Do we have any candidates?
  JAE  DONE   // If we don't then jump straight to DONE

  // Each loop of CANDIDATE is 1 candidate, c. c is i in sparseNeighbors32Go, so
  // this is:
  //
  //   start, end := offsets[i], offsets[i+1]
  //
  // We also get the norm2 + norms[i] part of the distance going here since it
  // doesn't need the dot product, and write c into output the same way FILTER
  // does. R11 is start and R12 is end - start. X0 and X1 are the 2 sums from
  // sparseDot32Go:
  //
  //   var a, b float32
  CANDIDATE:
    MOVLQSX (DI),          R10      // R10 = c, the candidate we are looking at
    MOVL    R10,           (R8)     // Write c where the next neighbor goes, this only sticks if it is close enough
    VADDSS  (SI)(R10*4),   X14, X3  // X3 = our norm + norms[c], the first part of the distance
    MOVLQSX (DX)(R10*4),   R11      // R11 = offsets[c], where c's entries start
    MOVLQSX 4(DX)(R10*4),  R12      // R12 = offsets[c+1], where c's entries end
    SUBQ    R11,           R12      // R12 = how many entries c has

    VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
    VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

    CMPQ R12, $2 // Do we have at least 2 entries?
    JB   SINGLE  // If we don't then jump straight to SINGLE

    // PAIR is the same loop as __sparseDot32Scalar_AVX, see the notes there for
    // why it does 2 at a time. In sparseDot32Go it is this, where indices and
    // values there are indices[start:end] and values[start:end] here. So R11 is
    // start + i and R12 is count - i:
    //
    //   for ; i+2 <= count; i += 2 {
    //     a += dense[indices[i]] * values[i]
    //     b += dense[indices[i+1]] * values[i+1]
    //   }
    PAIR:
      MOVQ    (BX)(R11*4), R13          // Load the next 2 indicies into R13 at once, the first one is the low 32 bits and the second is the high 32 bits
      MOVLQSX R13,         R10          // R10 = the low 32 bits of R13 sign extended, this is the first index
      SARQ    $32,         R13          // Shift R13 right by 32 keeping the sign, now R13 is the second index
      VMOVSS  (AX)(R10*4), X2           // X2 = dense[R10]
      VMOVSS  (AX)(R13*4), X4           // X4 = dense[R13]
      VMULSS  (CX)(R11*4), X2,  X2      // X2 = values[R11] * X2
      VMULSS  4(CX)(R11*4), X4, X4      // X4 = values[R11+1] * X4
      VADDSS  X2,          X0,  X0      // X0 = X0 + X2
      VADDSS  X4,          X1,  X1      // X1 = X1 + X4
      ADDQ    $2,          R11          // Move R11 forward by 2 entries
      SUBQ    $2,          R12          // Subtract 2 from R12 since we just did 2 entries
      CMPQ    R12,         $2           // Are there still at least 2 left?
      JAE     PAIR                      // If there are then jump back to PAIR, otherwise fall through to SINGLE

    // If the count was odd there is 1 entry left over. In sparseDot32Go that
    // is:
    //
    //   if i < count {
    //     a += dense[indices[i]] * values[i]
    //   }
    SINGLE:
      TESTQ   R12,         R12     // Is R12 zero?
      JZ      DISTANCE             // If it is then there is nothing left, jump straight to DISTANCE
      MOVLQSX (BX)(R11*4), R10     // Load the last index into R10
      VMOVSS  (AX)(R10*4), X2      // X2 = dense[R10]
      VMULSS  (CX)(R11*4), X2, X2  // X2 = values[R11] * X2
      VADDSS  X2,          X0, X0  // X0 = X0 + X2

    // In sparseNeighbors32Go this is:
    //
    //   dot := sparseDot32Go(...) // which ends with return a + b
    //   distance := norm2 + norms[i] - 2*dot
    //   if distance <= epsilon {
    //     output[count] = int32(i)
    //     count++
    //   }
    //
    // c already got written into output at the top of CANDIDATE, so all the if
    // has to do here is move R8 forward. VUCOMISS sets CF if epsilon is less
    // than the distance, and also if either of them is NaN. So SETCC (set if CF
    // is clear) is the same as distance <= epsilon in Go, NaN included
    DISTANCE:
      VADDSS   X1,  X0, X0      // X0 = X0 + X1, this is our dot product
      VADDSS   X0,  X0, X0      // X0 = X0 + X0, which is exactly 2 * dot without needing a constant
      VSUBSS   X0,  X3, X3      // X3 = X3 - X0, this is the distance
      XORL     R13, R13         // Zero R13 so SETCC only has to set the low byte, this has to happen before the compare because XOR changes the flags
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is cleared if epsilon >= distance
      SETCC    R13B             // R13 = 1 if the distance <= epsilon, 0 if it isn't
      LEAQ     (R8)(R13*4), R8  // Move R8 forward by 1 int32 only if c was close enough
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // DONE is the return count at the end of sparseNeighbors32Go. R8 is
  // &output[count], so count is how far R8 got from the start of output
  DONE:
    MOVQ output_base+160(FP), DI  // Load the pointer of output into DI again
    SUBQ DI,                  R8  // R8 = how many bytes of neighbors we wrote
    SHRQ $2,                  R8  // Divide by 4 to get how many neighbors that is
    MOVQ R8,                  ret+184(FP) // Return how many neighbors we found
    RET

// func __sparseNeighbors32_AVX512VL(dense []float32, signature uint64, norm2, epsilon float32, signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32, output []int32) int
//
// Same as __sparseNeighbors32_AVX but the signatures get checked 8 at a time
// with AVX-512, for Skylake and Cascade Lake Xeons. Everything from FILTERED
// down is the exact same code
//
// What is different from the AVX version:
//
//   - There is no FILTER, BLOCK does 8 signatures per loop instead. See the
//     notes above BLOCK
//   - TAIL does the last few signatures 1 at a time the same way FILTER does
//   - BLOCK leaves the top halves of the YMM registers dirty, so there is a
//     VZEROUPPER before FILTERED
//
// This uses 256 bit registers (AVX-512VL) instead of 512 bit ones. On Skylake
// and Cascade Lake anything 512 bit drops the clock speed of the core for a
// while after, and CANDIDATE right after this would be stuck running slower
// too. 256 bit AVX-512 instructions like these don't do that. Anything newer
// gets __sparseNeighbors32_AVX512
// https://travisdowns.github.io/blog/2020/01/17/avxfreq1.html
TEXT ·__sparseNeighbors32_AVX512VL(SB), NOSPLIT, $0-192
  MOVQ signature+24(FP),       R8 // Load our signature into R8
  MOVQ signatures_base+40(FP), SI // Load the pointer of signatures into SI
  MOVQ signatures_len+48(FP),  CX // Load the number of vectors into CX
  MOVQ output_base+160(FP),    DI // Load the pointer of output into DI

  XORQ AX, AX // AX = i, the vector we are looking at
  XORQ DX, DX // DX = how many candidates we have written into output so far

  VPBROADCASTQ signature+24(FP),                Y0 // Y0 = our signature in all 4 lanes
  VMOVDQU      const_sparse_neighbors_lanes<>(SB), Y1 // Y1 = [0, 1, ..., 7], the position of each vector in the first block
  MOVL         $8,                              R9 // R9 = 8
  VPBROADCASTD R9,                              Y2 // Y2 = 8 in all 8 lanes, we add this to Y1 for each block

  CMPQ CX, $8 // Do we have at least 8 vectors?
  JB   TAIL   // If we don't then jump straight to TAIL

  // BLOCK does the same thing as FILTER in the AVX version, which is this part
  // of the loop in sparseNeighbors32Go:
  //
  //   for i := range signatures {
  //     if signature&signatures[i] == 0 {
  //       continue
  //     }
  //     ...
  //   }
  //
  // But 8 vectors at a time. Written out in Go each loop is:
  //
  //   for n := range 8 {
  //     if signature&signatures[i+n] != 0 {
  //       output[candidates] = int32(i + n)
  //       candidates++
  //     }
  //   }
  //   i += 8
  //
  // There are no branches in it though:
  //
  //   VPTESTMQ ANDs 4 signatures with ours and sets a bit in a mask register
  //   for each one that isn't zero. Doing it twice and shifting one of them up
  //   gets us 8 bits for 8 vectors in K1, that is the if
  //
  //   VPCOMPRESSD takes the lanes of Y1 that have their bit set in K1 and packs
  //   them down to the bottom of Y3. Y1 is [i, i+1, ..., i+7], so if only i+2
  //   and i+5 are candidates we get [i+2, i+5, 0, 0, 0, 0, 0, 0]. We store all
  //   8 lanes and then only move forward by how many bits were set, so the
  //   zeros just get written over by the next block. In a block we have looked
  //   at 8 more vectors than we could have written, so this can never write
  //   past the end of output
  //
  // VPCOMPRESSD can write straight to memory, but on Zen 4 that was 13 times
  // slower than compressing into Y3 and storing that. For 4096 signatures it
  // took 5451ns instead of 405ns on a 7950X. So it always goes through a
  // register
  BLOCK:
    VPTESTMQ      (SI)(AX*8),   Y0, K1 // K1 bit n = 1 if signatures[i+n] & our signature isn't zero, for the first 4
    VPTESTMQ      32(SI)(AX*8), Y0, K2 // K2 is the same thing for the next 4
    KSHIFTLW      $4,  K2, K2          // Shift K2 up by 4 so it lines up with vectors 4 to 7
    KORW          K2,  K1, K1          // K1 = K1 | K2, now K1 has 1 bit for each of the 8 vectors
    VPCOMPRESSD.Z Y1,  K1, Y3          // Pack the positions of the candidates down to the bottom of Y3
    VMOVDQU32     Y3,  (DI)(DX*4)      // Store all 8 lanes at output[DX], only the candidates at the bottom stick
    KMOVW         K1,  R10             // R10 = K1 so we can count the bits
    POPCNTL       R10, R10             // R10 = how many candidates were in this block
    ADDQ          R10, DX              // Move DX forward by that many
    VPADDD        Y2,  Y1, Y1          // Add 8 to every lane of Y1 for the next block
    ADDQ          $8,  AX              // i += 8
    LEAQ          8(AX), R10           // R10 = i + 8, where the next block would end
    CMPQ          R10, CX              // Is there a whole block left?
    JBE           BLOCK                // If there is then jump back to BLOCK

  // There are less than 8 vectors left, TAIL does them 1 at a time the same way
  // FILTER does in the AVX version. Written out in Go it is:
  //
  //   for ; i < len(signatures); i++ {
  //     output[candidates] = int32(i)
  //     if signature&signatures[i] != 0 {
  //       candidates++
  //     }
  //   }
  TAIL:
    CMPQ  AX, CX         // Have we checked every vector?
    JAE   FILTERED       // If we have then jump to FILTERED
    MOVQ  (SI)(AX*8), R9 // R9 = signatures[i]
    MOVL  AX, (DI)(DX*4) // output[DX] = i, this only sticks if i is a candidate
    XORL  R10, R10       // Zero R10 so that SETNE below only has to set the low byte
    ANDQ  R8,  R9        // R9 = R9 & our signature, this sets ZF if they have no bits in common
    SETNE R10B           // R10 = 1 if they have a bit in common, 0 if they don't
    ADDQ  R10, DX        // Move DX forward by 1 only if i was a candidate
    INCQ  AX             // i++
    JMP   TAIL           // Jump back to TAIL

  // We're done with the 256 bit registers, clear the top halves so the scalar
  // instructions below don't have to deal with them
  FILTERED:
  VZEROUPPER

  // Everything from here down is the same as the AVX version, see the notes
  // there. CANDIDATE is the rest of the loop in sparseNeighbors32Go, but only
  // for the candidates:
  //
  //   start, end := offsets[i], offsets[i+1]
  //   dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //   distance := norm2 + norms[i] - 2*dot
  //   if distance <= epsilon {
  //     output[count] = int32(i)
  //     count++
  //   }
  //
  // Written out in Go the way it is done here, with DI walking over the
  // candidates, R8 as &output[count] and R9 as where the candidates end:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     output[count] = c
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if distance <= epsilon {
  //       count++
  //     }
  //   }
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], where the first neighbor goes

  MOVQ   dense_base+0(FP),     AX  // Load the pointer of dense into AX
  MOVQ   indices_base+112(FP), BX  // Load the pointer of indices into BX
  MOVQ   values_base+136(FP),  CX  // Load the pointer of values into CX
  MOVQ   offsets_base+88(FP),  DX  // Load the pointer of offsets into DX
  MOVQ   norms_base+64(FP),    SI  // Load the pointer of norms into SI
  VMOVSS norm2+32(FP),         X14 // X14 = our squared norm
  VMOVSS epsilon+36(FP),       X15 // X15 = epsilon

  CMPQ DI, R9 // Do we have any candidates?
  JAE  DONE   // If we don't then jump straight to DONE

  // Same as CANDIDATE in the AVX version. c is i in sparseNeighbors32Go:
  //
  //   start, end := offsets[i], offsets[i+1]
  //
  // And X0 and X1 are the 2 sums from sparseDot32Go:
  //
  //   var a, b float32
  CANDIDATE:
    MOVLQSX (DI),          R10      // R10 = c, the candidate we are looking at
    MOVL    R10,           (R8)     // Write c where the next neighbor goes, this only sticks if it is close enough
    VADDSS  (SI)(R10*4),   X14, X3  // X3 = our norm + norms[c], the first part of the distance
    MOVLQSX (DX)(R10*4),   R11      // R11 = offsets[c], where c's entries start
    MOVLQSX 4(DX)(R10*4),  R12      // R12 = offsets[c+1], where c's entries end
    SUBQ    R11,           R12      // R12 = how many entries c has

    VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
    VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

    CMPQ R12, $2 // Do we have at least 2 entries?
    JB   SINGLE  // If we don't then jump straight to SINGLE

    // Same as PAIR in the AVX version. In sparseDot32Go it is:
    //
    //   for ; i+2 <= count; i += 2 {
    //     a += dense[indices[i]] * values[i]
    //     b += dense[indices[i+1]] * values[i+1]
    //   }
    PAIR:
      MOVQ    (BX)(R11*4), R13          // Load the next 2 indicies into R13 at once, the first one is the low 32 bits and the second is the high 32 bits
      MOVLQSX R13,         R10          // R10 = the low 32 bits of R13 sign extended, this is the first index
      SARQ    $32,         R13          // Shift R13 right by 32 keeping the sign, now R13 is the second index
      VMOVSS  (AX)(R10*4), X2           // X2 = dense[R10]
      VMOVSS  (AX)(R13*4), X4           // X4 = dense[R13]
      VMULSS  (CX)(R11*4), X2,  X2      // X2 = values[R11] * X2
      VMULSS  4(CX)(R11*4), X4, X4      // X4 = values[R11+1] * X4
      VADDSS  X2,          X0,  X0      // X0 = X0 + X2
      VADDSS  X4,          X1,  X1      // X1 = X1 + X4
      ADDQ    $2,          R11          // Move R11 forward by 2 entries
      SUBQ    $2,          R12          // Subtract 2 from R12 since we just did 2 entries
      CMPQ    R12,         $2           // Are there still at least 2 left?
      JAE     PAIR                      // If there are then jump back to PAIR, otherwise fall through to SINGLE

    // If the count was odd there is 1 entry left over. In sparseDot32Go that
    // is:
    //
    //   if i < count {
    //     a += dense[indices[i]] * values[i]
    //   }
    SINGLE:
      TESTQ   R12,         R12     // Is R12 zero?
      JZ      DISTANCE             // If it is then there is nothing left, jump straight to DISTANCE
      MOVLQSX (BX)(R11*4), R10     // Load the last index into R10
      VMOVSS  (AX)(R10*4), X2      // X2 = dense[R10]
      VMULSS  (CX)(R11*4), X2, X2  // X2 = values[R11] * X2
      VADDSS  X2,          X0, X0  // X0 = X0 + X2

    // Same as DISTANCE in the AVX version, see the notes there. In
    // sparseNeighbors32Go this is:
    //
    //   dot := sparseDot32Go(...) // which ends with return a + b
    //   distance := norm2 + norms[i] - 2*dot
    //   if distance <= epsilon {
    //     output[count] = int32(i)
    //     count++
    //   }
    DISTANCE:
      VADDSS   X1,  X0, X0      // X0 = X0 + X1, this is our dot product
      VADDSS   X0,  X0, X0      // X0 = X0 + X0, which is exactly 2 * dot without needing a constant
      VSUBSS   X0,  X3, X3      // X3 = X3 - X0, this is the distance
      XORL     R13, R13         // Zero R13 so SETCC only has to set the low byte, this has to happen before the compare because XOR changes the flags
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is cleared if epsilon >= distance
      SETCC    R13B             // R13 = 1 if the distance <= epsilon, 0 if it isn't
      LEAQ     (R8)(R13*4), R8  // Move R8 forward by 1 int32 only if c was close enough
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // DONE is the return count at the end of sparseNeighbors32Go
  DONE:
    MOVQ output_base+160(FP), DI  // Load the pointer of output into DI again
    SUBQ DI,                  R8  // R8 = how many bytes of neighbors we wrote
    SHRQ $2,                  R8  // Divide by 4 to get how many neighbors that is
    MOVQ R8,                  ret+184(FP) // Return how many neighbors we found
    RET

// func __sparseNeighbors32_AVX512(dense []float32, signature uint64, norm2, epsilon float32, signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32, output []int32) int
//
// Same as the AVX512VL version but with 512 bit ZMM registers, so BLOCK does 16
// vectors at a time instead of 8. This is the one for Zen 4 and Ice Lake or
// newer, where using 512 bit registers doesn't slow the core down
//
// What is different from the AVX512VL version:
//
//   - 1 VPTESTMQ covers 8 signatures instead of 4, and KUNPCKBW glues the 8
//     bits from 2 of them together into 16 bits for 16 vectors
//   - Z1 is [i, i+1, ..., i+15] and Z2 adds 16 to it for each block
//   - TAIL can have up to 15 vectors left instead of 7
//
// Zen 4 splits a 512 bit instruction into 2 256 bit halves, but it still came
// out ahead because there are half as many instructions to get through. For
// 4096 signatures BLOCK and TAIL took 272ns, vs 405ns with 256 bit registers on
// a 7950X. That is only about 5% of the whole call though, and less when most
// vectors make it through, since most of the time is in CANDIDATE
TEXT ·__sparseNeighbors32_AVX512(SB), NOSPLIT, $0-192
  MOVQ signature+24(FP),       R8 // Load our signature into R8
  MOVQ signatures_base+40(FP), SI // Load the pointer of signatures into SI
  MOVQ signatures_len+48(FP),  CX // Load the number of vectors into CX
  MOVQ output_base+160(FP),    DI // Load the pointer of output into DI

  XORQ AX, AX // AX = i, the vector we are looking at
  XORQ DX, DX // DX = how many candidates we have written into output so far

  VPBROADCASTQ signature+24(FP),                   Z0 // Z0 = our signature in all 8 lanes
  VMOVDQU32    const_sparse_neighbors_lanes<>(SB), Z1 // Z1 = [0, 1, ..., 15], the position of each vector in the first block
  MOVL         $16,                                R9 // R9 = 16
  VPBROADCASTD R9,                                 Z2 // Z2 = 16 in all 16 lanes, we add this to Z1 for each block

  CMPQ CX, $16 // Do we have at least 16 vectors?
  JB   TAIL    // If we don't then jump straight to TAIL

  // Same as BLOCK in the AVX512VL version but 16 vectors at a time, see the
  // notes there. It is still this part of the loop in sparseNeighbors32Go:
  //
  //   for i := range signatures {
  //     if signature&signatures[i] == 0 {
  //       continue
  //     }
  //     ...
  //   }
  //
  // And written out in Go each loop is:
  //
  //   for n := range 16 {
  //     if signature&signatures[i+n] != 0 {
  //       output[candidates] = int32(i + n)
  //       candidates++
  //     }
  //   }
  //   i += 16
  BLOCK:
    VPTESTMQ      (SI)(AX*8),   Z0, K1 // K1 bit n = 1 if signatures[i+n] & our signature isn't zero, for the first 8
    VPTESTMQ      64(SI)(AX*8), Z0, K2 // K2 is the same thing for the next 8
    KUNPCKBW      K1,  K2, K1          // K1 = K2 << 8 | K1, now K1 has 1 bit for each of the 16 vectors
    VPCOMPRESSD.Z Z1,  K1, Z3          // Pack the positions of the candidates down to the bottom of Z3
    VMOVDQU32     Z3,  (DI)(DX*4)      // Store all 16 lanes at output[DX], only the candidates at the bottom stick
    KMOVW         K1,  R10             // R10 = K1 so we can count the bits
    POPCNTL       R10, R10             // R10 = how many candidates were in this block
    ADDQ          R10, DX              // Move DX forward by that many
    VPADDD        Z2,  Z1, Z1          // Add 16 to every lane of Z1 for the next block
    ADDQ          $16, AX              // i += 16
    LEAQ          16(AX), R10          // R10 = i + 16, where the next block would end
    CMPQ          R10, CX              // Is there a whole block left?
    JBE           BLOCK                // If there is then jump back to BLOCK

  // There are less than 16 vectors left, TAIL does them 1 at a time the same
  // way FILTER does in the AVX version. Written out in Go it is:
  //
  //   for ; i < len(signatures); i++ {
  //     output[candidates] = int32(i)
  //     if signature&signatures[i] != 0 {
  //       candidates++
  //     }
  //   }
  TAIL:
    CMPQ  AX, CX         // Have we checked every vector?
    JAE   FILTERED       // If we have then jump to FILTERED
    MOVQ  (SI)(AX*8), R9 // R9 = signatures[i]
    MOVL  AX, (DI)(DX*4) // output[DX] = i, this only sticks if i is a candidate
    XORL  R10, R10       // Zero R10 so that SETNE below only has to set the low byte
    ANDQ  R8,  R9        // R9 = R9 & our signature, this sets ZF if they have no bits in common
    SETNE R10B           // R10 = 1 if they have a bit in common, 0 if they don't
    ADDQ  R10, DX        // Move DX forward by 1 only if i was a candidate
    INCQ  AX             // i++
    JMP   TAIL           // Jump back to TAIL

  // We're done with the 512 bit registers, clear everything above the bottom
  // 128 bits so the scalar instructions below don't have to deal with them
  FILTERED:
  VZEROUPPER

  // Everything from here down is the same as the AVX version, see the notes
  // there. CANDIDATE is the rest of the loop in sparseNeighbors32Go, but only
  // for the candidates:
  //
  //   start, end := offsets[i], offsets[i+1]
  //   dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //   distance := norm2 + norms[i] - 2*dot
  //   if distance <= epsilon {
  //     output[count] = int32(i)
  //     count++
  //   }
  //
  // Written out in Go the way it is done here, with DI walking over the
  // candidates, R8 as &output[count] and R9 as where the candidates end:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     output[count] = c
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if distance <= epsilon {
  //       count++
  //     }
  //   }
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], where the first neighbor goes

  MOVQ   dense_base+0(FP),     AX  // Load the pointer of dense into AX
  MOVQ   indices_base+112(FP), BX  // Load the pointer of indices into BX
  MOVQ   values_base+136(FP),  CX  // Load the pointer of values into CX
  MOVQ   offsets_base+88(FP),  DX  // Load the pointer of offsets into DX
  MOVQ   norms_base+64(FP),    SI  // Load the pointer of norms into SI
  VMOVSS norm2+32(FP),         X14 // X14 = our squared norm
  VMOVSS epsilon+36(FP),       X15 // X15 = epsilon

  CMPQ DI, R9 // Do we have any candidates?
  JAE  DONE   // If we don't then jump straight to DONE

  // Same as CANDIDATE in the AVX version. c is i in sparseNeighbors32Go:
  //
  //   start, end := offsets[i], offsets[i+1]
  //
  // And X0 and X1 are the 2 sums from sparseDot32Go:
  //
  //   var a, b float32
  CANDIDATE:
    MOVLQSX (DI),          R10      // R10 = c, the candidate we are looking at
    MOVL    R10,           (R8)     // Write c where the next neighbor goes, this only sticks if it is close enough
    VADDSS  (SI)(R10*4),   X14, X3  // X3 = our norm + norms[c], the first part of the distance
    MOVLQSX (DX)(R10*4),   R11      // R11 = offsets[c], where c's entries start
    MOVLQSX 4(DX)(R10*4),  R12      // R12 = offsets[c+1], where c's entries end
    SUBQ    R11,           R12      // R12 = how many entries c has

    VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
    VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

    CMPQ R12, $2 // Do we have at least 2 entries?
    JB   SINGLE  // If we don't then jump straight to SINGLE

    // Same as PAIR in the AVX version. In sparseDot32Go it is:
    //
    //   for ; i+2 <= count; i += 2 {
    //     a += dense[indices[i]] * values[i]
    //     b += dense[indices[i+1]] * values[i+1]
    //   }
    PAIR:
      MOVQ    (BX)(R11*4), R13          // Load the next 2 indicies into R13 at once, the first one is the low 32 bits and the second is the high 32 bits
      MOVLQSX R13,         R10          // R10 = the low 32 bits of R13 sign extended, this is the first index
      SARQ    $32,         R13          // Shift R13 right by 32 keeping the sign, now R13 is the second index
      VMOVSS  (AX)(R10*4), X2           // X2 = dense[R10]
      VMOVSS  (AX)(R13*4), X4           // X4 = dense[R13]
      VMULSS  (CX)(R11*4), X2,  X2      // X2 = values[R11] * X2
      VMULSS  4(CX)(R11*4), X4, X4      // X4 = values[R11+1] * X4
      VADDSS  X2,          X0,  X0      // X0 = X0 + X2
      VADDSS  X4,          X1,  X1      // X1 = X1 + X4
      ADDQ    $2,          R11          // Move R11 forward by 2 entries
      SUBQ    $2,          R12          // Subtract 2 from R12 since we just did 2 entries
      CMPQ    R12,         $2           // Are there still at least 2 left?
      JAE     PAIR                      // If there are then jump back to PAIR, otherwise fall through to SINGLE

    // If the count was odd there is 1 entry left over. In sparseDot32Go that
    // is:
    //
    //   if i < count {
    //     a += dense[indices[i]] * values[i]
    //   }
    SINGLE:
      TESTQ   R12,         R12     // Is R12 zero?
      JZ      DISTANCE             // If it is then there is nothing left, jump straight to DISTANCE
      MOVLQSX (BX)(R11*4), R10     // Load the last index into R10
      VMOVSS  (AX)(R10*4), X2      // X2 = dense[R10]
      VMULSS  (CX)(R11*4), X2, X2  // X2 = values[R11] * X2
      VADDSS  X2,          X0, X0  // X0 = X0 + X2

    // Same as DISTANCE in the AVX version, see the notes there. In
    // sparseNeighbors32Go this is:
    //
    //   dot := sparseDot32Go(...) // which ends with return a + b
    //   distance := norm2 + norms[i] - 2*dot
    //   if distance <= epsilon {
    //     output[count] = int32(i)
    //     count++
    //   }
    DISTANCE:
      VADDSS   X1,  X0, X0      // X0 = X0 + X1, this is our dot product
      VADDSS   X0,  X0, X0      // X0 = X0 + X0, which is exactly 2 * dot without needing a constant
      VSUBSS   X0,  X3, X3      // X3 = X3 - X0, this is the distance
      XORL     R13, R13         // Zero R13 so SETCC only has to set the low byte, this has to happen before the compare because XOR changes the flags
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is cleared if epsilon >= distance
      SETCC    R13B             // R13 = 1 if the distance <= epsilon, 0 if it isn't
      LEAQ     (R8)(R13*4), R8  // Move R8 forward by 1 int32 only if c was close enough
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // DONE is the return count at the end of sparseNeighbors32Go
  DONE:
    MOVQ output_base+160(FP), DI  // Load the pointer of output into DI again
    SUBQ DI,                  R8  // R8 = how many bytes of neighbors we wrote
    SHRQ $2,                  R8  // Divide by 4 to get how many neighbors that is
    MOVQ R8,                  ret+184(FP) // Return how many neighbors we found
    RET
