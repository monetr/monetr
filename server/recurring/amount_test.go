package recurring

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"golang.org/x/exp/slices"
)

func TestAmountDifferent(t *testing.T) {
	t.Run("KDE", func(t *testing.T) {
		items := GetFixtures(t, "amazon_sample_data_1.json")
		data := make([]float64, len(items))
		for i, item := range items {
			data[i] = float64(item.Amount)
		}

		freq := map[int64]int64{
			-3:    1,
			20045: 1,
			20076: 1,
			2141:  1,
			21578: 1,
			22882: 1,
			24159: 1,
			24189: 1,
			24754: 1,
			24791: 1,
			24819: 1,
			24838: 1,
			24:    1,
			7656:  1,
			8127:  1,
			8159:  1,
			8363:  1,
			8402:  1,
			8419:  1,
			9018:  1,
		}
		// freq = map[int64]int64{
		// 	1500: 10,
		// 	1700: 15,
		// }
		fmt.Sprint(freq)
		// data = make([]float64, 0)
		// for amount, count := range freq {
		// 	for i := 0; i < int(count); i++ {
		// 		data = append(data, float64(amount))
		// 	}
		// }

		data = append(data, data...)

		sort.Float64s(data)

		minimum, maximum := slices.Min(data), slices.Max(data)
		var bandwidth float64
		if maximum-minimum < 500 {
			bandwidth = (maximum - minimum) + 1
		} else {
			bandwidth = SilvermansRuleOfThumb(data)
		}

		// Estimate the bandwidth
		// fmt.Println("Silvermans Bandwidth:", bandwidth)
		//
		// bandwidths := make([]float64, 0)
		// maxPre := math.Max(5000, bandwidth+1)
		// maxBandwidth := int(math.Max(math.Min(maxPre, maximum-minimum), 600))
		// minBandwidth := int(math.Min(math.Max(10, maximum-minimum), 500))
		// fmt.Println("Max allowed bandwidth:", maxBandwidth)
		// for i := minBandwidth; i < maxBandwidth; i += 100 {
		// 	bandwidths = append(bandwidths, float64(i))
		// }
		// if bandwidth > float64(minBandwidth) {
		// 	bandwidths = append(bandwidths, bandwidth)
		// 	sort.Float64s(data)
		// }
		// scores := make([]float64, len(bandwidths))
		// for i, bandwidth := range bandwidths {
		// 	scores[i] = LSCVScore(data, bandwidth)
		// }
		// sort.Slice(bandwidths, func(i, j int) bool {
		// 	return scores[i] < scores[j]
		// })
		// bandwidth = bandwidths[0]
		fmt.Println("Chosen Bandwidth:", bandwidth)

		if maximum-minimum < bandwidth {
			avg := 0.0
			for _, amount := range data {
				avg += amount
			}
			avg /= float64(len(data))
			fmt.Printf("Peak: [%d] %d ±%d (%d - %d)\n", -1, int64(avg), int64(bandwidth), int64(avg)-int64(bandwidth), int64(avg)+int64(bandwidth))
			return
		}

		// fmt.Println("Chosen Bandwidth:", bandwidth)

		// Points where we want to estimate the density
		points := make([]float64, 0)
		for _, amount := range data {
			if slices.Contains(points, amount) {
				continue
			}

			points = append(points, amount)
		}
		// points := []float64{1395, 1610, 100, 500, 1000, 10000, 20000}

		// Perform KDE
		densities := KernelDensityEstimation(data, bandwidth, points)

		// Print the estimated densities
		// fmt.Println("Estimated Densities:", densities)

		// axisY := make([]opts.BarData, 0)
		// for _, item := range densities {
		// 	axisY = append(axisY, opts.BarData{
		// 		Value: item,
		// 	})
		// }
		// bar := charts.NewBar()
		// // set some global options like Title/Legend/ToolTip or anything else
		// bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		// 	Title: "KDE density graph",
		// }))

		// Put data into instance
		// bar.SetXAxis(points).
		// 	AddSeries("Category A", axisY)
		// f, _ := os.Create("density.html")
		// bar.Render(f)

		width := float64(len(points)) * 0.05
		fmt.Println("Width:", width)
		smoothed := GaussianSmooth(densities, width)
		// axisY = make([]opts.BarData, 0)
		// for _, item := range smoothed {
		// 	axisY = append(axisY, opts.BarData{
		// 		Value: item,
		// 	})
		// }
		//
		// bar = charts.NewBar()
		// // set some global options like Title/Legend/ToolTip or anything else
		// bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		// 	Title: "KDE density graph smoothed",
		// }))
		//
		// // Put data into instance
		// bar.SetXAxis(points).
		// 	AddSeries("Category A", axisY)
		// s, _ := os.Create("density_smoothed.html")
		// bar.Render(s)

		result := findPeaks(smoothed, int(width))
		for _, index := range result {
			value := int64(points[index])
			fmt.Printf("Peak: [%d] %d ±%d (%d - %d)\n", index, value, int64(bandwidth), value-int64(bandwidth), value+int64(bandwidth))
		}
	})
}

