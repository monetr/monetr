package calc

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFourierImplementation(t *testing.T) {
	// Based on: https://github.com/gonum/gonum/blob/1ca563a018b641e805317f1ac9ae0d37b32d162c/dsp/fourier/fourier_test.go#L65-L68
	t.Run("known #1", func(t *testing.T) {
		input := []float64{
			1, 0, 1, 0, 1, 0, 1, 0,
		}
		expected := []complex128{
			4, 0, 0, 0, 4,
			0, 0, 0, // Extra zeros for some reason?
		}

		series := make([]complex128, len(input))
		for x := range input {
			series[x] = complex(input[x], 0)
		}
		result := fft(series)
		assert.EqualValues(t, expected, result)
		fmt.Println(result)
	})

	t.Run("known #2", func(t *testing.T) {
		input := []float64{
			1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1, 0,
		}
		expected := []complex128{
			12,
			-2.301937735804838 - 1.108554787638881i,
			0.7469796037174659 + 0.9366827961047095i,
			-0.9450418679126271 - 4.140498958131061i,
			-0.9450418679126271 + 4.140498958131061i,
			0.7469796037174659 - 0.9366827961047095i,
			-2.301937735804838 + 1.108554787638881i,
			12,
		}

		series := make([]complex128, len(input))
		for x := range input {
			series[x] = complex(input[x], 0)
		}
		result := fft(series)
		assert.EqualValues(t, expected, result)
		fmt.Println(result)
	})

	t.Run("parsevals theorem", func(t *testing.T) {
		sineWave := func(length int, freq float64, sampleRate float64) []float64 {
			signal := make([]float64, length)
			for i := 0; i < length; i++ {
				t := float64(i) / sampleRate
				signal[i] = math.Sin(2 * math.Pi * freq * t)
			}
			return signal
		}

		sumOfSquaresSignal := func(signal []float64) float64 {
			sum := 0.0
			for _, v := range signal {
				sum += v * v
			}
			return sum
		}

		sumOfSquaresFrequency := func(result []complex128, n int) float64 {
			sum := 0.0
			for _, v := range result {
				magnitude := math.Sqrt(real(v)*real(v) + imag(v)*imag(v))
				sum += magnitude * magnitude
			}
			return sum / float64(n) // Scale by 1/N
		}

		sampleRate := 128.0
		frequency := 5.0
		length := 2048

		signal := sineWave(length, frequency, sampleRate)

		series := make([]complex128, len(signal))
		for x := range signal {
			series[x] = complex(signal[x], 0)
		}

		result := fft(series)

		timeDomain := sumOfSquaresSignal(signal)
		frequencyDomain := sumOfSquaresFrequency(result, len(series))

		fmt.Printf("Energy in time domain: %.6f\n", timeDomain)
		fmt.Printf("Energy in frequency domain: %.6f\n", frequencyDomain)

		assert.InDeltaf(t, timeDomain, frequencyDomain, 1e-6, "must validate Parseval's theorem")
	})
}

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
