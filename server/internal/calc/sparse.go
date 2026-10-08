package calc

// SparseVectorThreshold is how many entries a sparse vector needs before
// [SparseDot32] will hand it to the assembly, below this the plain Go loop is
// faster. This came from running BenchmarkSparseDot32_Go and
// BenchmarkSparseDot32_AVX512 against each other, they cross over at around 16
// entries
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

// SparseDot32 will calculate the dot product of a dense vector and a sparse
// one. The sparse vector is two slices, indices is every spot in the dense
// vector where the sparse vector has a value, and values is the value at each
// of those spots
//
// This is used to get the euclidean distance between two normalized vectors
// without having to touch every index of both of them:
//
//	||a - b||^2 = ||a||^2 + ||b||^2 - 2(a . b)
//
// a . b is the dot product, a[0]*b[0] + a[1]*b[1] + ... and so on. Anywhere
// either side is 0 that index adds nothing, so we only need the indicies the
// sparse side actually has. A transaction is only a handful of words out of
// every word in the account, so this is a tiny fraction of the work
// [EuclideanDistance32] would do on the same pair
//
// The caller has to bring ||a||^2 and ||b||^2 themselves. Each document gets
// compared against a lot of others so it is worth working those out once up
// front
func SparseDot32(dense []float32, indices []int32, values []float32) float32 {
	if len(indices) != len(values) {
		panic("invalid sparse vector provided, the number of indicies and values must match!")
	}
	// If there are only a few entries then the plain Go loop is faster. The
	// assembly always does a full gather and then has to add up all 16 lanes at
	// the end, that costs the same if there are 2 entries or 16 and it is more
	// than just doing a couple of multiplies in Go. Most transactions are well
	// under this so this is usually the path we take
	if len(indices) < SparseVectorThreshold {
		return sparseDot32Go(dense, indices, values)
	}
	return sparseDotImplementation32(dense, indices, values)
}
