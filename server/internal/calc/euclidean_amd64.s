#include "textflag.h"

// func __euclideanDistance(a, b []float64) float64
TEXT ·__euclideanDistance(SB), NOSPLIT, $0-56
  MOVQ a+8(FP), DX  // Load the length of a into the DX register
  MOVQ a+0(FP), AX  // Load the pointer of the first array
  MOVQ b+24(FP), BX // Load the pointer of the second array

  VXORPD Y0, Y0, Y0 // Clear the accumulator registers

  LOOP:
    VMOVUPD (AX), Y1  // Load the current 4 float64s from a into the 256-bit register Y1
    VMOVUPD (BX), Y2  // Load the current 4 float64s from b into the 256-bit register Y2
    VSUBPD Y2, Y1, Y1 // Subtract the 4 from a from the 4 from b and store the result in Y1
    VMULPD Y1, Y1, Y2 // Then square Y1 (the result of the subtraction) and store it in Y2
    VADDPD Y2, Y0, Y0 // Then add the result of Y2 into the accumulator register.

    ADDQ $32, AX      // Add 32 (4 * 8) to the AX and BX registers. This moves the pointer forward
    ADDQ $32, BX      // in the array by 4 items for the next simd operation to take place.

    SUBQ $4, DX       // Subtract 4 from the DX length register since we are going 4 at a time.
    JNZ LOOP          // If the DX register is not zero then jump to the beginning of the loop again.

  // Now the accumulator should have the distance as 4 different values. We need the sum of those
  // values.
  VHADDPD Y0, Y0, Y0
  VPERM2F128 $1, Y0, Y0, Y1
  VADDPD Y0, Y1, Y0
  MOVQ X0, ret+48(FP)
  RET
