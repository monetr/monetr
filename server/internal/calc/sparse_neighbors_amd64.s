//go:build amd64 && !nosimd

// vim: set expandtab tabstop=2 shiftwidth=2 textwidth=80 comments=\://:

#include "textflag.h"

// Some notes before reading this file
//
// The plain Go version of this is sparseNeighbors32Go in sparse_neighbors.go,
// and the dot product in the middle of it is sparseDot32Go in sparse.go. The
// notes above each part below have the Go that part lines up with. The
// assembly doesn't go in the same order as the Go though. It checks the
// signature of every vector first, then does the dot product for just the
// ones that made it, and then packs the ones that were close enough down to
// the front of output. So where the order is different the notes have the Go
// from the original loop, and then that same Go written out the way the
// assembly does it
//
// There are 3 versions of this. They only differ in how many vectors BLOCK
// and KEEP look at at once, CANDIDATE is the exact same code in all 3:
//
//   __sparseNeighbors32_AVX:      4 at a time with 128 bit registers, for Ivy
//                                 Bridge (the E5-2667 v2) and anything else
//                                 without AVX-512
//
//   __sparseNeighbors32_AVX512VL: 8 at a time with 256 bit registers, for
//                                 Skylake and Cascade Lake Xeons
//
//   __sparseNeighbors32_AVX512:   16 at a time with 512 bit registers, for
//                                 Zen 4 and Ice Lake or newer
//
// The dot product adds up in the exact same order as sparseDot32Go, X0 gets
// the even entries and X1 gets the odd ones. So the distance is the same down
// to the last bit as the Go version, and a pair that lands right on epsilon
// doesn't change sides depending on which version ran
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VSUBSS X0, X3, X3 is X3 = X3 - X0

// const_sparse_neighbors_lanes is 0 through 15 as int32s. BLOCK adds i to
// these to get the position of each vector it is looking at. The AVX version
// only uses the first 4 and the 256 bit version only uses the first 8
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

