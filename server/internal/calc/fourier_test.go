package calc

import (
	"fmt"
	"testing"
)

func TestFFT(t *testing.T) {
	input := []complex128{1, 1, 1, 1, 0, 0, 0, 0}
	// For the input declaration above, golang will generate the following
	// assembly code. Which is fine for pure go, but on an AVX system we could
	// optimize how we are stashing the complex128's
	//
	// LEAQ    type:[8]complex128(SB), AX
	// PCDATA  $1, $0
	// NOP
	// CALL    runtime.newobject(SB)
	// MOVSD   $f64.3ff0000000000000(SB), X0 // Create a register that is 1.0 in the low 64 bits
	// MOVSD   X0, (AX)		// Create the first half of the complex128 with 1.0
	// XORPS   X1, X1     // Create a register that is just 0.0
	// MOVSD   X1, 8(AX)  // Create the second half of the complex128 with 0.0
	// MOVSD   X0, 16(AX) // Repeat 3x more times
	// MOVSD   X1, 24(AX)
	// MOVSD   X0, 32(AX)
	// MOVSD   X1, 40(AX)
	// MOVSD   X0, 48(AX)
	// MOVSD   X1, 56(AX)
	// MOVSD   X1, 64(AX) // then just store 0.0 for the remaining bytes.
	// MOVSD   X1, 72(AX)
	// MOVSD   X1, 80(AX)
	// MOVSD   X1, 88(AX)
	// MOVSD   X1, 96(AX)
	// MOVSD   X1, 104(AX)
	// MOVSD   X1, 112(AX)
	// MOVSD   X1, 120(AX)

	// The optimized version
	//
	// LEAQ    type:[8]complex128(SB), AX
	// PCDATA  $1, $0
	// NOP
	// CALL    runtime.newobject(SB)
	// MOVSD   $f64.3ff0000000000000(SB), X0
	// VMOVUPD X0, (AX) // Move the entire 128 bit register into the first 8 bytes
	// VMOVUPD X0, 16(AX) // Repeat for each complex128(1)
	// VMOVUPD X0, 32(AX) // Repeat for each complex128(1)
	// VMOVUPD X0, 48(AX) // Repeat for each complex128(1)
	// XORPS   X0, X0 // Clean up after ourselves and now we have the 0.0 value
	// VMOVUPD X0, 64(AX)
	// VMOVUPD X0, 80(AX)
	// VMOVUPD X0, 96(AX)
	// VMOVUPD X0, 112(AX)
	//                     // The (AX) array should now be the same value but in
	//                     // far fewer instructions. We also only use a single
	//                     // SIMD register instead of 2.
	fmt.Sprint(input)

	input[7] = complex(2, 3)
	// Creating a complex number is also interesting
	//
	// MOVSD   $f64.4000000000000000(SB), X0
	// MOVSD   X0, 112(AX)
	// MOVSD   $f64.4008000000000000(SB), X0
	// MOVSD   X0, 120(AX)
	//
	// The instructions are staggered, so at first I thought this was storing the
	// 2.0 then a 0.0 then the 3.0 then another 0.0 but I was looking at it wrong.
	// This one takes over the X0 register previously used for the 1.0 and first
	// writes 2.0 to the low 64 bits and then stores the low 64 bits in the array
	// and then it overwrites X0 again with 3.0 in the low 64 bits and performs
	// the same operation.
	// ---
	// I need to check MOVSD but since X0 is 128 bits and (AX) only has 64 bits of
	// space left does this not overwrite into address space beyond (AX)? Or is
	// MOVSD clever and is doing the right thing here?
	// Okay so it is kind of clever? https://www.felixcloutier.com/x86/movsd the
	// destination can be a 128 bit register OR a 64 bit register. So it knows
	// that it's only writing 64 bits at a time here and thats why it doesnt write
	// more than it should.
}

