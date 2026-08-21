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

// const_negate_all flips the sign bit of every float64 in a ZMM register. The
// radix-4 passes use it to turn the splatted real half of w2 into the splatted
// imaginary half of w3, which is what multiplying a twiddle factor by -i comes
// down to once the halves have already been splatted apart.
DATA const_negate_all<>+0(SB)/8,  $0x8000000000000000
DATA const_negate_all<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_all<>+16(SB)/8, $0x8000000000000000
DATA const_negate_all<>+24(SB)/8, $0x8000000000000000
DATA const_negate_all<>+32(SB)/8, $0x8000000000000000
DATA const_negate_all<>+40(SB)/8, $0x8000000000000000
DATA const_negate_all<>+48(SB)/8, $0x8000000000000000
DATA const_negate_all<>+56(SB)/8, $0x8000000000000000
GLOBL const_negate_all<>(SB), (RODATA+NOPTR), $64

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
//   2. Five radix-4 passes over dst, in place, each one folding two of the ten
//      remaining radix-2 stages together. 4096 is 4^6, so the first pass plus
//      these five cover all twelve stages exactly with nothing left over.
//
// Folding pairs of stages together is the whole point of part two. On Zen 4 a
// 512 bit store only retires every other cycle, so what limits this transform
// is how many times it walks the buffer, not how much arithmetic it does.
// Halving the number of passes halves both the loads and the stores.
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
	// index, which is exactly a quarter, a half and three quarters of the way
	// in.
  PASS1:
    VMOVUPD (SI),        Z0 // Load a0, four complex numbers from the start of src.
    VMOVUPD (SI)(R8*1),  Z1 // Load a1 from halfway through src.
    VMOVUPD (SI)(R9*1),  Z2 // Load a2 from a quarter of the way through src.
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
    VPXORQ    Z16,   Z7, Z7 // Flip the sign of the new imaginary half, giving (b3.imag, -b3.real), which is -i*b3.

    VADDPD Z6, Z4, Z8  // c0 = b0 + b2
    VADDPD Z7, Z5, Z9  // c1 = b1 + (-i * b3)
    VSUBPD Z6, Z4, Z10 // c2 = b0 - b2
    VSUBPD Z7, Z5, Z11 // c3 = b1 - (-i * b3)

		// At this point Z8 through Z11 are laid out the wrong way round for
		// storing. Each register holds one output slot for four different groups,
		// but memory wants all four output slots of a single group contiguous.
		// That is a 4x4 transpose of 128-bit lanes, and VSHUFF64X2 does it in
		// eight instructions.
    VSHUFF64X2 $0x44, Z9,  Z8,  Z12 // t0 = c0 and c1 for the first two groups.
    VSHUFF64X2 $0xEE, Z9,  Z8,  Z13 // t1 = c0 and c1 for the last two groups.
    VSHUFF64X2 $0x44, Z11, Z10, Z14 // t2 = c2 and c3 for the first two groups.
    VSHUFF64X2 $0xEE, Z11, Z10, Z15 // t3 = c2 and c3 for the last two groups.
    VSHUFF64X2 $0x88, Z14, Z12, Z0  // The complete first group, c0 through c3.
    VSHUFF64X2 $0xDD, Z14, Z12, Z1  // The complete second group.
    VSHUFF64X2 $0x88, Z15, Z13, Z2  // The complete third group.
    VSHUFF64X2 $0xDD, Z15, Z13, Z3  // The complete fourth group.

		// The four groups all belong at unrelated places in dst, so read where
		// each one goes out of the precomputed table. Every store writes a full 64
		// byte group, which is one whole cache line when dst is aligned.
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
    SUBQ $1,  CX // One less group of four to go.
    JNZ  PASS1   // Keep going until first quarter of src has been read.

  // Everything from here on happens in place in dst.
  //
  // The ten radix-2 stages that are left get run as five radix-4 passes, each
  // one folding two stages together. That is worth doing because on Zen 4 a
  // 512 bit store only retires every other cycle, so what limits this transform
  // is how many times it walks the buffer rather than how much arithmetic it
  // does. Folding two stages into one pass halves both the loads and the
  // stores. See the measured store rate of 2.00 cycles here:
  // https://uops.info/html-instr/VMOVUPD_M512_ZMM.html
  //
  // A radix-4 butterfly takes four points spaced h apart and needs three
  // twiddle factors:
  //
  //   a' = a + w1*b        A = a' + w2*c'
  //   b' = a - w1*b        B = b' + w3*d'
  //   c' = c + w1*d        C = a' - w2*c'
  //   d' = c - w1*d        D = b' - w3*d'
  //
  // where w1 = W(2h)^j, w2 = W(4h)^j and w3 = W(4h)^(j+h). Only w1 and w2 are
  // in the table. w3 is always -i times w2, and once w2 has been splatted that
  // costs a single sign flip: the real half of w3 is the imaginary half of w2,
  // and the imaginary half of w3 is the negated real half of w2.
  LEAQ    ·fourierTwiddles4096(SB), BX // BX walks the twiddle table one pass at a time.
  VMOVUPD const_negate_all<>(SB),   Z31 // Sign flip mask that turns w2 into w3.

  MOVQ dst_len+8(FP), DX
  SHLQ $4,            DX  // DX = n*16, the total size of the buffer in bytes.
  LEAQ (DI)(DX*1),    AX  // AX = one byte past the end of the buffer.
  SHRQ $2,            DX  // DX = n*4, the byte width of h on the very last pass.
  MOVQ $64,           R12 // The first pass has h = 4 complex numbers, which is 64 bytes.

  PASS4:
    LEAQ (R12)(R12*2), R13 // R13 = 3h in bytes, the distance out to the fourth point.
    MOVQ DI,           R8  // R8 walks the start of each block.

		// One radix-4 butterfly is a long chain: load, swap, multiply, fused
		// multiply-add, add, then swap, multiply, fused multiply-add and add all
		// over again for the second folded stage. That comes to roughly 29 cycles
		// of latency against about 12.5 cycles of port throughput, so a single
		// butterfly in flight leaves the floating point pipes idle more than half
		// the time. Running two independent butterflies per iteration covers it.
		//
		// The first pass cannot be unrolled that way because its h is only four
		// complex numbers, which is one register, so it gets the plain body below
		// instead. It is one pass out of five.
    CMPQ R12, $64
    JEQ  NARROWBLOCK

    WIDEBLOCK:
      MOVQ R8,  R10 // R10 walks the butterflies inside this block.
      MOVQ BX,  R9  // The twiddle cursor restarts at the top of the pass for every block.
      MOVQ R12, SI  // SI counts down the bytes of butterflies left in this block.

      WIDEBUTTERFLY:
        // Twiddles for the first butterfly. A twiddle entry is a complex128,
        // the same 16 bytes as a data point, so the very same register indexes
        // both the table and the buffer.
        //
        // Both splats come out of a register rather than letting VMOVDDUP and
        // VPERMILPD carry their own memory operand. Folding the load into the
        // instruction looks free but it spends two load slots on the same 64
        // bytes, and Zen 4 only retires about 0.87 512 bit loads per cycle, so
        // one extra register instruction to save a load comes out ahead:
        //   https://uops.info/html-instr/VMOVUPD_ZMM_M512.html   (load, TP 1.00)
        //   https://uops.info/html-instr/VMOVDDUP_ZMM_M512.html  (splat from memory)
        //   https://uops.info/html-instr/VPERMILPD_ZMM_M512_I8.html
        //   https://uops.info/html-instr/VPERMILPD_ZMM_ZMM_I8.html (register form, TP 0.67)
        VMOVUPD   (R9),             Z30
        VMOVDDUP  Z30,              Z4  // (w1.real, w1.real)
        VPERMILPD $0xFF,       Z30, Z5  // (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),      Z30
        VMOVDDUP  Z30,              Z6  // (w2.real, w2.real)
        VPERMILPD $0xFF,       Z30, Z7  // (w2.imag, w2.imag), which is also (w3.real, w3.real).
        VPXORQ    Z31,         Z6,  Z8  // (w3.imag, w3.imag) = -(w2.real, w2.real)

        // And the same again for the second butterfly.
        VMOVUPD   64(R9),           Z30
        VMOVDDUP  Z30,              Z19
        VPERMILPD $0xFF,       Z30, Z20
        VMOVUPD   64(R9)(R12*1),    Z30
        VMOVDDUP  Z30,              Z21
        VPERMILPD $0xFF,       Z30, Z22
        VPXORQ    Z31,         Z21, Z23

        VMOVUPD (R10),             Z0  // a
        VMOVUPD (R10)(R12*1),      Z1  // b, h points further on.
        VMOVUPD (R10)(R12*2),      Z2  // c, 2h further on.
        VMOVUPD (R10)(R13*1),      Z3  // d, 3h further on.
        VMOVUPD 64(R10),           Z15 // The same four for the second butterfly.
        VMOVUPD 64(R10)(R12*1),    Z16
        VMOVUPD 64(R10)(R12*2),    Z17
        VMOVUPD 64(R10)(R13*1),    Z18

        // The first of the two folded stages. Both pairs turn on w1.
        VPERMILPD      $0x55, Z1,  Z9
        VPERMILPD      $0x55, Z3,  Z10
        VPERMILPD      $0x55, Z16, Z24
        VPERMILPD      $0x55, Z18, Z25
        VMULPD         Z5,    Z9,  Z9
        VMULPD         Z5,    Z10, Z10
        VMULPD         Z20,   Z24, Z24
        VMULPD         Z20,   Z25, Z25
        VFMADDSUB231PD Z4,    Z1,  Z9  // w1*b
        VFMADDSUB231PD Z4,    Z3,  Z10 // w1*d
        VFMADDSUB231PD Z19,   Z16, Z24
        VFMADDSUB231PD Z19,   Z18, Z25

        VADDPD Z9,  Z0,  Z11 // a' = a + w1*b
        VSUBPD Z9,  Z0,  Z12 // b' = a - w1*b
        VADDPD Z10, Z2,  Z13 // c' = c + w1*d
        VSUBPD Z10, Z2,  Z14 // d' = c - w1*d
        VADDPD Z24, Z15, Z26
        VSUBPD Z24, Z15, Z27
        VADDPD Z25, Z17, Z28
        VSUBPD Z25, Z17, Z29

        // The second folded stage. The lower pair turns on w2, the upper on w3.
        // a and b have both been consumed by now, so their registers are free
        // to hold the swapped copies.
        VPERMILPD      $0x55, Z13, Z0
        VPERMILPD      $0x55, Z14, Z1
        VPERMILPD      $0x55, Z28, Z15
        VPERMILPD      $0x55, Z29, Z16
        VMULPD         Z7,    Z0,  Z0
        VMULPD         Z8,    Z1,  Z1
        VMULPD         Z22,   Z15, Z15
        VMULPD         Z23,   Z16, Z16
        VFMADDSUB231PD Z6,    Z13, Z0  // w2*c'
        VFMADDSUB231PD Z7,    Z14, Z1  // w3*d', reusing w2.imag as w3.real.
        VFMADDSUB231PD Z21,   Z28, Z15
        VFMADDSUB231PD Z22,   Z29, Z16

        VADDPD Z0,  Z11, Z2  // A = a' + w2*c'
        VADDPD Z1,  Z12, Z3  // B = b' + w3*d'
        VSUBPD Z0,  Z11, Z9  // C = a' - w2*c'
        VSUBPD Z1,  Z12, Z10 // D = b' - w3*d'
        VADDPD Z15, Z26, Z17
        VADDPD Z16, Z27, Z18
        VSUBPD Z15, Z26, Z24
        VSUBPD Z16, Z27, Z25

        VMOVUPD Z2,  (R10)
        VMOVUPD Z3,  (R10)(R12*1)
        VMOVUPD Z9,  (R10)(R12*2)
        VMOVUPD Z10, (R10)(R13*1)
        VMOVUPD Z17, 64(R10)
        VMOVUPD Z18, 64(R10)(R12*1)
        VMOVUPD Z24, 64(R10)(R12*2)
        VMOVUPD Z25, 64(R10)(R13*1)

        ADDQ $128, R9  // Eight more twiddle factors.
        ADDQ $128, R10 // Eight more points.
        SUBQ $128, SI  // Eight fewer left to do in this block.
        JNZ  WIDEBUTTERFLY

      LEAQ (R8)(R12*4), R8 // The next block starts four h further along.
      CMPQ R8,          AX
      JCS  WIDEBLOCK       // Keep going while the block pointer is inside the buffer.
      JMP  PASSDONE

    // The h = 4 pass, one butterfly per block, otherwise identical.
    NARROWBLOCK:
      MOVQ R8,  R10
      MOVQ BX,  R9
      MOVQ R12, SI

      NARROWBUTTERFLY:
        VMOVUPD   (R9),          Z30
        VMOVDDUP  Z30,           Z4
        VPERMILPD $0xFF,    Z30, Z5
        VMOVUPD   (R9)(R12*1),   Z30
        VMOVDDUP  Z30,           Z6
        VPERMILPD $0xFF,    Z30, Z7
        VPXORQ    Z31,      Z6,  Z8

        VMOVUPD (R10),        Z0 // a
        VMOVUPD (R10)(R12*1), Z1 // b
        VMOVUPD (R10)(R12*2), Z2 // c
        VMOVUPD (R10)(R13*1), Z3 // d

        VPERMILPD      $0x55, Z1,  Z9
        VPERMILPD      $0x55, Z3,  Z10
        VMULPD         Z5,    Z9,  Z9
        VMULPD         Z5,    Z10, Z10
        VFMADDSUB231PD Z4,    Z1,  Z9  // w1*b
        VFMADDSUB231PD Z4,    Z3,  Z10 // w1*d

        VADDPD Z9,  Z0, Z11 // a'
        VSUBPD Z9,  Z0, Z12 // b'
        VADDPD Z10, Z2, Z13 // c'
        VSUBPD Z10, Z2, Z14 // d'

        VPERMILPD      $0x55, Z13, Z0
        VPERMILPD      $0x55, Z14, Z1
        VMULPD         Z7,    Z0,  Z0
        VMULPD         Z8,    Z1,  Z1
        VFMADDSUB231PD Z6,    Z13, Z0 // w2*c'
        VFMADDSUB231PD Z7,    Z14, Z1 // w3*d'

        VADDPD Z0, Z11, Z2  // A
        VADDPD Z1, Z12, Z3  // B
        VSUBPD Z0, Z11, Z9  // C
        VSUBPD Z1, Z12, Z10 // D

        VMOVUPD Z2,  (R10)
        VMOVUPD Z3,  (R10)(R12*1)
        VMOVUPD Z9,  (R10)(R12*2)
        VMOVUPD Z10, (R10)(R13*1)

        ADDQ $64, R9
        ADDQ $64, R10
        SUBQ $64, SI
        JNZ  NARROWBUTTERFLY

      LEAQ (R8)(R12*4), R8
      CMPQ R8,          AX
      JCS  NARROWBLOCK

    PASSDONE:
    LEAQ (BX)(R12*2), BX  // This pass consumed 2h twiddle entries.
    SHLQ $2,          R12 // And the next pass is four times as wide.
    CMPQ R12,         DX
    JLS  PASS4            // Keep going while the pass still fits in the transform.

  // Clear the dirty upper halves before handing control back. Go's ABI wrapper
  // runs a legacy SSE xorps on XMM15 the moment this returns, and reaching that
  // with the upper bits live costs a merge penalty on Intel parts and keeps the
  // core in its AVX512 frequency licence longer than it needs to be.
  VZEROUPPER
  RET // The transform is finished and sitting in dst.

