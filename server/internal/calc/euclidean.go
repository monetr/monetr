package calc

var (
	euclideanImplementation func(a, b []float64) float64 = euclideanDistanceGo
)

func euclideanDistanceGo(a, b []float64) float64 {
	var distance float64
	for i := range a {
		x := a[i] - b[i]
		distance += x * x
	}
	return distance
}

func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("invalid euclidean vectors provided, length must match!")
	}
	return euclideanImplementation(a, b)
}
