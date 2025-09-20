#include "textflag.h"

// func __euclideanDistance64_AVX(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX(SB), NOSPLIT, $48-56
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

  VXORPD Y0, Y0, Y0 // Clear the accumulator register.

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

// func __euclideanDistance32_AVX(a, b []float32) float32
TEXT ·__euclideanDistance32_AVX(SB), NOSPLIT, $48-52
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

  VXORPS Y0, Y0, Y0 // Clear the accumulator register.

  LOOP:
    VMOVUPS (AX), Y1  // Load the current 8 float32s from a into the 256-bit register Y1.
    VMOVUPS (BX), Y2  // Load the current 8 float32s from b into the 256-bit register Y2.
    VSUBPS Y1, Y2, Y1 // Subtract the 8 from a from the 8 from b and store the result in Y1.
    VMULPS Y1, Y1, Y2 // Then square Y1 (the result of the subtraction) and store it in Y2.
    VADDPS Y2, Y0, Y0 // Then add the result of Y2 into the accumulator register.

    ADDQ $32, AX      // Add 32 (4 * 8) to the AX and BX registers. This moves the pointer forward
    ADDQ $32, BX      // in the array by 4 items for the next simd operation to take place.

    SUBQ $8, DX       // Subtract 8 from the DX length register since we are going 8 at a time.
    JNZ LOOP          // If the DX register is not zero then jump to the beginning of the loop again.

  // Now we have an accumulator register that looks like [1, 2, 3, 4, 5, 6, 7, 8] and we need to get the sum of those 8 values.
  VHADDPS    Y0, Y0, Y0     // Add pairs, now we have [3, 7, 3, 7, 11, 15, 11, 15].
  VPERM2F128 $1, Y0, Y0, Y1 // Take the high 128 bits and store them in XMM1
  VADDPS     X0, X1, X0     // Add the low 128 bits and the high 128 bits as pairs.
  HADDPS     X0, X0         // Add the horizontal pairs together to get 4 equal values.
  MOVL       X0, ret+48(FP) // Extract the lowest 32 bits of the Y0 register via the low X0 register.
  RET                       // Return the euclidean distance.


// func __euclideanDistance64_AVX512(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX512(SB), NOSPLIT, $48-56
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

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
  VADDPD X0, X1, X0         // XMM0 += XMM1 We only care about the low 64 bits.
  VADDPD X0, X2, X0         // XMM0 += XMM2
  VADDPD X0, X3, X0         // XMM0 += XMM3
  MOVQ X0, ret+48(FP)       // Move the low 64 bits from YMM0 into the return address space.
  RET                       // Return

// func __euclideanDistance32_AVX512(a, b []float32) float32
TEXT ·__euclideanDistance32_AVX512(SB), NOSPLIT, $48-52
  MOVQ a_len+8(FP), DX  // Load the length of a into the DX register.
  MOVQ a_base+0(FP), AX  // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

  VXORPS Z0, Z0, Z0 // Clear the accumulator register.

  LOOP:
    VMOVUPS (AX), Z1  // Load the current 16 float64s from a into the 512-bit register ZMM1.
    VMOVUPS (BX), Z2  // Load the current 16 float64s from b into the 512-bit register ZMM2.
    VSUBPS Z1, Z2, Z1 // Subtract the 16 from a from the 16 from b and store the result in ZMM1.
    VMULPS Z1, Z1, Z2 // Then square ZMM1 (the result of the subtraction) and store it in ZMM2.
    VADDPS Z2, Z0, Z0 // Then add the result of ZMM2 into the accumulator register.

    ADDQ $64, AX      // Add 64 (16 * 4) to the AX and BX registers. This moves the pointer forward
    ADDQ $64, BX      // in the array by 16 items for the next simd operation to take place.

    SUBQ $16, DX      // Subtract 16 from the DX length register since we are going 16 at a time.
    JNZ LOOP          // If the DX register is not zero then jump to the beginning of the loop again.

  VEXTRACTF32X8 $1, Z0, Y1     // Extract the high 256-bits of ZMM0 into YMM1
  VHADDPS       Y0, Y0, Y0     // Do horizontal add on YMM0 (the lower 256-bits of ZMM0)
  VPERM2F128    $1, Y0, Y0, Y2 // Extract the high 128 bits of YMM0 to YMM2
  VHADDPS       Y1, Y1, Y1     // Do a horizontal add on YMM1 (from the first extract)
  VPERM2F128    $1, Y1, Y1, Y3 // Extract the high 128 bits of YMM1 into YMM3
  VADDPS        X0, X1, X0     // XMM0 += XMM1 We only care about the low 64 bits.
  VADDPS        X0, X2, X0     // XMM0 += XMM2
  VADDPS        X0, X3, X0     // XMM0 += XMM3
  VHADDPS       X0, X0, X0     // Add the 32-bit pairs. Now we will have all equal values
  MOVL          X0, ret+48(FP) // Move the low 32 bits from YMM0 into the return address space.
  RET                          // Return