// const_negate_imaginary_y is the 256 bit form of const_negate_imaginary. It
// flips the sign bit of the imaginary half of both complex numbers in a YMM.
DATA const_negate_imaginary_y<>+0(SB)/8,  $0x0000000000000000
DATA const_negate_imaginary_y<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_imaginary_y<>+16(SB)/8, $0x0000000000000000
DATA const_negate_imaginary_y<>+24(SB)/8, $0x8000000000000000
GLOBL const_negate_imaginary_y<>(SB), (RODATA+NOPTR), $32

// const_negate_all_y is the 256 bit form of const_negate_all, used to turn the
// splatted real half of w2 into the splatted imaginary half of w3.
DATA const_negate_all_y<>+0(SB)/8,  $0x8000000000000000
DATA const_negate_all_y<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_all_y<>+16(SB)/8, $0x8000000000000000
DATA const_negate_all_y<>+24(SB)/8, $0x8000000000000000
GLOBL const_negate_all_y<>(SB), (RODATA+NOPTR), $32

// func __fastFourierTransform_AVX_FMA(dst, src []complex128)
//
// The same transform as the AVX512 version above, worked in 256 bit registers.
// It reads the exact same generated tables: a twiddle entry is a complex128
// either way, and a pass's w2 sub-table still sits h entries past its w1, so
// the same register indexes both the table and the buffer no matter how wide
// the vectors are.
//
// Three things change coming down from ZMM to YMM:
//
//   1. A YMM holds two complex128 rather than four, so every loop covers half
//      as many points and runs twice as many iterations.
//   2. There are sixteen vector registers instead of thirty two, and the
//      radix-4 butterfly needs all sixteen, so there is no room to run two
//      butterflies at once the way the AVX512 version does. That costs very
//      little; the second butterfly was only worth a couple of percent there.
//   3. The 4x4 transpose of 128 bit lanes in the first pass becomes a 2x2 one,
//      because a YMM only has two lanes. Eight VSHUFF64X2 turn into four
//      VPERM2F128, which is actually cheaper per point.
//
// The sign flips use VXORPD rather than VPXORQ. VPXOR on YMM is an AVX2
// instruction, while VXORPD is plain AVX, so this stays inside the AVX feature
// set. That is the mirror image of the AVX512 version, where VXORPD is the one
// that needs an extension (AVX512DQ) and VPXORQ only needs the foundation set.
TEXT ·__fastFourierTransform_AVX_FMA(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI // The output buffer, which everything happens in after the first pass.
  MOVQ src_base+24(FP), SI // The input buffer, only ever read.
  MOVQ dst_len+8(FP),   AX // The number of complex128 points in the transform.

  VMOVUPD const_negate_imaginary_y<>(SB), Y12

  LEAQ ·fourierScatter4096(SB), BX

  MOVQ AX, R9
  SHLQ $2, R9  // R9 = n*4, a quarter of the way through src.
  MOVQ R9, R8
  ADDQ R9, R8  // R8 = n*8, halfway through src.
  MOVQ R8, R10
  ADDQ R9, R10 // R10 = n*12, three quarters of the way through src.

  MOVQ AX, CX
  SHRQ $3, CX  // Two source groups per iteration, so n/8 of them.

  PASS1_AVXFMA:
    VMOVUPD (SI),        Y0 // a0
    VMOVUPD (SI)(R8*1),  Y1 // a1
    VMOVUPD (SI)(R9*1),  Y2 // a2
    VMOVUPD (SI)(R10*1), Y3 // a3

    VADDPD Y1, Y0, Y4 // b0 = a0 + a1
    VSUBPD Y1, Y0, Y5 // b1 = a0 - a1
    VADDPD Y3, Y2, Y6 // b2 = a2 + a3
    VSUBPD Y3, Y2, Y7 // b3 = a2 - a3

    VPERMILPD $0x5, Y7, Y7 // (b3.imag, b3.real)
    VXORPD    Y12,  Y7, Y7 // (b3.imag, -b3.real), which is -i*b3.

    VADDPD Y6, Y4, Y8  // c0 = b0 + b2
    VADDPD Y7, Y5, Y9  // c1 = b1 + (-i * b3)
    VSUBPD Y6, Y4, Y10 // c2 = b0 - b2
    VSUBPD Y7, Y5, Y11 // c3 = b1 - (-i * b3)

    // A 2x2 transpose of 128 bit lanes. An output group is four complex
    // numbers, which is two YMM registers rather than the one ZMM above.
    VPERM2F128 $0x20, Y9,  Y8,  Y0 // [c0(q0), c1(q0)]
    VPERM2F128 $0x20, Y11, Y10, Y1 // [c2(q0), c3(q0)]
    VPERM2F128 $0x31, Y9,  Y8,  Y2 // [c0(q1), c1(q1)]
    VPERM2F128 $0x31, Y11, Y10, Y3 // [c2(q1), c3(q1)]

    MOVL 0(BX), R11
    MOVL 4(BX), R12
    VMOVUPD Y0, (DI)(R11*1)
    VMOVUPD Y1, 32(DI)(R11*1)
    VMOVUPD Y2, (DI)(R12*1)
    VMOVUPD Y3, 32(DI)(R12*1)

    ADDQ $32, SI // Two complex numbers.
    ADDQ $8,  BX // Two scatter entries.
    SUBQ $1,  CX
    JNZ  PASS1_AVXFMA

  LEAQ    ·fourierTwiddles4096(SB), BX
  VMOVUPD const_negate_all_y<>(SB),  Y15

  MOVQ dst_len+8(FP), DX
  SHLQ $4,            DX
  LEAQ (DI)(DX*1),    AX
  SHRQ $2,            DX
  MOVQ $64,           R12

  PASS4_AVXFMA:
    LEAQ (R12)(R12*2), R13
    MOVQ DI,           R8

    BLOCK_AVXFMA:
      MOVQ R8,  R10
      MOVQ BX,  R9
      MOVQ R12, SI

      BUTTERFLY_AVXFMA:
        // Y0 doubles as the scratch register for the raw twiddle factors before
        // it gets loaded with a. That is what keeps this inside sixteen
        // registers.
        VMOVUPD   (R9),         Y0
        VMOVDDUP  Y0,           Y4 // (w1.real, w1.real)
        VPERMILPD $0xF,     Y0, Y5 // (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),  Y0
        VMOVDDUP  Y0,           Y6 // (w2.real, w2.real)
        VPERMILPD $0xF,     Y0, Y7 // (w2.imag, w2.imag), also (w3.real, w3.real)
        VXORPD    Y15,      Y6, Y8 // (w3.imag, w3.imag) = -(w2.real, w2.real)

        VMOVUPD (R10),        Y0 // a
        VMOVUPD (R10)(R12*1), Y1 // b
        VMOVUPD (R10)(R12*2), Y2 // c
        VMOVUPD (R10)(R13*1), Y3 // d

        VPERMILPD      $0x5, Y1,  Y9
        VPERMILPD      $0x5, Y3,  Y10
        VMULPD         Y5,   Y9,  Y9
        VMULPD         Y5,   Y10, Y10
        VFMADDSUB231PD Y4,   Y1,  Y9  // w1*b
        VFMADDSUB231PD Y4,   Y3,  Y10 // w1*d

        VADDPD Y9,  Y0, Y11 // a' = a + w1*b
        VSUBPD Y9,  Y0, Y12 // b' = a - w1*b
        VADDPD Y10, Y2, Y13 // c' = c + w1*d
        VSUBPD Y10, Y2, Y14 // d' = c - w1*d

        VPERMILPD      $0x5, Y13, Y0
        VPERMILPD      $0x5, Y14, Y1
        VMULPD         Y7,   Y0,  Y0
        VMULPD         Y8,   Y1,  Y1
        VFMADDSUB231PD Y6,   Y13, Y0 // w2*c'
        VFMADDSUB231PD Y7,   Y14, Y1 // w3*d', reusing w2.imag as w3.real.

        VADDPD Y0, Y11, Y2  // A = a' + w2*c'
        VADDPD Y1, Y12, Y3  // B = b' + w3*d'
        VSUBPD Y0, Y11, Y9  // C = a' - w2*c'
        VSUBPD Y1, Y12, Y10 // D = b' - w3*d'

        VMOVUPD Y2,  (R10)
        VMOVUPD Y3,  (R10)(R12*1)
        VMOVUPD Y9,  (R10)(R12*2)
        VMOVUPD Y10, (R10)(R13*1)

        ADDQ $32, R9
        ADDQ $32, R10
        SUBQ $32, SI
        JNZ  BUTTERFLY_AVXFMA

      LEAQ (R8)(R12*4), R8
      CMPQ R8,          AX
      JCS  BLOCK_AVXFMA

    LEAQ (BX)(R12*2), BX
    SHLQ $2,          R12
    CMPQ R12,         DX
    JLS  PASS4_AVXFMA

  VZEROUPPER // Leave no dirty upper state behind for whatever SSE code runs next.
  RET

// func __fastFourierTransform_AVX(dst, src []complex128)
//
// The same 256 bit transform without any fused multiply-add, for hosts that
// have AVX but not FMA3. That is Sandy Bridge and Ivy Bridge, including the
// Xeon E5 v1 and v2 parts; everything from Haswell on has FMA.
//
// AVX512 dropped VADDSUBPD, but plain AVX still has it, and it does exactly the
// subtract-on-real, add-on-imaginary that finishes a complex multiply. So the
// multiply here is two VMULPD and one VADDSUBPD, against one VMULPD and one
// VFMADDSUB231PD with FMA: one extra instruction, slightly looser rounding, and
// no extra registers, because each product overwrites the operand that just
// died.
//
// On Sandy and Ivy Bridge this ends up limited by port 1, which has to carry
// both the VADDSUBPD from every complex multiply and all eight butterfly
// adds and subtracts. There is no way around that on those cores, VADDPD only
// issues on port 1.
TEXT ·__fastFourierTransform_AVX(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI
  MOVQ src_base+24(FP), SI
  MOVQ dst_len+8(FP),   AX

  VMOVUPD const_negate_imaginary_y<>(SB), Y12

  LEAQ ·fourierScatter4096(SB), BX

  MOVQ AX, R9
  SHLQ $2, R9
  MOVQ R9, R8
  ADDQ R9, R8
  MOVQ R8, R10
  ADDQ R9, R10

  MOVQ AX, CX
  SHRQ $3, CX

  PASS1_AVX:
    VMOVUPD (SI),        Y0 // a0
    VMOVUPD (SI)(R8*1),  Y1 // a1
    VMOVUPD (SI)(R9*1),  Y2 // a2
    VMOVUPD (SI)(R10*1), Y3 // a3

    VADDPD Y1, Y0, Y4 // b0
    VSUBPD Y1, Y0, Y5 // b1
    VADDPD Y3, Y2, Y6 // b2
    VSUBPD Y3, Y2, Y7 // b3

    VPERMILPD $0x5, Y7, Y7
    VXORPD    Y12,  Y7, Y7 // -i*b3

    VADDPD Y6, Y4, Y8  // c0
    VADDPD Y7, Y5, Y9  // c1
    VSUBPD Y6, Y4, Y10 // c2
    VSUBPD Y7, Y5, Y11 // c3

    VPERM2F128 $0x20, Y9,  Y8,  Y0
    VPERM2F128 $0x20, Y11, Y10, Y1
    VPERM2F128 $0x31, Y9,  Y8,  Y2
    VPERM2F128 $0x31, Y11, Y10, Y3

    MOVL 0(BX), R11
    MOVL 4(BX), R12
    VMOVUPD Y0, (DI)(R11*1)
    VMOVUPD Y1, 32(DI)(R11*1)
    VMOVUPD Y2, (DI)(R12*1)
    VMOVUPD Y3, 32(DI)(R12*1)

    ADDQ $32, SI
    ADDQ $8,  BX
    SUBQ $1,  CX
    JNZ  PASS1_AVX

  LEAQ    ·fourierTwiddles4096(SB), BX
  VMOVUPD const_negate_all_y<>(SB),  Y15

  MOVQ dst_len+8(FP), DX
  SHLQ $4,            DX
  LEAQ (DI)(DX*1),    AX
  SHRQ $2,            DX
  MOVQ $64,           R12

  PASS4_AVX:
    LEAQ (R12)(R12*2), R13
    MOVQ DI,           R8

    BLOCK_AVX:
      MOVQ R8,  R10
      MOVQ BX,  R9
      MOVQ R12, SI

      BUTTERFLY_AVX:
        VMOVUPD   (R9),         Y0
        VMOVDDUP  Y0,           Y4 // (w1.real, w1.real)
        VPERMILPD $0xF,     Y0, Y5 // (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),  Y0
        VMOVDDUP  Y0,           Y6 // (w2.real, w2.real)
        VPERMILPD $0xF,     Y0, Y7 // (w2.imag, w2.imag), also (w3.real, w3.real)
        VXORPD    Y15,      Y6, Y8 // (w3.imag, w3.imag)

        VMOVUPD (R10),        Y0 // a
        VMOVUPD (R10)(R12*1), Y1 // b
        VMOVUPD (R10)(R12*2), Y2 // c
        VMOVUPD (R10)(R13*1), Y3 // d

        VPERMILPD $0x5, Y1,  Y9
        VPERMILPD $0x5, Y3,  Y10
        VMULPD    Y5,   Y9,  Y9  // swap(b)*w1.imag
        VMULPD    Y5,   Y10, Y10 // swap(d)*w1.imag
        VMULPD    Y4,   Y1,  Y1  // b*w1.real, b is dead after this
        VMULPD    Y4,   Y3,  Y3  // d*w1.real, d is dead after this
        VADDSUBPD Y9,   Y1,  Y9  // w1*b
        VADDSUBPD Y10,  Y3,  Y10 // w1*d

        VADDPD Y9,  Y0, Y11 // a'
        VSUBPD Y9,  Y0, Y12 // b'
        VADDPD Y10, Y2, Y13 // c'
        VSUBPD Y10, Y2, Y14 // d'

        VPERMILPD $0x5, Y13, Y0
        VPERMILPD $0x5, Y14, Y1
        VMULPD    Y7,   Y0,  Y0  // swap(c')*w2.imag
        VMULPD    Y8,   Y1,  Y1  // swap(d')*w3.imag
        VMULPD    Y6,   Y13, Y13 // c'*w2.real, c' is dead after this
        VMULPD    Y7,   Y14, Y14 // d'*w3.real, which is w2.imag
        VADDSUBPD Y0,   Y13, Y0  // w2*c'
        VADDSUBPD Y1,   Y14, Y1  // w3*d'

        VADDPD Y0, Y11, Y2  // A
        VADDPD Y1, Y12, Y3  // B
        VSUBPD Y0, Y11, Y9  // C
        VSUBPD Y1, Y12, Y10 // D

        VMOVUPD Y2,  (R10)
        VMOVUPD Y3,  (R10)(R12*1)
        VMOVUPD Y9,  (R10)(R12*2)
        VMOVUPD Y10, (R10)(R13*1)

        ADDQ $32, R9
        ADDQ $32, R10
        SUBQ $32, SI
        JNZ  BUTTERFLY_AVX

      LEAQ (R8)(R12*4), R8
      CMPQ R8,          AX
      JCS  BLOCK_AVX

    LEAQ (BX)(R12*2), BX
    SHLQ $2,          R12
    CMPQ R12,         DX
    JLS  PASS4_AVX

  VZEROUPPER
  RET
