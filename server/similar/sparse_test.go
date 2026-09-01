package similar

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func documentsForFixture(t testing.TB, name string) []Document {
	fixtureJson := fixtures.LoadFile(t, name)
	var data []models.Transaction
	require.NoError(t, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")

	processor := NewTransactionTFIDF()
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	return processor.GetDocuments(context.Background())
}

// TestSparseVectorMatchesDense makes sure the sparse form of each document is a
// faithful copy of the dense vector it was derived from.
func TestSparseVectorMatchesDense(t *testing.T) {
	for _, name := range []string{"monetr_sample_data_1.json", "amazon_sample_data_1.json"} {
		t.Run(name, func(t *testing.T) {
			documents := documentsForFixture(t, name)
			require.NotEmpty(t, documents, "must have produced documents to compare")

			for _, document := range documents {
				require.Len(t, document.Values, len(document.Indices),
					"every index must have a value to go with it")

				// The sparse form has to name every non-zero index of the dense
				// vector, and nothing else.
				expected := make([]int32, 0, len(document.Indices))
				for index, value := range document.Vector {
					if value == 0 {
						continue
					}
					expected = append(expected, int32(index))
				}
				assert.Equal(t, expected, document.Indices,
					"the indicies must be exactly the non-zero positions, in ascending order")

				var previous int32 = -1
				for i, index := range document.Indices {
					assert.Greater(t, index, previous, "the indicies must be sorted ascending")
					previous = index

					assert.Equal(t, document.Vector[index], document.Values[i],
						"the sparse value must be bit for bit the dense value")
					assert.NotZero(t, document.Signature&(1<<(uint64(index)%64)),
						"every index must be represented in the signature")
				}
			}
		})
	}
}

// TestSparseDistanceAgreesWithDense is the regression guard for swapping the
// dense euclidean distance out of DBSCAN. It walks every pair of documents in
// the fixtures and asserts that the sparse identity reaches the same conclusion
// about the epsilon threshold that EuclideanDistance32 would have, because that
// comparison is the only decision getNeighbors actually makes.
func TestSparseDistanceAgreesWithDense(t *testing.T) {
	for _, name := range []string{"monetr_sample_data_1.json", "amazon_sample_data_1.json"} {
		t.Run(name, func(t *testing.T) {
			documents := documentsForFixture(t, name)
			require.NotEmpty(t, documents, "must have produced documents to compare")

			scratch := make([]float32, len(documents[0].Vector))
			var compared, skipped int
			for _, point := range documents {
				for i, vectorIndex := range point.Indices {
					scratch[vectorIndex] = point.Values[i]
				}

				for _, counterpoint := range documents {
					dense := calc.EuclideanDistance32(point.Vector, counterpoint.Vector)

					// A pair the signature rejects must never have been a neighbor,
					// otherwise the prefilter is silently dropping matches.
					if point.Signature&counterpoint.Signature == 0 {
						skipped++
						assert.Greater(t, dense, float32(Epsilon),
							"the signature prefilter must only skip pairs that are beyond epsilon")
						continue
					}

					dot := calc.SparseDot32(scratch, counterpoint.Indices, counterpoint.Values)
					sparse := point.Norm2 + counterpoint.Norm2 - 2*dot

					compared++
					assert.Equal(t, dense <= Epsilon, sparse <= Epsilon,
						"the sparse and dense distances must agree on the epsilon threshold, got %v against %v",
						sparse, dense)
				}

				for _, vectorIndex := range point.Indices {
					scratch[vectorIndex] = 0
				}
			}

			t.Logf("compared %d pairs, signature prefilter skipped %d (%.1f%%)",
				compared, skipped, 100*float64(skipped)/float64(compared+skipped))
		})
	}
}

func BenchmarkDBSCAN_Amazon(b *testing.B) {
	documents := documentsForFixture(b, "amazon_sample_data_1.json")
	ctx := context.Background()

	for b.Loop() {
		_ = NewDBSCAN(documents, Epsilon, MinNeighbors).Calculate(ctx)
	}
}
