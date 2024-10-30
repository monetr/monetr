package calc_test

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
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
		result := calc.FastFourierTransform(series)
		assert.EqualValues(t, expected, result)
		fmt.Println(result)
	})

	// This test does not work because gonum has a different fourier transform
	// implementation from mine. I'm using one that requires the input be a length
	// that is a power of 2. The implementation they are using seems to just
	// double the length of the input and pad it that way instead. So as a result
	// they are not quite the same. I don't understand it enough to adjust my
	// input to match their output yet.
	t.Run("known #2", func(t *testing.T) {
		t.Skip("not the same")
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
		result := calc.FastFourierTransform(series)
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
			return sum / float64(n)
		}

		sampleRate := 128.0
		frequency := 5.0
		length := 2048

		signal := sineWave(length, frequency, sampleRate)

		series := make([]complex128, len(signal))
		for x := range signal {
			series[x] = complex(signal[x], 0)
		}

		result := calc.FastFourierTransform(series)

		timeDomain := sumOfSquaresSignal(signal)
		frequencyDomain := sumOfSquaresFrequency(result, len(series))

		fmt.Printf("Energy in time domain: %.6f\n", timeDomain)
		fmt.Printf("Energy in frequency domain: %.6f\n", frequencyDomain)

		// This is based on https://en.wikipedia.org/wiki/Parseval%27s_theorem and
		// should be a way to validate that my implementation is still correct
		// without needing to have some predefined expected input and output
		// results?
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

type FrequencyTests struct {
	Rule                  string
	AcceptableFrequencies []int
	NumberOfTransactions  int
}

func TestFFTMess(t *testing.T) {
	t.Run("txns", func(t *testing.T) {
		// Number of transactions here seems to be the minimum needed to accurately
		// detect the recurrence. So something cuold be done where we only detect
		// transactions once there are a sufficient number of them? Like maybe we
		// only try to detect once it reaches 6.
		tests := []FrequencyTests{
			{
				Rule: "DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=FR",
				AcceptableFrequencies: []int{
					7,
				},
				NumberOfTransactions: 2,
			},
			{
				Rule: "DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR",
				AcceptableFrequencies: []int{
					14, 15,
				},
				NumberOfTransactions: 5,
			},
			{
				Rule: "DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
				AcceptableFrequencies: []int{
					15, 16,
				},
				NumberOfTransactions: 2,
			},
			{
				Rule: "DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1",
				AcceptableFrequencies: []int{
					30,
				},
				NumberOfTransactions: 6,
			},
			{
				Rule: "DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
				AcceptableFrequencies: []int{
					30,
				},
				NumberOfTransactions: 6,
			},
			{
				Rule: "DTSTART:20230401T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=1",
				AcceptableFrequencies: []int{
					60,
				},
				NumberOfTransactions: 3,
			},
			{
				Rule: "DTSTART:20230401T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1",
				AcceptableFrequencies: []int{
					90,
				},
				NumberOfTransactions: 4,
			},
		}
		for _, test := range tests {
			rule, err := models.NewRuleSet(test.Rule)
			assert.NoError(t, err)

			clampSettings := []bool{
				// true,
				false,
			}
			for _, clampWindow := range clampSettings {
				items := make([]models.Transaction, test.NumberOfTransactions)
				date := rule.After(time.Now().AddDate(-1, 0, 0), false)
				for i := 0; i < len(items); i++ {
					items[i] = models.Transaction{
						TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
						Amount:        100,
						Date:          date,
					}
					date = rule.After(date, false)
				}

				start := items[0].Date
				end := items[len(items)-1].Date
				diff := int64(end.Sub(start).Hours() / 24)
				fmt.Println("number of days observed:", diff)
				size := nextPowerOf2(diff)
				if clampWindow {
					size = prevPowerOf2(nextPowerOf2(diff) - 1)
				}
				padding := 3 // len(items) // 0 // (size - diff) / 2
				actualStart := start.AddDate(0, 0, -int(padding))

				series := make([]complex128, size)
				included := 0
				for i := range items {
					txn := items[i]
					if txn.Date.After(end) {
						continue
					}
					days := int64(txn.Date.Sub(actualStart).Hours() / 24)
					if int64(size) < days {
						fmt.Printf("[%d/%d] transaction %v [OUTSIDE WINDOW]\n", i, days, txn.Date)
						continue
					}
					included++
					series[days] = complex(float64(txn.Amount), 0)
					fmt.Printf("[%d/%d] transaction %v\n", i, days, txn.Date)
				}

				actualEnd := actualStart.AddDate(0, 0, int(size))
				result := calc.FastFourierTransform(series)
				fmt.Printf("Series created: %v -> %v\n", start, end)
				fmt.Printf(" Actual window: %v -> %v\n", actualStart, actualEnd)
				fmt.Printf("         Count: %d\n", len(series))
				fmt.Printf("    Total Txns: %d\n", len(items))
				fmt.Printf("      Included: %d\n", included)
				fmt.Printf("       Clamped: %t\n", clampWindow)
				fmt.Printf("      Result N: %d\n", len(result))

				// maxI, maxM := 0, 0.0
				// // clamp := (size / 2) - int64(math.Ceil(math.Sqrt(float64(size))))
				// // clamp := size / int64(included)
				// for i := 0; i < int(size/2); i++ {
				// 	c := result[i]
				// 	magnitude := math.Sqrt((real(c) * real(c)) + (imag(c) * imag(c)))
				// 	if magnitude > maxM && i != 0 {
				// 		maxI = i
				// 		maxM = magnitude
				// 	}
				//
				// 	// fmt.Println("index:", i, "magnitude:", magnitude, "freq:", float64(i)/float64(size))
				// }
				// fmt.Printf("Best: %d %f, estimated frequency: every %d days or %d days\n", maxI, maxM, int(float64(size)*(float64(maxI)/float64(size))), int(float64(size)/float64(maxI)))

				frequencies := []int{
					7,
					14,
					15,
					16,
					30,
					60,
					90,
				}
				type Freq struct {
					Frequency  int
					Concluded  float64
					Cumulative float64
					Confidence float64
					Magnitudes []float64
					Indexes    []int
				}
				final := make([]Freq, len(frequencies))
				for f := range frequencies {
					frequency := frequencies[f]
					if size < int64(frequency) {
						fmt.Println("---------")
						fmt.Printf("     Frequency: Every %d days\n", frequency)
						fmt.Printf("     Skipping due to lack of data\n")
						continue
					}
					estimatedCoordinates := (1 / float64(frequency)) * float64(size)
					primary := math.Ceil(estimatedCoordinates)
					confidence := ((float64(size) / 2) - primary) / (float64(size) / 2)
					indicies := []int{
						int(primary) - 1,
						int(primary),
						int(primary) + 1,
					}
					item := Freq{
						Frequency:  frequency,
						Confidence: confidence,
						Magnitudes: make([]float64, len(indicies)),
						Indexes:    indicies,
					}
					for i, index := range item.Indexes {
						cplx := result[index]
						magnitude := math.Sqrt((real(cplx) * real(cplx)) + (imag(cplx) * imag(cplx)))
						item.Magnitudes[i] = magnitude
						item.Concluded += magnitude
						item.Cumulative += magnitude
					}
					item.Concluded /= float64(len(indicies))
					item.Concluded *= confidence
					final[f] = item
				}
				sort.Slice(final, func(i, j int) bool {
					return final[i].Concluded > final[j].Concluded
				})

				fmt.Println(test.Rule)
				assert.Contains(t, test.AcceptableFrequencies, final[0].Frequency, "best frequency must be one of the acceptable frequencies for this test rule")
				for f := range final {
					item := final[f]
					fmt.Println("---------")
					fmt.Printf("     Frequency: Every %d days\n", item.Frequency)
					// fmt.Printf("    Cumulative: %f\n", item.Cumulative)
					fmt.Printf("    Conclusion: %f\n", item.Concluded)
					// fmt.Printf("    Confidence: %f\n", item.Confidence)
					// fmt.Printf("    Magnitudes: %+v\n", item.Magnitudes)
					// fmt.Printf("       Indexes: %+v\n", item.Indexes)
				}
				fmt.Println()
				fmt.Println()
				fmt.Println()
			}

		}
	})
}

func prevPowerOf2(n int64) int64 {
	if n < 1 {
		return 0 // No valid power of 2 for numbers less than 1
	}

	// The largest power of 2 less than or equal to n
	return 1 << (bitLength(n) - 1)
}

// Helper function to find the number of bits needed to represent n
func bitLength(n int64) int64 {
	length := int64(0)
	for n > 0 {
		length++
		n >>= 1
	}
	return length
}

func nextPowerOf2(n int64) int64 {
	n = n - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

// This test is the same as the big test above except that this one will not
// truncate data. Instead this one takes the dataset of transactions and spreads
// them out evenly over the time series. With the minimum length being the
// number of transactions * 2, then the nearest power of 2 that is greater than
// or equal to that length. This way there is not any "padding" per se because
// the time series has had its length adjusted for the minimum size of data. It
// could also just always adjust to a set data length such as 2048.
func TestFFTEvenDistribution(t *testing.T) {
	// rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
	// rule, err := models.NewRuleSet("DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=2,4,12,15,-1")
	// rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")
	// rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
	rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=DAILY;INTERVAL=2")
	assert.NoError(t, err)
	numberOfTransactions := 3
	size := 4096
	date := rule.After(time.Now().AddDate(-1, 0, 0), false)
	transactions := make([]models.Transaction, numberOfTransactions)
	for i := range transactions {
		transactions[i] = models.Transaction{
			TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
			Amount:        1,
			Date:          date,
		}
		date = rule.After(date, false)
	}
	padding := 3 // Number of days to have on each end of the series.
	// Get the start date with the padding placed before it.
	start := transactions[0].Date.AddDate(0, 0, -padding)
	// And the end date with the padding placed after it.
	end := transactions[len(transactions)-1].Date.AddDate(0, 0, padding)
	// How many seconds between the start and end?
	window := int64(end.Sub(start).Seconds())
	// How many seconds elapse for each point in the series.
	segment := float64(window) / float64(size)

	fmt.Println("segment:", segment)

	series := make([]complex128, size)
	for i := range transactions {
		txn := transactions[i]
		// Calculate the index by taking the number of seconds after the start
		// timestamp. Multiplying that by our segment size, and rounding down to
		// get our index.
		secondsSinceStart := float64(txn.Date.Sub(start).Seconds())
		// Then we can divide the number of seconds by our segment size; this will
		// tell us the index we want to use.
		index := int(math.Round(secondsSinceStart / segment))
		// Store the transaction and its amount at that index in the series.
		series[index] = complex(float64(txn.Amount), 0)
		fmt.Printf("[%02d/%04d] transaction %v\n", i, index, txn.Date)
	}

	result := calc.FastFourierTransform(series)

	// for i := 0; i < numberOfTransactions+2; i++ {
	// 	c := result[i]
	// 	magnitude := math.Sqrt((real(c) * real(c)) + (imag(c) * imag(c)))
	//
	// 	fmt.Println("index:", i, "magnitude:", magnitude, "freq:", float64(i)/float64(size))
	// }

	mean := meanMagnitude(result[1 : numberOfTransactions+1])
	fmt.Println("Mean magnitude: ", mean)
	stdDev := standardDeviation(result[1:numberOfTransactions+1], mean)
	fmt.Println("Standard deviation: ", stdDev)
	zScores := zScore(result[1 : numberOfTransactions+1])
	fmt.Println("Z-Scores: ", zScores)
	zscorePadding := 8 - len(zScores)%8
	scores := make([]float64, len(zScores)+zscorePadding)
	copy(scores, zScores)
	calc.NormalizeVector64(scores)

	frequencies := []int{
		7,
		14,
		15,
		16,
		30,
		60,
		90,
	}
	type Freq struct {
		Frequency      int
		Concluded      float64
		Confidence     float64
		EstimatedIndex float64
	}
	final := make([]Freq, len(frequencies))
	for f := range frequencies {
		frequency := frequencies[f]
		period := ((time.Duration(frequency) * 24 * time.Hour).Seconds()) / segment
		estimatedIndex := (1 / period) * float64(size)
		rounded := math.Round(estimatedIndex)
		primary := int(rounded)
		item := Freq{
			Frequency:      frequency,
			EstimatedIndex: estimatedIndex,
		}
		if rounded > float64(numberOfTransactions)+1 || primary == 0 {
			item.Concluded = 0.0
			final[f] = item
			continue
		}
		cplx := result[int(primary)]
		magnitude := math.Sqrt((real(cplx) * real(cplx)) + (imag(cplx) * imag(cplx)))
		item.Confidence = scores[primary-1]
		item.Concluded = magnitude
		final[f] = item
	}
	sort.Slice(final, func(i, j int) bool {
		return final[i].Concluded > final[j].Concluded
	})

	for f := range final {
		item := final[f]
		fmt.Println("---------")
		fmt.Printf("      Frequency: Every %d days\n", item.Frequency)
		fmt.Printf("     Conclusion: %f\n", item.Concluded)
		fmt.Printf("     Confidence: %f\n", item.Confidence)
		fmt.Printf("Estimated Index: %f\n", item.EstimatedIndex)
	}
}

func meanMagnitude(data []complex128) float64 {
	sum := 0.0
	for _, item := range data {
		sum += math.Sqrt((real(item) * real(item)) + (imag(item) * imag(item)))
	}
	return sum / float64(len(data))
}

func standardDeviation(data []complex128, mean float64) float64 {
	var varianceSum float64
	for _, item := range data {
		magnitude := math.Sqrt((real(item) * real(item)) + (imag(item) * imag(item)))
		varianceSum += (magnitude - mean) * (magnitude - mean)
	}
	return math.Sqrt(varianceSum / float64(len(data)))
}

func zScore(data []complex128) []float64 {
	mean := meanMagnitude(data)
	stdDev := standardDeviation(data, mean)
	zScores := make([]float64, len(data))
	for i, item := range data {
		magnitude := math.Sqrt((real(item) * real(item)) + (imag(item) * imag(item)))
		zScores[i] = (magnitude - mean) / stdDev
	}
	return zScores
}
