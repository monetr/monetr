#include "textflag.h"

// const_negate_imaginary flips the sign bit of the imaginary half of all four
// complex numbers held in a ZMM register and leaves the real halves alone. A
// complex128 is stored as (real, imaginary), so the odd float64 in every pair
// is the one that gets its sign bit toggled.
DATA const_negate_imaginary<>+0(SB)/8,  $0x0000000000000000
DATA const_negate_imaginary<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_imaginary<>+16(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+24(SB)/8, $0x8000000000000000
DATA const_negate_imaginary<>+32(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+40(SB)/8, $0x8000000000000000
DATA const_negate_imaginary<>+48(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+56(SB)/8, $0x8000000000000000
GLOBL const_negate_imaginary<>(SB), (RODATA+NOPTR), $64

// func __fastFourierTransform_AVX512(dst, src []complex128)
//
// An iterative Cooley-Tukey fast Fourier transform. Nothing recurses, nothing
// allocates and nothing trigonometric happens here. Every root of unity and
// every bit reversal offset was worked out at build time and lives in
// fourier_twiddles_amd64.s, so all this does is stream through those tables.
//
// A complex128 is two float64s laid out as (real, imaginary), so a 512-bit ZMM
// register holds exactly four complex numbers and every butterfly below works
// on four points at a time.
//
// The transform runs in two parts:
//
//   1. One pass that reads src, does the bit reversal permutation, and folds
//      the first two radix-2 stages into it as a single radix-4 butterfly. The
//      twiddle factors for those two stages are all either 1 or -i so there is
//      no table to read, and doing the permutation on the way out of src means
//      no scratch buffer is ever needed.
//
//   2. A stage loop over dst that runs the remaining radix-2 stages in place,
//      from eight points wide up to the full transform. Every one of those
//      stages has a half width of at least four complex numbers, so they all
//      vectorize cleanly with no in register shuffling.
TEXT ·__fastFourierTransform_AVX512(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI // Load the pointer to the output buffer, this is the buffer everything happens in after the first pass.
  MOVQ src_base+24(FP), SI // Load the pointer to the input buffer, it is only ever read.
  MOVQ dst_len+8(FP),   AX // Load the number of complex128 points in the transform.

  VMOVUPD const_negate_imaginary<>(SB), Z16 // Keep the sign flip mask in a register for the whole of the first pass.

  LEAQ ·fourierScatter4096(SB), BX // Point BX at the precomputed bit reversal offsets.

  // The first pass reads four points that are each a quarter of the transform
  // apart, so precompute those three distances in bytes. A complex128 is 16
  // bytes wide, which is where the multipliers come from: n/4 points is n*4
  // bytes, n/2 points is n*8 bytes and 3n/4 points is n*12 bytes.
  MOVQ AX, R9
  SHLQ $2, R9   // R9 = n*4, the byte distance to the point a quarter of the way through src.
  MOVQ R9, R8
  ADDQ R9, R8   // R8 = n*8, the byte distance to the halfway point of src.
  MOVQ R8, R10
  ADDQ R9, R10  // R10 = n*12, the byte distance to the point three quarters of the way through src.

  MOVQ AX, CX
  SHRQ $4, CX   // The pass covers the first quarter of src four points at a time, so n/16 iterations.

  // Every iteration of this loop takes four source groups and produces four
  // finished groups of four outputs. Reading a0 through a3 with a stride of a
  // quarter of the transform is what the bit reversal actually asks for: the
  // low two bits of a destination index become the top two bits of the source
  // index, which is exactly a quarter, a half and three quarters of the way in.
  PASS1:
    VMOVUPD (SI),       Z0 // Load a0, four complex numbers from the start of src.
    VMOVUPD (SI)(R8*1), Z1 // Load a1 from halfway through src.
    VMOVUPD (SI)(R9*1), Z2 // Load a2 from a quarter of the way through src.
    VMOVUPD (SI)(R10*1), Z3 // Load a3 from three quarters of the way through src.

    // The first radix-2 stage. Its twiddle factor is 1 for every butterfly so
    // this is just four sums and four differences with no multiplies at all.
    VADDPD Z1, Z0, Z4 // b0 = a0 + a1
    VSUBPD Z1, Z0, Z5 // b1 = a0 - a1
    VADDPD Z3, Z2, Z6 // b2 = a2 + a3
    VSUBPD Z3, Z2, Z7 // b3 = a2 - a3

    // The second radix-2 stage needs the twiddle factors 1 and -i. Multiplying
    // a complex number by -i is (x + iy) * -i = y - ix, which is a swap of the
    // real and imaginary halves followed by negating the new imaginary half.
    // Two instructions and no multiplier involved.
    VPERMILPD $0x55, Z7, Z7 // Swap the real and imaginary halves of b3, giving (b3.imag, b3.real).
    VPXORQ    Z16, Z7, Z7   // Flip the sign of the new imaginary half, giving (b3.imag, -b3.real), which is -i*b3.

    VADDPD Z6, Z4, Z8  // c0 = b0 + b2
    VADDPD Z7, Z5, Z9  // c1 = b1 + (-i * b3)
    VSUBPD Z6, Z4, Z10 // c2 = b0 - b2
    VSUBPD Z7, Z5, Z11 // c3 = b1 - (-i * b3)

    // At this point Z8 through Z11 are laid out the wrong way round for storing.
    // Each register holds one output slot for four different groups, but memory
    // wants all four output slots of a single group contiguous. That is a 4x4
    // transpose of 128-bit lanes, and VSHUFF64X2 does it in eight instructions.
    VSHUFF64X2 $0x44, Z9,  Z8,  Z12 // t0 = c0 and c1 for the first two groups.
    VSHUFF64X2 $0xEE, Z9,  Z8,  Z13 // t1 = c0 and c1 for the last two groups.
    VSHUFF64X2 $0x44, Z11, Z10, Z14 // t2 = c2 and c3 for the first two groups.
    VSHUFF64X2 $0xEE, Z11, Z10, Z15 // t3 = c2 and c3 for the last two groups.
    VSHUFF64X2 $0x88, Z14, Z12, Z0  // The complete first group, c0 through c3.
    VSHUFF64X2 $0xDD, Z14, Z12, Z1  // The complete second group.
    VSHUFF64X2 $0x88, Z15, Z13, Z2  // The complete third group.
    VSHUFF64X2 $0xDD, Z15, Z13, Z3  // The complete fourth group.

    // The four groups all belong at unrelated places in dst, so read where each
    // one goes out of the precomputed table. Every store writes a full 64 byte
    // group, which is one whole cache line when dst is aligned.
    MOVL 0(BX),  R11 // Byte offset of the first group.
    MOVL 4(BX),  R12 // Byte offset of the second group.
    MOVL 8(BX),  R13 // Byte offset of the third group.
    MOVL 12(BX), DX  // Byte offset of the fourth group.
    VMOVUPD Z0, (DI)(R11*1)
    VMOVUPD Z1, (DI)(R12*1)
    VMOVUPD Z2, (DI)(R13*1)
    VMOVUPD Z3, (DI)(DX*1)

    ADDQ $64, SI // Advance src by four complex numbers.
    ADDQ $16, BX // Advance the offset table by four entries.
    SUBQ $1, CX  // One less group of four to go.
    JNZ  PASS1   // Keep going until the whole first quarter of src has been read.

  // Everything from here on happens in place in dst.
  //
  // Every remaining stage is the same butterfly: the second half of each block
  // gets multiplied by a twiddle factor, then the sum and the difference of the
  // two halves are written back. The complex multiply is done as
  //
  //   t.real = v.real*w.real - v.imag*w.imag
  //   t.imag = v.imag*w.real + v.real*w.imag
  //
  // which VFMADDSUB231PD does in a single instruction once the twiddle has been
  // splatted into (w.real, w.real) and (w.imag, w.imag), because it subtracts on
  // the even slots and adds on the odd ones. The twiddle factors are stored
  // interleaved exactly like the data, so VMOVDDUP and VPERMILPD produce both
  // splats out of the same 64 bytes without any extra loads.
  LEAQ ·fourierTwiddles4096(SB), BX // Point BX at the base of the twiddle table.
  MOVQ dst_len+8(FP), DX
  SHLQ $4, DX         // DX = n*16, the total size of the buffer in bytes.
  LEAQ (DI)(DX*1), AX // AX = one byte past the end of the buffer, the stopping point for every block loop.
  SHRQ $1, DX         // DX = n*8, the half width of the final stage and so the bound for the stage loop.

  // The eight point stage gets its own loop. Its half is only four complex
  // numbers, which is a single register, so the generic loop below would spend
  // more instructions setting up each block than doing work in it. It also uses
  // the same four twiddle factors for every one of its blocks, so they can be
  // hoisted into registers once and the whole stage becomes a flat walk over
  // the buffer, two blocks at a time.
  VMOVDDUP  (BX), Z7        // (w.real, w.real) for the four twiddle factors of the eight point stage.
  VPERMILPD $0xFF, (BX), Z8 // (w.imag, w.imag) for the same four.
  MOVQ DI, R8               // R8 walks the buffer.

  STAGE8:
    VMOVUPD        64(R8), Z1     // v for the first block.
    VMOVUPD        192(R8), Z11   // v for the second block.
    VPERMILPD      $0x55, Z1, Z2  // (v.imag, v.real) for the first block.
    VPERMILPD      $0x55, Z11, Z12 // (v.imag, v.real) for the second block.
    VMULPD         Z8, Z2, Z2     // (v.imag*w.imag, v.real*w.imag)
    VMULPD         Z8, Z12, Z12
    VFMADDSUB231PD Z7, Z1, Z2     // t = v*w for the first block.
    VFMADDSUB231PD Z7, Z11, Z12   // t = v*w for the second block.
    VMOVUPD        (R8), Z0       // u for the first block.
    VMOVUPD        128(R8), Z10   // u for the second block.
    VADDPD         Z2, Z0, Z5     // u + t
    VADDPD         Z12, Z10, Z15
    VSUBPD         Z2, Z0, Z6     // u - t
    VSUBPD         Z12, Z10, Z16
    VMOVUPD        Z5, (R8)
    VMOVUPD        Z15, 128(R8)
    VMOVUPD        Z6, 64(R8)
    VMOVUPD        Z16, 192(R8)
    ADDQ $256, R8 // Two blocks of eight points each.
    CMPQ R8, AX
    JCS  STAGE8

  // And now every stage from sixteen points wide up to the whole transform.
  // From here on a half is at least eight complex numbers, so the butterfly
  // loop always runs an even number of registers and can be unrolled two at a
  // time with no leftovers to clean up afterwards.
  MOVQ $128, R12 // The half of the sixteen point stage is eight complex numbers, or 128 bytes.

  STAGE:
    // A stage whose half is h entries reads h twiddle factors starting at entry
    // h-4, because the halves double every stage and 4 + 8 + ... + h/2 is h-4.
    // In bytes that is just the byte width of the stage's half minus one
    // register, so there is no table of table offsets to keep around.
    LEAQ -64(BX)(R12*1), R13 // R13 = where this stage's twiddle factors start.
    MOVQ DI, R8              // R8 walks the start of each block in the buffer.

    BLOCK:
      MOVQ R13, R9          // The twiddle cursor restarts at the top for every block.
      MOVQ R8,  R10         // R10 walks the first half of the block.
      LEAQ (R8)(R12*1), R11 // R11 walks the second half of the block.
      MOVQ R12, SI          // SI counts down the bytes of butterflies left in this block.

      BUTTERFLY:
        // Two independent butterflies per iteration. They share nothing, so the
        // CPU can overlap them and hide the latency of the multiplies.
        VMOVUPD        (R11), Z1        // v, four complex numbers from the second half of the block.
        VMOVUPD        64(R11), Z11     // The next four.
        VPERMILPD      $0x55, Z1, Z2    // (v.imag, v.real), the swapped copy the imaginary term needs.
        VPERMILPD      $0x55, Z11, Z12
        VPERMILPD      $0xFF, (R9), Z4  // (w.imag, w.imag) for each of the four twiddle factors.
        VPERMILPD      $0xFF, 64(R9), Z14
        VMULPD         Z4, Z2, Z2       // (v.imag*w.imag, v.real*w.imag)
        VMULPD         Z14, Z12, Z12
        VMOVDDUP       (R9), Z3         // (w.real, w.real) for each of the four twiddle factors.
        VMOVDDUP       64(R9), Z13
        VFMADDSUB231PD Z3, Z1, Z2       // t = v*w. Subtracting on the real slots and adding on the imaginary ones is exactly a complex multiply.
        VFMADDSUB231PD Z13, Z11, Z12

        VMOVUPD (R10), Z0     // u, four complex numbers from the first half of the block.
        VMOVUPD 64(R10), Z10  // The next four.
        VADDPD  Z2, Z0, Z5    // u + t
        VADDPD  Z12, Z10, Z15
        VSUBPD  Z2, Z0, Z6    // u - t
        VSUBPD  Z12, Z10, Z16
        VMOVUPD Z5, (R10)     // The first half of the block becomes u + t.
        VMOVUPD Z15, 64(R10)
        VMOVUPD Z6, (R11)     // The second half of the block becomes u - t.
        VMOVUPD Z16, 64(R11)

        ADDQ $128, R9  // Advance to the next eight twiddle factors.
        ADDQ $128, R10 // Advance the first half cursor by eight complex numbers.
        ADDQ $128, R11 // Advance the second half cursor by eight complex numbers.
        SUBQ $128, SI  // That many fewer bytes of butterflies left in this block.
        JNZ  BUTTERFLY

      LEAQ (R8)(R12*2), R8 // Step to the next block, which is two halves further along.
      CMPQ R8, AX
      JCS  BLOCK           // Keep going while the block pointer is still inside the buffer.

    SHLQ $1, R12 // The next stage is twice as wide.
    CMPQ R12, DX
    JLS  STAGE   // Keep going while the stage still fits inside the transform.

  RET // The transform is finished and sitting in dst.
