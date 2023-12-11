package recurring

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/guptarohit/asciigraph"
)

func TestAmountBinning(t *testing.T) {
	// Sample amounts and frequency

	t.Run("first", func(t *testing.T) {
		t.Skip("hfjgkashjkl")
		data := map[int64]int64{
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

		var low int64 = math.MaxInt64
		var high int64 = math.MinInt64
		for amount := range data {
			if amount < low {
				low = amount
			}
			if amount > high {
				high = amount
			}
		}

		fmt.Println("Min:", low, "Max:", high)

		axis := make([]float64, high-low)

		for i := range axis {
			val := int64(i) + low
			if freq, ok := data[val]; ok {
				axis[i] = float64(freq)
			}
		}

		graph := asciigraph.Plot(axis, asciigraph.Height(30), asciigraph.Width(200))

		fmt.Println(graph)

		gaps := make([]float64, 0)
		count := 0.0
		for _, value := range axis {
			if value == 0 {
				count++
				continue
			}

			if count == 0 {
				continue
			}

			gaps = append(gaps, count)
			count = 0
		}
		if count > 0 {
			gaps = append(gaps, count)
			count = 0
		}

		sort.Float64s(gaps)
		graph = asciigraph.Plot(gaps, asciigraph.Height(30), asciigraph.Width(200))

		fmt.Println(graph)
		fmt.Println("Gaps", gaps)

		elbow := 1501

		fmt.Println("wooo")
		i := 0
		for {
			end := i + elbow
			fmt.Println("\tindex:", i, "end:", end)
			chunk := axis[i : i+elbow]

			if axis[end+1] == 0 {
				i = end + 1
			}

			if slices.Max(chunk) > 0 {
				for x := len(chunk) - 1; x > 0; x-- {
					if chunk[x] != 0 {
						i += x
						break
					}
				}
				continue
			}

			fmt.Println("Found index", i, axis[end:])
			fmt.Println("chunk:", chunk)
			return
		}

	})

	t.Run("different approach", func(t *testing.T) {
		// items := GetFixtures(t, "amazon_sample_data_1.json")
		// data := map[int64]int64{}
		// for _, item := range items {
		// 	data[item.Amount] += 1
		// }

		data := map[int64]int64{
			// -3:    1,
			20045: 1,
			20076: 1,
			// 2141:  1,
			21578: 1,
			22882: 1,
			24159: 1,
			24189: 1,
			24754: 1,
			24791: 1,
			24819: 1,
			24838: 1,
			// 24:    1,
			// 7656:  1,
			// 8127:  1,
			// 8159:  1,
			// 8363:  1,
			// 8402:  1,
			// 8419:  1,
			// 9018:  1,
		}

		amounts := make([]float64, 0, len(data))
		for amount := range data {
			amounts = append(amounts, float64(amount))
		}

		sort.Float64s(amounts)

		fmt.Println("Original amounts graph")
		graph := asciigraph.PlotMany([][]float64{
			amounts,
		}, asciigraph.Height(35), asciigraph.Width(200))
		fmt.Println(graph)

		fmt.Println()

		{
			elbowIndex := findElbow(amounts)
			lower := amounts[:elbowIndex]
			fmt.Println("Below elbow")
			graph = asciigraph.PlotMany([][]float64{
				lower,
			}, asciigraph.Height(35), asciigraph.Width(200))
			fmt.Println(graph)
		}

		fmt.Println()

		{
			elbowIndex := findElbow(amounts)
			higher := amounts[elbowIndex:]
			fmt.Println("Above elbow")
			graph = asciigraph.PlotMany([][]float64{
				higher,
			}, asciigraph.Height(35), asciigraph.Width(200))
			fmt.Println(graph)
		}

	})
}

func findElbow(input []float64) int {
	// normalized_numbers = (numbers - np.min(numbers)) / (np.max(numbers) - np.min(numbers))
	dup := make([]float64, len(input))
	copy(dup, input)
	normalizeShit(dup)

	// graph := asciigraph.PlotMany([][]float64{
	// 	dup,
	// }, asciigraph.Height(35), asciigraph.Width(200))
	// fmt.Println(graph)

	elbowIndex := findElbowPoint(dup)
	// value := dup[elbowIndex]

	fmt.Println("ELBOW:", elbowIndex, input[elbowIndex], len(input))

	return elbowIndex

}

func normalizeShit(input []float64) {
	low, high := slices.Min(input), slices.Max(input)

	for i := range input {
		input[i] = (input[i] - low) / (high - low)
	}
}

func findElbowPoint(points []float64) int {
	distances := calculateDistances(points)

	maxDistance, elbowIndex := 0.0, 0
	for i, distance := range distances {
		if distance > maxDistance {
			maxDistance = distance
			elbowIndex = i
		}
	}

	return elbowIndex
}

func calculateDistances(points []float64) []float64 {
	// Line endpoints
	startPointX, startPointY := 1.0, points[0]
	endPointX, endPointY := float64(len(points)), points[len(points)-1]

	// Line vector
	lineVecX, lineVecY := endPointX-startPointX, endPointY-startPointY

	// Slice to store distances
	distances := make([]float64, len(points))

	// Calculate distance for each point
	for i, point := range points {
		pointVecX, pointVecY := float64(i+1)-startPointX, point-startPointY
		lineLength := math.Sqrt(lineVecX*lineVecX + lineVecY*lineVecY)
		distance := math.Abs(lineVecX*pointVecY-lineVecY*pointVecX) / lineLength
		distances[i] = distance
	}

	return distances
}

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

		sort.Float64s(data)

		// Estimate the bandwidth
		bandwidth := SilvermansRuleOfThumb(data)

		bandwidths := make([]float64, 0)
		for i := 500; i < 5000; i += 10 {
			bandwidths = append(bandwidths, float64(i))
		}
		scores := make([]float64, len(bandwidths))
		for i, bandwidth := range bandwidths {
			scores[i] = LSCVScore(data, bandwidth)
		}
		sort.Slice(bandwidths, func(i, j int) bool {
			return scores[i] > scores[j]
		})
		bandwidth = bandwidths[0]

		fmt.Println("Chosen Bandwidth:", bandwidth)

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
		fmt.Println("Estimated Densities:", densities)

		graph := asciigraph.PlotMany([][]float64{
			densities,
		}, asciigraph.Height(35), asciigraph.Width(225))
		fmt.Println(graph)

		maxIndex, maxDensity := -1, 0.0
		for i, density := range densities {
			if density > maxDensity {
				maxIndex = i
				maxDensity = density
			}
		}
		fmt.Println("bandwidth:", bandwidth)
		fmt.Println("Densist value:", points[maxIndex])
	})
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

// LSCVScore calculates the Least Squares Cross-Validation score for a given bandwidth
func LSCVScore(data []float64, bandwidth float64) float64 {
	n := float64(len(data))
	sum := 0.0
	for _, xi := range data {
		sum += KernelDensityEstimation(data, xi, []float64{bandwidth})[0]
	}
	score := (1 / (n * bandwidth)) - (2/(n*n))*sum
	return score
}
