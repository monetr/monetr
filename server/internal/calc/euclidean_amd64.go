//go:build amd64

package calc

import (
	"fmt"
)

//go:noescape
func __euclideanDistanceAVX(len int32, a, b, c []float64) int32

func euclideanDistanceAVX(a, b []float64) float64 {
	if len(a)%4 != 0 {
		panic("input vector must be divisible by 4")
	}
	if len(b)%4 != 0 {
		panic("input vector must be divisible by 4")
	}

	result := make([]float64, 4, 4)
	test := __euclideanDistanceAVX(int32(len(a)), a, b, result)
	fmt.Println(result, test)
	return 0
}
