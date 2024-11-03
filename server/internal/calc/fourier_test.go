package calc_test

import (
	"fmt"
	"math"
	"math/cmplx"
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
	numberOfTransactions := 9
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
		fmt.Printf("Estimated Index: %f\n", item.EstimatedIndex)
	}
}

func TestFFTReverse(t *testing.T) {
	rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
	// rule, err := models.NewRuleSet("DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
	// rule, err := models.NewRuleSet("DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=4,12,15,23,-1")
	// rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")
	// rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=FR")
	// rule, err := models.NewRuleSet("DTSTART:20230401T050000Z\nRRULE:FREQ=DAILY;INTERVAL=2")
	assert.NoError(t, err)
	numberOfTransactions := 10
	size := 4096
	individualMagnitude := 1024
	date := rule.After(time.Now().AddDate(-1, 0, 0), false)
	transactions := make([]models.Transaction, numberOfTransactions)
	for i := range transactions {
		transactions[i] = models.Transaction{
			TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
			Amount:        int64(individualMagnitude),
			Date:          date,
		}
		date = rule.After(date, false)
	}
	padding := 2 // Number of days to have on each end of the series.
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
		series[index] = complex(float64(individualMagnitude), 0)
		fmt.Printf("[%02d/%04d] transaction %v\n", i, index, txn.Date)
	}

	result := calc.FastFourierTransform(series)

	// for i := 0; i < numberOfTransactions+2; i++ {
	// 	c := result[i]
	// 	magnitude := math.Sqrt((real(c) * real(c)) + (imag(c) * imag(c)))
	//
	// 	fmt.Println("index:", i, "magnitude:", magnitude, "freq:", float64(i)/float64(size))
	// }

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
		item.Confidence = magnitude / (float64(individualMagnitude) * float64(numberOfTransactions))
		item.Concluded = magnitude
		final[f] = item
	}
	sort.Slice(final, func(i, j int) bool {
		return final[i].Confidence > final[j].Confidence
	})

	for f := range final {
		item := final[f]
		fmt.Println("---------")
		fmt.Printf("      Frequency: Every %d days\n", item.Frequency)
		fmt.Printf("     Conclusion: %f\n", item.Concluded)
		fmt.Printf("     Confidence: %f\n", item.Confidence)
		fmt.Printf("Estimated Index: %f\n", item.EstimatedIndex)
	}

	frequencyToIsolate := final[0]

	if frequencyToIsolate.Confidence < 0.2 {
		fmt.Println("Confidence is too low, frequency is likely wrong")
	}

	index := int(math.Round(frequencyToIsolate.EstimatedIndex))
	inverseSeries := IsolateFrequencyComponent(result, index)

	// Now reconstruct our data
	reconstructed := calc.InverseFastFourierTransform(inverseSeries)
	reconSignal := make([]float64, len(reconstructed))
	for i := range reconstructed {
		reconSignal[i] = real(reconstructed[i])
	}

	// TODO Recon signal is a decent waveform showing the frequency that we want,
	// if I can add a peak detection algorithm into it to isolate the regions of
	// the waveform where we would want to cherry pick data from the original
	// dataset that would be great. Another option might be to do some kind of
	// 75th percentile measurement and then find all of the transactions that are
	// part of the waveform where the value is above that percentile.

	// TODO Implement a zscore peak detection algorithm, with the window size or
	// lag being half the number of segments equal to the frequency size. So if
	// the frequency is 7 days, and thats 4 segments, the the window size should
	// be 2 segments. It wont work out exactly like that but yyou get the idea.
	// No idea how to calculate the threshold yet.

	var threshold float64
	{
		duplicate := make([]float64, len(reconSignal))
		copy(duplicate, reconSignal)
		sort.Float64s(duplicate)
		// Sort the points from the resulting signal least to greatest, then take
		// the value of the point at 66% through the sorted results. This value
		// represents a point that is higher than 66% of the magnitudes, and thus
		// any value greater than this might be considered a peak.
		cut := int(float64(len(duplicate)) / 3)
		threshold = duplicate[len(duplicate)-cut]
		fmt.Println("CUT:", cut)
	}
	fmt.Println("Threshold:", threshold)
	// Then build an array of ranges that are above that threshold, these are our
	// "peaks" which we can then use to rebuild which transactions belong to which
	// peaks.
	ranges := make([][2]int, 0, numberOfTransactions)
	startOfRange := -1
	for x, y := range reconSignal {
		// If we are under the threshold and not currently observing a range of
		// indicies then just keep going.
		if y < threshold && startOfRange == -1 {
			continue
		}

		if startOfRange == -1 {
			startOfRange = x
		} else if threshold > y {
			ranges = append(ranges, [2]int{
				startOfRange,
				x,
			})
			startOfRange = -1
		}
	}
	if startOfRange != -1 {
		ranges = append(ranges, [2]int{
			startOfRange,
			len(reconSignal) - 1,
		})
	}

	fmt.Println(ranges)

	for _, txnRange := range ranges {
		a, b := txnRange[0], txnRange[1]
		for i := range transactions {
			txn := transactions[i]
			secondsSinceStart := float64(txn.Date.Sub(start).Seconds())
			index := int(math.Round(secondsSinceStart / segment))

			if index >= a && index <= b {
				fmt.Printf("Transaction %02d - %s is part of %d day frequency, index: %d range: %v\n", i, txn.Date, frequencyToIsolate.Frequency, index, txnRange)
			}
		}
	}
}

func IsolateFrequencyComponent(result []complex128, index int) []complex128 {
	n := len(result)
	inverseSeries := make([]complex128, n)
	copy(inverseSeries, result)

	// Zero out everything except the index and its symmetric counterpart
	for i, _ := range result {
		if i == index || i == (n-index)%n {
			continue
		}
		inverseSeries[i] = complex(0, 0)
	}

	return inverseSeries
}

func TestFFTRoundTrip(t *testing.T) {
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
			magnitude := cmplx.Abs(v)
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

	inverse := calc.InverseFastFourierTransform(result)

	// Convert back to real domain for energy comparison
	reconstructedSignal := make([]float64, len(inverse))
	for i := range inverse {
		reconstructedSignal[i] = real(inverse[i]) // Only take the real part
	}

	reconstructedEnergy := sumOfSquaresSignal(reconstructedSignal)
	fmt.Printf("Energy in reconstructed time domain: %.6f\n", reconstructedEnergy)

	assert.InDeltaf(t, timeDomain, reconstructedEnergy, 1e-6, "must validate Parseval's theorem for the inverse as well")
}
