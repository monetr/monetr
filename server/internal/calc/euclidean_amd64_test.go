//go:build amd64

package calc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEuclideanAMD64_AVX(t *testing.T) {
	a := []float64{
		1,
		2,
		3,
		8,
		1,
		2,
		3,
		8,
	}
	b := []float64{
		1.5,
		2,
		2.7,
		3,
		1.5,
		2,
		2.6,
		3,
	}
	normal := euclideanDistanceVanilla(a, b)
	simd := euclideanDistanceAVX(a, b)
	fmt.Printf("simd: %f\n", simd)
	assert.EqualValues(t, normal, simd)
}
