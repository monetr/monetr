#include "textflag.h"

// func __normalizeVector64_AVX(input []float64)
TEXT ·__normalizeVector64_AVX(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPD Y0, Y0, Y0 // Clear the YMM0 register to store the normalization weight.

  LOOP:
    VMOVUPD (AX), Y1   // Load the current 4 float64s from the input vector into the YMM1 register.
    VMULPD  Y1, Y1, Y1 // Square the current 4 float64s and overwrite the YMM1 register with the result.
    VADDPD  Y1, Y0, Y0 // Add the values to the normalization weight register.
    ADDQ    $32, AX    // Add 32 (4 * 8) to the AX register. This moves the pointer forward.
    SUBQ    $4, CX     // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ     LOOP       // If the CX register is not zero then jump to the beginning of the loop again.

  // TODO: I think this bit could somehow be improved but I have no idea how.
  VHADDPD     Y0, Y0, Y0     // Do a horizontal sum of the two 128 bit pairs in YMM0.
  VPERM2F128  $1, Y0, Y0, Y1 // Take the high 128 bits of YMM0 and stash them in the low 128 of YMM1.
  ADDPD       X1, X0         // Add the low 128 bits of each as XMM0 and XMM1.
  SQRTPD      X0, X0         // Calculate the square root of the low 128.
  VINSERTF128 $1, X0, Y0, Y0 // Copy the result in the low 128 of YMM0 into the high of YMM0 so the register is 4 equal values.

  LOOP2:
    VMOVUPD (BX), Y1   // Load the current 4 float64s into the YMM1 register again.
    VDIVPD  Y0, Y1, Y1 // Divide them by the normalization weight we calculated above.
    VMOVUPD Y1, (BX)   // Store the values back in the array.
    ADDQ    $32, BX    // Add 32 (4 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $4, DX     // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector32_AVX(input []float32)
TEXT ·__normalizeVector32_AVX(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPS Y0, Y0, Y0 // Clear the YMM0 register to store the normalization weight.

  LOOP:
    VMOVUPS (AX), Y1   // Load the current 8 float32s from the input vector into the YMM1 register.
    VMULPS  Y1, Y1, Y1 // Square the current 8 float32s and overwrite the YMM1 register with the result.
    VADDPS  Y1, Y0, Y0 // Add the values to the normalization weight register.
    ADDQ    $32, AX    // Add 32 (4 * 8) to the AX register. This moves the pointer forward.
    SUBQ    $8, CX     // Subtract 8 from the CX length register since we are going 8 at a time.
    JNZ     LOOP       // If the CX register is not zero then jump to the beginning of the loop again.

  VHADDPS     Y0, Y0, Y0     // Do a horizontal sum of the two 128 bit pairs in YMM0.
  VPERM2F128  $1, Y0, Y0, Y1 // Take the high 128 bits of YMM0 and stash them in the low 128 of YMM1.
  VADDPS      X0, X1, X0     // Add the high 128 bits that we took from YMM0 with the low 128 bits of YMM0.
  HADDPS      X0, X0         // Do a horizontal add of the 4 32 bit values. They should now be 4 equal values.
  SQRTPS      X0, X0         // Calculate the square root of the low 128.
  VINSERTF128 $1, X0, Y0, Y0 // Copy the result in the low 128 of YMM0 into the high of YMM0 so the register is 8 equal values.

  LOOP2:
    VMOVUPS (BX), Y1   // Load the current 8 float32s into the YMM1 register again.
    VDIVPS  Y0, Y1, Y1 // Divide them by the normalization weight we calculated above.
    VMOVUPS Y1, (BX)   // Store the values back in the array.
    ADDQ    $32, BX    // Add 32 (4 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $8, DX     // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector64_AVX512(input []float64)
TEXT ·__normalizeVector64_AVX512(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPD Z0, Z0, Z0 // Clear the ZMM0 register to store the normalization weight.

  LOOP:
    VMOVUPD (AX), Z1  // Load the current 8 float64s from the input vector into the ZMM1 register.
    VMULPD Z1, Z1, Z1 // Square the current 8 float64s and overwrite the ZMM1 register with the result.
    VADDPD Z1, Z0, Z0 // Add the values to the normalization weight register.
    ADDQ $64, AX      // Add 64 (8 * 8) to the AX register. This moves the pointer forward.
    SUBQ $8, CX       // Subtract 8 from the CX length register since we are going 8 at a time.
    JNZ LOOP          // If the CX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF64X4 $1, Z0, Y1     // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPD       Y0, Y0, Y0     // Do horizontal add on YMM0 (the lower 256-bits of ZMM0)
  VPERM2F128    $1, Y0, Y0, Y2 // Extract the high 128 bits of YMM0 to YMM2
  VHADDPD       Y1, Y1, Y1     // Do a horizontal add on YMM1 (from the first extract)
  VPERM2F128    $1, Y1, Y1, Y3 // Extract the high 128 bits of YMM1 into YMM3
  VADDPD        Y0, Y1, Y0     // YMM0 += YMM1
  VADDPD        Y0, Y2, Y0     // YMM0 += YMM2
  VADDPD        Y0, Y3, Y0     // YMM0 += YMM3
  VSQRTPD       Y0, Y0         // Get the square root of the YMM0 register.
  VINSERTF64X4  $1, Y0, Z0, Z0 // Copy the 256-bits of YMM0 into the high 256 of ZMM0.

  LOOP2:
    VMOVUPD (BX), Z1   // Load the current 8 float64s into the ZMM1 register again.
    VDIVPD  Z0, Z1, Z1 // Divide them by the normalization weight we calculated above.
    VMOVUPD Z1, (BX)   // Store the values back in the array.
    ADDQ    $64, BX    // Add 64 (8 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $8, DX     // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector32_AVX512(input []float32)
TEXT ·__normalizeVector32_AVX512(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPS Z0, Z0, Z0 // Clear the ZMM0 register to store the normalization weight.

  LOOP:
    VMOVUPS (AX), Z1   // Load the current 16 float32s from the input vector into the ZMM1 register.
    VMULPS  Z1, Z1, Z1 // Square the current 16 float32s and overwrite the ZMM1 register with the result.
    VADDPS  Z1, Z0, Z0 // Add the values to the normalization weight register.
    ADDQ    $64, AX    // Add 64 (4 * 16) to the AX register. This moves the pointer forward.
    SUBQ    $16, CX    // Subtract 16 from the CX length register since we are going 16 at a time.
    JNZ     LOOP       // If the CX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF32X8 $1, Z0, Y1     // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPS       Y0, Y1, Y0     // Pairwise add between YMM0 and YMM1
  VHADDPS       Y0, Y0, Y0     // Then do a pairwise add betwen YMM0 and itself.
  VPERM2F128    $1, Y0, Y0, Y1 // Extract the high 128 bits of YMM0 to YMM1
  VADDPS        X0, X1, X0     // Do an add between XMM0 and XMM1
  VHADDPS       X0, X0, X0     // Do a pairwise add between XMM0 and itself.
  VSQRTPS       X0, X0         // Squre our register.
  VINSERTF128   $1, X0, Y0, Y0 // Copy the 4 float32s in our low 128 into the high 128 of YMM0
  VINSERTF32X8  $1, Y0, Z0, Z0 // Now we have 8 equal values in the low 256, copy them into the high 256.

  LOOP2:
    VMOVUPS (BX), Z1   // Load the current 16 float32s into the ZMM1 register again.
    VDIVPS  Z0, Z1, Z1 // Divide them by the normalization weight we calculated above.
    VMOVUPS Z1, (BX)   // Store the values back in the array.
    ADDQ    $64, BX    // Add 64 (4 * 16) to the BX register. This moves the pointer forward.
    SUBQ    $16, DX    // Subtract 16 from the DX length register since we are going 16 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

