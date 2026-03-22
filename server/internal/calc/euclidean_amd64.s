#include "textflag.h"

// func __euclideanDistance64_AVX(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX(SB), NOSPLIT, $48-56
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

	// We use two accumulator registers here because this loop now processes two
  // independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the multiply/add work before we
	// combine the partial sums at the end.
  VXORPD Y0, Y0, Y0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPD Y3, Y3, Y3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPD 0(AX), Y1  // Load the current 4 float64s from a into the 256-bit register Y1.
    VMOVUPD 0(BX), Y2  // Load the current 4 float64s from b into the 256-bit register Y2.
    VSUBPD  Y1, Y2, Y1 // Subtract the 4 from a from the 4 from b and store the result in Y1.
    VMULPD  Y1, Y1, Y2 // Then square Y1 (the result of the subtraction) and store it in Y2.
    VADDPD  Y2, Y0, Y0 // Then add the result of Y2 into the first accumulator register.
    VMOVUPD 32(AX), Y4 // Load the next 4 float64s from a into the 256-bit register Y4.
    VMOVUPD 32(BX), Y5 // Load the next 4 float64s from b into the 256-bit register Y5.
    VSUBPD  Y4, Y5, Y4 // Subtract the next 4 from a from the next 4 from b and store the result in Y4.
    VMULPD  Y4, Y4, Y5 // Then square Y4 (the result of the subtraction) and store it in Y5.
    VADDPD  Y5, Y3, Y3 // Then add the result of Y5 into the second accumulator register.

    ADDQ $64, AX       // Add 64 (8 * 8) to AX. This moves the pointer forward in a by 8 items for the next SIMD operation to take place.
    ADDQ $64, BX       // Add 64 (8 * 8) to BX. This moves the pointer forward in b by 8 items for the next SIMD operation to take place.

    SUBQ $8, DX        // Subtract 8 from the DX length register since we are now going 8 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPD Y3, Y0, Y0   // Add the second accumulator register into the first so we have one final accumulated vector.

  // Now we have an accumulator register that looks like [1, 2, 3, 4] and we need to get the sum of those 4 values.
  VHADDPD Y0, Y0, Y0        // Add 1 + 2 and 3 + 4, now we have [3, 3, 7, 7].
  VPERM2F128 $1, Y0, Y0, Y1 // Take the 7, 7 and move them into the lower 128 bits of the Y1 register.
  VADDPD Y0, Y1, Y0         // Add Y0 and Y1 (really just 3, 3 + 7, 7) and output to the Y0 register.
  MOVQ X0, ret+48(FP)       // Extract the lowest 64 bits of the Y0 register via the low X0 register.
  RET                       // Return the euclidean distance.

