#include "textflag.h"

// func __euclideanDistanceAVX(len int32, a, b, c []float64) int32
TEXT ·__euclideanDistanceAVX(SB), NOSPLIT, $0-80
  MOVQ len+0(SP), CX
  MOVQ a+8(FP), AX      // Move pointer to first element of a to AX
  MOVQ b+32(FP), BX     // Move pointer to first element of b to BX
  MOVSD (CX), X0        // Load first element of a into X0
  MOVSD (BX), X1        // Load first element of b into X1
  ADDSD X0, X0          // Add X1 to X0
  MOVSD X0, ret+80(SP)  // Store result in return value
  RET                   // Return

//  MOVQ a+0(FP), AX      // Move the pointer of the first element of a into AX
//  MOVQ b+24(FP), BX     // Move the pointer of the first element of b into BX
//  MOVQ c+48(FP), DX     // Move the pointer of the first element of c into DX
//  MOVQ a+8(FP), CX      // Load the length of the vectors, assume they are equal; what could go wrong?
//  ADDQ $0x10, CX
//  MOVQ CX, X0
//
//
//
//  VXORPD Y0, Y0, Y0     // Clear the accumulator register
//
//  // Iterate over the vectors in batches of 4 (since I'm using YMM registers)
//  LOOP:
//    VMOVUPD (AX), Y1    // Load 4 flaot64 elements from vector a
//    VMOVUPD (BX), Y2    // Load 4 float64 elements from vector b
//    VSUBPD Y1, Y1, Y2   // Subtract a from b and store in a's register
//    VMULPD Y1, Y1 ,Y1   // Square the result and overwrite the buffer TODO: does that even work, might need to overwrite 2 instead
//    VADDPD Y0, Y0, Y1   // Add the results to the accumulator.
//
//    ADDQ $32, AX        // Move to the next 4 elements in vector a
//    ADDQ $32, BX        // Move to the next 4 elements in vector b
//    SUBQ $4, CX         // Decement the length counter (4 elements processed)
//    JNZ LOOP            // Continue the loop if there are more elements
//
//  // Now we have an accumulator register that looks like [1, 2, 3, 4] and we need to get the sum of those 4 values
//  // MOVQ (DX), AX
//  // VMOVUPD (DX), Y0
//  MOVSD X0, ret+72(FP)
//  RET
//  // VHADDPD Y0, Y0, Y0    // Add 1 + 2 and 3 + 4, now we have [3, 3, 7, 7]
//  // VPERM2F128 $0x01, Y1, Y0, Y0 // Take the 7, 7 and move it to the lower 128 bits of the Y1 register
//  // VADDPD Y0, Y1, Y0     // Add the 3, 3 and the 7, 7
//  // VEXTRACTF128 $0x0, Y0, X0 // Extract the 10, 10
//  // MOVQ X0, CX
//  // MOVQ CX, ret+48(FP) // Move the 10 into the return
//  // // MOVQ T, ret+48(FP) // Pluck the lower 64 bits and return them
//  // RET                   // Return