func findPeaksBandwidth(data []float64, bandwidth float64) []float64 {
	// Initialize buckets and the current bucket
	var buckets [][]float64
	currentBucket := []float64{data[0]}

	for i := 1; i < len(data); i++ {
		// Check if the current element is within the threshold of the last element in the current bucket
		if data[i]-currentBucket[len(currentBucket)-1] <= bandwidth {
			currentBucket = append(currentBucket, data[i])
		} else {
			// If not, the current bucket is complete and we start a new one
			buckets = append(buckets, currentBucket)
			currentBucket = []float64{data[i]}
		}
	}

	// Add the last bucket
	buckets = append(buckets, currentBucket)

	result := make([]float64, len(buckets))
	for i := range result {
		avg := 0.0
		for _, item := range buckets[i] {
			avg += item
		}
		avg /= float64(len(buckets[i]))
		result[i] = avg
	}

	return result
}

func findPeaks(data []float64, windowSize int) []int {
	var peaks []int
	n := len(data)
	for i := windowSize; i < n-windowSize; i++ {
		isPeak := true
		for j := -windowSize; j <= windowSize; j++ {
			if j != 0 && data[i] <= data[i+j] {
				isPeak = false
				break
			}
		}
		if isPeak {
			peaks = append(peaks, i)
		}
	}
	return peaks
}

func GaussianSmoothKernel(size int, sigma float64) []float64 {
	kernel := make([]float64, size)
	sum := 0.0
	m := size / 2

	for i := 0; i < size; i++ {
		diff := float64(i - m)
		kernel[i] = math.Exp(-(diff * diff) / (2 * sigma * sigma))
		sum += kernel[i]
	}

	// Normalize the kernel
	for i := range kernel {
		kernel[i] /= sum
	}

	return kernel
}

func GaussianSmooth(data []float64, sigma float64) []float64 {
	size := int(sigma * 6) // a common choice for kernel size
	if size%2 == 0 {
		size++ // ensure kernel size is odd
	}

	kernel := GaussianSmoothKernel(size, sigma)
	halfSize := size / 2
	smoothedData := make([]float64, len(data))

	for i := range data {
		var weightedSum float64
		var weightSum float64

		for j := -halfSize; j <= halfSize; j++ {
			if i+j >= 0 && i+j < len(data) {
				weight := kernel[halfSize+j]
				weightedSum += data[i+j] * weight
				weightSum += weight
			}
		}

		smoothedData[i] = weightedSum / weightSum
	}

	return smoothedData
}

func GaussianKernel(x float64) float64 {
	return (1 / math.Sqrt(2*math.Pi)) * math.Exp(-0.5*x*x)
}

func KernelDensityEstimation(data []float64, bandwidth float64, points []float64) []float64 {
	densities := make([]float64, len(points))

	for i, x := range points {
		sum := 0.0
		for _, xi := range data {
			sum += GaussianKernel((x - xi) / bandwidth)
		}
		densities[i] = sum / (float64(len(data)) * bandwidth)
	}

	return densities
}

func SilvermansRuleOfThumb(data []float64) float64 {
	var mean, variance float64
	for _, value := range data {
		mean += value
	}
	mean /= float64(len(data))

	for _, value := range data {
		variance += (value - mean) * (value - mean)
	}
	variance /= float64(len(data))

	stdDev := math.Sqrt(variance)
	return 1.06 * stdDev * math.Pow(float64(len(data)), -1.0/5.0)
}

// // LSCVScore calculates the Least Squares Cross-Validation score for a given bandwidth
// func LSCVScore(data []float64, bandwidth float64) float64 {
// 	n := float64(len(data))
// 	sum := 0.0
// 	for _, xi := range data {
// 		sum += KernelDensityEstimation(data, xi, []float64{bandwidth})[0]
// 	}
// 	score := (1 / (n * bandwidth)) - (2/(n*n))*sum
// 	return score
// }