// func __euclideanDistance64_AVX_FMA(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX_FMA(SB), NOSPLIT, $48-56
  MOVQ a_len+8(FP),   DX // Load the length of A into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

	// We use two accumulator registers here because this loop now processes two
  // independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the fused multiply-add work
	// before we combine the partial sums at the end.
  VXORPD Y0, Y0, Y0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPD Y3, Y3, Y3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPD     0(AX), Y1    // Load the current 4 float64s from a into the 256-bit register Y1.
    VMOVUPD     0(BX), Y2    // Load the current 4 float64s from b into the 256-bit register Y2.
    VSUBPD      Y1, Y2, Y1   // Subtract the 4 from a from the 4 from b and store the result in Y1.
    VFMADD231PD Y1, Y1, Y0   // Square Y1 and add the result to the first accumulator register in a single fused instruction.
    VMOVUPD     32(AX), Y4   // Load the next 4 float64s from a into the 256-bit register Y4.
    VMOVUPD     32(BX), Y5   // Load the next 4 float64s from b into the 256-bit register Y5.
    VSUBPD      Y4, Y5, Y4   // Subtract the next 4 from a from the next 4 from b and store the result in Y4.
    VFMADD231PD Y4, Y4, Y3   // Square Y4 and add the result to the second accumulator register in a single fused instruction.

    ADDQ $64, AX       // Add 64 (8 * 8) to AX. This moves the pointer forward in a by 8 items for the next SIMD operation to take place.
    ADDQ $64, BX       // Add 64 (8 * 8) to BX. This moves the pointer forward in b by 8 items for the next SIMD operation to take place.

    SUBQ $8, DX        // Subtract 8 from the DX length register since we are now going 8 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPD Y3, Y0, Y0   // Add the second accumulator register into the first so we have one final accumulated vector.

  // Now we have an accumulator register that looks like [1, 2, 3, 4] and we
  // need to get the sum of those 4 values.
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

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the multiply/add work before we
	// combine the partial sums at the end.
  VXORPS Y0, Y0, Y0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPS Y3, Y3, Y3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPS 0(AX), Y1  // Load the current 8 float32s from a into the 256-bit register Y1.
    VMOVUPS 0(BX), Y2  // Load the current 8 float32s from b into the 256-bit register Y2.
    VSUBPS  Y1, Y2, Y1 // Subtract the 8 from a from the 8 from b and store the result in Y1.
    VMULPS  Y1, Y1, Y2 // Then square Y1 (the result of the subtraction) and store it in Y2.
    VADDPS  Y2, Y0, Y0 // Then add the result of Y2 into the first accumulator register.
    VMOVUPS 32(AX), Y4 // Load the next 8 float32s from a into the 256-bit register Y4.
    VMOVUPS 32(BX), Y5 // Load the next 8 float32s from b into the 256-bit register Y5.
    VSUBPS  Y4, Y5, Y4 // Subtract the next 8 from a from the next 8 from b and store the result in Y4.
    VMULPS  Y4, Y4, Y5 // Then square Y4 (the result of the subtraction) and store it in Y5.
    VADDPS  Y5, Y3, Y3 // Then add the result of Y5 into the second accumulator register.

    ADDQ $64, AX       // Add 64 (16 * 4) to AX. This moves the pointer forward in a by 16 items for the next SIMD operation to take place.
    ADDQ $64, BX       // Add 64 (16 * 4) to BX. This moves the pointer forward in b by 16 items for the next SIMD operation to take place.

    SUBQ $16, DX       // Subtract 16 from the DX length register since we are now going 16 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPS Y3, Y0, Y0   // Add the second accumulator register into the first so we have one final accumulated vector.

  // Now we have an accumulator register that looks like [1, 2, 3, 4, 5, 6, 7, 8] and we need to get the sum of those 8 values.
  VHADDPS    Y0, Y0, Y0     // Add pairs, now we have [3, 7, 3, 7, 11, 15, 11, 15].
  VPERM2F128 $1, Y0, Y0, Y1 // Take the high 128 bits and store them in XMM1
  VADDPS     X0, X1, X0     // Add the low 128 bits and the high 128 bits as pairs.
  HADDPS     X0, X0         // Add the horizontal pairs together to get 4 equal values.
  MOVL       X0, ret+48(FP) // Extract the lowest 32 bits of the Y0 register via the low X0 register.
  RET                       // Return the euclidean distance.

