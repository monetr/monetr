package calc

import "unsafe"

var (
	sparseDotImplementation32 func(
		dense *float32,
		denseLength int,
		indices *int32,
		values *float32,
		count int,
	) float32 = sparseDot32GoPointers
)

func sparseDot32Go(dense []float32, indices []int32, values []float32) float32 {
	// This adds into 2 separate sums instead of just 1. Every add has to wait for
	// the one before it to finish when they all go into the same sum, and an add
	// takes 3 cycles on Ivy Bridge and Zen 4. With 2 sums there are always 2 adds
	// that don't depend on each other. 4 sums was worse for the 2 to 8 entries
	// most transactions have, setting up and adding together the extra sums
	// cost more than it saved. On a 7950X 8 entries took 4.4ns with 1 sum, 4.2ns
	// with 4 sums and 3.9ns with 2 sums
	//   https://uops.info/html-instr/ADDSS_XMM_XMM.html
	var a, b float32
	count := len(indices)
	values = values[:count]
	i := 0
	for ; i+2 <= count; i += 2 {
		a += dense[indices[i]] * values[i]
		b += dense[indices[i+1]] * values[i+1]
	}

	// If the count was odd there is 1 entry left over
	if i < count {
		a += dense[indices[i]] * values[i]
	}

	return a + b
}

// sparseDot32GoPointers lets sparseDot32Go sit behind
// sparseDotImplementation32, which takes pointers instead of slices because of
// the assembly. It just turns them back into the same slices SparseDot32 was
// given, so every bounds check in sparseDot32Go still works
func sparseDot32GoPointers(
	dense *float32,
	denseLength int,
	indices *int32,
	values *float32,
	count int,
) float32 {
	return sparseDot32Go(
		unsafe.Slice(dense, denseLength),
		unsafe.Slice(indices, count),
		unsafe.Slice(values, count),
	)
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

	// This takes pointers and counts for the assembly, see the notes above
	// __sparseDot32Scalar_AVX in sparse_amd64.s. Keeping this down to the length
	// check and 1 call also means Go can inline SparseDot32, so DBSCAN calls
	// straight into the assembly. If SparseDot32 had to call it instead that is 2
	// calls, and the second one cost about 0.7ns on a 7950X, which is more than
	// the assembly saves on a short vector
	return sparseDotImplementation32(
		unsafe.SliceData(dense),
		len(dense),
		unsafe.SliceData(indices),
		unsafe.SliceData(values),
		len(indices),
	)
}
