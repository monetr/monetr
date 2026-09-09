#include "textflag.h"

DATA  Pow10+0(SB)/8,  $1000000000
DATA  Pow10+8(SB)/8,  $100000000
DATA  Pow10+16(SB)/8, $10000000
DATA  Pow10+24(SB)/8, $1000000
DATA  Pow10+32(SB)/8, $100000
DATA  Pow10+40(SB)/8, $10000
DATA  Pow10+48(SB)/8, $1000
DATA  Pow10+56(SB)/8, $100
DATA  Pow10+64(SB)/8, $10
DATA  Pow10+72(SB)/8, $1
GLOBL Pow10(SB),      RODATA, $80 // 10 * 8

DATA  UnicodeZero+0(SB)/8, $0x3030303030303030
GLOBL UnicodeZero(SB),     RODATA, $8

// func __atoi_AVX512(input *[64]byte) int64
TEXT ·__atoi_AVX512(SB), NOSPLIT, $8-16
	// Load the pointer of the start of the input array.
	MOVQ input_base+0(FP), AX

	// Clear these registers so we can use them below.
	VXORPD Z0, Z0, Z0
	VXORPD Z2, Z2, Z2
	VXORPD Z3, Z3, Z3 // Clears the lower bits for YMM3 too!
	VXORPD Z4, Z4, Z4
	VXORPD Z5, Z5, Z5

	// Load the entire input array (since its 64 bytes) into the ZMM0 register.
	VMOVUPS (AX), Z1

	// Compare our input to an all zero register so we can create a mask. 4 is !=.
	VPCMPUB $4, Z1, Z0, K1

	// Put our mask into a general purpose register so we can turn it into an
	// integer.
	KMOVQ K1, BX
	// Get the index of the most significant bit to convert the mask into an
	// integer. This will be equal to the length of the string number provided
	// minus 1.
	BSRQ BX, BX
	// Add one to our index so we are no longer zero indexed.
	ADDQ $1, BX
	// Multiply it by 8 to get the byte offset we need to get our powers of 10
	// from above.
	IMULQ $8, BX
	// Subtract our offset from 72 so we know where to start in our power of 10
	// constant.
	MOVQ $80, CX
	SUBQ BX, CX
	// Load our powers of 10 into our ZMM6 register using our offset.
	LEAQ Pow10(SB), DX
	VMOVUPS (DX)(CX*1), Y5

	// Broadcast our zero unicode bytes into ZMM2 so we can use this to calculate
	// the numeric values of each. Using our mask we calculated above so we don't
	// overcalculate anything.
	VPBROADCASTB.Z UnicodeZero(SB), K1, Z2
	// Subtract our unicode register from our unicode zero register. After this
	// ZMM1 will now have the numeric representations of digits in each individual
	// byte.
	VXORPD Z1, Z2, Z1

	// Yeet each of our lower 4 bytes from XMM1 into 4 packed 64-bit integers in
	// YMM3. Note: If you are debugging this the value in YMM3 will be little
	// endian!
	VPMOVZXBQ X1, Y3
	// Multiply our numeric values by their respective powers of 10, this now
	// gives us each REAL digit of the final number as an individual int64. The
	// sum of all of these ints will give us our final number.
	VPMULLQ Y3, Y5, Y3

	// TODO Repeat the process above but with the next 4 bytes and moving the
	// power of 10 register along.

	// Take the high 128 bits from YMM3 and put them into XMM4
	VEXTRACTI128 $1, Y3, X4
	// Add XMM3 and XMM4 and write the result to XMM3
	VPADDQ X4, X3, X3           // X0 = [a+c, b+d]
	// Shuffle XXM3 into XMM4 so we can add the remaining 64 bit pairs together.
	VPSHUFD $0b11101110, X3, X4
	// Add the final pairs together. The low 64 bits of XMM3 now is our final
	// number.
	VPADDQ X4, X3, X3

	// Move the low 64 bits from the XMM3 register into the return address space.
	MOVQ X3, ret+8(FP)
	// Return
	RET

// 	VMOVDQU64 X3, (AX)
//
//
//
//
// // Shift our bytes register down by 4 bytes since we just took some off.
// VPSRLDQ $4, Z1, Z1
// // Take 4 more bytes off
// VPMOVZXBQ X1, Y4
//
//
//
// // VMOVDQU64 Z3, (AX)
//
// //KMOVQ K1, BX
// //VMOVDQU64 Z0, (AX)
// //MOVQ BX, (AX)
//
//
// // VPOPCNTB Z0, Z2
// // Subtract Z0 from Z1 and store it in Z1
// // VPTERNLOGD $0x9, Z0, Z1, Z2
//
// // Move the data from the register back to the memory space that the input
// // array was in.
// // VMOVDQU64 Z2, (AX)
//
// RET
//
//
//
// // XOR the entire array against itself for now.
// //	VXORPS Z1, Z1, Z1
