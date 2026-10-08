package calc

// SparseVectorThreshold is how many entries a sparse vector needs before
// [SparseDot32] will hand it off to sparseDotImplementation32, which is the
// assembly when the CPU has AVX512 and sparseDot32Go otherwise. 16 is the
// smallest count where the assembly actually gets to use its gather, below
// that it is just doing scalar loads same as Go but it also has to pay for
// the call into assembly
const SparseVectorThreshold = 16

var (
	sparseDotImplementation32 func(dense []float32, indices []int32, values []float32) float32 = sparseDot32Go
)

func sparseDot32Go(dense []float32, indices []int32, values []float32) float32 {
	// This is what runs for longer vectors when the CPU doesn't have AVX512, like
	// Ivy Bridge, see SparseDot32 for the short ones
	//
	// This adds into 4 separate sums instead of just 1. Every add has to wait for
	// the one before it to finish when they all go into the same sum, and an add
	// takes 3 cycles on Ivy Bridge and Zen 4 and 4 cycles on Skylake-SP. So with
	// 1 sum we can only ever do 1 entry every 3 or 4 cycles no matter how fast
	// the loads are. With 4 sums there are always 4 adds that don't depend on
	// each other, and the loads become the limit instead. On a 7950X this took
	// 128 entries from 58ns to 43ns
	//   https://uops.info/html-instr/ADDSS_XMM_XMM.html
	var a, b, c, d float32
	count := len(indices)
	values = values[:count]
	i := 0
	for ; i+4 <= count; i += 4 {
		a += dense[indices[i]] * values[i]
		b += dense[indices[i+1]] * values[i+1]
		c += dense[indices[i+2]] * values[i+2]
		d += dense[indices[i+3]] * values[i+3]
	}

	// Then whatever is left over, there are only ever 0 to 3 of these
	for ; i < count; i++ {
		a += dense[indices[i]] * values[i]
	}

	return (a + b) + (c + d)
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
	// Most transactions are only 2 to 8 words, so this is usually the path we
	// take. For that few entries the plain loop with 1 sum beats the 4 sums in
	// sparseDot32Go, setting up and adding together the extra sums costs more
	// than it saves. It is written out here instead of calling a function
	// because the call alone was about half a nanosecond on a 7950X, which is a
	// lot when the whole thing only takes 2 or 3
	if len(indices) < SparseVectorThreshold {
		var dot float32
		for i, index := range indices {
			dot += dense[index] * values[i]
		}
		return dot
	}
	return sparseDotImplementation32(dense, indices, values)
}
