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