// func __euclideanDistance32_AVX_FMA(a, b []float32) float32
TEXT ·__euclideanDistance32_AVX_FMA(SB), NOSPLIT, $48-52
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the fused multiply-add work
	// before we combine the partial sums at the end.
  VXORPS Y0, Y0, Y0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPS Y3, Y3, Y3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPS     0(AX), Y1    // Load the current 8 float32s from a into the 256-bit register Y1.
    VMOVUPS     0(BX), Y2    // Load the current 8 float32s from b into the 256-bit register Y2.
    VSUBPS      Y1, Y2, Y1   // Subtract the 8 from a from the 8 from b and store the result in Y1.
    VFMADD231PS Y1, Y1, Y0   // Square Y1 and add the result to the first accumulator register in a single fused instruction.
    VMOVUPS     32(AX), Y4   // Load the next 8 float32s from a into the 256-bit register Y4.
    VMOVUPS     32(BX), Y5   // Load the next 8 float32s from b into the 256-bit register Y5.
    VSUBPS      Y4, Y5, Y4   // Subtract the next 8 from a from the next 8 from b and store the result in Y4.
    VFMADD231PS Y4, Y4, Y3   // Square Y4 and add the result to the second accumulator register in a single fused instruction.

    ADDQ $64, AX       // Add 64 (16 * 4) to AX. This moves the pointer forward in a by 16 items for the next SIMD operation to take place.
    ADDQ $64, BX       // Add 64 (16 * 4) to BX. This moves the pointer forward in b by 16 items for the next SIMD operation to take place.

    SUBQ $16, DX       // Subtract 16 from the DX length register since we are now going 16 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPS Y3, Y0, Y0   // Add the second accumulator register into the first so we have one final accumulated vector.

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

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the multiply/add work before we
	// combine the partial sums at the end.
  VXORPD Z0, Z0, Z0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPD Z3, Z3, Z3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPD 0(AX), Z1  // Load the current 8 float64s from a into the 512-bit register ZMM1.
    VMOVUPD 0(BX), Z2  // Load the current 8 float64s from b into the 512-bit register ZMM2.
    VSUBPD  Z1, Z2, Z1 // Subtract the 8 from a from the 8 from b and store the result in ZMM1.
    VMULPD  Z1, Z1, Z2 // Then square ZMM1 (the result of the subtraction) and store it in ZMM2.
    VADDPD  Z2, Z0, Z0 // Then add the result of ZMM2 into the first accumulator register.
    VMOVUPD 64(AX), Z4 // Load the next 8 float64s from a into the 512-bit register ZMM4.
    VMOVUPD 64(BX), Z5 // Load the next 8 float64s from b into the 512-bit register ZMM5.
    VSUBPD  Z4, Z5, Z4 // Subtract the next 8 from a from the next 8 from b and store the result in ZMM4.
    VMULPD  Z4, Z4, Z5 // Then square ZMM4 (the result of the subtraction) and store it in ZMM5.
    VADDPD  Z5, Z3, Z3 // Then add the result of ZMM5 into the second accumulator register.

    ADDQ $128, AX      // Add 128 (16 * 8) to AX. This moves the pointer forward in a by 16 items for the next SIMD operation to take place.
    ADDQ $128, BX      // Add 128 (16 * 8) to BX. This moves the pointer forward in b by 16 items for the next SIMD operation to take place.

    SUBQ $16, DX       // Subtract 16 from the DX length register since we are now going 16 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPD Z3, Z0, Z0   // Add the second accumulator register into the first so we have one final accumulated vector.

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

// func __euclideanDistance64_AVX512_FMA(a, b []float64) float64
TEXT ·__euclideanDistance64_AVX512_FMA(SB), NOSPLIT, $48-56
  MOVQ a_len+8(FP),   DX // Load the length of a into the DX register.
  MOVQ a_base+0(FP),  AX // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the fused multiply-add work
	// before we combine the partial sums at the end.
  VXORPD Z0, Z0, Z0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPD Z3, Z3, Z3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPD     0(AX), Z1    // Load the current 8 float64s from a into the 512-bit register ZMM1.
    VMOVUPD     0(BX), Z2    // Load the current 8 float64s from b into the 512-bit register ZMM2.
    VSUBPD      Z1, Z2, Z1   // Subtract the 8 from a from the 8 from b and store the result in ZMM1.
    VFMADD231PD Z1, Z1, Z0   // Square ZMM1 and add the result to the first accumulator register in a single fused instruction.
    VMOVUPD     64(AX), Z4   // Load the next 8 float64s from a into the 512-bit register ZMM4.
    VMOVUPD     64(BX), Z5   // Load the next 8 float64s from b into the 512-bit register ZMM5.
    VSUBPD      Z4, Z5, Z4   // Subtract the next 8 from a from the next 8 from b and store the result in ZMM4.
    VFMADD231PD Z4, Z4, Z3   // Square ZMM4 and add the result to the second accumulator register in a single fused instruction.

    ADDQ $128, AX      // Add 128 (16 * 8) to AX. This moves the pointer forward in a by 16 items for the next SIMD operation to take place.
    ADDQ $128, BX      // Add 128 (16 * 8) to BX. This moves the pointer forward in b by 16 items for the next SIMD operation to take place.

    SUBQ $16, DX       // Subtract 16 from the DX length register since we are now going 16 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPD Z3, Z0, Z0   // Add the second accumulator register into the first so we have one final accumulated vector.

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

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the multiply/add work before we
	// combine the partial sums at the end.
  VXORPS Z0, Z0, Z0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPS Z3, Z3, Z3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPS 0(AX), Z1  // Load the current 16 float32s from a into the 512-bit register ZMM1.
    VMOVUPS 0(BX), Z2  // Load the current 16 float32s from b into the 512-bit register ZMM2.
    VSUBPS  Z1, Z2, Z1 // Subtract the 16 from a from the 16 from b and store the result in ZMM1.
    VMULPS  Z1, Z1, Z2 // Then square ZMM1 (the result of the subtraction) and store it in ZMM2.
    VADDPS  Z2, Z0, Z0 // Then add the result of ZMM2 into the first accumulator register.
    VMOVUPS 64(AX), Z4 // Load the next 16 float32s from a into the 512-bit register ZMM4.
    VMOVUPS 64(BX), Z5 // Load the next 16 float32s from b into the 512-bit register ZMM5.
    VSUBPS  Z4, Z5, Z4 // Subtract the next 16 from a from the next 16 from b and store the result in ZMM4.
    VMULPS  Z4, Z4, Z5 // Then square ZMM4 (the result of the subtraction) and store it in ZMM5.
    VADDPS  Z5, Z3, Z3 // Then add the result of ZMM5 into the second accumulator register.

    ADDQ $128, AX      // Add 128 (32 * 4) to AX. This moves the pointer forward in a by 32 items for the next SIMD operation to take place.
    ADDQ $128, BX      // Add 128 (32 * 4) to BX. This moves the pointer forward in b by 32 items for the next SIMD operation to take place.

    SUBQ $32, DX       // Subtract 32 from the DX length register since we are now going 32 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPS Z3, Z0, Z0   // Add the second accumulator register into the first so we have one final accumulated vector.

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

