#include "textflag.h"

// func __euclideanDistance64(a, b []float64) float64
TEXT ·__euclideanDistance64(SB), NOSPLIT, $0-56
  MOVQ a+8(FP), DX  // Load the length of a into the DX register.
  MOVQ a+0(FP), AX  // Load the pointer of the first array.
  MOVQ b+24(FP), BX // Load the pointer of the second array.

  VXORPD Y0, Y0, Y0 // Clear the accumulator registers.

  LOOP:
    VMOVUPD (AX), Y1  // Load the current 4 float64s from a into the 256-bit register Y1.
    VMOVUPD (BX), Y2  // Load the current 4 float64s from b into the 256-bit register Y2.
    VSUBPD Y1, Y2, Y1 // Subtract the 4 from a from the 4 from b and store the result in Y1.
    VMULPD Y1, Y1, Y2 // Then square Y1 (the result of the subtraction) and store it in Y2.
    VADDPD Y2, Y0, Y0 // Then add the result of Y2 into the accumulator register.

    ADDQ $32, AX      // Add 32 (4 * 8) to the AX and BX registers. This moves the pointer forward
    ADDQ $32, BX      // in the array by 4 items for the next simd operation to take place.

    SUBQ $4, DX       // Subtract 4 from the DX length register since we are going 4 at a time.
    JNZ LOOP          // If the DX register is not zero then jump to the beginning of the loop again.

  // Now we have an accumulator register that looks like [1, 2, 3, 4] and we need to get the sum of those 4 values.
  VHADDPD Y0, Y0, Y0        // Add 1 + 2 and 3 + 4, now we have [3, 3, 7, 7].
  VPERM2F128 $1, Y0, Y0, Y1 // Take the 7, 7 and move them into the lower 128 bits of the Y1 register.
  VADDPD Y0, Y1, Y0         // Add Y0 and Y1 (really just 3, 3 + 7, 7) and output to the Y0 register.
  MOVQ X0, ret+48(FP)       // Extract the lowest 64 bits of the Y0 register via the low X0 register.
  RET                       // Return the euclidean distance.


// func __euclideanDistance64_AVX512(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX512(SB), NOSPLIT, $0-56
  MOVQ a+8(FP), DX  // Load the length of a into the DX register.
  MOVQ a+0(FP), AX  // Load the pointer of the first array.
  MOVQ b+24(FP), BX // Load the pointer of the second array.

  VXORPD Z0, Z0, Z0 // Clear the accumulator register.

  LOOP:
    VMOVUPD (AX), Z1  // Load the current 8 float64s from a into the 512-bit register ZMM1.
    VMOVUPD (BX), Z2  // Load the current 8 float64s from b into the 512-bit register ZMM2.
    VSUBPD Z1, Z2, Z1 // Subtract the 8 from a from the 8 from b and store the result in ZMM1.
    VMULPD Z1, Z1, Z2 // Then square ZMM1 (the result of the subtraction) and store it in ZMM2.
    VADDPD Z2, Z0, Z0 // Then add the result of ZMM2 into the accumulator register.

    ADDQ $64, AX      // Add 64 (8 * 8) to the AX and BX registers. This moves the pointer forward
    ADDQ $64, BX      // in the array by 8 items for the next simd operation to take place.

    SUBQ $8, DX       // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ LOOP          // If the DX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF64X4 $1, Z0, Y1  // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPD Y0, Y0, Y0        // Do horizontal add on YMM0 (the lower 256-bits of ZMM0)
  VPERM2F128 $1, Y0, Y0, Y2 // Extract the high 128 bits of YMM0 to YMM2
  VHADDPD Y1, Y1, Y1        // Do a horizontal add on YMM1 (from the first extract)
  VPERM2F128 $1, Y1, Y1, Y3 // Extract the high 128 bits of YMM1 into YMM3
  VADDPD Y0, Y1, Y0         // YMM0 += YMM1 We only care about the low 64 bits.
  VADDPD Y0, Y2, Y0         // YMM0 += YMM2
  VADDPD Y0, Y3, Y0         // YMM0 += YMM3
  MOVQ X0, ret+48(FP)       // Move the low 64 bits from YMM0 into the return address space.
  RET                       // Return
