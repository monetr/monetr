#include "textflag.h"

// func __normalizeVector64_AVX(input []float64)
TEXT ·__normalizeVector64_AVX(SB), NOSPLIT, $0-24
  MOVQ input+0(FP), AX  // Load the pointer of the first item in the input vector array.
  MOVQ input+0(FP), BX  // Load the pointer of the first item in the input vector array.
  MOVQ input+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPD Y0, Y0, Y0 // Clear the YMM0 register to store the normalization weight.

  LOOP:
    VMOVUPD (AX), Y1  // Load the current 4 float64s from the input vector into the YMM1 register.
    VMULPD Y1, Y1, Y1 // Square the current 4 float64s and overwrite the YMM1 register with the result.
    VADDPD Y1, Y0, Y0 // Add the values to the normalization weight register.
    ADDQ $32, AX      // Add 32 (4 * 8) to the AX register. This moves the pointer forward.
    SUBQ $4, CX       // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ LOOP          // If the CX register is not zero then jump to the beginning of the loop again.

  // I think this bit could somehow be improved but I have no idea how.
  VHADDPD Y0, Y0, Y0         // Do a horizontal sum of the two 128 bit pairs in YMM0.
  VPERM2F128 $1, Y0, Y0, Y1  // Take the high 128 bits of YMM0 and stash them in the low 128 of YMM1.
  ADDPD X1, X0               // Add the low 128 bits of each as XMM0 and XMM1.
  SQRTPD X0, X0              // Calculate the square root of the low 128.
  VINSERTF128 $1, X0, Y0, Y0 // Copy the result in the low 128 of YMM0 into the high of YMM0 so the register is 4 equal values.

  LOOP2:
    VMOVUPD (BX), Y1  // Load the current 4 float64s into the YMM1 register again.
    VDIVPD Y0, Y1, Y1 // Divide them by the normalization weight we calculated above.
    VMOVUPD Y1, (BX)  // Store the values back in the array.
    ADDQ $32, BX      // Add 32 (4 * 8) to the BX register. This moves the pointer forward.
    SUBQ $4, DX       // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ LOOP2         // If the CX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.