// func __euclideanDistance32_AVX512_FMA(a, b []float32) float32
TEXT ·__euclideanDistance32_AVX512_FMA(SB), NOSPLIT, $48-52
  MOVQ a_len+8(FP), DX  // Load the length of a into the DX register.
  MOVQ a_base+0(FP), AX  // Load the pointer of the first array.
  MOVQ b_base+24(FP), BX // Load the pointer of the second array.

	// We use two accumulator registers here because this loop now processes two
	// independent vector chunks per iteration.
	// Keeping two running sums reduces the dependency chain on a single
	// accumulator register, which gives the CPU more instruction-level
	// parallelism and better hides the latency of the fused multiply-add work
	// before we combine the partial sums at the end.
  VXORPS Z0, Z0, Z0 // Clear the first accumulator register so it can store the first running partial sum.
  VXORPS Z3, Z3, Z3 // Clear the second accumulator register so it can store the second running partial sum.

  LOOP:
    VMOVUPS     0(AX), Z1    // Load the current 16 float32s from a into the 512-bit register ZMM1.
    VMOVUPS     0(BX), Z2    // Load the current 16 float32s from b into the 512-bit register ZMM2.
    VSUBPS      Z1, Z2, Z1   // Subtract the 16 from a from the 16 from b and store the result in ZMM1.
    VFMADD231PS Z1, Z1, Z0   // Square ZMM1 and add the result to the first accumulator register in a single fused instruction.
    VMOVUPS     64(AX), Z4   // Load the next 16 float32s from a into the 512-bit register ZMM4.
    VMOVUPS     64(BX), Z5   // Load the next 16 float32s from b into the 512-bit register ZMM5.
    VSUBPS      Z4, Z5, Z4   // Subtract the next 16 from a from the next 16 from b and store the result in ZMM4.
    VFMADD231PS Z4, Z4, Z3   // Square ZMM4 and add the result to the second accumulator register in a single fused instruction.

    ADDQ $128, AX      // Add 128 (32 * 4) to AX. This moves the pointer forward in a by 32 items for the next SIMD operation to take place.
    ADDQ $128, BX      // Add 128 (32 * 4) to BX. This moves the pointer forward in b by 32 items for the next SIMD operation to take place.

    SUBQ $32, DX       // Subtract 32 from the DX length register since we are now going 32 at a time.
    JNZ LOOP           // If the DX register is not zero then jump to the beginning of the loop again.

  VADDPS Z3, Z0, Z0   // Add the second accumulator register into the first so we have one final accumulated vector.

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
