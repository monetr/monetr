package calc

var (
	sparseNeighborsImplementation32 func(
		dense []float32,
		signature uint64,
		norm2, epsilon float32,
		signatures []uint64,
		norms []float32,
		offsets []int32,
		indices []int32,
		values []float32,
		output []int32,
	) int = sparseNeighbors32Go
)

func sparseNeighbors32Go(
	dense []float32,
	signature uint64,
	norm2, epsilon float32,
	signatures []uint64,
	norms []float32,
	offsets []int32,
	indices []int32,
	values []float32,
	output []int32,
) int {
	count := 0
	for i := range signatures {
		// No bits in common means no words in common, so they can't be close
		if signature&signatures[i] == 0 {
			continue
		}

		start, end := offsets[i], offsets[i+1]
		dot := sparseDot32Go(dense, indices[start:end], values[start:end])
		distance := norm2 + norms[i] - 2*dot
		if distance <= epsilon {
			output[count] = int32(i)
			count++
		}
	}

	return count
}

// SparseNeighbors32 will find every sparse vector that is within epsilon of the
// dense one, and write their positions into output. It returns how many it
// wrote. This is the whole inner loop of DBSCAN's getNeighbors in one call,
// instead of calling [SparseDot32] once for every pair.
//
// The sparse vectors are laid out end to end. Vector i is
// indices[offsets[i]:offsets[i+1]] and values[offsets[i]:offsets[i+1]], with
// its squared norm in norms[i] and its signature in signatures[i]. The
// signature is a 64 bit bloom filter of the indices, any vector whose signature
// has no bits in common with signature is skipped without doing the dot
// product. The distance is the same squared euclidean distance [SparseDot32]
// describes, worked out from norm2, norms[i] and the dot product.
//
// The dense vector is included in the results if it is also one of the sparse
// vectors, the caller has to filter that out if they don't want it. output
// needs to be at least as long as signatures, since every vector might be close
// enough. Like [SparseDot32] the indices are not checked against dense.
func SparseNeighbors32(
	dense []float32,
	signature uint64,
	norm2, epsilon float32,
	signatures []uint64,
	norms []float32,
	offsets []int32,
	indices []int32,
	values []float32,
	output []int32,
) int {
	if len(norms) != len(signatures) || len(offsets) != len(signatures)+1 {
		panic("invalid sparse vectors provided, there must be a norm for every signature and one more offset than signatures!")
	}
	if len(indices) != len(values) || int(offsets[len(signatures)]) != len(indices) {
		panic("invalid sparse vectors provided, the number of indicies and values must match the last offset!")
	}
	if len(output) < len(signatures) {
		panic("output must be large enough to hold every sparse vector!")
	}

	return sparseNeighborsImplementation32(
		dense,
		signature,
		norm2,
		epsilon,
		signatures,
		norms,
		offsets,
		indices,
		values,
		output,
	)
}
