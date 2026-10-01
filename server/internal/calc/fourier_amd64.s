#include "textflag.h"

// Some notes before reading this file
//
// A complex128 is two float64s next to each other, the real half first and
// then the imaginary half. So a ZMM register (512 bits) holds 4 complex
// numbers and a YMM register (256 bits) holds 2. In the comments below (real,
// imag) is one complex number, and [x, y, z, w] is the lanes of a whole
// register where each lane is one complex number (128 bits)
//
// Go assembly puts the operands in the opposite order from the Intel docs, the
// destination is always last. So VSUBPD Z1, Z0, Z5 is Z5 = Z0 - Z1. This
// matters most for the shuffles, so the comments on those spell out exactly
// what ends up in each lane
//
// The plain Go version of this is fastFourierTransformGo in fourier.go. PASS1
// below is the same as the first loop there. After that the Go version does
// one radix-2 stage per loop, while this does two stages per loop (radix-4)
//
// The tables this reads (fourierScatter4096 and fourierTwiddles4096) are in
// fourier_twiddles_amd64.s, which is written by gen/main.go

// const_negate_imaginary is used to flip the sign of the imaginary half of
// every complex number in a ZMM register. 0x8000000000000000 is just the sign
// bit of a float64, so XORing with it flips the sign and XORing with 0 leaves
// the value alone. The imaginary half is every other float64, so the mask is
// [0, sign, 0, sign, ...]
DATA const_negate_imaginary<>+0(SB)/8,  $0x0000000000000000
DATA const_negate_imaginary<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_imaginary<>+16(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+24(SB)/8, $0x8000000000000000
DATA const_negate_imaginary<>+32(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+40(SB)/8, $0x8000000000000000
DATA const_negate_imaginary<>+48(SB)/8, $0x0000000000000000
DATA const_negate_imaginary<>+56(SB)/8, $0x8000000000000000
GLOBL const_negate_imaginary<>(SB), (RODATA+NOPTR), $64

// const_negate_all flips the sign of every float64 in a ZMM register. We use it
// to get w3 out of w2 in PASS4, see the notes above PASS4
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
// A 4096 point FFT that does not recurse and does not do any trig. All of the
// twiddle factors and bit reversal offsets come out of the generated tables
//
// A 4096 point FFT has 12 radix-2 stages (2^12 = 4096). We do them like this:
//
//   PASS1: Reads src, does the bit reversal, and does the first 2 stages all
//          at once. The twiddle factors for those 2 stages are only ever 1 or
//          -i so there is nothing to look up. The results go into dst
//
//   PASS4: Runs 5 times over dst in place, doing 2 stages each time
//
// 2 + (5 * 2) = 12 stages
TEXT ·__fastFourierTransform_AVX512(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI // Load the pointer of dst into DI. Everything after PASS1 reads and writes here
  MOVQ src_base+24(FP), SI // Load the pointer of src into SI. We only ever read from this
  MOVQ dst_len+8(FP),   AX // Load the length of dst into AX. This is n, the number of complex numbers (4096)

  VMOVUPD const_negate_imaginary<>(SB), Z16 // Load the imaginary sign mask into Z16, PASS1 uses it every loop

  LEAQ ·fourierScatter4096(SB), BX // Point BX at the start of the bit reversal table. Each entry is a 4 byte offset into dst

	// PASS1 reads 4 spots in src at once, each a quarter of the way apart. We
	// need those distances in bytes. A complex128 is 16 bytes, so n/4 points is
	// (n/4) * 16 = n*4 bytes
  MOVQ AX, R9   // Copy n into R9
  SHLQ $2, R9   // Shift left by 2 to multiply by 4. R9 = n*4, the byte distance to a quarter of the way through src
  MOVQ R9, R8   // Copy n*4 into R8
  ADDQ R9, R8   // R8 = n*4 + n*4 = n*8, the byte distance to halfway through src
  MOVQ R8, R10  // Copy n*8 into R10
  ADDQ R9, R10  // R10 = n*8 + n*4 = n*12, the byte distance to three quarters of the way through src

  MOVQ AX, CX   // Copy n into CX, this is our loop counter
  SHRQ $4, CX   // Shift right by 4 to divide by 16. We go through the first quarter of src (n/4) 4 at a time, so n/16 loops (256)

  // Each loop of PASS1 does 4 radix-4 butterflies side by side, one in each
  // lane of the ZMM registers. Call them q0 through q3, where q0 is wherever SI
  // is pointing right now. Butterfly q reads these 4 points:
  //
  //   a0 = src[q]
  //   a1 = src[q + n/2]
  //   a2 = src[q + n/4]
  //   a3 = src[q + 3n/4]
  //
	// After the bit reversal those 4 points end up right next to each other in
	// the order a0, a1, a2, a3. So instead of shuffling src into bit reversed
	// order first, we just read them from where they already are and only have
	// to figure out where to write the results
  PASS1:
    VMOVUPD (SI),        Z0 // Load a0 for q0 through q3 into Z0
    VMOVUPD (SI)(R8*1),  Z1 // Load a1 for q0 through q3 into Z1, from SI + n*8 (halfway)
    VMOVUPD (SI)(R9*1),  Z2 // Load a2 for q0 through q3 into Z2, from SI + n*4 (a quarter)
    VMOVUPD (SI)(R10*1), Z3 // Load a3 for q0 through q3 into Z3, from SI + n*12 (three quarters)

    // First stage. The twiddle factor is always 1 here so it is just adds and
    // subtracts
    VADDPD Z1, Z0, Z4 // b0 = a0 + a1, stored in Z4
    VSUBPD Z1, Z0, Z5 // b1 = a0 - a1, stored in Z5
    VADDPD Z3, Z2, Z6 // b2 = a2 + a3, stored in Z6
    VSUBPD Z3, Z2, Z7 // b3 = a2 - a3, stored in Z7

		// Second stage. b2 gets multiplied by 1 and b3 gets multiplied by -i. We
		// don't need an actual multiply for -i though:
    //
    //   (x + iy) * -i = -ix - i*i*y = y - ix
    //
		// So the new real half is y and the new imaginary half is -x. That is just
		// swapping the two halves and then flipping the sign of the imaginary
		// half
    //
		// VPERMILPD picks the real half (0) or the imaginary half (1) for each
		// float64 using one bit each, read from the right. 0x55 is 01010101, so
		// the first float64 of each pair gets the imaginary half and the second
		// gets the real half, which is a swap
    VPERMILPD $0x55, Z7, Z7 // Swap the halves of b3, so (x, y) becomes (y, x)
    VPXORQ    Z16,   Z7, Z7 // XOR with the imaginary sign mask, so (y, x) becomes (y, -x). Z7 is now -i * b3

    VADDPD Z6, Z4, Z8  // c0 = b0 + b2, stored in Z8
    VADDPD Z7, Z5, Z9  // c1 = b1 + (-i * b3), stored in Z9
    VSUBPD Z6, Z4, Z10 // c2 = b0 - b2, stored in Z10
    VSUBPD Z7, Z5, Z11 // c3 = b1 - (-i * b3), stored in Z11

		// Now we have all 16 outputs but they are in the wrong registers. Z8 has
		// c0 for all 4 butterflies, Z9 has c1 for all 4, and so on:
    //
    //   Z8  = [c0(q0), c0(q1), c0(q2), c0(q3)]
    //   Z9  = [c1(q0), c1(q1), c1(q2), c1(q3)]
    //   Z10 = [c2(q0), c2(q1), c2(q2), c2(q3)]
    //   Z11 = [c3(q0), c3(q1), c3(q2), c3(q3)]
    //
		// But dst wants c0 through c3 of one butterfly next to each other, like
		// [c0(q0), c1(q0), c2(q0), c3(q0)]. So we need to transpose the 4x4 above
    //
		// VSHUFF64X2 builds a register out of the lanes of two other registers. In
		// Go order it is VSHUFF64X2 $imm, high, low, dst. The low 2 lanes of dst
		// come from low and the high 2 lanes come from high. The immediate is 4
		// pairs of bits read from the right, and each pair picks lane 0, 1, 2 or
		// 3
    //
    //   0x44 = 01 00 01 00 -> lanes 0, 1 of low, then lanes 0, 1 of high
    //   0xEE = 11 10 11 10 -> lanes 2, 3 of low, then lanes 2, 3 of high
    //   0x88 = 10 00 10 00 -> lanes 0, 2 of low, then lanes 0, 2 of high
    //   0xDD = 11 01 11 01 -> lanes 1, 3 of low, then lanes 1, 3 of high
    VSHUFF64X2 $0x44, Z9,  Z8,  Z12 // Z12 = [c0(q0), c0(q1), c1(q0), c1(q1)]
    VSHUFF64X2 $0xEE, Z9,  Z8,  Z13 // Z13 = [c0(q2), c0(q3), c1(q2), c1(q3)]
    VSHUFF64X2 $0x44, Z11, Z10, Z14 // Z14 = [c2(q0), c2(q1), c3(q0), c3(q1)]
    VSHUFF64X2 $0xEE, Z11, Z10, Z15 // Z15 = [c2(q2), c2(q3), c3(q2), c3(q3)]
    VSHUFF64X2 $0x88, Z14, Z12, Z0  // Z0  = [c0(q0), c1(q0), c2(q0), c3(q0)], all of q0
    VSHUFF64X2 $0xDD, Z14, Z12, Z1  // Z1  = [c0(q1), c1(q1), c2(q1), c3(q1)], all of q1
    VSHUFF64X2 $0x88, Z15, Z13, Z2  // Z2  = [c0(q2), c1(q2), c2(q2), c3(q2)], all of q2
    VSHUFF64X2 $0xDD, Z15, Z13, Z3  // Z3  = [c0(q3), c1(q3), c2(q3), c3(q3)], all of q3

		// Each of those registers is one finished group of 4 outputs, and each
		// group goes somewhere different in dst. The scatter table has the byte
		// offset for each one already worked out (it is 64 * the bit reversal of
		// q). MOVL loads 4 bytes and zeroes the top half of the register, so it is
		// safe to use as a 64 bit offset
    MOVL 0(BX),  R11 // Load the dst byte offset for q0 into R11
    MOVL 4(BX),  R12 // Load the dst byte offset for q1 into R12
    MOVL 8(BX),  R13 // Load the dst byte offset for q2 into R13
    MOVL 12(BX), DX  // Load the dst byte offset for q3 into DX
    VMOVUPD Z0, (DI)(R11*1) // Store all of q0 at dst + R11
    VMOVUPD Z1, (DI)(R12*1) // Store all of q1 at dst + R12
    VMOVUPD Z2, (DI)(R13*1) // Store all of q2 at dst + R13
    VMOVUPD Z3, (DI)(DX*1)  // Store all of q3 at dst + DX

    ADDQ $64, SI // Add 64 (4 * 16) to SI. This moves src forward by 4 complex numbers
    ADDQ $16, BX // Add 16 (4 * 4) to BX. This moves the scatter table forward by 4 entries
    SUBQ $1,  CX // Subtract 1 from the CX loop counter
    JNZ  PASS1   // If CX is not zero then jump back to the start of PASS1

	// dst now has the first 2 stages done. Everything from here on reads and
	// writes dst in place
  //
	// We do the other 10 stages 2 at a time. Each pass has a size h, which is
	// how far apart the 4 points of one butterfly are. h starts at 4 and gets 4
	// times bigger each pass: 4, 16, 64, 256, 1024. That is 5 passes. R12 holds
	// h in bytes (h * 16) the whole way through
  //
	// Doing 2 stages per pass means we only go through dst 5 times instead of
	// 10. On Zen 4 a 512 bit store can only happen every other cycle, so cutting
	// the number of loads and stores in half is a big deal here:
	// https://uops.info/html-instr/VMOVUPD_M512_ZMM.html
  //
	// dst gets split up into blocks of 4h points. In each block, butterfly j (0
	// through h-1) reads 4 points a, b, c and d that are each h apart, and does:
  //
  //   a' = a + w1*b        A = a' + w2*c'
  //   b' = a - w1*b        B = b' + w3*d'
  //   c' = c + w1*d        C = a' - w2*c'
  //   d' = c - w1*d        D = b' - w3*d'
  //
	// The left side is the first stage and the right side is the second stage.
	// Then A, B, C and D get written back to where a, b, c and d came from
  //
  // w1, w2 and w3 are the twiddle factors for butterfly j. W(N) is
  // e^(-2*pi*i/N), the same as complexExponential(-2 * math.Pi / N) in Go
  //
  //   w1 = W(2h)^j      The first h entries for this pass in the twiddle table
  //   w2 = W(4h)^j      The next h entries in the twiddle table
  //   w3 = W(4h)^(j+h)  Not in the table at all
  //
	// We don't store w3 because it is always -i * w2, and like in PASS1 -i * (x
	// + iy) = y - ix. So the real half of w3 is the imaginary half of w2, and
	// the imaginary half of w3 is the real half of w2 with the sign flipped
  //
	// A twiddle entry is a complex128, 16 bytes, same as a point in dst. So
	// adding h bytes gets us from w1 to w2 in the table the same way it gets us
	// from a to b in dst
  LEAQ    ·fourierTwiddles4096(SB), BX  // Point BX at the start of the twiddle table. It moves forward after each pass
  VMOVUPD const_negate_all<>(SB),   Z31 // Load the sign mask for every float64 into Z31, used to get w3 from w2

  MOVQ dst_len+8(FP), DX  // Load n into DX again, PASS1 used DX for an offset
  SHLQ $4,            DX  // Shift left by 4 to multiply by 16. DX = n*16, the size of dst in bytes
  LEAQ (DI)(DX*1),    AX  // AX = DI + DX, the address right after the end of dst. Blocks stop when they get here
  SHRQ $2,            DX  // Shift right by 2 to divide by 4. DX = n*4, which is h in bytes on the last pass (1024 * 16)
  MOVQ $64,           R12 // R12 is h in bytes. The first pass has h = 4, so 4 * 16 = 64 bytes

  PASS4:
    LEAQ (R12)(R12*2), R13 // R13 = R12 + R12*2 = 3h in bytes, the distance from a to d
    MOVQ DI,           R8  // R8 is the start of the current block, start at the beginning of dst

		// WIDEBUTTERFLY does 2 sets of registers per loop, j through j+3 in one
		// set and j+4 through j+7 in the other. This is the same idea as the 2
		// accumulators in the euclidean distance code. Each set is a long chain
		// where every instruction waits on the one before it, so having 2
		// independent chains gives the CPU something else to work on while it
		// waits
    //
		// That needs at least 8 values of j in a block. When h is 4 there are only
		// 4, so that pass goes through NARROWBLOCK instead which only does 1 set
    CMPQ R12, $64    // Is h 64 bytes (4 points)?
    JEQ  NARROWBLOCK // If it is then jump to NARROWBLOCK for this pass

    WIDEBLOCK:
      MOVQ R8,  R10 // R10 points at a for the current j, start at the beginning of the block
      MOVQ BX,  R9  // R9 points at w1 for the current j. Every block in a pass uses the same twiddles so start over at BX
      MOVQ R12, SI  // SI counts down the bytes of j left in this block, starting at h

      WIDEBUTTERFLY:
				// To multiply by a twiddle factor w = (wr, wi) we need (wr, wr) in one
				// register and (wi, wi) in another. VMOVDDUP copies the real half over
				// the imaginary half. VPERMILPD with 0xFF (all 1s) picks the imaginary
				// half for both
        //
				// Both of those can load straight from memory, but then we would be
				// loading the same 64 bytes twice. Zen 4 can't do that many 512 bit
				// loads per cycle, so we load once into Z30 and split it from there
        //   https://uops.info/html-instr/VMOVUPD_ZMM_M512.html
        //   https://uops.info/html-instr/VMOVDDUP_ZMM_M512.html
        //   https://uops.info/html-instr/VPERMILPD_ZMM_M512_I8.html
        //   https://uops.info/html-instr/VPERMILPD_ZMM_ZMM_I8.html
        VMOVUPD   (R9),             Z30 // Load w1 for j through j+3 into Z30
        VMOVDDUP  Z30,              Z4  // Z4 = (w1.real, w1.real)
        VPERMILPD $0xFF,       Z30, Z5  // Z5 = (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),      Z30 // Load w2 for j through j+3 into Z30, it is h bytes after w1
        VMOVDDUP  Z30,              Z6  // Z6 = (w2.real, w2.real)
        VPERMILPD $0xFF,       Z30, Z7  // Z7 = (w2.imag, w2.imag), which is also (w3.real, w3.real)
        VPXORQ    Z31,         Z6,  Z8  // Z8 = (-w2.real, -w2.real), which is (w3.imag, w3.imag)

        // Same thing for j+4 through j+7, which is 64 bytes further along
        VMOVUPD   64(R9),           Z30 // Load w1 for j+4 through j+7 into Z30
        VMOVDDUP  Z30,              Z19 // Z19 = (w1.real, w1.real)
        VPERMILPD $0xFF,       Z30, Z20 // Z20 = (w1.imag, w1.imag)
        VMOVUPD   64(R9)(R12*1),    Z30 // Load w2 for j+4 through j+7 into Z30
        VMOVDDUP  Z30,              Z21 // Z21 = (w2.real, w2.real)
        VPERMILPD $0xFF,       Z30, Z22 // Z22 = (w2.imag, w2.imag), also (w3.real, w3.real)
        VPXORQ    Z31,         Z21, Z23 // Z23 = (w3.imag, w3.imag)

        // Z16 was the mask from PASS1, we don't need it anymore so it gets reused
        VMOVUPD (R10),             Z0  // Load a for j through j+3 into Z0
        VMOVUPD (R10)(R12*1),      Z1  // Load b into Z1, h bytes after a
        VMOVUPD (R10)(R12*2),      Z2  // Load c into Z2, 2h bytes after a
        VMOVUPD (R10)(R13*1),      Z3  // Load d into Z3, 3h bytes after a
        VMOVUPD 64(R10),           Z15 // Load a for j+4 through j+7 into Z15
        VMOVUPD 64(R10)(R12*1),    Z16 // Load b for j+4 through j+7 into Z16
        VMOVUPD 64(R10)(R12*2),    Z17 // Load c for j+4 through j+7 into Z17
        VMOVUPD 64(R10)(R13*1),    Z18 // Load d for j+4 through j+7 into Z18

        // First stage, we need w1*b and w1*d
        //
        // Multiplying 2 complex numbers w = (wr, wi) and b = (br, bi) gives:
        //
        //   real = br*wr - bi*wi
        //   imag = bi*wr + br*wi
        //
        // We get there in 3 steps:
        //
        //   1. Swap the halves of b to get (bi, br)
        //   2. Multiply that by (wi, wi) to get (bi*wi, br*wi)
        //   3. VFMADDSUB231PD multiplies b by (wr, wr) to get (br*wr, bi*wr),
        //      then subtracts step 2 on the real half and adds step 2 on the
        //      imaginary half. That gives (br*wr - bi*wi, bi*wr + br*wi)
        VPERMILPD      $0x55, Z1,  Z9  // Z9 = b with its halves swapped
        VPERMILPD      $0x55, Z3,  Z10 // Z10 = d with its halves swapped
        VPERMILPD      $0x55, Z16, Z24 // Z24 = b swapped for j+4 through j+7
        VPERMILPD      $0x55, Z18, Z25 // Z25 = d swapped for j+4 through j+7
        VMULPD         Z5,    Z9,  Z9  // Z9 = swapped b * w1.imag
        VMULPD         Z5,    Z10, Z10 // Z10 = swapped d * w1.imag
        VMULPD         Z20,   Z24, Z24 // Z24 = swapped b * w1.imag for j+4 through j+7
        VMULPD         Z20,   Z25, Z25 // Z25 = swapped d * w1.imag for j+4 through j+7
        VFMADDSUB231PD Z4,    Z1,  Z9  // Z9 = (b * w1.real) -/+ Z9, which is w1*b
        VFMADDSUB231PD Z4,    Z3,  Z10 // Z10 = (d * w1.real) -/+ Z10, which is w1*d
        VFMADDSUB231PD Z19,   Z16, Z24 // Z24 = w1*b for j+4 through j+7
        VFMADDSUB231PD Z19,   Z18, Z25 // Z25 = w1*d for j+4 through j+7

        VADDPD Z9,  Z0,  Z11 // Z11 = a' = a + w1*b
        VSUBPD Z9,  Z0,  Z12 // Z12 = b' = a - w1*b
        VADDPD Z10, Z2,  Z13 // Z13 = c' = c + w1*d
        VSUBPD Z10, Z2,  Z14 // Z14 = d' = c - w1*d
        VADDPD Z24, Z15, Z26 // Z26 = a' for j+4 through j+7
        VSUBPD Z24, Z15, Z27 // Z27 = b' for j+4 through j+7
        VADDPD Z25, Z17, Z28 // Z28 = c' for j+4 through j+7
        VSUBPD Z25, Z17, Z29 // Z29 = d' for j+4 through j+7

				// Second stage, we need w2*c' and w3*d'. This is the same 3 step
				// multiply as above. We are done with a and b now so Z0, Z1, Z15 and
				// Z16 get reused for the swapped copies
        //
				// For w3 we use Z7 (w2.imag) as w3.real and Z8 (-w2.real) as w3.imag,
				// see the notes above PASS4
        VPERMILPD      $0x55, Z13, Z0  // Z0 = c' swapped
        VPERMILPD      $0x55, Z14, Z1  // Z1 = d' swapped
        VPERMILPD      $0x55, Z28, Z15 // Z15 = c' swapped for j+4 through j+7
        VPERMILPD      $0x55, Z29, Z16 // Z16 = d' swapped for j+4 through j+7
        VMULPD         Z7,    Z0,  Z0  // Z0 = swapped c' * w2.imag
        VMULPD         Z8,    Z1,  Z1  // Z1 = swapped d' * w3.imag
        VMULPD         Z22,   Z15, Z15 // Z15 = swapped c' * w2.imag for j+4 through j+7
        VMULPD         Z23,   Z16, Z16 // Z16 = swapped d' * w3.imag for j+4 through j+7
        VFMADDSUB231PD Z6,    Z13, Z0  // Z0 = w2*c'
        VFMADDSUB231PD Z7,    Z14, Z1  // Z1 = w3*d', Z7 is w2.imag which is the same as w3.real
        VFMADDSUB231PD Z21,   Z28, Z15 // Z15 = w2*c' for j+4 through j+7
        VFMADDSUB231PD Z22,   Z29, Z16 // Z16 = w3*d' for j+4 through j+7

        VADDPD Z0,  Z11, Z2  // Z2 = A = a' + w2*c'
        VADDPD Z1,  Z12, Z3  // Z3 = B = b' + w3*d'
        VSUBPD Z0,  Z11, Z9  // Z9 = C = a' - w2*c'
        VSUBPD Z1,  Z12, Z10 // Z10 = D = b' - w3*d'
        VADDPD Z15, Z26, Z17 // Z17 = A for j+4 through j+7
        VADDPD Z16, Z27, Z18 // Z18 = B for j+4 through j+7
        VSUBPD Z15, Z26, Z24 // Z24 = C for j+4 through j+7
        VSUBPD Z16, Z27, Z25 // Z25 = D for j+4 through j+7

        VMOVUPD Z2,  (R10)          // Store A where a came from
        VMOVUPD Z3,  (R10)(R12*1)   // Store B where b came from
        VMOVUPD Z9,  (R10)(R12*2)   // Store C where c came from
        VMOVUPD Z10, (R10)(R13*1)   // Store D where d came from
        VMOVUPD Z17, 64(R10)        // Store A for j+4 through j+7
        VMOVUPD Z18, 64(R10)(R12*1) // Store B for j+4 through j+7
        VMOVUPD Z24, 64(R10)(R12*2) // Store C for j+4 through j+7
        VMOVUPD Z25, 64(R10)(R13*1) // Store D for j+4 through j+7

        ADDQ $128, R9      // Add 128 (8 * 16) to R9. This moves the twiddle pointer forward by 8 entries
        ADDQ $128, R10     // Add 128 (8 * 16) to R10. This moves forward by 8 points to j+8
        SUBQ $128, SI      // Subtract 128 from SI since we just did 8 values of j
        JNZ  WIDEBUTTERFLY // If SI is not zero then there is more of this block left, jump back to WIDEBUTTERFLY

      LEAQ (R8)(R12*4), R8 // R8 = R8 + R12*4. This moves R8 forward by 4h to the start of the next block
      CMPQ R8,          AX // Compare R8 against the end of dst
      JCS  WIDEBLOCK       // If R8 is below the end of dst then jump back to WIDEBLOCK. JCS is the unsigned "less than"
      JMP  PASSDONE        // Otherwise this pass is done, skip over NARROWBLOCK

		// NARROWBLOCK is the same as WIDEBLOCK but with only 1 set of registers,
		// for the h = 4 pass. See WIDEBUTTERFLY for how each step works
    NARROWBLOCK:
      MOVQ R8,  R10 // R10 points at a, start at the beginning of the block
      MOVQ BX,  R9  // R9 points at w1, start over at the top of this pass in the twiddle table
      MOVQ R12, SI  // SI counts down the bytes of j left in this block. h is 64 bytes here so this only loops once

      NARROWBUTTERFLY:
        VMOVUPD   (R9),          Z30 // Load w1 into Z30
        VMOVDDUP  Z30,           Z4  // Z4 = (w1.real, w1.real)
        VPERMILPD $0xFF,    Z30, Z5  // Z5 = (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),   Z30 // Load w2 into Z30
        VMOVDDUP  Z30,           Z6  // Z6 = (w2.real, w2.real)
        VPERMILPD $0xFF,    Z30, Z7  // Z7 = (w2.imag, w2.imag), also (w3.real, w3.real)
        VPXORQ    Z31,      Z6,  Z8  // Z8 = (w3.imag, w3.imag)

        VMOVUPD (R10),        Z0 // Load a into Z0
        VMOVUPD (R10)(R12*1), Z1 // Load b into Z1
        VMOVUPD (R10)(R12*2), Z2 // Load c into Z2
        VMOVUPD (R10)(R13*1), Z3 // Load d into Z3

        VPERMILPD      $0x55, Z1,  Z9  // Z9 = b swapped
        VPERMILPD      $0x55, Z3,  Z10 // Z10 = d swapped
        VMULPD         Z5,    Z9,  Z9  // Z9 = swapped b * w1.imag
        VMULPD         Z5,    Z10, Z10 // Z10 = swapped d * w1.imag
        VFMADDSUB231PD Z4,    Z1,  Z9  // Z9 = w1*b
        VFMADDSUB231PD Z4,    Z3,  Z10 // Z10 = w1*d

        VADDPD Z9,  Z0, Z11 // Z11 = a' = a + w1*b
        VSUBPD Z9,  Z0, Z12 // Z12 = b' = a - w1*b
        VADDPD Z10, Z2, Z13 // Z13 = c' = c + w1*d
        VSUBPD Z10, Z2, Z14 // Z14 = d' = c - w1*d

        VPERMILPD      $0x55, Z13, Z0 // Z0 = c' swapped
        VPERMILPD      $0x55, Z14, Z1 // Z1 = d' swapped
        VMULPD         Z7,    Z0,  Z0 // Z0 = swapped c' * w2.imag
        VMULPD         Z8,    Z1,  Z1 // Z1 = swapped d' * w3.imag
        VFMADDSUB231PD Z6,    Z13, Z0 // Z0 = w2*c'
        VFMADDSUB231PD Z7,    Z14, Z1 // Z1 = w3*d'

        VADDPD Z0, Z11, Z2  // Z2 = A = a' + w2*c'
        VADDPD Z1, Z12, Z3  // Z3 = B = b' + w3*d'
        VSUBPD Z0, Z11, Z9  // Z9 = C = a' - w2*c'
        VSUBPD Z1, Z12, Z10 // Z10 = D = b' - w3*d'

        VMOVUPD Z2,  (R10)        // Store A where a came from
        VMOVUPD Z3,  (R10)(R12*1) // Store B where b came from
        VMOVUPD Z9,  (R10)(R12*2) // Store C where c came from
        VMOVUPD Z10, (R10)(R13*1) // Store D where d came from

        ADDQ $64, R9         // Add 64 (4 * 16) to R9. This moves the twiddle pointer forward by 4 entries
        ADDQ $64, R10        // Add 64 (4 * 16) to R10. This moves forward by 4 points
        SUBQ $64, SI         // Subtract 64 from SI since we just did 4 values of j
        JNZ  NARROWBUTTERFLY // If SI is not zero then jump back to NARROWBUTTERFLY

      LEAQ (R8)(R12*4), R8 // Move R8 forward by 4h to the start of the next block
      CMPQ R8,          AX // Compare R8 against the end of dst
      JCS  NARROWBLOCK     // If R8 is below the end of dst then jump back to NARROWBLOCK

    PASSDONE:
    LEAQ (BX)(R12*2), BX  // BX = BX + R12*2. This pass used 2h twiddle entries (w1 and w2), so move BX to where the next pass starts
    SHLQ $2,          R12 // Shift left by 2 to multiply h by 4 for the next pass
    CMPQ R12,         DX  // Compare the new h against h for the last pass (n*4 bytes)
    JLS  PASS4            // If it is less than or the same then jump back to PASS4. JLS is the unsigned "less than or equal"

	// We have to clear the top halves of the vector registers before going back
	// to Go. Right after this returns Go zeroes X15 with XORPS, which is an old
	// SSE instruction, and on Intel running SSE instructions while the top
	// halves are dirty is slow
  VZEROUPPER // Zero the upper bits of the vector registers
  RET        // We are done, return

// const_negate_imaginary_y is const_negate_imaginary for a YMM register, so 2
// complex numbers instead of 4
DATA const_negate_imaginary_y<>+0(SB)/8,  $0x0000000000000000
DATA const_negate_imaginary_y<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_imaginary_y<>+16(SB)/8, $0x0000000000000000
DATA const_negate_imaginary_y<>+24(SB)/8, $0x8000000000000000
GLOBL const_negate_imaginary_y<>(SB), (RODATA+NOPTR), $32

// const_negate_all_y is const_negate_all for a YMM register
DATA const_negate_all_y<>+0(SB)/8,  $0x8000000000000000
DATA const_negate_all_y<>+8(SB)/8,  $0x8000000000000000
DATA const_negate_all_y<>+16(SB)/8, $0x8000000000000000
DATA const_negate_all_y<>+24(SB)/8, $0x8000000000000000
GLOBL const_negate_all_y<>(SB), (RODATA+NOPTR), $32

// func __fastFourierTransform_AVX_FMA(dst, src []complex128)
//
// Same as the AVX512 version but with 256 bit YMM registers, for CPUs that have
// AVX and FMA but not AVX512. It uses the exact same tables
//
// What is different from the AVX512 version:
//
//   - A YMM register only holds 2 complex numbers instead of 4, so every loop
//     does half as much and runs twice as many times
//   - There are only 16 YMM registers and 1 set of butterfly registers uses all
//     16 of them. So there is no WIDEBLOCK, every pass uses 1 set at a time
//   - The transpose in PASS1 is 2x2 instead of 4x4 since a YMM only has 2 lanes
//   - The sign flips use VXORPD instead of VPXORQ. VPXOR on YMM registers needs
//     AVX2 but VXORPD only needs AVX. On ZMM registers it is the other way
//     around, VXORPD needs AVX512DQ, which is why the AVX512 version uses
//     VPXORQ
TEXT ·__fastFourierTransform_AVX_FMA(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI // Load the pointer of dst into DI
  MOVQ src_base+24(FP), SI // Load the pointer of src into SI
  MOVQ dst_len+8(FP),   AX // Load the length of dst into AX, this is n

  VMOVUPD const_negate_imaginary_y<>(SB), Y12 // Load the imaginary sign mask into Y12

  LEAQ ·fourierScatter4096(SB), BX // Point BX at the start of the bit reversal table

  MOVQ AX, R9  // Copy n into R9
  SHLQ $2, R9  // R9 = n*4, the byte distance to a quarter of the way through src
  MOVQ R9, R8  // Copy n*4 into R8
  ADDQ R9, R8  // R8 = n*8, the byte distance to halfway through src
  MOVQ R8, R10 // Copy n*8 into R10
  ADDQ R9, R10 // R10 = n*12, the byte distance to three quarters of the way through src

  MOVQ AX, CX  // Copy n into CX, this is our loop counter
  SHRQ $3, CX  // Shift right by 3 to divide by 8. We go through the first quarter of src 2 at a time, so n/8 loops (512)

	// Same as PASS1 in the AVX512 version, but only 2 butterflies side by side
	// (q0 and q1) instead of 4
  PASS1_AVXFMA:
    VMOVUPD (SI),        Y0 // Load a0 for q0 and q1 into Y0
    VMOVUPD (SI)(R8*1),  Y1 // Load a1 into Y1, from halfway through src
    VMOVUPD (SI)(R9*1),  Y2 // Load a2 into Y2, from a quarter of the way through src
    VMOVUPD (SI)(R10*1), Y3 // Load a3 into Y3, from three quarters of the way through src

    VADDPD Y1, Y0, Y4 // b0 = a0 + a1, stored in Y4
    VSUBPD Y1, Y0, Y5 // b1 = a0 - a1, stored in Y5
    VADDPD Y3, Y2, Y6 // b2 = a2 + a3, stored in Y6
    VSUBPD Y3, Y2, Y7 // b3 = a2 - a3, stored in Y7

    VPERMILPD $0x5, Y7, Y7 // Swap the halves of b3. 0x5 is 0101, the YMM version of 0x55
    VXORPD    Y12,  Y7, Y7 // Flip the sign of the new imaginary half. Y7 is now -i * b3

    VADDPD Y6, Y4, Y8  // c0 = b0 + b2, stored in Y8
    VADDPD Y7, Y5, Y9  // c1 = b1 + (-i * b3), stored in Y9
    VSUBPD Y6, Y4, Y10 // c2 = b0 - b2, stored in Y10
    VSUBPD Y7, Y5, Y11 // c3 = b1 - (-i * b3), stored in Y11

    // Same transpose problem as the AVX512 version but 2x2:
    //
    //   Y8  = [c0(q0), c0(q1)]
    //   Y9  = [c1(q0), c1(q1)]
    //   Y10 = [c2(q0), c2(q1)]
    //   Y11 = [c3(q0), c3(q1)]
    //
		// A group of 4 outputs is 64 bytes, which is 2 YMM registers. So for q0 we
		// want [c0(q0), c1(q0)] and [c2(q0), c3(q0)], and the same for q1
    //
		// VPERM2F128 builds a register out of the lanes of two other registers. In
		// Go order it is VPERM2F128 $imm, b, a, dst. The right hex digit of the
		// immediate picks the low lane of dst and the left digit picks the high
		// lane: 0 is the low lane of a, 1 the high lane of a, 2 the low lane of b
		// and 3 the high lane of b
    //
    //   0x20 -> low lane of a, then low lane of b
    //   0x31 -> high lane of a, then high lane of b
    VPERM2F128 $0x20, Y9,  Y8,  Y0 // Y0 = [c0(q0), c1(q0)]
    VPERM2F128 $0x20, Y11, Y10, Y1 // Y1 = [c2(q0), c3(q0)]
    VPERM2F128 $0x31, Y9,  Y8,  Y2 // Y2 = [c0(q1), c1(q1)]
    VPERM2F128 $0x31, Y11, Y10, Y3 // Y3 = [c2(q1), c3(q1)]

    MOVL 0(BX), R11           // Load the dst byte offset for q0 into R11
    MOVL 4(BX), R12           // Load the dst byte offset for q1 into R12
    VMOVUPD Y0, (DI)(R11*1)   // Store the first half of q0 at dst + R11
    VMOVUPD Y1, 32(DI)(R11*1) // Store the second half of q0 32 bytes after that
    VMOVUPD Y2, (DI)(R12*1)   // Store the first half of q1 at dst + R12
    VMOVUPD Y3, 32(DI)(R12*1) // Store the second half of q1 32 bytes after that

    ADDQ $32, SI      // Add 32 (2 * 16) to SI. This moves src forward by 2 complex numbers
    ADDQ $8,  BX      // Add 8 (2 * 4) to BX. This moves the scatter table forward by 2 entries
    SUBQ $1,  CX      // Subtract 1 from the CX loop counter
    JNZ  PASS1_AVXFMA // If CX is not zero then jump back to the start of PASS1_AVXFMA

  LEAQ    ·fourierTwiddles4096(SB), BX  // Point BX at the start of the twiddle table
  VMOVUPD const_negate_all_y<>(SB),  Y15 // Load the sign mask for every float64 into Y15, used to get w3 from w2

  MOVQ dst_len+8(FP), DX  // Load n into DX again
  SHLQ $4,            DX  // DX = n*16, the size of dst in bytes
  LEAQ (DI)(DX*1),    AX  // AX = the address right after the end of dst
  SHRQ $2,            DX  // DX = n*4, which is h in bytes on the last pass
  MOVQ $64,           R12 // R12 is h in bytes, starting at 4 * 16 = 64

  // Same passes as PASS4 in the AVX512 version, see the notes there
  PASS4_AVXFMA:
    LEAQ (R12)(R12*2), R13 // R13 = 3h in bytes, the distance from a to d
    MOVQ DI,           R8  // R8 is the start of the current block, start at the beginning of dst

    BLOCK_AVXFMA:
      MOVQ R8,  R10 // R10 points at a, start at the beginning of the block
      MOVQ BX,  R9  // R9 points at w1, start over at the top of this pass in the twiddle table
      MOVQ R12, SI  // SI counts down the bytes of j left in this block, starting at h

      BUTTERFLY_AVXFMA:
				// We are short on registers, so Y0 holds the raw twiddles first and
				// then gets a loaded into it once the twiddles have been split out
        VMOVUPD   (R9),         Y0 // Load w1 for j and j+1 into Y0
        VMOVDDUP  Y0,           Y4 // Y4 = (w1.real, w1.real)
        VPERMILPD $0xF,     Y0, Y5 // Y5 = (w1.imag, w1.imag). 0xF is 1111, the YMM version of 0xFF
        VMOVUPD   (R9)(R12*1),  Y0 // Load w2 for j and j+1 into Y0
        VMOVDDUP  Y0,           Y6 // Y6 = (w2.real, w2.real)
        VPERMILPD $0xF,     Y0, Y7 // Y7 = (w2.imag, w2.imag), also (w3.real, w3.real)
        VXORPD    Y15,      Y6, Y8 // Y8 = (-w2.real, -w2.real), which is (w3.imag, w3.imag)

        VMOVUPD (R10),        Y0 // Load a for j and j+1 into Y0
        VMOVUPD (R10)(R12*1), Y1 // Load b into Y1, h bytes after a
        VMOVUPD (R10)(R12*2), Y2 // Load c into Y2, 2h bytes after a
        VMOVUPD (R10)(R13*1), Y3 // Load d into Y3, 3h bytes after a

        // Same 3 step complex multiply as WIDEBUTTERFLY in the AVX512 version
        VPERMILPD      $0x5, Y1,  Y9  // Y9 = b swapped
        VPERMILPD      $0x5, Y3,  Y10 // Y10 = d swapped
        VMULPD         Y5,   Y9,  Y9  // Y9 = swapped b * w1.imag
        VMULPD         Y5,   Y10, Y10 // Y10 = swapped d * w1.imag
        VFMADDSUB231PD Y4,   Y1,  Y9  // Y9 = (b * w1.real) -/+ Y9, which is w1*b
        VFMADDSUB231PD Y4,   Y3,  Y10 // Y10 = (d * w1.real) -/+ Y10, which is w1*d

        VADDPD Y9,  Y0, Y11 // Y11 = a' = a + w1*b
        VSUBPD Y9,  Y0, Y12 // Y12 = b' = a - w1*b
        VADDPD Y10, Y2, Y13 // Y13 = c' = c + w1*d
        VSUBPD Y10, Y2, Y14 // Y14 = d' = c - w1*d

        VPERMILPD      $0x5, Y13, Y0 // Y0 = c' swapped
        VPERMILPD      $0x5, Y14, Y1 // Y1 = d' swapped
        VMULPD         Y7,   Y0,  Y0 // Y0 = swapped c' * w2.imag
        VMULPD         Y8,   Y1,  Y1 // Y1 = swapped d' * w3.imag
        VFMADDSUB231PD Y6,   Y13, Y0 // Y0 = w2*c'
        VFMADDSUB231PD Y7,   Y14, Y1 // Y1 = w3*d', Y7 is w2.imag which is the same as w3.real

        VADDPD Y0, Y11, Y2  // Y2 = A = a' + w2*c'
        VADDPD Y1, Y12, Y3  // Y3 = B = b' + w3*d'
        VSUBPD Y0, Y11, Y9  // Y9 = C = a' - w2*c'
        VSUBPD Y1, Y12, Y10 // Y10 = D = b' - w3*d'

        VMOVUPD Y2,  (R10)        // Store A where a came from
        VMOVUPD Y3,  (R10)(R12*1) // Store B where b came from
        VMOVUPD Y9,  (R10)(R12*2) // Store C where c came from
        VMOVUPD Y10, (R10)(R13*1) // Store D where d came from

        ADDQ $32, R9          // Add 32 (2 * 16) to R9. This moves the twiddle pointer forward by 2 entries
        ADDQ $32, R10         // Add 32 (2 * 16) to R10. This moves forward by 2 points
        SUBQ $32, SI          // Subtract 32 from SI since we just did 2 values of j
        JNZ  BUTTERFLY_AVXFMA // If SI is not zero then jump back to BUTTERFLY_AVXFMA

      LEAQ (R8)(R12*4), R8 // Move R8 forward by 4h to the start of the next block
      CMPQ R8,          AX // Compare R8 against the end of dst
      JCS  BLOCK_AVXFMA    // If R8 is below the end of dst then jump back to BLOCK_AVXFMA

    LEAQ (BX)(R12*2), BX  // Move BX forward by 2h entries to where the next pass's twiddles start
    SHLQ $2,          R12 // Multiply h by 4 for the next pass
    CMPQ R12,         DX  // Compare the new h against h for the last pass
    JLS  PASS4_AVXFMA     // If it is less than or the same then jump back to PASS4_AVXFMA

  VZEROUPPER // Zero the upper bits of the vector registers before going back to Go, see the AVX512 version
  RET        // We are done, return

// func __fastFourierTransform_AVX(dst, src []complex128)
//
// Same as the AVX_FMA version but without FMA, for CPUs that have AVX but not
// FMA3. That is pretty much just Sandy Bridge and Ivy Bridge (including the
// Xeon E5 v1 and v2), everything from Haswell on has FMA
//
// The only thing that changes is the complex multiply. Without VFMADDSUB231PD
// we do the (wr, wr) multiply on its own and then use VADDSUBPD to finish it.
// VADDSUBPD subtracts on the real half and adds on the imaginary half, which
// is the same thing VFMADDSUB231PD does after its multiply. AVX512 doesn't
// have a ZMM version of VADDSUBPD, but AVX has a YMM one
//
// On Sandy and Ivy Bridge VADDPD, VSUBPD and VADDSUBPD all only run on port 1,
// so that port ends up being the bottleneck. There isn't really a way around
// that on those CPUs
TEXT ·__fastFourierTransform_AVX(SB), NOSPLIT, $0-48
  MOVQ dst_base+0(FP),  DI // Load the pointer of dst into DI
  MOVQ src_base+24(FP), SI // Load the pointer of src into SI
  MOVQ dst_len+8(FP),   AX // Load the length of dst into AX, this is n

  VMOVUPD const_negate_imaginary_y<>(SB), Y12 // Load the imaginary sign mask into Y12

  LEAQ ·fourierScatter4096(SB), BX // Point BX at the start of the bit reversal table

  MOVQ AX, R9  // Copy n into R9
  SHLQ $2, R9  // R9 = n*4, the byte distance to a quarter of the way through src
  MOVQ R9, R8  // Copy n*4 into R8
  ADDQ R9, R8  // R8 = n*8, the byte distance to halfway through src
  MOVQ R8, R10 // Copy n*8 into R10
  ADDQ R9, R10 // R10 = n*12, the byte distance to three quarters of the way through src

  MOVQ AX, CX  // Copy n into CX, this is our loop counter
  SHRQ $3, CX  // Shift right by 3 to divide by 8, so n/8 loops (512)

  // PASS1 is exactly the same as PASS1_AVXFMA, there is no multiply in it
  PASS1_AVX:
    VMOVUPD (SI),        Y0 // Load a0 for q0 and q1 into Y0
    VMOVUPD (SI)(R8*1),  Y1 // Load a1 into Y1, from halfway through src
    VMOVUPD (SI)(R9*1),  Y2 // Load a2 into Y2, from a quarter of the way through src
    VMOVUPD (SI)(R10*1), Y3 // Load a3 into Y3, from three quarters of the way through src

    VADDPD Y1, Y0, Y4 // b0 = a0 + a1, stored in Y4
    VSUBPD Y1, Y0, Y5 // b1 = a0 - a1, stored in Y5
    VADDPD Y3, Y2, Y6 // b2 = a2 + a3, stored in Y6
    VSUBPD Y3, Y2, Y7 // b3 = a2 - a3, stored in Y7

    VPERMILPD $0x5, Y7, Y7 // Swap the halves of b3
    VXORPD    Y12,  Y7, Y7 // Flip the sign of the new imaginary half. Y7 is now -i * b3

    VADDPD Y6, Y4, Y8  // c0 = b0 + b2, stored in Y8
    VADDPD Y7, Y5, Y9  // c1 = b1 + (-i * b3), stored in Y9
    VSUBPD Y6, Y4, Y10 // c2 = b0 - b2, stored in Y10
    VSUBPD Y7, Y5, Y11 // c3 = b1 - (-i * b3), stored in Y11

    VPERM2F128 $0x20, Y9,  Y8,  Y0 // Y0 = [c0(q0), c1(q0)]
    VPERM2F128 $0x20, Y11, Y10, Y1 // Y1 = [c2(q0), c3(q0)]
    VPERM2F128 $0x31, Y9,  Y8,  Y2 // Y2 = [c0(q1), c1(q1)]
    VPERM2F128 $0x31, Y11, Y10, Y3 // Y3 = [c2(q1), c3(q1)]

    MOVL 0(BX), R11           // Load the dst byte offset for q0 into R11
    MOVL 4(BX), R12           // Load the dst byte offset for q1 into R12
    VMOVUPD Y0, (DI)(R11*1)   // Store the first half of q0 at dst + R11
    VMOVUPD Y1, 32(DI)(R11*1) // Store the second half of q0 32 bytes after that
    VMOVUPD Y2, (DI)(R12*1)   // Store the first half of q1 at dst + R12
    VMOVUPD Y3, 32(DI)(R12*1) // Store the second half of q1 32 bytes after that

    ADDQ $32, SI   // Add 32 (2 * 16) to SI. This moves src forward by 2 complex numbers
    ADDQ $8,  BX   // Add 8 (2 * 4) to BX. This moves the scatter table forward by 2 entries
    SUBQ $1,  CX   // Subtract 1 from the CX loop counter
    JNZ  PASS1_AVX // If CX is not zero then jump back to the start of PASS1_AVX

  LEAQ    ·fourierTwiddles4096(SB), BX  // Point BX at the start of the twiddle table
  VMOVUPD const_negate_all_y<>(SB),  Y15 // Load the sign mask for every float64 into Y15, used to get w3 from w2

  MOVQ dst_len+8(FP), DX  // Load n into DX again
  SHLQ $4,            DX  // DX = n*16, the size of dst in bytes
  LEAQ (DI)(DX*1),    AX  // AX = the address right after the end of dst
  SHRQ $2,            DX  // DX = n*4, which is h in bytes on the last pass
  MOVQ $64,           R12 // R12 is h in bytes, starting at 4 * 16 = 64

  PASS4_AVX:
    LEAQ (R12)(R12*2), R13 // R13 = 3h in bytes, the distance from a to d
    MOVQ DI,           R8  // R8 is the start of the current block, start at the beginning of dst

    BLOCK_AVX:
      MOVQ R8,  R10 // R10 points at a, start at the beginning of the block
      MOVQ BX,  R9  // R9 points at w1, start over at the top of this pass in the twiddle table
      MOVQ R12, SI  // SI counts down the bytes of j left in this block, starting at h

      BUTTERFLY_AVX:
        VMOVUPD   (R9),         Y0 // Load w1 for j and j+1 into Y0
        VMOVDDUP  Y0,           Y4 // Y4 = (w1.real, w1.real)
        VPERMILPD $0xF,     Y0, Y5 // Y5 = (w1.imag, w1.imag)
        VMOVUPD   (R9)(R12*1),  Y0 // Load w2 for j and j+1 into Y0
        VMOVDDUP  Y0,           Y6 // Y6 = (w2.real, w2.real)
        VPERMILPD $0xF,     Y0, Y7 // Y7 = (w2.imag, w2.imag), also (w3.real, w3.real)
        VXORPD    Y15,      Y6, Y8 // Y8 = (w3.imag, w3.imag)

        VMOVUPD (R10),        Y0 // Load a for j and j+1 into Y0
        VMOVUPD (R10)(R12*1), Y1 // Load b into Y1, h bytes after a
        VMOVUPD (R10)(R12*2), Y2 // Load c into Y2, 2h bytes after a
        VMOVUPD (R10)(R13*1), Y3 // Load d into Y3, 3h bytes after a

        // w1*b without FMA, for w = (wr, wi) and b = (br, bi):
        //
				//   1. Swap the halves of b and multiply by (wi, wi) to get (bi*wi,
				//      br*wi)
        //   2. Multiply b by (wr, wr) to get (br*wr, bi*wr)
				//   3. VADDSUBPD does step 2 - step 1 on the real half and step 2 +
				//      step 1 on the imaginary half. That gives (br*wr - bi*wi, bi*wr
				//      + br*wi)
        //
        // Step 2 writes over b since we are done with it, so this doesn't need
        // any more registers than the FMA version
        VPERMILPD $0x5, Y1,  Y9  // Y9 = b swapped
        VPERMILPD $0x5, Y3,  Y10 // Y10 = d swapped
        VMULPD    Y5,   Y9,  Y9  // Y9 = swapped b * w1.imag
        VMULPD    Y5,   Y10, Y10 // Y10 = swapped d * w1.imag
        VMULPD    Y4,   Y1,  Y1  // Y1 = b * w1.real, overwriting b
        VMULPD    Y4,   Y3,  Y3  // Y3 = d * w1.real, overwriting d
        VADDSUBPD Y9,   Y1,  Y9  // Y9 = Y1 -/+ Y9, which is w1*b
        VADDSUBPD Y10,  Y3,  Y10 // Y10 = Y3 -/+ Y10, which is w1*d

        VADDPD Y9,  Y0, Y11 // Y11 = a' = a + w1*b
        VSUBPD Y9,  Y0, Y12 // Y12 = b' = a - w1*b
        VADDPD Y10, Y2, Y13 // Y13 = c' = c + w1*d
        VSUBPD Y10, Y2, Y14 // Y14 = d' = c - w1*d

        VPERMILPD $0x5, Y13, Y0  // Y0 = c' swapped
        VPERMILPD $0x5, Y14, Y1  // Y1 = d' swapped
        VMULPD    Y7,   Y0,  Y0  // Y0 = swapped c' * w2.imag
        VMULPD    Y8,   Y1,  Y1  // Y1 = swapped d' * w3.imag
        VMULPD    Y6,   Y13, Y13 // Y13 = c' * w2.real, overwriting c'
        VMULPD    Y7,   Y14, Y14 // Y14 = d' * w3.real, overwriting d'. Y7 is w2.imag which is the same as w3.real
        VADDSUBPD Y0,   Y13, Y0  // Y0 = Y13 -/+ Y0, which is w2*c'
        VADDSUBPD Y1,   Y14, Y1  // Y1 = Y14 -/+ Y1, which is w3*d'

        VADDPD Y0, Y11, Y2  // Y2 = A = a' + w2*c'
        VADDPD Y1, Y12, Y3  // Y3 = B = b' + w3*d'
        VSUBPD Y0, Y11, Y9  // Y9 = C = a' - w2*c'
        VSUBPD Y1, Y12, Y10 // Y10 = D = b' - w3*d'

        VMOVUPD Y2,  (R10)        // Store A where a came from
        VMOVUPD Y3,  (R10)(R12*1) // Store B where b came from
        VMOVUPD Y9,  (R10)(R12*2) // Store C where c came from
        VMOVUPD Y10, (R10)(R13*1) // Store D where d came from

        ADDQ $32, R9       // Add 32 (2 * 16) to R9. This moves the twiddle pointer forward by 2 entries
        ADDQ $32, R10      // Add 32 (2 * 16) to R10. This moves forward by 2 points
        SUBQ $32, SI       // Subtract 32 from SI since we just did 2 values of j
        JNZ  BUTTERFLY_AVX // If SI is not zero then jump back to BUTTERFLY_AVX

      LEAQ (R8)(R12*4), R8 // Move R8 forward by 4h to the start of the next block
      CMPQ R8,          AX // Compare R8 against the end of dst
      JCS  BLOCK_AVX       // If R8 is below the end of dst then jump back to BLOCK_AVX

    LEAQ (BX)(R12*2), BX  // Move BX forward by 2h entries to where the next pass's twiddles start
    SHLQ $2,          R12 // Multiply h by 4 for the next pass
    CMPQ R12,         DX  // Compare the new h against h for the last pass
    JLS  PASS4_AVX        // If it is less than or the same then jump back to PASS4_AVX

  VZEROUPPER // Zero the upper bits of the vector registers before going back to Go, see the AVX512 version
  RET        // We are done, return
