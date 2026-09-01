package calc

// SparseVectorThreshold is the number of non-zero entries a sparse vector needs
// before SparseDot32 is worth handing to a vector implementation. It was picked
// by benchmarking the two against each other, see BenchmarkSparseDot32_Go and
// BenchmarkSparseDot32_AVX512, where they cross over at around 16 entries.
const SparseVectorThreshold = 16

var (
	sparseDotImplementation32 func(dense []float32, indices []int32, values []float32) float32 = sparseDot32Go
)

func sparseDot32Go(dense []float32, indices []int32, values []float32) float32 {
	var dot float32
	for i, index := range indices {
		dot += dense[index] * values[i]
	}
	return dot
}

// SparseDot32 calculates the dot product of a dense vector and a sparse one.
// The sparse vector is provided as a pair of arrays, one of the indicies within
// the dense vector that the sparse vector actually occupies, and one of the
// values at each of those indicies.
//
// This exists so that the euclidean distance between two normalized vectors can
// be derived without touching every index of both of them. Because:
//
//	||a - b||^2 == ||a||^2 + ||b||^2 - 2(a . b)
//
// And because a[i] * b[i] is zero wherever either side is zero, the dot product
// only needs the indicies that the sparse side occupies. For monetr's
// transaction vectors that is a handful of words out of a vocabulary that grows
// with the size of the account, so this ends up being a very small fraction of
// the work that EuclideanDistance32 would do on the same pair.
//
// Note that the caller is responsible for the norms. Squaring them ahead of time
// is worthwhile since each document is compared against many others.
func SparseDot32(dense []float32, indices []int32, values []float32) float32 {
	if len(indices) != len(values) {
		panic("invalid sparse vector provided, the number of indicies and values must match!")
	}
	// Below a certain number of entries the plain go loop beats the vector
	// kernel. A gather and the horizontal reduce that has to follow it cost the
	// same whether there are two entries or sixteen, and that fixed cost is more
	// than a couple of multiply-adds. Transaction vectors normally land well
	// under this threshold, so this is the branch that usually gets taken.
	if len(indices) < SparseVectorThreshold {
		return sparseDot32Go(dense, indices, values)
	}
	return sparseDotImplementation32(dense, indices, values)
}