// const_sparse_neighbors_pack is how the AVX version packs 4 int32s down to
// the front of a register, since VPCOMPRESSD needs AVX-512. The first 256
// bytes are 16 VPSHUFB masks, one for each combination of the 4 lanes we want
// to drop. Mask m is at m * 16, and it moves every lane whose bit in m is 0
// down to the bottom in order. The rest of the bytes are 0x80, which VPSHUFB
// turns into zeros. The 16 bytes after the masks are how many lanes each mask
// keeps, so we don't need POPCNT
//
// The bit for lane 0 is on the right, so 0001 drops lane 0 and keeps the rest
DATA const_sparse_neighbors_pack<>+0(SB)/8,   $0x0706050403020100 // 0000: drop nothing, keep 0, 1, 2, 3
DATA const_sparse_neighbors_pack<>+8(SB)/8,   $0x0f0e0d0c0b0a0908
DATA const_sparse_neighbors_pack<>+16(SB)/8,  $0x0b0a090807060504 // 0001: drop 0, keep 1, 2, 3
DATA const_sparse_neighbors_pack<>+24(SB)/8,  $0x808080800f0e0d0c
DATA const_sparse_neighbors_pack<>+32(SB)/8,  $0x0b0a090803020100 // 0010: drop 1, keep 0, 2, 3
DATA const_sparse_neighbors_pack<>+40(SB)/8,  $0x808080800f0e0d0c
DATA const_sparse_neighbors_pack<>+48(SB)/8,  $0x0f0e0d0c0b0a0908 // 0011: drop 0, 1, keep 2, 3
DATA const_sparse_neighbors_pack<>+56(SB)/8,  $0x8080808080808080
DATA const_sparse_neighbors_pack<>+64(SB)/8,  $0x0706050403020100 // 0100: drop 2, keep 0, 1, 3
DATA const_sparse_neighbors_pack<>+72(SB)/8,  $0x808080800f0e0d0c
DATA const_sparse_neighbors_pack<>+80(SB)/8,  $0x0f0e0d0c07060504 // 0101: drop 0, 2, keep 1, 3
DATA const_sparse_neighbors_pack<>+88(SB)/8,  $0x8080808080808080
DATA const_sparse_neighbors_pack<>+96(SB)/8,  $0x0f0e0d0c03020100 // 0110: drop 1, 2, keep 0, 3
DATA const_sparse_neighbors_pack<>+104(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+112(SB)/8, $0x808080800f0e0d0c // 0111: drop 0, 1, 2, keep 3
DATA const_sparse_neighbors_pack<>+120(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+128(SB)/8, $0x0706050403020100 // 1000: drop 3, keep 0, 1, 2
DATA const_sparse_neighbors_pack<>+136(SB)/8, $0x808080800b0a0908
DATA const_sparse_neighbors_pack<>+144(SB)/8, $0x0b0a090807060504 // 1001: drop 0, 3, keep 1, 2
DATA const_sparse_neighbors_pack<>+152(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+160(SB)/8, $0x0b0a090803020100 // 1010: drop 1, 3, keep 0, 2
DATA const_sparse_neighbors_pack<>+168(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+176(SB)/8, $0x808080800b0a0908 // 1011: drop 0, 1, 3, keep 2
DATA const_sparse_neighbors_pack<>+184(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+192(SB)/8, $0x0706050403020100 // 1100: drop 2, 3, keep 0, 1
DATA const_sparse_neighbors_pack<>+200(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+208(SB)/8, $0x8080808007060504 // 1101: drop 0, 2, 3, keep 1
DATA const_sparse_neighbors_pack<>+216(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+224(SB)/8, $0x8080808003020100 // 1110: drop 1, 2, 3, keep 0
DATA const_sparse_neighbors_pack<>+232(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+240(SB)/8, $0x8080808080808080 // 1111: drop everything
DATA const_sparse_neighbors_pack<>+248(SB)/8, $0x8080808080808080
DATA const_sparse_neighbors_pack<>+256(SB)/1, $4
DATA const_sparse_neighbors_pack<>+257(SB)/1, $3
DATA const_sparse_neighbors_pack<>+258(SB)/1, $3
DATA const_sparse_neighbors_pack<>+259(SB)/1, $2
DATA const_sparse_neighbors_pack<>+260(SB)/1, $3
DATA const_sparse_neighbors_pack<>+261(SB)/1, $2
DATA const_sparse_neighbors_pack<>+262(SB)/1, $2
DATA const_sparse_neighbors_pack<>+263(SB)/1, $1
DATA const_sparse_neighbors_pack<>+264(SB)/1, $3
DATA const_sparse_neighbors_pack<>+265(SB)/1, $2
DATA const_sparse_neighbors_pack<>+266(SB)/1, $2
DATA const_sparse_neighbors_pack<>+267(SB)/1, $1
DATA const_sparse_neighbors_pack<>+268(SB)/1, $2
DATA const_sparse_neighbors_pack<>+269(SB)/1, $1
DATA const_sparse_neighbors_pack<>+270(SB)/1, $1
DATA const_sparse_neighbors_pack<>+271(SB)/1, $0
GLOBL const_sparse_neighbors_pack<>(SB), RODATA|NOPTR, $272

// func __sparseNeighbors32_AVX(dense []float32, signature uint64, norm2, epsilon float32, signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32, output []int32) int
//
// This is the whole inner loop of DBSCAN's getNeighbors in one call. It
// writes the position of every vector within epsilon of dense into output and
// returns how many there were. We go over the vectors in 3 passes:
//
//   BLOCK:     Checks 4 signatures at a time against ours and writes the
//              position of every vector with at least 1 bit in common into
//              output. TAIL does whatever is left 1 at a time
//
//   CANDIDATE: Goes over just the vectors BLOCK kept. PAIR and SINGLE do the
//              dot product, then DISTANCE works out the distance and turns
//              the ones that are too far away into -1 in output
//
//   KEEP:      Packs everything in output that isn't -1 down to the front, 4
//              at a time. KEEPTAIL does whatever is left 1 at a time
//
// The arguments are slices instead of pointers like __sparseDot32Scalar_AVX.
// This gets called once for each point instead of once for each pair, so the
// cost of copying the arguments doesn't matter anymore
//
// Nothing in here needs more than AVX. The 128 bit integer instructions in
// BLOCK and KEEP are the VEX versions of SSE ones (VPCMPEQQ is SSE4.1 and
// VPSHUFB is SSSE3), and there is no VPBROADCASTD since that needs AVX2. That
// way this runs on Ivy Bridge (E5-2667 v2) which doesn't have AVX2 or BMI, and
// everything newer too
TEXT ·__sparseNeighbors32_AVX(SB), NOSPLIT, $0-192
  MOVQ signature+24(FP),       R8 // Load our signature into R8
  MOVQ signatures_base+40(FP), SI // Load the pointer of signatures into SI
  MOVQ signatures_len+48(FP),  CX // Load the number of vectors into CX
  MOVQ output_base+160(FP),    DI // Load the pointer of output into DI

  XORQ AX, AX // AX = i, the vector we are looking at
  XORQ DX, DX // DX = how many candidates we have written into output so far

  VMOVDDUP signature+24(FP),                   X0  // X0 = our signature in both 64 bit lanes
  VPXOR    X7,  X7, X7                             // X7 = 0, BLOCK compares against it
  VMOVDQU  const_sparse_neighbors_lanes<>(SB), X6  // X6 = [0, 1, 2, 3], the position of each vector in the first block
  MOVL     $4,  R10                                // R10 = 4
  VMOVD    R10, X8                                 // X8 = 4 in the bottom lane
  VPSHUFD  $0,  X8, X8                             // Copy the bottom lane into all 4, X8 = 4 in every lane. We add this to X6 for each block
  LEAQ     const_sparse_neighbors_pack<>(SB), R11  // R11 = the start of the pack table

  CMPQ CX, $4 // Do we have at least 4 vectors?
  JB   TAIL   // If we don't then jump straight to TAIL

  // BLOCK is this part of the loop in sparseNeighbors32Go:
  //
  //   for i := range signatures {
  //     if signature&signatures[i] == 0 {
  //       continue
  //     }
  //     ...
  //   }
  //
  // But it doesn't do the dot product right away, it keeps the candidates in
  // output for CANDIDATE. And it does 4 vectors at a time. Written out in Go
  // each loop is:
  //
  //   for n := range 4 {
  //     if signature&signatures[i+n] != 0 {
  //       output[candidates] = int32(i + n)
  //       candidates++
  //     }
  //   }
  //   i += 4
  //
  // There are no branches in it though. Whether a pair has a bit in common is
  // pretty much random, so a branch would guess wrong a lot and every wrong
  // guess costs about 15 to 20 cycles:
  //
  //   VPAND and VPCMPEQQ against 0 turn each signature into a 64 bit lane
  //   that is all 1s if it has nothing in common with ours. VPACKSSDW squishes
  //   2 registers of those into 1, so we get a 32 bit lane per vector that is
  //   still all 1s or all 0s. VMOVMSKPS takes the top bit of each of those, so
  //   R10 gets a 4 bit mask of the vectors to drop. That is the if
  //
  //   The mask picks 1 of the 16 VPSHUFB masks in const_sparse_neighbors_pack.
  //   X6 is [i, i+1, i+2, i+3], so if only i+1 and i+3 are candidates we get
  //   [i+1, i+3, 0, 0]. We store all 4 lanes and then only move forward by how
  //   many we kept, so the zeros just get written over by the next block. In a
  //   block we have looked at 4 more vectors than we could have written, so
  //   this can never write past the end of output
  //
  // This used to be 1 vector at a time with no SIMD at all. Every vector was 9
  // uops on Ivy Bridge, which can only get 4 through a cycle, and storing to an
  // address with an index in it is 2 uops there on its own. This is about 5
  // uops a vector. On Zen 4 checking 4096 signatures went from 1440ns to 606ns
  // https://uops.info/html-instr/MOV_M32_R32.html
  BLOCK:
    VPAND     (SI)(AX*8),   X0, X1     // X1 = the signatures of i and i+1 & our signature
    VPAND     16(SI)(AX*8), X0, X2     // X2 = the signatures of i+2 and i+3 & our signature
    VPCMPEQQ  X7,  X1, X1              // Each 64 bit lane of X1 is all 1s if it is 0 (nothing in common), all 0s if it isn't
    VPCMPEQQ  X7,  X2, X2              // Same thing for X2
    VPACKSSDW X2,  X1, X1              // Pack X1 and X2 into X1, now each 32 bit lane is all 1s if that vector has nothing in common
    VMOVMSKPS X1,  R10                 // R10 = the top bit of each 32 bit lane, bit n = 1 if vector i+n gets dropped
    MOVQ      R10, R12                 // Copy the mask into R12
    SHLQ      $4,  R12                 // Multiply by 16, R12 = where the VPSHUFB mask for this combination starts in the table
    VPSHUFB   (R11)(R12*1), X6, X3     // Pack the positions of the candidates down to the bottom of X3
    VMOVDQU   X3,  (DI)(DX*4)          // Store all 4 lanes at output[DX], only the candidates at the bottom stick
    MOVBLZX   256(R11)(R10*1), R10     // R10 = how many candidates were in this block, from the counts after the masks
    ADDQ      R10, DX                  // Move DX forward by that many
    VPADDD    X8,  X6, X6              // Add 4 to every lane of X6 for the next block
    ADDQ      $4,  AX                  // i += 4
    LEAQ      4(AX), R10               // R10 = i + 4, where the next block would end
    CMPQ      R10, CX                  // Is there a whole block left?
    JBE       BLOCK                    // If there is then jump back to BLOCK

  // There are less than 4 vectors left, TAIL does them 1 at a time. It always
  // writes i into output and then only moves forward if i made it, otherwise
  // the next one just writes over it. Written out in Go it is:
  //
  //   for ; i < len(signatures); i++ {
  //     output[candidates] = int32(i)
  //     if signature&signatures[i] != 0 {
  //       candidates++
  //     }
  //   }
  //
  // The if is a SETNE and an ADDQ though, not a jump
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
  // But it doesn't pack the neighbors down to the front as it goes, it only
  // marks the ones that are too far away with -1 right where they are. KEEP
  // packs them afterwards. Written out in Go that is this, with DI walking
  // over the candidates and R9 as where the candidates end:
  //
  //   for j, c := range output[:candidates] {
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if !(distance <= epsilon) {
  //       output[j] = -1
  //     }
  //   }
  //
  // This used to write c to output[count] at the top of every loop, and only
  // move count forward if it was close enough. But count isn't known until
  // DISTANCE is done with the candidate before it, which is the end of a long
  // chain of loads and math. Until then the CPU doesn't know where that store
  // goes, and on Zen 4 that held up the next candidate too. Now the store
  // always goes to where we just read c from, which is known right away. I
  // couldn't check that with perf counters, but on Zen 4 it made a 4096
  // vector call to __sparseNeighbors32_AVX512 go from 3710ns to 2950ns, KEEP
  // included
  FILTERED:
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], KEEP writes the neighbors starting from here

  MOVQ   dense_base+0(FP),     AX  // Load the pointer of dense into AX
  MOVQ   indices_base+112(FP), BX  // Load the pointer of indices into BX
  MOVQ   values_base+136(FP),  CX  // Load the pointer of values into CX
  MOVQ   offsets_base+88(FP),  DX  // Load the pointer of offsets into DX
  MOVQ   norms_base+64(FP),    SI  // Load the pointer of norms into SI
  VMOVSS norm2+32(FP),         X14 // X14 = our squared norm
  VMOVSS epsilon+36(FP),       X15 // X15 = epsilon

  CMPQ DI, R9 // Do we have any candidates?
  JAE  DONE   // If we don't then jump straight to DONE

  // Each loop of CANDIDATE is 1 candidate, c. c is i in sparseNeighbors32Go,
  // so this is:
  //
  //   start, end := offsets[i], offsets[i+1]
  //
  // We also get the norm2 + norms[i] part of the distance going here since it
  // doesn't need the dot product. R11 is start and R12 is end - start. X0 and
  // X1 are the 2 sums from sparseDot32Go:
  //
  //   var a, b float32
  CANDIDATE:
    MOVLQSX (DI),          R10      // R10 = c, the candidate we are looking at
    VADDSS  (SI)(R10*4),   X14, X3  // X3 = our norm + norms[c], the first part of the distance
    MOVLQSX (DX)(R10*4),   R11      // R11 = offsets[c], where c's entries start
    MOVLQSX 4(DX)(R10*4),  R12      // R12 = offsets[c+1], where c's entries end
    SUBQ    R11,           R12      // R12 = how many entries c has

    VXORPS X0, X0, X0 // Zero out X0, the sum for the first entry of each pair
    VXORPS X1, X1, X1 // Zero out X1, the sum for the second entry of each pair

    CMPQ R12, $2 // Do we have at least 2 entries?
    JB   SINGLE  // If we don't then jump straight to SINGLE

    // PAIR is the same loop as __sparseDot32Scalar_AVX, see the notes there
    // for why it does 2 at a time. In sparseDot32Go it is this, where indices
    // and values there are indices[start:end] and values[start:end] here. So
    // R11 is start + i and R12 is count - i:
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
    // Here the if is the output[j] = -1 from the second Go loop above, and
    // there is no branch for it. VUCOMISS sets CF if epsilon is less than the
    // distance, and also if either of them is NaN. SBBL of a register from
    // itself turns CF into -1 or 0, and ORing that into output turns c into -1
    // or leaves it alone. So c only stays if distance <= epsilon in Go, NaN
    // included. c is never negative so it can't be mixed up with the -1
    DISTANCE:
      VADDSS   X1,  X0, X0      // X0 = X0 + X1, this is our dot product
      VADDSS   X0,  X0, X0      // X0 = X0 + X0, which is exactly 2 * dot without needing a constant
      VSUBSS   X0,  X3, X3      // X3 = X3 - X0, this is the distance
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is set if epsilon < distance or either is NaN
      SBBL     R13, R13         // R13 = R13 - R13 - CF, so -1 if c is too far away and 0 if it is close enough
      ORL      R13, (DI)        // output[j] = output[j] | R13, c stays c if it is close enough and becomes -1 if it isn't
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // KEEP packs every candidate that isn't -1 down to the front of output.
  // Written out in Go it is this, with DI walking over the candidates again
  // and R8 as &output[count]:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     if c >= 0 {
  //       output[count] = c
  //       count++
  //     }
  //   }
  //
  // It's the same trick as BLOCK, 4 at a time. -1 is the only thing in output
  // with the top bit set, so VMOVMSKPS gives us the mask of what to drop
  // straight from the candidates with nothing else to do. The store at R8
  // covers 4 lanes, but R8 is never ahead of DI. So it only ever writes over
  // candidates we have already loaded, and never goes past R9
  MOVQ output_base+160(FP),               DI  // DI = &output[0] again, KEEP reads from the start
  LEAQ const_sparse_neighbors_pack<>(SB), R11 // R11 = the start of the pack table again, CANDIDATE used R11 for something else
  LEAQ 16(DI),                            R10 // R10 = where the first 4 would end
  CMPQ R10,                               R9  // Are there at least 4 candidates?
  JA   KEEPTAIL                               // If there aren't then jump straight to KEEPTAIL

  KEEP:
    VMOVDQU   (DI), X4                 // X4 = the next 4 candidates, each one is either c or -1
    VMOVMSKPS X4,   R10                // R10 = the top bit of each, bit n = 1 if that one is -1 and gets dropped
    MOVQ      R10,  R12                // Copy the mask into R12
    SHLQ      $4,   R12                // Multiply by 16, R12 = where the VPSHUFB mask for this combination starts in the table
    VPSHUFB   (R11)(R12*1), X4, X5     // Pack the ones we are keeping down to the bottom of X5
    VMOVDQU   X5,   (R8)               // Store all 4 lanes where the next neighbor goes, only the neighbors at the bottom stick
    MOVBLZX   256(R11)(R10*1), R10     // R10 = how many we kept, from the counts after the masks
    LEAQ      (R8)(R10*4), R8          // Move R8 forward by that many
    ADDQ      $16,  DI                 // Move DI forward to the next 4
    LEAQ      16(DI), R10              // R10 = where the next 4 would end
    CMPQ      R10,  R9                 // Are there at least 4 left?
    JBE       KEEP                     // If there are then jump back to KEEP

  // There are less than 4 candidates left, KEEPTAIL does them 1 at a time.
  // Same trick as TAIL, it always writes the candidate and only moves R8
  // forward if it isn't -1
  KEEPTAIL:
    CMPQ DI,  R9          // Have we looked at every candidate?
    JAE  DONE             // If we have then jump to DONE
    MOVL (DI), R10        // R10 = the next candidate, either c or -1
    MOVL R10, (R8)        // Write it where the next neighbor goes, this only sticks if it isn't -1
    NOTL R10              // Flip every bit of R10, now the top bit is 1 if it was c and 0 if it was -1
    SHRL $31, R10         // Shift the top bit down, R10 = 1 if we are keeping it and 0 if we aren't
    LEAQ (R8)(R10*4), R8  // Move R8 forward by 1 int32 only if we kept it
    ADDQ $4,  DI          // Move DI forward to the next candidate
    JMP  KEEPTAIL         // Jump back to KEEPTAIL

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
// Same as __sparseNeighbors32_AVX but BLOCK and KEEP do 8 at a time with
// AVX-512, for Skylake and Cascade Lake Xeons. CANDIDATE is the exact same
// code
//
// What is different from the AVX version:
//
//   - BLOCK uses VPTESTMQ for the if and VPCOMPRESSD to do the packing, so
//     there is no pack table. See the notes above BLOCK
//   - KEEP does the same thing with VPTESTNMD and VPCOMPRESSD
//   - BLOCK and KEEP leave the top halves of the YMM registers dirty, so there
//     is a VZEROUPPER after each of them
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

  // BLOCK does the same thing as BLOCK in the AVX version, which is this part
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
  //   for each one that isn't zero. Doing it twice and shifting one of them
  //   up gets us 8 bits for 8 vectors in K1, that is the if
  //
  //   VPCOMPRESSD takes the lanes of Y1 that have their bit set in K1 and
  //   packs them down to the bottom of Y3. Y1 is [i, i+1, ..., i+7], so if
  //   only i+2 and i+5 are candidates we get [i+2, i+5, 0, 0, 0, 0, 0, 0]. We
  //   store all 8 lanes and then only move forward by how many bits were set,
  //   so the zeros just get written over by the next block. In a block we
  //   have looked at 8 more vectors than we could have written, so this can
  //   never write past the end of output
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

  // There are less than 8 vectors left, TAIL does them 1 at a time the same
  // way TAIL does in the AVX version. Written out in Go it is:
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

  // We're done with the 256 bit registers for now, clear the top halves so the
  // scalar instructions below don't have to deal with them
  FILTERED:
  VZEROUPPER

  // Everything from here to KEEP is the same as the AVX version, see the notes
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
  // candidates and R9 as where the candidates end:
  //
  //   for j, c := range output[:candidates] {
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if !(distance <= epsilon) {
  //       output[j] = -1
  //     }
  //   }
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], KEEP writes the neighbors starting from here

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
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is set if epsilon < distance or either is NaN
      SBBL     R13, R13         // R13 = R13 - R13 - CF, so -1 if c is too far away and 0 if it is close enough
      ORL      R13, (DI)        // output[j] = output[j] | R13, c stays c if it is close enough and becomes -1 if it isn't
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // KEEP is the same as KEEP in the AVX version, see the notes there. Written
  // out in Go it is:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     if c >= 0 {
  //       output[count] = c
  //       count++
  //     }
  //   }
  //
  // But 8 at a time, and with VPCOMPRESSD instead of the pack table. VPTESTNMD
  // against just the top bit sets a bit in K1 for every lane that isn't -1,
  // and those are the ones VPCOMPRESSD keeps
  MOVQ         output_base+160(FP), DI  // DI = &output[0] again, KEEP reads from the start
  MOVL         $0x80000000,         R10 // R10 = just the top bit of an int32
  VPBROADCASTD R10,                 Y5  // Y5 = just the top bit in all 8 lanes
  LEAQ         32(DI),              R10 // R10 = where the first 8 would end
  CMPQ         R10,                 R9  // Are there at least 8 candidates?
  JA           KEEPTAIL                 // If there aren't then jump straight to KEEPTAIL

  KEEP:
    VMOVDQU32     (DI), Y4             // Y4 = the next 8 candidates, each one is either c or -1
    VPTESTNMD     Y5,   Y4, K1         // K1 bit n = 1 if lane n doesn't have the top bit set, so it is c and not -1
    VPCOMPRESSD.Z Y4,   K1, Y6         // Pack the ones we are keeping down to the bottom of Y6
    VMOVDQU32     Y6,   (R8)           // Store all 8 lanes where the next neighbor goes, only the neighbors at the bottom stick
    KMOVW         K1,   R10            // R10 = K1 so we can count the bits
    POPCNTL       R10,  R10            // R10 = how many we kept
    LEAQ          (R8)(R10*4), R8      // Move R8 forward by that many
    ADDQ          $32,  DI             // Move DI forward to the next 8
    LEAQ          32(DI), R10          // R10 = where the next 8 would end
    CMPQ          R10,  R9             // Are there at least 8 left?
    JBE           KEEP                 // If there are then jump back to KEEP

  // There are less than 8 candidates left, KEEPTAIL does them 1 at a time the
  // same way KEEPTAIL does in the AVX version
  KEEPTAIL:
    CMPQ DI,  R9          // Have we looked at every candidate?
    JAE  KEPT             // If we have then jump to KEPT
    MOVL (DI), R10        // R10 = the next candidate, either c or -1
    MOVL R10, (R8)        // Write it where the next neighbor goes, this only sticks if it isn't -1
    NOTL R10              // Flip every bit of R10, now the top bit is 1 if it was c and 0 if it was -1
    SHRL $31, R10         // Shift the top bit down, R10 = 1 if we are keeping it and 0 if we aren't
    LEAQ (R8)(R10*4), R8  // Move R8 forward by 1 int32 only if we kept it
    ADDQ $4,  DI          // Move DI forward to the next candidate
    JMP  KEEPTAIL         // Jump back to KEEPTAIL

  // KEEP used the 256 bit registers again, so clear the top halves before we
  // go back to Go
  KEPT:
  VZEROUPPER

  // DONE is the return count at the end of sparseNeighbors32Go
  DONE:
    MOVQ output_base+160(FP), DI  // Load the pointer of output into DI again
    SUBQ DI,                  R8  // R8 = how many bytes of neighbors we wrote
    SHRQ $2,                  R8  // Divide by 4 to get how many neighbors that is
    MOVQ R8,                  ret+184(FP) // Return how many neighbors we found
    RET

// func __sparseNeighbors32_AVX512(dense []float32, signature uint64, norm2, epsilon float32, signatures []uint64, norms []float32, offsets []int32, indices []int32, values []float32, output []int32) int
//
// Same as the AVX512VL version but with 512 bit ZMM registers, so BLOCK and
// KEEP do 16 at a time instead of 8. This is the one for Zen 4 and Ice Lake or
// newer, where using 512 bit registers doesn't slow the core down
//
// What is different from the AVX512VL version:
//
//   - 1 VPTESTMQ covers 8 signatures instead of 4, and KUNPCKBW glues the 8
//     bits from 2 of them together into 16 bits for 16 vectors
//   - Z1 is [i, i+1, ..., i+15] and Z2 adds 16 to it for each block
//   - TAIL can have up to 15 vectors left instead of 7, and KEEPTAIL up to 15
//     candidates
//
// Zen 4 splits a 512 bit instruction into 2 256 bit halves, but it still came
// out ahead because there are half as many instructions to get through. For
// 4096 signatures BLOCK and TAIL took 268ns, vs 395ns with 256 bit registers
// on a 7950X. That is only a small part of the whole call though, most of the
// time is in CANDIDATE
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
  // way TAIL does in the AVX version. Written out in Go it is:
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

  // We're done with the 512 bit registers for now, clear everything above the
  // bottom 128 bits so the scalar instructions below don't have to deal with
  // them
  FILTERED:
  VZEROUPPER

  // Everything from here to KEEP is the same as the AVX version, see the notes
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
  // candidates and R9 as where the candidates end:
  //
  //   for j, c := range output[:candidates] {
  //     start, end := offsets[c], offsets[c+1]
  //     dot := sparseDot32Go(dense, indices[start:end], values[start:end])
  //     distance := norm2 + norms[c] - 2*dot
  //     if !(distance <= epsilon) {
  //       output[j] = -1
  //     }
  //   }
  LEAQ (DI)(DX*4), R9 // R9 = &output[DX], where the candidates end
  MOVQ DI,         R8 // R8 = &output[0], KEEP writes the neighbors starting from here

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
      VUCOMISS X3,  X15         // Compare epsilon to the distance, CF is set if epsilon < distance or either is NaN
      SBBL     R13, R13         // R13 = R13 - R13 - CF, so -1 if c is too far away and 0 if it is close enough
      ORL      R13, (DI)        // output[j] = output[j] | R13, c stays c if it is close enough and becomes -1 if it isn't
      ADDQ     $4,  DI          // Move DI forward to the next candidate
      CMPQ     DI,  R9          // Have we done every candidate?
      JB       CANDIDATE        // If we haven't then jump back to CANDIDATE

  // Same as KEEP in the AVX512VL version but 16 at a time, see the notes
  // there. Written out in Go it is:
  //
  //   count := 0
  //   for _, c := range output[:candidates] {
  //     if c >= 0 {
  //       output[count] = c
  //       count++
  //     }
  //   }
  MOVQ         output_base+160(FP), DI  // DI = &output[0] again, KEEP reads from the start
  MOVL         $0x80000000,         R10 // R10 = just the top bit of an int32
  VPBROADCASTD R10,                 Z5  // Z5 = just the top bit in all 16 lanes
  LEAQ         64(DI),              R10 // R10 = where the first 16 would end
  CMPQ         R10,                 R9  // Are there at least 16 candidates?
  JA           KEEPTAIL                 // If there aren't then jump straight to KEEPTAIL

  KEEP:
    VMOVDQU32     (DI), Z4             // Z4 = the next 16 candidates, each one is either c or -1
    VPTESTNMD     Z5,   Z4, K1         // K1 bit n = 1 if lane n doesn't have the top bit set, so it is c and not -1
    VPCOMPRESSD.Z Z4,   K1, Z6         // Pack the ones we are keeping down to the bottom of Z6
    VMOVDQU32     Z6,   (R8)           // Store all 16 lanes where the next neighbor goes, only the neighbors at the bottom stick
    KMOVW         K1,   R10            // R10 = K1 so we can count the bits
    POPCNTL       R10,  R10            // R10 = how many we kept
    LEAQ          (R8)(R10*4), R8      // Move R8 forward by that many
    ADDQ          $64,  DI             // Move DI forward to the next 16
    LEAQ          64(DI), R10          // R10 = where the next 16 would end
    CMPQ          R10,  R9             // Are there at least 16 left?
    JBE           KEEP                 // If there are then jump back to KEEP

  // There are less than 16 candidates left, KEEPTAIL does them 1 at a time
  // the same way KEEPTAIL does in the AVX version
  KEEPTAIL:
    CMPQ DI,  R9          // Have we looked at every candidate?
    JAE  KEPT             // If we have then jump to KEPT
    MOVL (DI), R10        // R10 = the next candidate, either c or -1
    MOVL R10, (R8)        // Write it where the next neighbor goes, this only sticks if it isn't -1
    NOTL R10              // Flip every bit of R10, now the top bit is 1 if it was c and 0 if it was -1
    SHRL $31, R10         // Shift the top bit down, R10 = 1 if we are keeping it and 0 if we aren't
    LEAQ (R8)(R10*4), R8  // Move R8 forward by 1 int32 only if we kept it
    ADDQ $4,  DI          // Move DI forward to the next candidate
    JMP  KEEPTAIL         // Jump back to KEEPTAIL

  // KEEP used the 512 bit registers again, so clear everything above the
  // bottom 128 bits before we go back to Go
  KEPT:
  VZEROUPPER

  // DONE is the return count at the end of sparseNeighbors32Go
  DONE:
    MOVQ output_base+160(FP), DI  // Load the pointer of output into DI again
    SUBQ DI,                  R8  // R8 = how many bytes of neighbors we wrote
    SHRQ $2,                  R8  // Divide by 4 to get how many neighbors that is
    MOVQ R8,                  ret+184(FP) // Return how many neighbors we found
    RET
