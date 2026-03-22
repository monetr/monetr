#include "textflag.h"

// Constant for 1.0 as float64, used to compute the reciprocal of the norm.
DATA const_one_f64<>+0(SB)/8, $1.0
GLOBL const_one_f64<>(SB), (RODATA+NOPTR), $8

// Constant for 1.0 as float32, used to compute the reciprocal of the norm.
DATA const_one_f32<>+0(SB)/4, $1.0
GLOBL const_one_f32<>(SB), (RODATA+NOPTR), $4

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
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSD        const_one_f64<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSD       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSD        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSD 0(SP), Y0               // Broadcast the reciprocal to all 4 lanes of YMM0.

  LOOP2:
    VMOVUPD (BX), Y1   // Load the current 4 float64s into the YMM1 register again.
    VMULPD  Y0, Y1, Y1 // Multiply by the reciprocal of the norm we calculated above.
    VMOVUPD Y1, (BX)   // Store the values back in the array.
    ADDQ    $32, BX    // Add 32 (4 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $4, DX     // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector64_AVX_FMA(input []float64)
TEXT ·__normalizeVector64_AVX_FMA(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPD Y0, Y0, Y0 // Clear the YMM0 register to store the normalization weight.

  LOOP:
    VMOVUPD     (AX), Y1  // Load the current 4 float64s from the input vector into the YMM1 register.
    VFMADD231PD Y1,   Y1, Y0 // Square the Y1 register and add the result to the Y0 accumulator register.
    ADDQ        $32,  AX  // Add 32 (4 * 8) to the AX register. This moves the pointer forward.
    SUBQ        $4,   CX  // Subtract 4 from the CX length register since we are going 4 at a time.
    JNZ         LOOP  // If the CX register is not zero then jump to the beginning of the loop again.

  // TODO: I think this bit could somehow be improved but I have no idea how.
  VHADDPD     Y0, Y0, Y0     // Do a horizontal sum of the two 128 bit pairs in YMM0.
  VPERM2F128  $1, Y0, Y0, Y1 // Take the high 128 bits of YMM0 and stash them in the low 128 of YMM1.
  ADDPD       X1, X0         // Add the low 128 bits of each as XMM0 and XMM1.
  SQRTPD      X0, X0         // Calculate the square root of the low 128.
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSD        const_one_f64<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSD       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSD        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSD 0(SP), Y0               // Broadcast the reciprocal to all 4 lanes of YMM0.

  LOOP2:
    VMOVUPD (BX), Y1   // Load the current 4 float64s into the YMM1 register again.
    VMULPD  Y0, Y1, Y1 // Multiply by the reciprocal of the norm we calculated above.
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
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSS        const_one_f32<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSS       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSS        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSS 0(SP), Y0               // Broadcast the reciprocal to all 8 lanes of YMM0.

  LOOP2:
    VMOVUPS (BX), Y1   // Load the current 8 float32s into the YMM1 register again.
    VMULPS  Y0, Y1, Y1 // Multiply by the reciprocal of the norm we calculated above.
    VMOVUPS Y1, (BX)   // Store the values back in the array.
    ADDQ    $32, BX    // Add 32 (4 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $8, DX     // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector32_AVX_FMA(input []float32)
TEXT ·__normalizeVector32_AVX_FMA(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPS Y0, Y0, Y0 // Clear the YMM0 register to store the normalization weight.

  LOOP:
    VMOVUPS     (AX), Y1  // Load the current 8 float32s from the input vector into the YMM1 register.
    VFMADD231PS Y1,   Y1, Y0 // Square the Y1 register and add the result to the Y0 accumulator register.
    ADDQ        $32,  AX  // Add 32 (4 * 8) to the AX register. This moves the pointer forward.
    SUBQ        $8,   CX  // Subtract 8 from the CX length register since we are going 8 at a time.
    JNZ         LOOP  // If the CX register is not zero then jump to the beginning of the loop again.

  VHADDPS     Y0, Y0, Y0     // Do a horizontal sum of the two 128 bit pairs in YMM0.
  VPERM2F128  $1, Y0, Y0, Y1 // Take the high 128 bits of YMM0 and stash them in the low 128 of YMM1.
  VADDPS      X0, X1, X0     // Add the high 128 bits that we took from YMM0 with the low 128 bits of YMM0.
  HADDPS      X0, X0         // Do a horizontal add of the 4 32 bit values. They should now be 4 equal values.
  SQRTPS      X0, X0         // Calculate the square root of the low 128.
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSS        const_one_f32<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSS       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSS        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSS 0(SP), Y0               // Broadcast the reciprocal to all 8 lanes of YMM0.

  LOOP2:
    VMOVUPS (BX), Y1   // Load the current 8 float32s into the YMM1 register again.
    VMULPS  Y0, Y1, Y1 // Multiply by the reciprocal of the norm we calculated above.
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
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSD        const_one_f64<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSD       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSD        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSD 0(SP), Z0               // Broadcast the reciprocal to all 8 lanes of ZMM0.

  LOOP2:
    VMOVUPD (BX), Z1   // Load the current 8 float64s into the ZMM1 register again.
    VMULPD  Z0, Z1, Z1 // Multiply by the reciprocal of the norm we calculated above.
    VMOVUPD Z1, (BX)   // Store the values back in the array.
    ADDQ    $64, BX    // Add 64 (8 * 8) to the BX register. This moves the pointer forward.
    SUBQ    $8, DX     // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector64_AVX512_FMA(input []float64)
TEXT ·__normalizeVector64_AVX512_FMA(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPD Z0, Z0, Z0 // Clear the ZMM0 register to store the normalization weight.

  LOOP:
    VMOVUPD     (AX), Z1  // Load the current 8 float64s from the input vector into the ZMM1 register.
    VFMADD231PD Z1,   Z1, Z0 // Square the ZMM1 register and add it to the ZMM0 accumulator register.
    ADDQ        $64,  AX  // Add 64 (8 * 8) to the AX register. This moves the pointer forward.
    SUBQ        $8,   CX  // Subtract 8 from the CX length register since we are going 8 at a time.
    JNZ         LOOP  // If the CX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF64X4 $1, Z0, Y1     // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPD       Y0, Y0, Y0     // Do horizontal add on YMM0 (the lower 256-bits of ZMM0)
  VPERM2F128    $1, Y0, Y0, Y2 // Extract the high 128 bits of YMM0 to YMM2
  VHADDPD       Y1, Y1, Y1     // Do a horizontal add on YMM1 (from the first extract)
  VPERM2F128    $1, Y1, Y1, Y3 // Extract the high 128 bits of YMM1 into YMM3
  VADDPD        Y0, Y1, Y0     // YMM0 += YMM1
  VADDPD        Y0, Y2, Y0     // YMM0 += YMM2
  VADDPD        Y0, Y3, Y0     // YMM0 += YMM3
  VSQRTPD       Y0, Y0         // Get the square root of the YMM0 register.
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSD        const_one_f64<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSD       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSD        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSD 0(SP), Z0               // Broadcast the reciprocal to all 8 lanes of ZMM0.

  LOOP2:
    VMOVUPD (BX), Z1   // Load the current 8 float64s into the ZMM1 register again.
    VMULPD  Z0, Z1, Z1 // Multiply by the reciprocal of the norm we calculated above.
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
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSS        const_one_f32<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSS       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSS        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSS 0(SP), Z0               // Broadcast the reciprocal to all 16 lanes of ZMM0.

  LOOP2:
    VMOVUPS (BX), Z1   // Load the current 16 float32s into the ZMM1 register again.
    VMULPS  Z0, Z1, Z1 // Multiply by the reciprocal of the norm we calculated above.
    VMOVUPS Z1, (BX)   // Store the values back in the array.
    ADDQ    $64, BX    // Add 64 (4 * 16) to the BX register. This moves the pointer forward.
    SUBQ    $16, DX    // Subtract 16 from the DX length register since we are going 16 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.

// func __normalizeVector32_AVX512_FMA(input []float32)
TEXT ·__normalizeVector32_AVX512_FMA(SB), NOSPLIT, $24-0
  MOVQ input_base+0(FP), AX // Load the pointer of the first item in the input vector array.
  MOVQ input_base+0(FP), BX // Load the pointer of the first item in the input vector array.
  MOVQ input_len+8(FP), CX  // Load the length of the input vector into the CX register.
  MOVQ input_len+8(FP), DX  // Load the length of the input vector into the DX register.

  VXORPS Z0, Z0, Z0 // Clear the ZMM0 register to store the normalization weight.

  LOOP:
    VMOVUPS     (AX), Z1  // Load the current 16 float32s from the input vector into the ZMM1 register.
    VFMADD231PS Z1,   Z1, Z0 // Square the ZMM1 register and add it to the ZMM0 accumulator register.
    ADDQ        $64,  AX  // Add 64 (4 * 16) to the AX register. This moves the pointer forward.
    SUBQ        $16,  CX  // Subtract 16 from the CX length register since we are going 16 at a time.
    JNZ         LOOP  // If the CX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF32X8 $1, Z0, Y1     // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPS       Y0, Y1, Y0     // Pairwise add between YMM0 and YMM1
  VHADDPS       Y0, Y0, Y0     // Then do a pairwise add betwen YMM0 and itself.
  VPERM2F128    $1, Y0, Y0, Y1 // Extract the high 128 bits of YMM0 to YMM1
  VADDPS        X0, X1, X0     // Do an add between XMM0 and XMM1
  VHADDPS       X0, X0, X0     // Do a pairwise add between XMM0 and itself.
  VSQRTPS       X0, X0         // Squre our register.
  // Compute the reciprocal of the norm (1.0 / norm) so we can multiply instead of divide.
  MOVSS        const_one_f32<>(SB), X1 // Load 1.0 into XMM1.
  VDIVSS       X0, X1, X1              // Compute 1.0 / norm as a single scalar division.
  MOVSS        X1, 0(SP)               // Store the reciprocal to the stack.
  VBROADCASTSS 0(SP), Z0               // Broadcast the reciprocal to all 16 lanes of ZMM0.

  LOOP2:
    VMOVUPS (BX), Z1   // Load the current 16 float32s into the ZMM1 register again.
    VMULPS  Z0, Z1, Z1 // Multiply by the reciprocal of the norm we calculated above.
    VMOVUPS Z1, (BX)   // Store the values back in the array.
    ADDQ    $64, BX    // Add 64 (4 * 16) to the BX register. This moves the pointer forward.
    SUBQ    $16, DX    // Subtract 16 from the DX length register since we are going 16 at a time.
    JNZ     LOOP2      // If the DX register is not zero then jump to the beginning of the loop again.

  RET // We are done, return.
